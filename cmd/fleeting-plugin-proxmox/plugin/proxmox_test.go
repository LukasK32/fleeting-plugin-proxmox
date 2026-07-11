package plugin

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInstanceGroup_getProxmoxClient(t *testing.T) {
	tempDir := t.TempDir()
	ig := InstanceGroup{
		Settings: Settings{
			URL:                   "https://example.com/proxmox",
			InsecureSkipTLSVerify: false,
			CredentialsFilePath:   path.Join(tempDir, "prox_credentials.json"),
		},
	}

	err := os.WriteFile(
		ig.CredentialsFilePath,
		[]byte(`{"realm": "pve","username": "03Ewl6rENi","password": "-rx£N503o_8(%\"l+=*4,YD"}`),
		0o600,
	)
	require.NoError(t, err)

	_, err = ig.getProxmoxClient()
	require.NoError(t, err)
}

func TestInstanceGroup_getProxmoxCredentials(t *testing.T) {
	tempDir := t.TempDir()
	ig := InstanceGroup{
		Settings: Settings{
			CredentialsFilePath: path.Join(tempDir, "sample_credentials.json"),
		},
	}

	// Missing credentials file
	_, err := ig.getProxmoxCredentials()
	require.ErrorIs(t, err, os.ErrNotExist)

	// Malformed credentials file
	err = os.WriteFile(
		ig.CredentialsFilePath,
		[]byte(`{"realm": 'pve',`),
		0o600,
	)
	require.NoError(t, err)

	_, err = ig.getProxmoxCredentials()
	require.Error(t, err)

	// Correct credentials file
	err = os.WriteFile(
		ig.CredentialsFilePath,
		[]byte(`{"realm": "pve","username": "oQcW8N246FODI6Qui","password": "88u3[kKLJ{gU7A£fhWq"}`),
		0o600,
	)
	require.NoError(t, err)

	credentials, err := ig.getProxmoxCredentials()
	require.NoError(t, err)
	require.Equal(t, "pve", credentials.Realm)
	require.Equal(t, "oQcW8N246FODI6Qui", credentials.Username)
	require.Equal(t, `88u3[kKLJ{gU7A£fhWq`, credentials.Password)
}
