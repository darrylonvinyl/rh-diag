package diag

import (
	"testing"
)

func TestFirewallRunDiagnostic(t *testing.T) {
	// The firewall diagnostics rely heavily on executing `firewall-cmd`.
	// In our Dockerized test environment, or generic CI environments, this
	// daemon will almost universally be missing or turned off.
	// Therefore, our primary test is ensuring it fails gracefully and
	// reports "not running" rather than panicking or crashing when parsing.

	info := FirewallRunDiagnostic()

	// In an isolated CI or container, it should safely return not running
	if info.State != "not running" {
		t.Logf("Notice: firewalld is actually running in this test environment. State: %s", info.State)
	} else {
		t.Log("Standard test environment detected: firewalld is safely reported as 'not running'.")
	}

	// We expect the parser to skip polling zones if the daemon is dead.
	if info.State == "not running" && len(info.ActiveZones) > 0 {
		t.Errorf("Firewalld is not running, but ActiveZones were returned: %v", info.ActiveZones)
	}

	if info.State == "not running" && len(info.Errors) > 0 {
		t.Errorf("Firewalld is not running, but Errors were populated: %v", info.Errors)
	}
}
