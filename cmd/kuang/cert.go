package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zoobz-io/kuang/cli"
)

func certCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cert",
		Short: "Manage certificates",
	}
	cmd.AddCommand(certInitCmd())
	cmd.AddCommand(certIssueCmd())
	return cmd
}

func certInitCmd() *cobra.Command {
	var dir string
	var scopes []string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Generate a new CA and server certificate",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := cli.InitCA(dir, scopes); err != nil {
				return err
			}
			fmt.Printf("CA and server cert created in %s/\n", dir)
			fmt.Println("files:")
			fmt.Println("  ca.pem, ca-key.pem       — certificate authority")
			fmt.Println("  server.pem, server-key.pem — server certificate")
			return nil
		},
	}

	cmd.Flags().StringVar(&dir, "dir", cli.DefaultCertDir, "directory for certificates")
	cmd.Flags().StringSliceVar(&scopes, "scopes", nil, "server permission ceiling (OU entries)")

	return cmd
}

func certIssueCmd() *cobra.Command {
	var dir string
	var scopes []string

	cmd := &cobra.Command{
		Use:   "issue <name>",
		Short: "Issue a client certificate for an agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			name := args[0]
			if err := cli.IssueCert(dir, name, scopes); err != nil {
				return err
			}
			fmt.Printf("cert issued for %q in %s/\n", name, dir)
			fmt.Printf("  %s.pem, %s-key.pem\n", name, name)
			if len(scopes) > 0 {
				fmt.Printf("  scopes: %s\n", strings.Join(scopes, ", "))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&dir, "dir", cli.DefaultCertDir, "directory for certificates")
	cmd.Flags().StringSliceVar(&scopes, "scopes", nil, "agent scopes (OU entries)")

	return cmd
}
