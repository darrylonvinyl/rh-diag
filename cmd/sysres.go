package cmd

import (
	"fmt"

	"github.com/darrylonvinyl/rh-diag/pkg/diag"
	"github.com/spf13/cobra"
)

var sysResCmd = &cobra.Command{
	Use:   "sysres",
	Short: "Diagnose system resource capacity (Memory, CPU Load, Disk)",
	Long: `Reads natively from /proc and issues Statfs syscalls to gather
system resource usage, returning high-level utilization metrics
to instantly identify capacity bottlenecks (e.g., OOM risks or full disks).`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running System Resource diagnostics...")

		results := diag.SysResRunDiagnostic()

		if len(results.Errors) > 0 {
			fmt.Println("\n--- Diagnostic Errors Encountered ---")
			for _, err := range results.Errors {
				fmt.Printf(" - %s\n", err)
			}
		}

		fmt.Printf("\n--- Memory & Swap ---\n")
		fmt.Printf("Memory Total      : %.2f GB\n", float64(results.MemoryTotal)/(1024*1024*1024))
		fmt.Printf("Memory Available  : %.2f GB\n", float64(results.MemoryAvailable)/(1024*1024*1024))
		if results.SwapTotal > 0 {
			fmt.Printf("Swap Total        : %.2f GB\n", float64(results.SwapTotal)/(1024*1024*1024))
			fmt.Printf("Swap Free         : %.2f GB\n", float64(results.SwapFree)/(1024*1024*1024))
		} else {
			fmt.Println("Swap Enabled      : No")
		}

		fmt.Printf("\n--- CPU Load Average ---\n")
		// Load average above the number of CPU cores means processes are waiting
		fmt.Printf("1-Minute Load     : %.2f\n", results.LoadAvg1)
		fmt.Printf("5-Minute Load     : %.2f\n", results.LoadAvg5)
		fmt.Printf("15-Minute Load    : %.2f\n", results.LoadAvg15)

		fmt.Printf("\n--- Persistent Local Disks ---\n")
		capacityWarning := false
		if len(results.DiskUsage) == 0 {
			fmt.Println("No persistent local disks matched (/proc/mounts mapping may have failed)")
		} else {
			for _, disk := range results.DiskUsage {
				totalGB := float64(disk.TotalBytes) / (1024 * 1024 * 1024)
				usedGB := float64(disk.UsedBytes) / (1024 * 1024 * 1024)
				fmt.Printf("[%s] Usage: %.1f%% (%.2f GB / %.2f GB)\n", disk.MountPoint, disk.UsagePercent, usedGB, totalGB)

				if disk.UsagePercent >= 90.0 {
					capacityWarning = true
					fmt.Printf("   -> [WARNING] Disk %s is nearly full!\n", disk.MountPoint)
				}
			}
		}

		// Calculate overall summary
		if capacityWarning {
			fmt.Println("\nSummary: ACTION REQUIRED. One or more disks are at capacity limits.")
		} else {
			memPercent := 0.0
			if results.MemoryTotal > 0 {
				memPercent = 100 - ((float64(results.MemoryAvailable) / float64(results.MemoryTotal)) * 100)
			}
			if memPercent > 95 {
				fmt.Println("\nSummary: WARNING. System is experiencing extreme memory pressure (>95%). OOM kills may occur.")
			} else {
				fmt.Println("\nSummary: Core system capacity (Memory/Disks) appears within normal operating thresholds.")
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(sysResCmd)
}
