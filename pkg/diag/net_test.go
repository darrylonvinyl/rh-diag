package diag

import (
	"testing"
)

// We rely heavily on OS execution (`ip` and `ping`), making traditional unit
// testing difficult without heavy mocking or dependency injection.
// For `rh-diag`, we treat these as integration tests. They assert that the
// OS commands exist and our parsing logic doesn't panic on a valid Linux system.

func TestGetDefaultGateway(t *testing.T) {
	// Let's ensure our parsing logic doesn't panic when we run it.
	// We do not assert a specific gateway IP because it changes per environment.
	gw, err := getDefaultGateway()
	if err != nil {
		// Depending on the test environment (e.g. isolated CI), a default gateway
		// might not exist. We just log the error instead of failing the build.
		t.Logf("getDefaultGateway returned error (normal in isolated test environments): %v", err)
	} else {
		t.Logf("Successfully parsed default gateway: %s", gw)
		if gw == "" {
			t.Errorf("Expected a gateway string or error, got empty string")
		}
	}
}

func TestPingHost(t *testing.T) {
	// Test a guaranteed localhost ping to ensure the exec flow and args are correct
	passed := pingHost("127.0.0.1")
	if !passed {
		t.Errorf("Expected ping to 127.0.0.1 to pass, but it failed")
	}

	// Test an invalid host to ensure failure state logic works
	passedInvalid := pingHost("255.255.255.255")
	if passedInvalid {
		t.Errorf("Expected ping to 255.255.255.255 to fail, but it passed")
	}
}

func TestNetRunDiagnostic(t *testing.T) {
	// Run the full battery
	results := NetRunDiagnostic()

	// Broad assertions
	if len(results.Interfaces) == 0 {
		t.Errorf("Expected to find at least one network interface, got 0")
	}

	// The `if results.GatewayPing` bool should only be true if we actually found a GW
	if results.GatewayPing && results.DefaultGateway == "" {
		t.Errorf("GatewayPing was true, but DefaultGateway was empty!")
	}
}
