package cmd

import (
	"fmt"

	"github.com/darrylonvinyl/rh-diag/pkg/diag"
	"github.com/spf13/cobra"
)

var netCmd = &cobra.Command{
	Use:   "net",
	Short: "Diagnose network interfaces, routing, and connectivity",
	Long: `Inspects local network interfaces for states and IP assignments, 
identifies the active default gateway via the OS routing table, 
and sequentially validates ICMP connectivity to both the gateway and the internet.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running Networking & Routing diagnostics...")

		results := diag.NetRunDiagnostic()

		fmt.Println("\n--- Active Interfaces ---")
		if len(results.Interfaces) == 0 {
			fmt.Println("No active interfaces found!")
		} else {
			for _, iface := range results.Interfaces {
				fmt.Printf("[%s] Flags: %s\n", iface.Name, iface.Flags)
				for _, ip := range iface.Addresses {
					fmt.Printf("  -> IP: %s\n", ip)
				}
			}
		}

		fmt.Println("\n--- Routing & Connectivity ---")

		if results.DefaultGateway != "" {
			fmt.Printf("Default Gateway : %s\n", results.DefaultGateway)
		} else {
			fmt.Println("Default Gateway : [MISSING/UNREACHABLE]")
		}

		if results.DefaultGateway != "" {
			if results.GatewayPing {
				fmt.Printf("Gateway Ping    : [PASS] (Reachable)\n")
			} else {
				fmt.Printf("Gateway Ping    : [FAIL] (Unreachable - check local firewall or hypervisor switches)\n")
			}
		}

		if results.InternetPing {
			fmt.Printf("Internet Ping   : [PASS] (Reachable against 8.8.8.8)\n")
		} else {
			fmt.Printf("Internet Ping   : [FAIL] (No outbound ICMP. Check NAT routing or edge firewall rules)\n")
		}

		if len(results.Errors) > 0 {
			fmt.Println("\n--- Diagnostic Errors Encountered ---")
			for _, err := range results.Errors {
				fmt.Printf(" - %s\n", err)
			}
		}

		// Calculate overall summary
		if results.DefaultGateway != "" && results.GatewayPing && results.InternetPing {
			fmt.Println("\nSummary: Hardware and basic system routing look HEALTHY.")
		} else {
			fmt.Println("\nSummary: ACTION REQUIRED. Networking checks failed.")
			fmt.Println("         - Use 'nmcli connection show' or 'ip a' to investigate interfaces.")
			fmt.Println("         - Ensure your VM network adapter is attached (e.g., bridged vs NAT in KVM).")
		}
	},
}

func init() {
	rootCmd.AddCommand(netCmd)
}
