package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/ossm-org/orchid/cmd/orchid/frontend"
)

var rootCmd = &cobra.Command{
	Use: "Orchid",
}

func run() {
	rootCmd.AddCommand(frontend.FrontendCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
