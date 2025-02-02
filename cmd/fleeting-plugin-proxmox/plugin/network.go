package plugin

import (
	"errors"
	"net"

	"github.com/luthermonson/go-proxmox"
)

var ErrNoIPAddress = errors.New("failed to determine IP address for instance")

// Determines internal and external address for given interfaces.
func determineAddresses(networkInterfaces []*proxmox.AgentNetworkIface, requestedInterface string, requestedProtocol NetworkProtocol) (string, string, error) {
	internalIPv4, externalIPv4, internalIPv6, externalIPv6 := determinePossibleAddresses(networkInterfaces, requestedInterface)

	// IPv6 (or Any)
	if requestedProtocol == NetworkProtocolIPv6 || requestedProtocol == NetworkProtocolAny {
		// External address is required so use internal if needed
		if externalIPv6 == "" {
			externalIPv6 = internalIPv6
		}

		if externalIPv6 != "" {
			return internalIPv6, externalIPv6, nil
		}
	}

	// IPv4 (or Any)
	if requestedProtocol == NetworkProtocolIPv4 || requestedProtocol == NetworkProtocolAny {
		// External address is required so use internal if needed
		if externalIPv4 == "" {
			externalIPv4 = internalIPv4
		}

		if externalIPv4 != "" {
			return internalIPv4, externalIPv4, nil
		}
	}

	// Did not find the address in requested protocol
	return "", "", ErrNoIPAddress
}

// Finds possible IPv4 and IPv6 addresses for given interfaces.
//
//nolint:nakedret,nonamedreturns
func determinePossibleAddresses(networkInterfaces []*proxmox.AgentNetworkIface, requestedInterface string) (internalIPv4, externalIPv4, internalIPv6, externalIPv6 string) {
	for _, networkInterface := range networkInterfaces {
		if networkInterface.Name != requestedInterface {
			continue
		}

		for _, address := range networkInterface.IPAddresses {
			parsedAddress := net.ParseIP(address.IPAddress)

			if parsedAddress == nil {
				continue
			}

			if parsedAddress.IsLoopback() || parsedAddress.IsUnspecified() {
				continue
			}

			if address.IPAddressType == "ipv4" {
				if parsedAddress.IsPrivate() {
					internalIPv4 = address.IPAddress
				} else if parsedAddress.IsGlobalUnicast() {
					externalIPv4 = address.IPAddress
				}
			}

			if address.IPAddressType == "ipv6" {
				if parsedAddress.IsPrivate() {
					internalIPv6 = address.IPAddress
				} else if parsedAddress.IsGlobalUnicast() {
					externalIPv6 = address.IPAddress
				}
			}
		}

		// We found requested interface so we can break the loop
		return
	}

	return
}
