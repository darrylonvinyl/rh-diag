package diag

import (
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"strings"
)

// NetInfo encapsulates the results of our network diagnostics
type NetInfo struct {
	Interfaces     []InterfaceInfo
	DefaultGateway string
	GatewayPing    bool
	InternetPing   bool
	Errors         []string
}

// InterfaceInfo holds basic data about a network interface
type InterfaceInfo struct {
	Name      string
	Flags     string
	Addresses []string
}

// NetRunDiagnostic executes the battery of network checks and returns the aggregated results
func NetRunDiagnostic() NetInfo {
	info := NetInfo{}

	// 1. Enumerate Interfaces
	ifaces, err := net.Interfaces()
	if err != nil {
		info.Errors = append(info.Errors, fmt.Sprintf("Failed to list interfaces: %v", err))
	} else {
		for _, i := range ifaces {
			// Skip inactive interfaces and loopback for cleaner output, though we log loopback
			if i.Flags&net.FlagUp == 0 {
				continue
			}

			ifaceInfo := InterfaceInfo{
				Name:  i.Name,
				Flags: i.Flags.String(),
			}

			addrs, err := i.Addrs()
			if err != nil {
				info.Errors = append(info.Errors, fmt.Sprintf("Failed to get addresses for %s: %v", i.Name, err))
				continue
			}

			for _, addr := range addrs {
				// We only care about the IP, not the subnet mask for the display
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				ifaceInfo.Addresses = append(ifaceInfo.Addresses, ip.String())
			}
			info.Interfaces = append(info.Interfaces, ifaceInfo)
		}
	}

	// 2. Determine Default Gateway via OS execution
	// We use execution here because Go does not have a native, cross-platform, non-root
	// way to read the routing table. Executing 'ip route' is the standard on Linux.
	gw, err := getDefaultGateway()
	if err != nil {
		info.Errors = append(info.Errors, fmt.Sprintf("Failed to determine default gateway: %v", err))
	} else {
		info.DefaultGateway = gw
	}

	// 3. Gateway Connectivity Check
	if info.DefaultGateway != "" {
		info.GatewayPing = pingHost(info.DefaultGateway)
	}

	// 4. Outbound Connectivity Check (Reliable external host)
	info.InternetPing = pingHost("8.8.8.8")

	return info
}

// getDefaultGateway parses the output of `ip route show default`
func getDefaultGateway() (string, error) {
	cmd := exec.Command("ip", "route", "show", "default")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}

	output := strings.TrimSpace(out.String())
	if output == "" {
		return "", fmt.Errorf("no default route found")
	}

	// Example output: "default via 192.168.1.1 dev eth0 proto dhcp metric 100"
	parts := strings.Fields(output)
	for i, part := range parts {
		if part == "via" && i+1 < len(parts) {
			return parts[i+1], nil
		}
	}

	return "", fmt.Errorf("could not parse gateway from: %s", output)
}

// pingHost attempts a 1-packet, 2-second timeout ICMP ping using the system ping command
func pingHost(host string) bool {
	// -c 1 (1 packet), -W 2 (2 seconds timeout)
	cmd := exec.Command("ping", "-c", "1", "-W", "2", host)
	err := cmd.Run()
	// Ping returns exit code 0 on success, non-zero on failure
	return err == nil
}
