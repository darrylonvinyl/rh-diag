package diag

import (
	"runtime"
	"testing"
)

func TestSysResRunDiagnostic(t *testing.T) {
	// The sysres routines read from /proc. These are native and stable
	// heavily relied upon endpoints universally available on all Linux.

	if runtime.GOOS != "linux" {
		t.Skip("Skipping /proc tests, test environment is not Linux")
	}

	info := SysResRunDiagnostic()

	// Generic assertions: we expect a standard linux host to have memory
	if info.MemoryTotal == 0 {
		t.Errorf("Expected MemoryTotal > 0, got 0")
	}

	// Available should never realistically be higher than total
	if info.MemoryAvailable > info.MemoryTotal {
		t.Errorf("MemoryAvailable (%d) > MemoryTotal (%d), parsing is wrong", info.MemoryAvailable, info.MemoryTotal)
	}

	// Loads are floats that might occasionally be exactly 0 on an idle isolated env,
	// but we should verify we didn't populate errors.
	if len(info.Errors) > 0 {
		t.Errorf("SysRes routine encountered parsing errors: %v", info.Errors)
	}

	// We should have hit at least the root fallback mount point /
	if len(info.DiskUsage) == 0 {
		t.Errorf("Expected at least one disk mount parsed, got 0")
	} else {
		for _, disk := range info.DiskUsage {
			if disk.TotalBytes == 0 {
				t.Errorf("Disk %s reported 0 bytes total", disk.MountPoint)
			}
		}
	}
}
