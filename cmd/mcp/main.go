// Package main is the entry point for the kuang MCP bridge.
package main

import (
	"log"
	"os"

	"github.com/zoobz-io/kuang/mcp"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	baseURL := envOrDefault("KUANG_URL", "https://localhost:8080")
	caCert := envOrDefault("KUANG_CA_CERT", "certs/ca.pem")
	cert := envOrDefault("KUANG_CERT", "certs/client.pem")
	key := envOrDefault("KUANG_KEY", "certs/client-key.pem")

	client, err := mcp.NewClient(baseURL, caCert, cert, key)
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

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
