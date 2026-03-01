package cmd

import (
	"fmt"
	"strings"

	"github.com/darrylonvinyl/rh-diag/pkg/diag"
	"github.com/spf13/cobra"
)

var firewallCmd = &cobra.Command{
	Use:   "firewall",
	Short: "Diagnose firewalld state, active zones, and exposed rules",
	Long: `Queries the dynamic firewall daemon (firewalld) to confirm if it is running,
extracts all active zones bound to network interfaces, and lists the 
currently allowed services and ports to detect rogue exposures or dropped traffic.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running Firewalld diagnostics...")

		results := diag.FirewallRunDiagnostic()

		fmt.Printf("\n--- Firewall Daemon State ---\n")
		if results.State == "running" {
			fmt.Println("Status: [ACTIVE] (running)")
		} else {
			fmt.Println("Status: [INACTIVE] (not running)")
			fmt.Println("\nSummary: firewalld is not active. All network traffic may be allowed by default depending on raw iptables/nftables state.")
			return
		}

		if len(results.Errors) > 0 {
			fmt.Println("\n--- Diagnostic Errors Encountered ---")
			for _, err := range results.Errors {
				fmt.Printf(" - %s\n", err)
			}
		}

		fmt.Printf("\n--- Active Zones & Rules ---\n")
		if len(results.ActiveZones) == 0 {
			fmt.Println("No active zones bound to interfaces detected.")
		} else {
			for _, zone := range results.ActiveZones {
				fmt.Printf("\nZone: %s\n", zone)
				rules, ok := results.ZoneRules[zone]
				if !ok {
					continue
				}

				if len(rules.Services) > 0 {
					fmt.Printf("  Allowed Services: %s\n", strings.Join(rules.Services, ", "))
				} else {
					fmt.Printf("  Allowed Services: (none)\n")
				}

				if len(rules.Ports) > 0 {
					fmt.Printf("  Allowed Ports:    %s\n", strings.Join(rules.Ports, ", "))
				} else {
					fmt.Printf("  Allowed Ports:    (none)\n")
				}
			}
		}

		fmt.Println("\nSummary: Firewalld is active and inspecting traffic. Review the allowed services/ports above to ensure expected traffic is not being dropped.")
	},
}

func init() {
	rootCmd.AddCommand(firewallCmd)
}
