package diag

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

// DNSResult contains the test result for a single nameserver
type DNSResult struct {
	Nameserver string
	HostTested string
	Passed     bool
	ErrorMsg   string
}

// DNSRunDiagnostic is the core orchestrator for testing local DNS resolution
func DNSRunDiagnostic() ([]DNSResult, error) {
	// Step 1: Parse /etc/resolv.conf
	// In RHEL 9, NetworkManager typically populates this file dynamically.
	// We read it directly rather than using Go's default resolver to guarantee
	// we are testing the actual configured state of the host.
	nameservers, err := parseResolvConf("/etc/resolv.conf")
	if err != nil {
		return nil, fmt.Errorf("failed to read /etc/resolv.conf: %w", err)
	}

	if len(nameservers) == 0 {
		return nil, fmt.Errorf("no nameservers configured in /etc/resolv.conf. Is NetworkManager managing DNS?")
	}

	var results []DNSResult
	targetHost := "redhat.com"

	// Step 2: Test each nameserver individually
	// We do this to catch "split-brain" DNS issues where the primary is down
	// but the OS is silently falling back to the secondary, delaying resolution.
	for _, ns := range nameservers {
		result := testNameserver(ns, targetHost)
		results = append(results, result)
	}

	return results, nil
}

// parseResolvConf manually extracts the `nameserver` directives
func parseResolvConf(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var nameservers []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "nameserver ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				nameservers = append(nameservers, parts[1])
			}
		}
	}
	return nameservers, scanner.Err()
}

// testNameserver synthetically routes a DNS query directly to the target IP
func testNameserver(nameserverIP, host string) DNSResult {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(2000), // 2-second timeout per RHEL defaults
			}
			// Force traffic over UDP/TCP port 53 to the specific nameserver
			return d.DialContext(ctx, "udp", net.JoinHostPort(nameserverIP, "53"))
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ips, err := r.LookupIPAddr(ctx, host)

	res := DNSResult{
		Nameserver: nameserverIP,
		HostTested: host,
		Passed:     true,
	}

	if err != nil {
		res.Passed = false
		res.ErrorMsg = err.Error()
	} else if len(ips) == 0 {
		res.Passed = false
		res.ErrorMsg = "no records returned"
	}

	return res
}
