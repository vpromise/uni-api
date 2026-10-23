package toolconv

import (
	"testing"

	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/relayconvert/convmeta"
	kitutil "github.com/QuantumNous/new-api/relaykit/relayconvert/kitutil"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func hasDiagnosticCode(diagnostics []types.ConversionDiagnostic, code string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}

func TestResponsePhaseNeverRejectsEvenUnderStrictPolicy(t *testing.T) {
	t.Parallel()

	text := "hello"
	resp := &dto.ClaudeResponse{
		Type:       "message",
		Role:       "assistant",
		StopReason: "pause_turn",
		Content: []dto.ClaudeMediaMessage{
			{Type: "redacted_thinking", Data: "secret"},
			{Type: "text", Text: &text},
		},
	}
	diagnostics := InspectResponse(types.RelayFormatClaude, types.RelayFormatOpenAI, resp)
	require.True(t, hasDiagnosticCode(diagnostics, "continuation_state_lost"))
	require.Error(t, types.RejectConversionLoss(types.ConversionLossPolicyStrict, diagnostics))

	_, hosted, err := ExtractHostedResponse(types.RelayFormatClaude, resp)
	require.NoError(t, err)
	out, _, err := AttachHostedResponse(
		types.RelayFormatOpenAI,
		&dto.OpenAITextResponse{},
		hosted,
		&convmeta.Options{ToolLossPolicy: types.ConversionLossPolicyStrict},
	)
	require.NoError(t, err)
	require.NotNil(t, out)
}

func codexResponsesToolsRequest(t *testing.T, toolChoice any, extraTools ...map[string]any) *dto.OpenAIResponsesRequest {
	t.Helper()
	tools := []map[string]any{
		{
			"type":        "custom",
			"name":        "exec",
			"description": "Run code.",
			"format":      map[string]any{"type": "grammar", "syntax": "lark", "definition": "start: /.+/"},
		},
		{"type": "function", "name": "wait", "parameters": map[string]any{"type": "object"}},
	}
	tools = append(tools, extraTools...)
	rawTools, err := kitutil.Marshal(tools)
	require.NoError(t, err)
	request := &dto.OpenAIResponsesRequest{Model: "gpt-test", Tools: rawTools}
	if toolChoice != nil {
		request.ToolChoice, err = kitutil.Marshal(toolChoice)
		require.NoError(t, err)
	}
	return request
}

// responsesCustomToolTargets are the upstream protocols that receive a
// Responses custom tool as a function taking one string argument.
var responsesCustomToolTargets = []struct {
	format  types.RelayFormat
	request func() any
}{
	{types.RelayFormatOpenAI, func() any { return &dto.GeneralOpenAIRequest{Model: "gpt-test"} }},
	{types.RelayFormatClaude, func() any { return &dto.ClaudeRequest{Model: "claude-test"} }},
}

func attachedFunctions(t *testing.T, out any) ([]map[string]any, any) {
	t.Helper()
	var (
		functions []map[string]any
		choice    any
		err       error
	)
	switch target := out.(type) {
	case *dto.GeneralOpenAIRequest:
		for _, tool := range target.Tools {
			function, err := kitutil.Any2Type[map[string]any](tool.Function)
			require.NoError(t, err)
			functions = append(functions, function)
		}
		choice, err = kitutil.Any2Type[any](target.ToolChoice)
	case *dto.ClaudeRequest:
		functions, err = kitutil.Any2Type[[]map[string]any](target.Tools)
		require.NoError(t, err)
		choice, err = kitutil.Any2Type[any](target.ToolChoice)

	default:
		require.FailNow(t, "unexpected request type", "%T", out)
	}
	require.NoError(t, err)
	return functions, choice
}

func TestResponsesCustomToolReachesEachTargetAsStringInputFunction(t *testing.T) {
	t.Parallel()

	const execDescription = "Run code.\n\nThis tool takes freeform text. Put the complete raw text in the \"input\" argument.\n\nThe input must match this Lark grammar:\nstart: /.+/"
	inputSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"input": map[string]any{"type": "string", "description": "Raw input for the tool."},
		},
		"required":             []any{"input"},
		"additionalProperties": false,
	}
	expected := map[types.RelayFormat]struct {
		functions []map[string]any
		choice    any
	}{
		types.RelayFormatOpenAI: {
			functions: []map[string]any{
				{"name": "exec", "description": execDescription, "parameters": inputSchema},
				{"name": "wait", "parameters": map[string]any{"type": "object"}},
			},
			choice: map[string]any{"type": "function", "function": map[string]any{"name": "exec"}},
		},
		types.RelayFormatClaude: {
			functions: []map[string]any{
				{"name": "exec", "description": execDescription, "input_schema": inputSchema},
				{"name": "wait", "input_schema": map[string]any{"type": "object", "properties": map[string]any{}}},
			},
			choice: map[string]any{"type": "tool", "name": "exec"},
		},
	}

	_, set, err := ExtractRequest(types.RelayFormatOpenAIResponses, codexResponsesToolsRequest(t, map[string]any{"type": "custom", "name": "exec"}))
	require.NoError(t, err)
	assert.Equal(t, map[string]struct{}{"exec": {}}, ResponsesCustomToolNames(set))

	for _, target := range responsesCustomToolTargets {
		t.Run(string(target.format), func(t *testing.T) {
			out, diagnostics, err := AttachRequest(target.format, target.request(), set, &convmeta.Options{ToolLossPolicy: types.ConversionLossPolicySafe})
			require.NoError(t, err)
			functions, choice := attachedFunctions(t, out)
			assert.Equal(t, expected[target.format].functions, functions)
			assert.Equal(t, expected[target.format].choice, choice)
			assert.True(t, hasDiagnosticCode(diagnostics, "custom_tool_as_function"))
			assert.False(t, hasDiagnosticCode(diagnostics, "unsupported_hosted_tool"))
			assert.False(t, hasDiagnosticCode(diagnostics, "unsupported_tool_choice"))

			_, _, err = AttachRequest(target.format, target.request(), set, &convmeta.Options{ToolLossPolicy: types.ConversionLossPolicyStrict})
			var loss *types.ConversionLossError
			require.ErrorAs(t, err, &loss)
			assert.True(t, hasDiagnosticCode(loss.Diagnostics, "custom_tool_as_function"))
		})
	}
}

