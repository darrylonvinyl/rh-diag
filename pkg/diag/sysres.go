package diag

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

// SysResInfo returns the aggregate hardware capacity state
type SysResInfo struct {
	MemoryTotal     uint64 // in bytes
	MemoryAvailable uint64 // in bytes (actual usable without swapping)
	SwapTotal       uint64
	SwapFree        uint64
	LoadAvg1        float64 // 1-minute load average
	LoadAvg5        float64 // 5-minute load average
	LoadAvg15       float64 // 15-minute load average
	DiskUsage       []DiskInfo
	Errors          []string
}

// DiskInfo captures usage metrics for a specific mount point
type DiskInfo struct {
	MountPoint   string
	TotalBytes   uint64
	UsedBytes    uint64
	FreeBytes    uint64
	UsagePercent float64
}

// SysResRunDiagnostic orchestrates gathering capacity data without shelling out
func SysResRunDiagnostic() SysResInfo {
	var info SysResInfo

	// 1. Process Memory
	memInfo, err := parseProcMeminfo()
	if err != nil {
		info.Errors = append(info.Errors, "Failed to read memory info: "+err.Error())
	} else {
		// /proc/meminfo values are in kilobytes, multiplying by 1024 for bytes
		info.MemoryTotal = memInfo["MemTotal:"] * 1024

		// MemAvailable is a more accurate representation of usable RAM than Free in Linux
		if val, ok := memInfo["MemAvailable:"]; ok {
			info.MemoryAvailable = val * 1024
		} else {
			// Fallback for very old kernels
			info.MemoryAvailable = memInfo["MemFree:"] * 1024
		}

		info.SwapTotal = memInfo["SwapTotal:"] * 1024
		info.SwapFree = memInfo["SwapFree:"] * 1024
	}

	// 2. Process CPU Load Averages
	l1, l5, l15, err := parseProcLoadavg()
	if err != nil {
		info.Errors = append(info.Errors, "Failed to parse load averages: "+err.Error())
	} else {
		info.LoadAvg1 = l1
		info.LoadAvg5 = l5
		info.LoadAvg15 = l15
	}

	// 3. Process primary local file systems
	mounts, err := getMountConfigs()
	if err != nil {
		info.Errors = append(info.Errors, "Failed to parse mounts: "+err.Error())
	} else {
		for _, m := range mounts {
			di, err := getDiskStats(m)
			if err != nil {
				info.Errors = append(info.Errors, fmt.Sprintf("Failed to stat disk %s: %v", m, err))
				continue
			}
			info.DiskUsage = append(info.DiskUsage, di)
		}
	}

	return info
}

func parseProcMeminfo() (map[string]uint64, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	memMap := make(map[string]uint64)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 {
			val, _ := strconv.ParseUint(fields[1], 10, 64)
			memMap[fields[0]] = val
		}
	}
	return memMap, scanner.Err()
}

func parseProcLoadavg() (float64, float64, float64, error) {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, 0, 0, err
	}

	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		return 0, 0, 0, fmt.Errorf("unexpected /proc/loadavg format")
	}

	l1, _ := strconv.ParseFloat(fields[0], 64)
	l5, _ := strconv.ParseFloat(fields[1], 64)
	l15, _ := strconv.ParseFloat(fields[2], 64)

	return l1, l5, l15, nil
}

// getMountConfigs reads /proc/mounts to uniquely identify persistent local file systems
// (e.g., ext4, xfs, btrfs) and ignores virtual ones like sysfs or tmpfs for noise reduction.
func getMountConfigs() ([]string, error) {
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var mounts []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 3 {
			fsType := fields[2]
			// We only want to alert on actual physical or virtualized block disks
			if fsType == "ext3" || fsType == "ext4" || fsType == "xfs" || fsType == "btrfs" || fsType == "vfat" {
				mounts = append(mounts, fields[1])
			}
		}
	}

	// If we somehow didn't parse any (e.g. running in an odd container overlay), fallback to Root
	if len(mounts) == 0 {
		mounts = append(mounts, "/")
	}

	return mounts, scanner.Err()
}

// getDiskStats relies on the native `sys/unix` package to issue a Statfs syscall
func getDiskStats(mountPath string) (DiskInfo, error) {
	var stat unix.Statfs_t
	err := unix.Statfs(mountPath, &stat)
	if err != nil {
		return DiskInfo{}, err
	}

	// Blocks * BlockSize = Bytes
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize) // Bavail is free blocks unprivileged users can use
	used := total - free

	var usagePct float64
	if total > 0 {
		usagePct = (float64(used) / float64(total)) * 100
	}

	return DiskInfo{
		MountPoint:   mountPath,
		TotalBytes:   total,
		UsedBytes:    used,
		FreeBytes:    free,
		UsagePercent: usagePct,
	}, nil
}
