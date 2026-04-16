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

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Generate a new CA and server certificate",
		Long:  "Reads scopes from kuang.yaml to set the server's permission ceiling.",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			if err := cli.InitCA(dir, cfg.Scopes); err != nil {
				return err
			}

			fmt.Printf("CA and server cert created in %s/\n", dir)
			fmt.Printf("  server scopes: %s\n", strings.Join(cfg.Scopes, ", "))
			return nil
		},
	}

	cmd.Flags().StringVar(&dir, "dir", cli.DefaultCertDir, "directory for certificates")
	return cmd
}

func certIssueCmd() *cobra.Command {
	var dir string

	cmd := &cobra.Command{
		Use:   "issue <agent>",
		Short: "Issue a client certificate for an agent",
		Long:  "Reads the agent's scopes from kuang.yaml and resolves glob patterns.",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			name := args[0]

			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			scopes, err := cfg.ResolveAgentScopes(name)
			if err != nil {
				return err
			}

			if err := cli.IssueCert(dir, name, scopes); err != nil {
				return err
			}

			fmt.Printf("cert issued for %q in %s/\n", name, dir)
			fmt.Printf("  scopes: %s\n", strings.Join(scopes, ", "))
			return nil
		},
	}

	cmd.Flags().StringVar(&dir, "dir", cli.DefaultCertDir, "directory for certificates")
	return cmd
}

func loadConfig() (*cli.Config, error) {
	path, err := cli.FindConfig()
	if err != nil {
		return nil, err
	}
	return cli.LoadConfig(path)
}
