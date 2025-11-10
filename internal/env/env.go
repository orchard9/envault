package env

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/orchard9/envault/internal/config"
	"github.com/orchard9/envault/internal/crypto"
)

// Load decrypts and writes environment secrets to configured target files
func Load(envName string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	environment, err := cfg.GetEnvironment(envName)
	if err != nil {
		return err
	}

	// Decrypt secrets
	plaintext, err := crypto.Decrypt(envName)
	if err != nil {
		return fmt.Errorf("failed to decrypt %s: %w", envName, err)
	}

	// Write to each target
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	for _, target := range environment.Targets {
		targetPath := filepath.Join(cwd, target.Path)

		// Create parent directory if it doesn't exist
		dir := filepath.Dir(targetPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		// Write file atomically (write to temp file, then rename)
		tempPath := targetPath + ".tmp"
		if err := os.WriteFile(tempPath, plaintext, 0600); err != nil {
			return fmt.Errorf("failed to write %s: %w", targetPath, err)
		}

		if err := os.Rename(tempPath, targetPath); err != nil {
			os.Remove(tempPath) // Clean up temp file on error
			return fmt.Errorf("failed to rename %s: %w", targetPath, err)
		}
	}

	return nil
}

// Validate checks if all target paths are valid
func Validate(envName string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	environment, err := cfg.GetEnvironment(envName)
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	for _, target := range environment.Targets {
		targetPath := filepath.Join(cwd, target.Path)

		// Check if path is absolute (should be relative)
		if filepath.IsAbs(target.Path) {
			return fmt.Errorf("target path %s should be relative, not absolute", target.Path)
		}

		// Check if parent directory exists or can be created
		dir := filepath.Dir(targetPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("cannot create directory %s: %w", dir, err)
		}
	}

	return nil
}

// ListTargets shows where secrets will be written for an environment
func ListTargets(envName string) ([]string, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	environment, err := cfg.GetEnvironment(envName)
	if err != nil {
		return nil, err
	}

	var targets []string
	for _, target := range environment.Targets {
		targets = append(targets, target.Path)
	}

	return targets, nil
}
