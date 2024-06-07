package integration_test

import (
	"encoding/json"
	"flag"
	"os"
	"testing"

	"github.com/lukask32/fleeting-plugin-proxmox/cmd/fleeting-plugin-proxmox/plugin"
	"gitlab.com/gitlab-org/fleeting/fleeting/integration"
	"gitlab.com/gitlab-org/fleeting/fleeting/provider"
)

//nolint:gochecknoglobals
var (
	pluginBinaryPath = flag.String("plugin-binary-path", "", "Path to the plugin binary")
	configFilePath   = flag.String("config-path", "", "Path to the configuration file")
)

// See integration.Config.
type IntegrationTestConfig struct {
	PluginSettings  plugin.Settings          `json:"plugin_settings"`
	ConnectorConfig provider.ConnectorConfig `json:"connector_config"`
	MaxInstances    int                      `json:"max_instances"`
	UseExternalAddr bool                     `json:"use_external_addr"`
}

func TestIntegration(t *testing.T) {
	if *pluginBinaryPath == "" {
		t.Skip("plugin binary path is missing, skipping")
	}

	if *configFilePath == "" {
		t.Skip("config file path is missing, skipping")
	}

	configFile, err := os.Open(*configFilePath)
	if err != nil {
		t.Errorf("failed to open config file: %v", err)
	}
	defer configFile.Close()

	config := new(IntegrationTestConfig)
	if err := json.NewDecoder(configFile).Decode(&config); err != nil {
		t.Errorf("failed to read config file: %v", err)
	}

	integration.TestProvisioning(
		t,
		*pluginBinaryPath,
		integration.Config{
			PluginConfig: plugin.InstanceGroup{
				PluginSettings: config.PluginSettings,
			},
			ConnectorConfig: config.ConnectorConfig,
			MaxInstances:    config.MaxInstances,
			UseExternalAddr: config.UseExternalAddr,
		},
	)
}
