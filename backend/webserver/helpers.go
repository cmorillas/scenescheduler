// backend/webserver/helpers.go
//
// This file contains pure, stateless utility functions for the WebServer package.
// These are functions (not methods) that can be called without a receiver.
//
// Contents:
// - Network Utilities

package webserver

import (
	"net"
)

// =============================================================================
// Network Utilities
// =============================================================================

// getLocalIPs retrieves all non-loopback IPv4 addresses from the host's
// network interfaces. This is useful for displaying connection info in logs
// and status events.
func getLocalIPs() []string {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		// Returning an empty slice is a safe fallback.
		return ips
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}
	return ips
}