package reasoning

import (
	"testing"

	"github.com/QuantumNous/new-api/setting/model_setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCanonicalBillingModelNames(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{
			name: "thinking on",
			in:   "gpt-custom@thinking:on",
			want: []string{"gpt-custom@thinking:on"},
		},
		{
			name: "shuffled temperature and thinking",
			in:   "gpt-custom@temperature:0.2@thinking:on",
			want: []string{"gpt-custom@thinking:on"},
		},
		{
			name: "thinking first then temperature",
			in:   "gpt-custom@thinking:on@temperature:0.2",
			want: []string{"gpt-custom@thinking:on"},
		},
		{
			name: "budget normalizes to on",
			in:   "gpt-custom@thinking:8192",
			want: []string{"gpt-custom@thinking:on"},
		},
		{
			name: "minus one normalizes to on",
			in:   "gpt-custom@thinking:-1",
			want: []string{"gpt-custom@thinking:on"},
		},
		{
			name: "adaptive normalizes to on",
			in:   "gpt-custom@thinking:adaptive",
			want: []string{"gpt-custom@thinking:on"},
		},
		{
			name: "thinking off",
			in:   "gpt-custom@thinking:off",
			want: []string{"gpt-custom@thinking:off"},
		},
		{
			name: "effort none becomes thinking off",
			in:   "gpt-custom@effort:none",
			want: []string{"gpt-custom@thinking:off"},
		},
		{
			name: "effort high implies thinking on",
			in:   "gpt-custom@effort:high",
			want: []string{"gpt-custom@effort:high@thinking:on", "gpt-custom@thinking:on"},
		},
		{
			name: "effort and thinking keys sorted",
			in:   "gpt-custom@thinking:on@effort:high@temperature:0.2",
			want: []string{"gpt-custom@effort:high@thinking:on", "gpt-custom@thinking:on"},
		},
		{
			name: "duplicate last wins then normalize",
			in:   "gpt-custom@thinking:off@thinking:on@effort:low@effort:high",
			want: []string{"gpt-custom@effort:high@thinking:on", "gpt-custom@thinking:on"},
		},
		{
			name: "legacy thinking alias",
			in:   "claude-3-7-sonnet-thinking",
			want: []string{"claude-3-7-sonnet@thinking:on"},
		},

		{
			name: "legacy nothinking",
			in:   "claude-3-7-sonnet-nothinking",
			want: []string{"claude-3-7-sonnet@thinking:off"},
		},
		{
			name: "temperature only has no reasoning state",
			in:   "gpt-custom@temperature:0.7",
			want: nil,
		},
	}

	t.Cleanup(func() {})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, CanonicalBillingModelNames(tt.in))
		})
	}

	assert.Equal(t, "gpt-5.1-codex-max", BaseModelName("gpt-5.1-codex-max"))
	assert.Empty(t, CanonicalBillingModelNames("gpt-5.1-codex-max"))
}

func TestParseOpenAIReasoningEffortPreservesCodexMax(t *testing.T) {
	effort, base := ParseOpenAIReasoningEffortFromModelSuffix("gpt-5.1-codex-max")
	assert.Empty(t, effort)
	assert.Equal(t, "gpt-5.1-codex-max", base)
}

func TestBaseModelNameStripsModifiers(t *testing.T) {
	require.Equal(t, "gpt-custom", BaseModelName("gpt-custom@thinking:on@temperature:0.2"))
}

func TestExemptAtNameIsOpaqueForBillingIdentity(t *testing.T) {
	settings := model_setting.GetGlobalSettings()
	original := append([]string(nil), settings.ThinkingModelBlacklist...)
	t.Cleanup(func() { settings.ThinkingModelBlacklist = original })
	settings.ThinkingModelBlacklist = append(original, "re:.*@sha256:.*")

	const model = "opaque@sha256:deadbeef"
	assert.Equal(t, model, BaseModelName(model))
	assert.Empty(t, CanonicalBillingModelNames(model))
	assert.Equal(t, "custom-base-thinking", BaseModelName("custom-base-thinking"))
	assert.Empty(t, CanonicalBillingModelNames("custom-base-thinking"))
}
