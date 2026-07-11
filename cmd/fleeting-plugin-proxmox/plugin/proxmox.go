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

var ErrNotFound = errors.New("not found")

func (ig *InstanceGroup) getProxmoxPool(ctx context.Context) (*proxmox.Pool, error) {
	pool, err := ig.proxmox.Pool(ctx, ig.Pool)
	if err != nil {
		return nil, fmt.Errorf("failed to get pool id='%s': %w", ig.Pool, err)
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

func (ig *InstanceGroup) getProxmoxClient() (*proxmox.Client, error) {
	url, err := url.Parse(ig.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL='%s': %w", ig.URL, err)
	}

	credentials, err := ig.getProxmoxCredentials()
	if err != nil {
		return nil, err
	}

	httpClient := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				//nolint:gosec
				InsecureSkipVerify: ig.InsecureSkipTLSVerify,
			},
		},
	}

	return proxmox.NewClient(
		url.JoinPath("/api2/json").String(),
		proxmox.WithCredentials(credentials),
		proxmox.WithHTTPClient(&httpClient),
	), nil
}

func (ig *InstanceGroup) getProxmoxCredentials() (*proxmox.Credentials, error) {
	credentialsFile, err := os.Open(ig.CredentialsFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open credentials file from path='%s': %w", ig.CredentialsFilePath, err)
	}
	defer credentialsFile.Close()

	credentials := proxmox.Credentials{}

	err = json.NewDecoder(credentialsFile).Decode(&credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to decode credentials file from path='%s': %w", ig.CredentialsFilePath, err)
	}

	return &credentials, nil
}
