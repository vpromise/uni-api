package controller

import (
	"testing"

	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/pkg/jsplugin"
	"github.com/stretchr/testify/require"
)

func TestValidateNewAPIChannelTaskPluginExtensions(t *testing.T) {
	// gateway-ext-vendor never declares new_api, so it can only run against
	// its vendor and must be refused on a New API channel.
	for key, upstreams := range map[string]string{"gateway-ext-a": `upstreams: ["vendor", "new_api"],`, "gateway-ext-b": `upstreams: ["new_api"],`, "gateway-ext-vendor": ""} {
		source := `
export const meta = {apiVersion: 1,  key: "` + key + `", name: "Gateway", version: "1.0.0", author: {name: "Test"}, models: ["doc-` + key + `"], fetchMode: "per_task", ` + upstreams + `};
export function buildSubmitRequest() { return {}; }
export function parseSubmitResponse() { return {}; }
export function buildQueryRequest() { return {}; }
export function parseTaskResult() { return {}; }
`
		_, err := jsplugin.DefaultRegistry.Register(source, jsplugin.Options{})
		require.NoError(t, err)
		t.Cleanup(func() { jsplugin.DefaultRegistry.Unregister(key) })
	}
	baseURL := "https://gateway.example"
	cases := []struct {
		name, setting, wantErr string
	}{
		{"single and extension keys", `{"task_plugin_key":"gateway-ext-a","task_extend_plugin_keys":["gateway-ext-b"]}`, ""},
		{"extension keys only", `{"task_extend_plugin_keys":["gateway-ext-a","gateway-ext-b"]}`, ""},
		{"unregistered", `{"task_extend_plugin_keys":["gateway-ext-a","missing"]}`, `task plugin "missing" is not registered`},
		{"vendor-only plugin as extension", `{"task_extend_plugin_keys":["gateway-ext-a","gateway-ext-vendor"]}`, ""},
		{"vendor-only plugin as single key", `{"task_plugin_key":"gateway-ext-vendor"}`, ""},
		{"duplicate", `{"task_extend_plugin_keys":["gateway-ext-a","gateway-ext-a"]}`, "Invalid parameters"},
		{"duplicate of single key", `{"task_plugin_key":"gateway-ext-a","task_extend_plugin_keys":["gateway-ext-a"]}`, "Invalid parameters"},
		{"padded", `{"task_extend_plugin_keys":[" gateway-ext-a"]}`, "Invalid parameters"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			setting := tc.setting
			channel := &model.Channel{Type: constant.ChannelTypeOpenAI, BaseURL: &baseURL, Setting: &setting}
			err := validateChannel(channel, false)
			if tc.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tc.wantErr)
		})
	}

}
