package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the .envault/config.yaml structure
type Config struct {
	Environments map[string]Environment `yaml:"environments"`
}

// Environment defines an environment's configuration
type Environment struct {
	EncryptedFile string   `yaml:"encrypted_file"`
	Targets       []Target `yaml:"targets"`
}

// Target defines where decrypted secrets should be written
type Target struct {
	Path string `yaml:"path"`
}

// EnvaultDir returns the path to .envault directory
func EnvaultDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}
	return filepath.Join(cwd, ".envault"), nil
}

// Load reads and parses the config.yaml file
func Load() (*Config, error) {
	envaultDir, err := EnvaultDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(envaultDir, "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config.yaml: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config.yaml: %w", err)
	}

	return &cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if len(c.Environments) == 0 {
		return fmt.Errorf("no environments defined in config.yaml")
	}

	for name, env := range c.Environments {
		if env.EncryptedFile == "" {
			return fmt.Errorf("environment %s: encrypted_file is required", name)
		}
		if len(env.Targets) == 0 {
			return fmt.Errorf("environment %s: at least one target is required", name)
		}
		for i, target := range env.Targets {
			if target.Path == "" {
				return fmt.Errorf("environment %s: target %d has empty path", name, i)
			}
		}
	}

	return nil
}

// GetEnvironment returns the configuration for a specific environment
func (c *Config) GetEnvironment(name string) (*Environment, error) {
	env, ok := c.Environments[name]
	if !ok {
		return nil, fmt.Errorf("environment %s not found in config.yaml", name)
	}
	return &env, nil
}

// DefaultConfig returns a default configuration for initialization
func DefaultConfig() *Config {
	return &Config{
		Environments: map[string]Environment{
			"dev": {
				EncryptedFile: "dev.age",
				Targets: []Target{
					{Path: ".env"},
				},
			},
		},
	}
}

// Save writes the configuration to config.yaml
func (c *Config) Save() error {
	envaultDir, err := EnvaultDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(envaultDir, "config.yaml")
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config.yaml: %w", err)
	}

	return nil
}
