package plugin

import (
	"errors"
	"fmt"
)

var ErrRequiredSettingMissing  = errors.New("required setting is missing")
var ErrSettingInvalidParameter = errors.New("setting has invalid parameter")

// Default values for plugin settings.
const (
	DefaultInstanceNetworkInterface = "ens18"
	DefaultInstanceNetworkProtocol  = "ipv4"

	DefaultInstanceNameCreating = "fleeting-creating"
	DefaultInstanceNameRunning  = "fleeting-running"
	DefaultInstanceNameRemoving = "fleeting-removing"
)

// Plguin settings.
type Settings struct {
	// Proxmox VE URL.
	URL string `json:"url"`

	// If true then TLS certificate verification is disabled.
	InsecureSkipTLSVerify bool `json:"insecure_skip_tls_verify"`

	// Path to Proxmox VE credentials file.
	CredentialsFilePath string `json:"credentials_file_path"`

	// Name of the Proxmox VE pool to use.
	Pool string `json:"pool"`

	// Name of the Proxmox VE storage to use.
	Storage string `json:"storage"`

	// ID of the Proxmox VE VM to create instances from.
	TemplateID *int `json:"template_id,omitempty"`

	// Maximum instances than can be deployed.
	MaxInstances *int `json:"max_instances,omitempty"`

	// Network interface to read instance's IP address from.
	InstanceNetworkInterface string `json:"instance_network_interface"`

	// Network protocol (ipv4, ipv6 or any)
	//   - "ipv4" tries to find one internal and one external IPv4 address
	//   - "ipv6" tries to find one internal (ULA) and one global (GUA) IPv6 address
	//   - "any" will prioritize IPv6 but return IPv4 if there is no IPv6.
	// (link-local adresses are apparently not supported by Gitlab-Runner)
	// Default is "ipv4" to not break existing setups - might be switched to any in the future.
	InstanceNetworkProtocol string `json:"instance_network_protocol"`

	// Name to set for instances during creation.
	InstanceNameCreating string `json:"instance_name_creating"`

	// Name to set for running instances.
	InstanceNameRunning string `json:"instance_name_running"`

	// Name to set for instances during removal.
	InstanceNameRemoving string `json:"instance_name_removing"`
}

func (s *Settings) FillWithDefaults() {
	if s.InstanceNetworkInterface == "" {
		s.InstanceNetworkInterface = DefaultInstanceNetworkInterface
	}

	if s.InstanceNameCreating == "" {
		s.InstanceNameCreating = DefaultInstanceNameCreating
	}

	if s.InstanceNameRunning == "" {
		s.InstanceNameRunning = DefaultInstanceNameRunning
	}

	if s.InstanceNameRemoving == "" {
		s.InstanceNameRemoving = DefaultInstanceNameRemoving
	}

	if s.InstanceNetworkProtocol == "" {
		s.InstanceNetworkProtocol = DefaultInstanceNetworkProtocol
	}
}

func (s *Settings) CheckRequiredFields() error {
	if s.URL == "" {
		return fmt.Errorf("%w: url", ErrRequiredSettingMissing)
	}

	if s.CredentialsFilePath == "" {
		return fmt.Errorf("%w: credentials_file_path", ErrRequiredSettingMissing)
	}

	if s.Pool == "" {
		return fmt.Errorf("%w: pool", ErrRequiredSettingMissing)
	}

	if s.TemplateID == nil {
		return fmt.Errorf("%w: template_id", ErrRequiredSettingMissing)
	}

	if s.MaxInstances == nil {
		return fmt.Errorf("%w: max_instances", ErrRequiredSettingMissing)
	}

	if s.InstanceNetworkProtocol != "" && s.InstanceNetworkProtocol != "ipv4" && s.InstanceNetworkProtocol != "ipv6" && s.InstanceNetworkProtocol != "any" {
		return fmt.Errorf("%w: instance_network_protocol: must be ipv4, ipv6 or any", ErrSettingInvalidParameter)
	}

	return nil
}
