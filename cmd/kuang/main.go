// Package main is the entry point for the kuang CLI.
package main

import (
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "kuang",
		Short: "CLI for managing kuang API servers",
	}

	root.AddCommand(certCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
