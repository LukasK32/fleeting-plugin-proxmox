package plugin

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/luthermonson/go-proxmox"
)

var (
	ErrNotFound = errors.New("not found")

	// Credentials file validation errors.
	ErrNoCredentials        = errors.New("credentials file specifies neither an API token nor a username/password")
	ErrIncompleteAPIToken   = errors.New("API token authentication requires both token_id and token_secret")
	ErrIncompletePassword   = errors.New("password authentication requires both username and password")
	ErrAmbiguousCredentials = errors.New("credentials file specifies both an API token and a username/password; use one or the other")
)

// proxmoxCredentials is the on-disk credentials file. It supports two
// mutually exclusive authentication methods:
//
//   - An API token (preferred). Tokens can be scoped to the plugin's pool,
//     don't expire, and need no session refresh — so no privileged
//     password sits in the runner-manager's memory or on disk.
//   - A username/password login (legacy). Retained for backwards
//     compatibility; it drives the hourly session-ticket refresher.
type proxmoxCredentials struct {
	// API token authentication. TokenID is the full token identifier in
	// user@realm!tokenname form; TokenSecret is the token's secret value.
	TokenID     string `json:"token_id"`
	TokenSecret string `json:"token_secret"`

	// Username/password authentication.
	Realm    string `json:"realm"`
	Username string `json:"username"`
	Password string `json:"password"`
	Otp      string `json:"otp,omitempty"`
}

// usesAPIToken reports whether the file selects API token authentication.
func (c *proxmoxCredentials) usesAPIToken() bool {
	return c.TokenID != "" || c.TokenSecret != ""
}

// needsSessionTicketRefresher reports whether this authentication method
// relies on a session ticket that must be periodically refreshed. Session
// tickets expire, so username/password authentication needs the refresher;
// API tokens don't expire, so they don't.
func (c *proxmoxCredentials) needsSessionTicketRefresher() bool {
	return !c.usesAPIToken()
}

func (c *proxmoxCredentials) validate() error {
	usesPassword := c.Username != "" || c.Password != ""

	switch {
	case c.usesAPIToken() && usesPassword:
		return ErrAmbiguousCredentials
	case c.usesAPIToken():
		if c.TokenID == "" || c.TokenSecret == "" {
			return ErrIncompleteAPIToken
		}
	case usesPassword:
		if c.Username == "" || c.Password == "" {
			return ErrIncompletePassword
		}
	default:
		return ErrNoCredentials
	}

	return nil
}

// authOption returns the go-proxmox client option implementing the
// selected authentication method.
func (c *proxmoxCredentials) authOption() proxmox.Option {
	if c.usesAPIToken() {
		return proxmox.WithAPIToken(c.TokenID, c.TokenSecret)
	}

	return proxmox.WithCredentials(c.toProxmoxCredentials())
}

// toProxmoxCredentials maps the username/password fields onto the library's
// credentials type, used for session-ticket login and refresh.
func (c *proxmoxCredentials) toProxmoxCredentials() *proxmox.Credentials {
	return &proxmox.Credentials{
		Username: c.Username,
		Password: c.Password,
		Otp:      c.Otp,
		Realm:    c.Realm,
	}
}

func (ig *InstanceGroup) getProxmoxPool(ctx context.Context) (*proxmox.Pool, error) {
	pool, err := ig.proxmox.Pool(ctx, ig.Settings.Pool)
	if err != nil {
		return nil, fmt.Errorf("failed to get pool id='%s': %w", ig.Settings.Pool, err)
	}

	return pool, nil
}

// Where possible, use getProxmoxVMOnNode instead as it makes less calls to API.
func (ig *InstanceGroup) getProxmoxVM(ctx context.Context, vmid int) (*proxmox.VirtualMachine, error) {
	pool, err := ig.getProxmoxPool(ctx)
	if err != nil {
		return nil, err
	}

	for _, member := range pool.Members {
		if member.Type != "qemu" {
			continue
		}

		if member.VMID == uint64(vmid) {
			return ig.getProxmoxVMOnNode(ctx, vmid, member.Node)
		}
	}

	return nil, ErrNotFound
}

func (ig *InstanceGroup) getProxmoxVMOnNode(ctx context.Context, vmid int, nodeName string) (*proxmox.VirtualMachine, error) {
	node, err := ig.proxmox.Node(ctx, nodeName)
	if err != nil {
		return nil, fmt.Errorf("failed to get node='%s': %w", nodeName, err)
	}

	vm, err := node.VirtualMachine(ctx, vmid)
	if err != nil {
		return nil, fmt.Errorf("failed to get vm='%d' on node='%s': %w", vmid, nodeName, err)
	}

	return vm, nil
}

func (ig *InstanceGroup) getProxmoxClient(credentials *proxmoxCredentials) (*proxmox.Client, error) {
	url, err := url.Parse(ig.Settings.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL='%s': %w", ig.Settings.URL, err)
	}

	httpClient := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				//nolint:gosec
				InsecureSkipVerify: ig.Settings.InsecureSkipTLSVerify,
			},
		},
	}

	return proxmox.NewClient(
		url.JoinPath("/api2/json").String(),
		credentials.authOption(),
		proxmox.WithHTTPClient(&httpClient),
	), nil
}

func (ig *InstanceGroup) getProxmoxCredentials() (*proxmoxCredentials, error) {
	credentialsFile, err := os.Open(ig.Settings.CredentialsFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open credentials file from path='%s': %w", ig.Settings.CredentialsFilePath, err)
	}
	defer credentialsFile.Close()

	credentials := proxmoxCredentials{}
	if err := json.NewDecoder(credentialsFile).Decode(&credentials); err != nil {
		return nil, fmt.Errorf("failed to decode credentials file from path='%s': %w", ig.Settings.CredentialsFilePath, err)
	}

	if err := credentials.validate(); err != nil {
		return nil, fmt.Errorf("invalid credentials file at path='%s': %w", ig.Settings.CredentialsFilePath, err)
	}

	return &credentials, nil
}
