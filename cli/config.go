package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// DefaultConfigFile is the default kuang configuration file.
const DefaultConfigFile = "kuang.yaml"

// Config is the top-level kuang configuration.
type Config struct {
	Agents  map[string]AgentConfig `yaml:"agents"`
	Server  ServerConfig           `yaml:"server"`
	Modules []string               `yaml:"modules"`
	Scopes  []string               `yaml:"scopes"`
}

// ServerConfig holds server settings.
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// AgentConfig defines an agent's scope grants.
type AgentConfig struct {
	Scopes []string `yaml:"scopes"`
}

// LoadConfig reads and parses a kuang.yaml file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path) //nolint:gosec // user config file
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Host == "" {
		cfg.Server.Host = "localhost"
	}

	return &cfg, nil
}

// FindConfig searches for kuang.yaml starting from the current directory
// and walking up to the filesystem root.
func FindConfig() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		path := filepath.Join(dir, DefaultConfigFile)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("kuang.yaml not found")
		}
		dir = parent
	}
}

// ResolveAgentScopes resolves an agent's scope patterns against the defined
// scopes. Patterns support * wildcards matching any segment.
func (c *Config) ResolveAgentScopes(agentName string) ([]string, error) {
	agent, ok := c.Agents[agentName]
	if !ok {
		return nil, fmt.Errorf("agent %q not defined in config", agentName)
	}

	var resolved []string
	for _, pattern := range agent.Scopes {
		matches := matchScopes(pattern, c.Scopes)
		if len(matches) == 0 {
			return nil, fmt.Errorf("scope pattern %q matches no defined scopes", pattern)
		}
		resolved = append(resolved, matches...)
	}

	return dedupe(resolved), nil
}

// matchScopes returns all scopes that match the given pattern.
// The pattern supports * as a wildcard for any segment between hyphens.
func matchScopes(pattern string, scopes []string) []string {
	// Literal "*" matches everything.
	if pattern == "*" {
		return scopes
	}

	var matches []string
	for _, scope := range scopes {
		if globMatch(pattern, scope) {
			matches = append(matches, scope)
		}
	}
	return matches
}

// globMatch matches a pattern against a value where * matches any substring
// between hyphens (segment-level wildcard).
func globMatch(pattern, value string) bool {
	patParts := strings.Split(pattern, "-")
	valParts := strings.Split(value, "-")

	if len(patParts) != len(valParts) {
		return false
	}

	for i, pp := range patParts {
		if pp == "*" {
			continue
		}
		if pp != valParts[i] {
			return false
		}
	}
	return true
}

func dedupe(ss []string) []string {
	seen := make(map[string]bool, len(ss))
	var result []string
	for _, s := range ss {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}
