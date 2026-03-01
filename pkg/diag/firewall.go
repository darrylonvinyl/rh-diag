package diag

import (
	"bytes"
	"os/exec"
	"strings"
)

// FirewallInfo returns the aggregate state of the host firewall
type FirewallInfo struct {
	State       string              // "running", "not running"
	ActiveZones []string            // e.g., ["public", "internal"]
	ZoneRules   map[string]ZoneData // Keys are zone names
	Errors      []string
}

// ZoneData captures allowed services and ports for a specific zone
type ZoneData struct {
	Services []string // e.g., ["ssh", "cockpit", "dhcpv6-client"]
	Ports    []string // e.g., ["8080/tcp", "443/tcp"]
}

// FirewallRunDiagnostic is the main orchestration function
func FirewallRunDiagnostic() FirewallInfo {
	info := FirewallInfo{
		ZoneRules: make(map[string]ZoneData),
	}

	info.State = "not running"
	if isFirewalldRunning() {
		info.State = "running"
	} else {
		return info // No need to query zones if the daemon is dead
	}

	zones, err := getActiveZones()
	if err != nil {
		info.Errors = append(info.Errors, "Failed to get active zones: "+err.Error())
		return info
	}
	info.ActiveZones = zones

	for _, zone := range zones {
		data, err := getZoneRules(zone)
		if err != nil {
			info.Errors = append(info.Errors, "Failed to get rules for zone "+zone+": "+err.Error())
			continue
		}
		info.ZoneRules[zone] = data
	}

	return info
}

// isFirewalldRunning executes `firewall-cmd --state`
func isFirewalldRunning() bool {
	cmd := exec.Command("firewall-cmd", "--state")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return false
	}
	return strings.TrimSpace(out.String()) == "running"
}

// getActiveZones executes `firewall-cmd --get-active-zones`
// It usually returns a format like:
// public
//
//	interfaces: eth0
func getActiveZones() ([]string, error) {
	cmd := exec.Command("firewall-cmd", "--get-active-zones")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	var zones []string
	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		// Active zones are printed without indentation. Interfaces under them are indented.
		if len(line) > 0 && !strings.HasPrefix(line, " ") {
			zones = append(zones, strings.TrimSpace(line))
		}
	}
	return zones, nil
}

// getZoneRules executes `firewall-cmd --zone=<zone> --list-all`
// and parses out the "services:" and "ports:" lines.
func getZoneRules(zone string) (ZoneData, error) {
	cmd := exec.Command("firewall-cmd", "--zone="+zone, "--list-all")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return ZoneData{}, err
	}

	var data ZoneData
	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "services:") {
			svcStr := strings.TrimSpace(strings.TrimPrefix(line, "services:"))
			if svcStr != "" {
				data.Services = strings.Fields(svcStr)
			}
		} else if strings.HasPrefix(line, "ports:") {
			portStr := strings.TrimSpace(strings.TrimPrefix(line, "ports:"))
			if portStr != "" {
				data.Ports = strings.Fields(portStr)
			}
		}
	}
	return data, nil
}
