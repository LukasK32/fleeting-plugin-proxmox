package plugin

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnmarshallingPluginSettings(t *testing.T) {
	settingsJSON := `{"url":"sample_url","template_id": 5}`
	instance := InstanceGroup{}

	err := json.Unmarshal([]byte(settingsJSON), &instance)
	require.NoError(t, err)

	require.Equal(t, "sample_url", instance.URL)
	require.Equal(t, 5, *instance.TemplateID)
}
