package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ritiksuman07/sqlpilot/internal/tui"
)

var (
	version = "0.2.0"
)

func main() {
	root := &cobra.Command{
		Use:   "sqlpilot",
		Short: "SQLPilot is a keyboard-first terminal SQL explorer",
		RunE: func(cmd *cobra.Command, args []string) error {
			dsn, _ := cmd.Flags().GetString("dsn")
			profile, _ := cmd.Flags().GetString("profile")
			return tui.Run(tui.Options{DSN: dsn, Profile: profile, Version: version})
		},
	}

	root.Flags().String("dsn", "", "Inline DSN string (e.g. postgres://user:pass@host/db)")
	root.Flags().String("profile", "", "Named connection profile")
	root.Version = version
	root.SetVersionTemplate("sqlpilot {{.Version}}\n")

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
