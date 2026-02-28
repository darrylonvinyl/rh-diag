package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "rh-diag",
	Short: "A robust system diagnostic CLI tool for Red Hat Enterprise Linux 9",
	Long: `rh-diag is a comprehensive system troubleshooting tool designed for RHEL 9 / Rocky Linux 9.
It analyzes system state, parses logs, and automatically identifies common misconfigurations
related to networking, DNS, and firewalld, providing actionable, solution-oriented data.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Root-level flags can be defined here (e.g., --verbose or --json output formats)
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
}