func TestResponsesCustomToolNameConflictIsDropped(t *testing.T) {
	t.Parallel()

	// The custom exec loses to the function exec, so a choice naming it has no
	// sent tool to point at and must stay unconverted.
	request := codexResponsesToolsRequest(t, map[string]any{"type": "custom", "name": "exec"},
		map[string]any{"type": "function", "name": "exec", "parameters": map[string]any{"type": "object"}},
		map[string]any{"type": "custom", "name": "apply_patch"},
		map[string]any{"type": "custom", "name": "apply_patch", "description": "duplicate"},
	)
	_, set, err := ExtractRequest(types.RelayFormatOpenAIResponses, request)
	require.NoError(t, err)
	assert.Equal(t, map[string]struct{}{"apply_patch": {}}, ResponsesCustomToolNames(set))

	for _, target := range responsesCustomToolTargets {
		t.Run(string(target.format), func(t *testing.T) {
			out, diagnostics, err := AttachRequest(target.format, target.request(), set, &convmeta.Options{})
			require.NoError(t, err)
			functions, choice := attachedFunctions(t, out)
			require.Len(t, functions, 3)
			names := make([]string, 0, len(functions))
			for _, function := range functions {
				names = append(names, function["name"].(string))
			}
			assert.Equal(t, []string{"wait", "exec", "apply_patch"}, names)
			// Only the custom definitions carry a description, so the kept exec
			// is the function one.
			assert.NotContains(t, functions[1], "description")
			assert.NotContains(t, functions[2]["description"], "duplicate")
			assert.Nil(t, choice)
			assert.True(t, hasDiagnosticCode(diagnostics, "custom_tool_name_conflict"))
			assert.True(t, hasDiagnosticCode(diagnostics, "unsupported_tool_choice"))

			_, _, err = AttachRequest(target.format, target.request(), set, &convmeta.Options{ToolLossPolicy: types.ConversionLossPolicySafe})
			require.Error(t, err)
		})
	}
}
