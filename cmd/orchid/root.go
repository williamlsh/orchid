package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/williamlsh/orchid/cmd/orchid/frontend"
)

var rootCmd = &cobra.Command{
	Use: "Orchid",
}

func run() {
	rootCmd.AddCommand(frontend.Cmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
