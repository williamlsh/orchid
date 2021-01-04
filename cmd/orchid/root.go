package main

import (
	"os"

	"github.com/ossm-org/orchid/cmd/orchid/frontend"
	"github.com/spf13/cobra"
)

var (
	frontendHost string
	frontendPort int
)

var rootCmd = &cobra.Command{
	Use: "Orchid",
}

func execute() {
	rootCmd.AddCommand(frontend.FrontendCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
