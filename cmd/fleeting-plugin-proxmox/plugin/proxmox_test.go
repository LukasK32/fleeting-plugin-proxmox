package plugin

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInstanceGroup_getProxmoxClient(t *testing.T) {
	tempDir := t.TempDir()

	for name, testCase := range map[string]struct {
		credentials  string
		wantAPIToken bool
	}{
		"password": {
			credentials:  `{"realm": "pve","username": "03Ewl6rENi","password": "-rx£N503o_8(%\"l+=*4,YD"}`,
			wantAPIToken: false,
		},
		"apiToken": {
			credentials:  `{"token_id": "user@pve!token","token_secret": "00000000-0000-0000-0000-000000000000"}`,
			wantAPIToken: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			ig := InstanceGroup{
				Settings: Settings{
					URL:                   "https://example.com/proxmox",
					InsecureSkipTLSVerify: false,
					CredentialsFilePath:   path.Join(tempDir, name+"_credentials.json"),
				},
			}

			err := os.WriteFile(ig.Settings.CredentialsFilePath, []byte(testCase.credentials), 0o600)
			require.NoError(t, err)

			creds, err := ig.getProxmoxCredentials()
			require.NoError(t, err)

			// The auth method the client is built with is security-relevant,
			// but go-proxmox stores it on unexported client fields, so we
			// assert the decision that selects it instead.
			require.Equal(t, testCase.wantAPIToken, creds.usesAPIToken())

			_, err = ig.getProxmoxClient(creds)
			require.NoError(t, err)
		})
	}
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
		ig.Settings.CredentialsFilePath,
		[]byte(`{"realm": 'pve',`),
		0o600,
	)
	require.NoError(t, err)

	_, err = ig.getProxmoxCredentials()
	require.Error(t, err)

	// Correct password credentials file
	err = os.WriteFile(
		ig.Settings.CredentialsFilePath,
		[]byte(`{"realm": "pve","username": "oQcW8N246FODI6Qui","password": "88u3[kKLJ{gU7A£fhWq"}`),
		0o600,
	)
	require.NoError(t, err)

	credentials, err := ig.getProxmoxCredentials()
	require.NoError(t, err)
	require.False(t, credentials.usesAPIToken())
	require.Equal(t, "pve", credentials.Realm)
	require.Equal(t, "oQcW8N246FODI6Qui", credentials.Username)
	require.Equal(t, `88u3[kKLJ{gU7A£fhWq`, credentials.Password)

	// Correct API token credentials file
	err = os.WriteFile(
		ig.Settings.CredentialsFilePath,
		[]byte(`{"token_id": "user@pve!token","token_secret": "3f2504e0-4f89-11d3-9a0c-0305e82c3301"}`),
		0o600,
	)
	require.NoError(t, err)

	credentials, err = ig.getProxmoxCredentials()
	require.NoError(t, err)
	require.True(t, credentials.usesAPIToken())
	require.Equal(t, "user@pve!token", credentials.TokenID)
	require.Equal(t, "3f2504e0-4f89-11d3-9a0c-0305e82c3301", credentials.TokenSecret)

	// Invalid (ambiguous) credentials file: validate() must reject it, and the
	// error must propagate through getProxmoxCredentials to the caller.
	err = os.WriteFile(
		ig.Settings.CredentialsFilePath,
		[]byte(`{"token_id": "user@pve!token","token_secret": "secret","username": "root","password": "secret"}`),
		0o600,
	)
	require.NoError(t, err)

	_, err = ig.getProxmoxCredentials()
	require.ErrorIs(t, err, ErrAmbiguousCredentials)
}

func TestProxmoxCredentials_validate(t *testing.T) {
	for name, testCase := range map[string]struct {
		credentials proxmoxCredentials
		wantErr     error
	}{
		"api token": {
			credentials: proxmoxCredentials{TokenID: "user@pve!token", TokenSecret: "secret"},
		},
		"password": {
			credentials: proxmoxCredentials{Username: "root", Password: "secret", Realm: "pam"},
		},
		"empty": {
			credentials: proxmoxCredentials{},
			wantErr:     ErrNoCredentials,
		},
		"token without secret": {
			credentials: proxmoxCredentials{TokenID: "user@pve!token"},
			wantErr:     ErrIncompleteAPIToken,
		},
		"secret without token id": {
			credentials: proxmoxCredentials{TokenSecret: "secret"},
			wantErr:     ErrIncompleteAPIToken,
		},
		"password without username": {
			credentials: proxmoxCredentials{Password: "secret"},
			wantErr:     ErrIncompletePassword,
		},
		"both methods": {
			credentials: proxmoxCredentials{TokenID: "user@pve!token", TokenSecret: "secret", Username: "root", Password: "secret"},
			wantErr:     ErrAmbiguousCredentials,
		},
	} {
		t.Run(name, func(t *testing.T) {
			err := testCase.credentials.validate()
			if testCase.wantErr != nil {
				require.ErrorIs(t, err, testCase.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestProxmoxCredentials_needsSessionTicketRefresher(t *testing.T) {
	for name, testCase := range map[string]struct {
		credentials proxmoxCredentials
		want        bool
	}{
		"password needs refresher": {
			credentials: proxmoxCredentials{Username: "root", Password: "secret", Realm: "pam"},
			want:        true,
		},
		"api token does not need refresher": {
			credentials: proxmoxCredentials{TokenID: "user@pve!token", TokenSecret: "secret"},
			want:        false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, testCase.want, testCase.credentials.needsSessionTicketRefresher())
		})
	}
}
