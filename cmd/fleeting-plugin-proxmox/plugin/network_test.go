package plugin

import (
	"testing"

	"github.com/luthermonson/go-proxmox"
	"github.com/stretchr/testify/require"
)

func Test_determineAddresses(t *testing.T) {
	tests := []struct {
		name string

		requestedInterface string
		requestedProtocol  NetworkProtocol
		networkInterfaces  []*proxmox.AgentNetworkIface

		expectedError           error
		expectedInternalAddress string
		expectedExternalAddress string
	}{
		{
			name: "No network interfaces",

			requestedInterface: DefaultInstanceNetworkInterface,
			requestedProtocol:  NetworkProtocolAny,
			networkInterfaces:  []*proxmox.AgentNetworkIface{},

			expectedError:           ErrNoIPAddress,
			expectedInternalAddress: "",
			expectedExternalAddress: "",
		},
		{
			name: "Any",

			requestedInterface: DefaultInstanceNetworkInterface,
			requestedProtocol:  NetworkProtocolAny,
			networkInterfaces: []*proxmox.AgentNetworkIface{
				{
					Name: DefaultInstanceNetworkInterface,
					IPAddresses: []*proxmox.AgentNetworkIPAddress{
						{
							IPAddressType: NetworkProtocolIPv4,
							IPAddress:     "8.8.8.8",
						},
						{
							IPAddressType: NetworkProtocolIPv4,
							IPAddress:     "192.168.0.1",
						},
						{
							IPAddressType: NetworkProtocolIPv6,
							IPAddress:     "2001:4860:4860::8888",
						},
						{
							IPAddressType: NetworkProtocolIPv6,
							IPAddress:     "fd3b:47fc:de09::1",
						},
					},
				},
			},

			expectedError:           nil,
			expectedInternalAddress: "fd3b:47fc:de09::1",
			expectedExternalAddress: "2001:4860:4860::8888",
		},
		{
			name: "Forced IPv4",

			requestedInterface: DefaultInstanceNetworkInterface,
			requestedProtocol:  NetworkProtocolIPv4,
			networkInterfaces: []*proxmox.AgentNetworkIface{
				{
					Name: DefaultInstanceNetworkInterface,
					IPAddresses: []*proxmox.AgentNetworkIPAddress{
						{
							IPAddressType: NetworkProtocolIPv4,
							IPAddress:     "8.8.8.8",
						},
						{
							IPAddressType: NetworkProtocolIPv4,
							IPAddress:     "192.168.0.1",
						},
						{
							IPAddressType: NetworkProtocolIPv6,
							IPAddress:     "2001:4860:4860::8888",
						},
						{
							IPAddressType: NetworkProtocolIPv6,
							IPAddress:     "fd3b:47fc:de09::1",
						},
					},
				},
			},

			expectedError:           nil,
			expectedInternalAddress: "192.168.0.1",
			expectedExternalAddress: "8.8.8.8",
		},
		{
			name: "Forced IPv6",

			requestedInterface: DefaultInstanceNetworkInterface,
			requestedProtocol:  NetworkProtocolIPv6,
			networkInterfaces: []*proxmox.AgentNetworkIface{
				{
					Name: DefaultInstanceNetworkInterface,
					IPAddresses: []*proxmox.AgentNetworkIPAddress{
						{
							IPAddressType: NetworkProtocolIPv4,
							IPAddress:     "8.8.8.8",
						},
						{
							IPAddressType: NetworkProtocolIPv4,
							IPAddress:     "192.168.0.1",
						},
						{
							IPAddressType: NetworkProtocolIPv6,
							IPAddress:     "2001:4860:4860::8888",
						},
						{
							IPAddressType: NetworkProtocolIPv6,
							IPAddress:     "fd3b:47fc:de09::1",
						},
					},
				},
			},

			expectedError:           nil,
			expectedInternalAddress: "fd3b:47fc:de09::1",
			expectedExternalAddress: "2001:4860:4860::8888",
		},
		{
			name: "Any with only internal address",

			requestedInterface: DefaultInstanceNetworkInterface,
			requestedProtocol:  NetworkProtocolAny,
			networkInterfaces: []*proxmox.AgentNetworkIface{
				{
					Name: DefaultInstanceNetworkInterface,
					IPAddresses: []*proxmox.AgentNetworkIPAddress{
						{
							IPAddressType: NetworkProtocolIPv4,
							IPAddress:     "192.168.0.1",
						},
						{
							IPAddressType: NetworkProtocolIPv6,
							IPAddress:     "fd3b:47fc:de09::1",
						},
					},
				},
			},

			expectedError:           nil,
			expectedInternalAddress: "fd3b:47fc:de09::1",
			expectedExternalAddress: "fd3b:47fc:de09::1",
		},
		{
			name: "Forced IPv4 with only internal address",

			requestedInterface: DefaultInstanceNetworkInterface,
			requestedProtocol:  NetworkProtocolIPv4,
			networkInterfaces: []*proxmox.AgentNetworkIface{
				{
					Name: DefaultInstanceNetworkInterface,
					IPAddresses: []*proxmox.AgentNetworkIPAddress{
						{
							IPAddressType: NetworkProtocolIPv4,
							IPAddress:     "192.168.0.1",
						},
						{
							IPAddressType: NetworkProtocolIPv6,
							IPAddress:     "fd3b:47fc:de09::1",
						},
					},
				},
			},

			expectedError:           nil,
			expectedInternalAddress: "192.168.0.1",
			expectedExternalAddress: "192.168.0.1",
		},
		{
			name: "Forced IPv6 with only internal address",

			requestedInterface: DefaultInstanceNetworkInterface,
			requestedProtocol:  NetworkProtocolIPv6,
			networkInterfaces: []*proxmox.AgentNetworkIface{
				{
					Name: DefaultInstanceNetworkInterface,
					IPAddresses: []*proxmox.AgentNetworkIPAddress{
						{
							IPAddressType: NetworkProtocolIPv4,
							IPAddress:     "192.168.0.1",
						},
						{
							IPAddressType: NetworkProtocolIPv6,
							IPAddress:     "fd3b:47fc:de09::1",
						},
					},
				},
			},

			expectedError:           nil,
			expectedInternalAddress: "fd3b:47fc:de09::1",
			expectedExternalAddress: "fd3b:47fc:de09::1",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			internalAddress, externalAddress, err := determineAddresses(testCase.networkInterfaces, testCase.requestedInterface, testCase.requestedProtocol)

			require.ErrorIs(t, err, testCase.expectedError)
			require.Equal(t, testCase.expectedInternalAddress, internalAddress)
			require.Equal(t, testCase.expectedExternalAddress, externalAddress)
		})
	}
}
