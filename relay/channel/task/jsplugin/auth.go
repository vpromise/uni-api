package jsplugin

import (
	"fmt"
	pluginruntime "github.com/QuantumNous/new-api/pkg/jsplugin"
	"strings"
)

func resolveAuth(meta pluginruntime.AuthMeta, apiKey, proxy string) (map[string]any, error) {
	switch strings.TrimSpace(meta.Type) {
	case "", "none", "api_key":
		return map[string]any{"authHeader": apiKey}, nil
	default:
		return nil, fmt.Errorf("unsupported task plugin authentication type: %s", meta.Type)
	}
}
