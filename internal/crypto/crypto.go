package crypto

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/orchard9/envault/internal/config"
	"github.com/orchard9/envault/internal/keys"
)

// Encrypt encrypts plaintext data for all authorized SSH keys
func Encrypt(envName string, plaintext []byte) error {
	envaultDir, err := config.EnvaultDir()
	if err != nil {
		return err
	}

	// Load config to get encrypted file path
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	env, err := cfg.GetEnvironment(envName)
	if err != nil {
		return err
	}

	encryptedPath := filepath.Join(envaultDir, env.EncryptedFile)

	// Get authorized_keys path
	authorizedKeysPath, err := keys.AuthorizedKeysPath()
	if err != nil {
		return err
	}

	// Verify authorized_keys has at least one key
	authorizedKeys, err := keys.Load()
	if err != nil {
		return err
	}

	if len(authorizedKeys) == 0 {
		return fmt.Errorf("no authorized keys found - run 'envault add-key' first")
	}

	// Run age encryption with authorized_keys file as recipient
	// age can read SSH public keys from a file with -R flag
	cmd := exec.Command("age", "-e", "-o", encryptedPath, "-R", authorizedKeysPath)
	cmd.Stdin = bytes.NewReader(plaintext)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("age encryption failed: %w\nStderr: %s", err, stderr.String())
	}

	return nil
}

// Decrypt decrypts an encrypted file using the user's SSH key
func Decrypt(envName string) ([]byte, error) {
	envaultDir, err := config.EnvaultDir()
	if err != nil {
		return nil, err
	}

	// Load config to get encrypted file path
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	env, err := cfg.GetEnvironment(envName)
	if err != nil {
		return nil, err
	}

	encryptedPath := filepath.Join(envaultDir, env.EncryptedFile)

	// Check if encrypted file exists
	if _, err := os.Stat(encryptedPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("encrypted file %s does not exist", env.EncryptedFile)
	}

	// Find user's SSH private key
	sshKeyPath, err := findSSHPrivateKey()
	if err != nil {
		return nil, err
	}

	// Run age decryption
	cmd := exec.Command("age", "-d", "-i", sshKeyPath, encryptedPath)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("age decryption failed: %w\nStderr: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// EncryptFile encrypts a plaintext file
func EncryptFile(envName string, plaintextPath string) error {
	data, err := os.ReadFile(plaintextPath)
	if err != nil {
		return fmt.Errorf("failed to read plaintext file: %w", err)
	}

	return Encrypt(envName, data)
}

// DecryptToWriter decrypts and writes to an io.Writer
func DecryptToWriter(envName string, w io.Writer) error {
	plaintext, err := Decrypt(envName)
	if err != nil {
		return err
	}

	if _, err := w.Write(plaintext); err != nil {
		return fmt.Errorf("failed to write decrypted data: %w", err)
	}

	return nil
}

// Reencrypt re-encrypts an environment with updated authorized_keys
func Reencrypt(envName string) error {
	// Decrypt with current key
	plaintext, err := Decrypt(envName)
	if err != nil {
		return fmt.Errorf("failed to decrypt: %w", err)
	}

	// Re-encrypt with all authorized keys
	if err := Encrypt(envName, plaintext); err != nil {
		return fmt.Errorf("failed to re-encrypt: %w", err)
	}

	return nil
}

// findSSHPrivateKey finds the user's SSH private key
func findSSHPrivateKey() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	sshDir := filepath.Join(homeDir, ".ssh")

	// Try common key names in order of preference
	keyNames := []string{
		"id_ed25519",
		"id_rsa",
		"id_ecdsa",
		"id_dsa",
	}

	for _, keyName := range keyNames {
		keyPath := filepath.Join(sshDir, keyName)
		if _, err := os.Stat(keyPath); err == nil {
			return keyPath, nil
		}
	}

	return "", fmt.Errorf("no SSH private key found in %s (tried: %s)", sshDir, strings.Join(keyNames, ", "))
}

// CheckAge verifies that the age tool is installed
func CheckAge() error {
	cmd := exec.Command("age", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("age is not installed - install with: brew install age")
	}
	return nil
}

// CanDecrypt checks if the current user can decrypt a specific environment
func CanDecrypt(envName string) error {
	_, err := Decrypt(envName)
	return err
}

// ReencryptAll re-encrypts all environments with updated authorized_keys
func ReencryptAll() ([]string, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	var reencrypted []string
	var errors []string

	for envName := range cfg.Environments {
		if err := Reencrypt(envName); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", envName, err))
		} else {
			reencrypted = append(reencrypted, envName)
		}
	}

	if len(errors) > 0 {
		return reencrypted, fmt.Errorf("failed to re-encrypt some environments:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return reencrypted, nil
}
