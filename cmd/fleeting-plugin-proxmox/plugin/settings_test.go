package plugin

import (
	"testing"

	"github.com/stretchr/testify/require"
)

//nolint:gochecknoglobals
var (
	sampleURL = "https://example.com"
	//nolint:gosec
	sampleCredentialsPath      = "/tmp/proxmox_credentials.json"
	samplePool                 = "sample_pool"
	sampleStorage              = "sample_storage"
	sampleTemplateID           = 20
	sampleMaxInstances         = 7
	sampleInstanceNameCreating = "proxmox-creating"
	sampleInstanceNameRunning  = "running-prox"
	sampleInstanceNameRemoving = "proxve-removing"
)

func TestSettings_fillWithDefaults(t *testing.T) {
	settings := Settings{}
	settings.FillWithDefaults()

	// Don't use consts here, we want to ensure they are not changed
	require.False(t, settings.InsecureSkipTLSVerify)
	require.Equal(t, "fleeting-creating", settings.InstanceNameCreating)
	require.Equal(t, "fleeting-running", settings.InstanceNameRunning)
	require.Equal(t, "fleeting-removing", settings.InstanceNameRemoving)
	require.Equal(t, "ens18", settings.InstanceNetworkInterface)

	settings2 := Settings{
		InstanceNameCreating: sampleInstanceNameCreating,
		InstanceNameRunning:  sampleInstanceNameRunning,
		InstanceNameRemoving: sampleInstanceNameRemoving,
	}
	settings2.FillWithDefaults()

	require.Equal(t, sampleInstanceNameCreating, settings2.InstanceNameCreating)
	require.Equal(t, sampleInstanceNameRunning, settings2.InstanceNameRunning)
	require.Equal(t, sampleInstanceNameRemoving, settings2.InstanceNameRemoving)
}

func TestSettings_checkRequiredFields(t *testing.T) {
	tests := []struct {
		name          string
		settings      Settings
		expectedError error
	}{
		{
			name: "Missing URL",
			settings: Settings{
				CredentialsFilePath: sampleCredentialsPath,
				Pool:                samplePool,
				Storage:             sampleStorage,
				TemplateID:          &sampleTemplateID,
				MaxInstances:        &sampleMaxInstances,
			},
			expectedError: ErrRequiredSettingMissing,
		},
		{
			name: "Missing credentials file path",
			settings: Settings{
				URL:          sampleURL,
				Pool:         samplePool,
				Storage:      sampleStorage,
				TemplateID:   &sampleTemplateID,
				MaxInstances: &sampleMaxInstances,
			},
			expectedError: ErrRequiredSettingMissing,
		},
		{
			name: "Missing pool",
			settings: Settings{
				URL:                 sampleURL,
				CredentialsFilePath: sampleCredentialsPath,
				Storage:             sampleStorage,
				TemplateID:          &sampleTemplateID,
				MaxInstances:        &sampleMaxInstances,
			},
			expectedError: ErrRequiredSettingMissing,
		},
		{
			name: "Missing storage",
			settings: Settings{
				URL:                 sampleURL,
				CredentialsFilePath: sampleCredentialsPath,
				Pool:                samplePool,
				TemplateID:          &sampleTemplateID,
				MaxInstances:        &sampleMaxInstances,
			},
			expectedError: nil,
		},
		{
			name: "Missing template id",
			settings: Settings{
				URL:                 sampleURL,
				CredentialsFilePath: sampleCredentialsPath,
				Pool:                samplePool,
				Storage:             sampleStorage,
				TemplateID:          nil,
				MaxInstances:        &sampleMaxInstances,
			},
			expectedError: ErrRequiredSettingMissing,
		},
		{
			name: "Missing max instances",
			settings: Settings{
				URL:                 sampleURL,
				CredentialsFilePath: sampleCredentialsPath,
				Pool:                samplePool,
				Storage:             sampleStorage,
				TemplateID:          nil,
				MaxInstances:        nil,
			},
			expectedError: ErrRequiredSettingMissing,
		},
		{
			name: "No missing parameters",
			settings: Settings{
				URL:                 sampleURL,
				CredentialsFilePath: sampleCredentialsPath,
				Pool:                samplePool,
				Storage:             sampleStorage,
				TemplateID:          &sampleTemplateID,
				MaxInstances:        &sampleMaxInstances,
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.settings.CheckRequiredFields()
			require.ErrorIs(t, err, tt.expectedError)
		})
	}
}
