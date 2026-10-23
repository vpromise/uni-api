package jsplugin

import (
	pluginruntime "github.com/QuantumNous/new-api/pkg/jsplugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestResolveAuth(t *testing.T) {
	for _, authType := range []string{"", "none", "api_key"} {
		auth, err := resolveAuth(pluginruntime.AuthMeta{Type: authType}, "test-key", "")
		require.NoError(t, err)
		assert.Equal(t, "test-key", auth["authHeader"])
	}
	_, err := resolveAuth(pluginruntime.AuthMeta{Type: "oauth2_jwt"}, "secret", "")
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "secret")
}
