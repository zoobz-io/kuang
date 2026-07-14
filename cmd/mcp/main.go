// Package main is the entry point for the kuang MCP bridge.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/zoobz-io/fig"
	"github.com/zoobzio/kuang/internal/mcp"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Load configuration.
	var cfg mcp.Config
	if err := fig.Load(&cfg); err != nil {
		return fmt.Errorf("load security config: %w", err)
	}

	client, err := mcp.NewClient(cfg)
	if err != nil {
		return err
	}

	server := mcp.NewServer(client)
	if err := server.LoadTools(); err != nil {
		return err
	}

	log.Printf("kuang-mcp: loaded %d tools", server.ToolCount())

	return server.Run(os.Stdin, os.Stdout)
}
