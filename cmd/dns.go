package cmd

import (
	"fmt"

	"github.com/darrylonvinyl/rh-diag/pkg/diag"
	"github.com/spf13/cobra"
)

var dnsCmd = &cobra.Command{
	Use:   "dns",
	Short: "Diagnose DNS resolution and /etc/resolv.conf",
	Long: `Parses /etc/resolv.conf for active nameservers and performs synthetic 
resolution tests against each configured server to detect silent failures.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running DNS diagnostics...")

		results, err := diag.DNSRunDiagnostic()
		if err != nil {
			fmt.Printf("[FAIL] DNS Verification Error: %v\n", err)
			return
		}

		allPassed := true
		for _, r := range results {
			if r.Passed {
				fmt.Printf("[PASS] Resolved '%s' via nameserver %s\n", r.HostTested, r.Nameserver)
			} else {
				allPassed = false
				fmt.Printf("[FAIL] Failed to resolve '%s' via nameserver %s. Error: %s\n", r.HostTested, r.Nameserver, r.ErrorMsg)
			}
		}

		if allPassed {
			fmt.Println("\nSummary: ALL nameservers are functioning correctly.")
		} else {
			fmt.Println("\nSummary: ACTION REQUIRED. One or more nameservers failed resolution.")
			fmt.Println("         - Check 'firewall-cmd --list-all' to ensure port 53 UDP/TCP is allowed if filtering outbound traffic.")
			fmt.Println("         - Use 'nmcli device show' or 'nmcli connection show <conn>' to verify provided nameservers.")
		}
	},
}

func init() {
	rootCmd.AddCommand(dnsCmd)
}
