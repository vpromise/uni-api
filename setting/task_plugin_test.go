package setting

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTaskPluginDisabledFactoryKeysTest(t *testing.T) {
	t.Helper()
	originalMap := common.OptionMap
	common.OptionMapRWMutex.Lock()
	common.OptionMap = map[string]string{}
	common.OptionMapRWMutex.Unlock()
	t.Cleanup(func() {
		common.OptionMapRWMutex.Lock()
		common.OptionMap = originalMap
		common.OptionMapRWMutex.Unlock()
	})
}

func TestTaskPluginDisabledFactoryKeysRoundTripAndDedupe(t *testing.T) {
	setupTaskPluginDisabledFactoryKeysTest(t)

	assert.Empty(t, GetTaskPluginDisabledFactoryKeys())
	assert.False(t, IsTaskPluginFactoryDisabled("sora"))

	require.NoError(t, SetTaskPluginDisabledFactoryKeysOption([]string{"sora", "sora", "sora", " custom-a "}))
	assert.Equal(t, []string{"custom-a", "sora"}, GetTaskPluginDisabledFactoryKeys())
	assert.Equal(t, `["custom-a","sora"]`, common.OptionMap[TaskPluginDisabledFactoryKeysKey])
	assert.True(t, IsTaskPluginFactoryDisabled("sora"))
	assert.True(t, IsTaskPluginFactoryDisabled("custom-a"))
	assert.False(t, IsTaskPluginFactoryDisabled("custom-b"))
}

func TestTaskPluginDisabledFactoryKeysBadJSONReturnsEmpty(t *testing.T) {
	setupTaskPluginDisabledFactoryKeysTest(t)

	for _, testCase := range []struct {
		name string
		raw  string
	}{
		{name: "absent", raw: ""},
		{name: "null", raw: "null"},
		{name: "object", raw: "{}"},
		{name: "number", raw: "1"},
		{name: "truncated", raw: `["sora"`},
		{name: "not json", raw: "sora"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			common.OptionMapRWMutex.Lock()
			if testCase.raw == "" {
				delete(common.OptionMap, TaskPluginDisabledFactoryKeysKey)
			} else {
				common.OptionMap[TaskPluginDisabledFactoryKeysKey] = testCase.raw
			}
			common.OptionMapRWMutex.Unlock()

			assert.Empty(t, GetTaskPluginDisabledFactoryKeys())
			assert.False(t, IsTaskPluginFactoryDisabled("sora"))
		})
	}
}
