package keys

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/orchard9/envault/internal/config"
)

// Key represents an SSH public key
type Key struct {
	Type        string // e.g., "ssh-rsa", "ssh-ed25519"
	Data        string // base64-encoded key data
	Comment     string // optional comment (usually email)
	Fingerprint string // SHA256 fingerprint
}

// AuthorizedKeysPath returns the path to authorized_keys file
func AuthorizedKeysPath() (string, error) {
	envaultDir, err := config.EnvaultDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(envaultDir, "authorized_keys"), nil
}

// Load reads all authorized keys
func Load() ([]Key, error) {
	keysPath, err := AuthorizedKeysPath()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(keysPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Key{}, nil
		}
		return nil, fmt.Errorf("failed to open authorized_keys: %w", err)
	}
	defer file.Close()

	var keys []Key
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, err := ParseKey(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}

		keys = append(keys, *key)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read authorized_keys: %w", err)
	}

	return keys, nil
}

// ParseKey parses an SSH public key from OpenSSH format
func ParseKey(line string) (*Key, error) {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid key format (expected at least 2 fields)")
	}

	keyType := parts[0]
	keyData := parts[1]
	var comment string
	if len(parts) > 2 {
		comment = strings.Join(parts[2:], " ")
	}

	// Generate fingerprint
	fingerprint := generateFingerprint(keyData)

	return &Key{
		Type:        keyType,
		Data:        keyData,
		Comment:     comment,
		Fingerprint: fingerprint,
	}, nil
}

// Add adds a new SSH public key to authorized_keys
func Add(keyString string) error {
	// Parse the key to validate it
	key, err := ParseKey(keyString)
	if err != nil {
		return fmt.Errorf("invalid key: %w", err)
	}

	// Check if key already exists
	existing, err := Load()
	if err != nil {
		return err
	}

	for _, k := range existing {
		if k.Data == key.Data {
			return fmt.Errorf("key already exists (fingerprint: %s)", k.Fingerprint)
		}
	}

	// Append to authorized_keys
	keysPath, err := AuthorizedKeysPath()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(keysPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open authorized_keys: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(keyString + "\n"); err != nil {
		return fmt.Errorf("failed to write key: %w", err)
	}

	return nil
}

// Remove removes an SSH public key by fingerprint
func Remove(fingerprint string) error {
	keys, err := Load()
	if err != nil {
		return err
	}

	// Filter out the key to remove
	var filtered []Key
	found := false
	for _, k := range keys {
		if k.Fingerprint == fingerprint {
			found = true
			continue
		}
		filtered = append(filtered, k)
	}

	if !found {
		return fmt.Errorf("key with fingerprint %s not found", fingerprint)
	}

	// Rewrite authorized_keys
	keysPath, err := AuthorizedKeysPath()
	if err != nil {
		return err
	}

	file, err := os.Create(keysPath)
	if err != nil {
		return fmt.Errorf("failed to create authorized_keys: %w", err)
	}
	defer file.Close()

	for _, k := range filtered {
		line := fmt.Sprintf("%s %s", k.Type, k.Data)
		if k.Comment != "" {
			line += " " + k.Comment
		}
		if _, err := file.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("failed to write key: %w", err)
		}
	}

	return nil
}

// generateFingerprint creates a SHA256 fingerprint of the key data
func generateFingerprint(keyData string) string {
	hash := sha256.Sum256([]byte(keyData))
	return fmt.Sprintf("%x", hash[:8]) // First 8 bytes for shorter fingerprint
}

// String returns a formatted string representation of the key
func (k *Key) String() string {
	s := fmt.Sprintf("%s (%s)", k.Fingerprint, k.Type)
	if k.Comment != "" {
		s += fmt.Sprintf(" - %s", k.Comment)
	}
	return s
}
