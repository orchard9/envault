package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/orchard9/envault/internal/config"
	"github.com/orchard9/envault/internal/crypto"
	"github.com/orchard9/envault/internal/env"
	"github.com/orchard9/envault/internal/keys"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Check if age is installed for crypto operations
	if needsAge(command) {
		if err := crypto.CheckAge(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	switch command {
	case "init":
		handleInit()
	case "dev", "staging", "prod":
		handleLoadEnv(command)
	case "add-key":
		handleAddKey()
	case "remove-key":
		handleRemoveKey()
	case "list-keys":
		handleListKeys()
	case "encrypt":
		handleEncrypt()
	case "decrypt":
		handleDecrypt()
	case "reencrypt":
		handleReencrypt()
	case "check":
		handleCheck()
	case "version", "--version", "-v":
		fmt.Printf("envault version %s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleInit() {
	envaultDir, err := config.EnvaultDir()
	if err != nil {
		fatal("Failed to determine .envault directory: %v", err)
	}

	// Check if .envault already exists
	if _, err := os.Stat(envaultDir); err == nil {
		fatal(".envault directory already exists")
	}

	// Create .envault directory
	if err := os.MkdirAll(envaultDir, 0755); err != nil {
		fatal("Failed to create .envault directory: %v", err)
	}

	// Create default config.yaml
	cfg := config.DefaultConfig()
	if err := cfg.Save(); err != nil {
		fatal("Failed to create config.yaml: %v", err)
	}

	// Create empty authorized_keys file
	keysPath, err := keys.AuthorizedKeysPath()
	if err != nil {
		fatal("Failed to determine authorized_keys path: %v", err)
	}

	if err := os.WriteFile(keysPath, []byte(""), 0644); err != nil {
		fatal("Failed to create authorized_keys: %v", err)
	}

	// Create .gitignore to ignore plaintext files
	gitignorePath := filepath.Join(envaultDir, ".gitignore")
	gitignoreContent := "*.plaintext\n*.plain\n*.decrypted\n"
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		fatal("Failed to create .gitignore: %v", err)
	}

	fmt.Println("✓ Initialized .envault directory")
	fmt.Println("✓ Created config.yaml with default configuration")
	fmt.Println("✓ Created authorized_keys file")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Add SSH public keys: envault add-key <public-key>")
	fmt.Println("  2. Create plaintext secrets file")
	fmt.Println("  3. Encrypt secrets: envault encrypt dev <plaintext-file>")
	fmt.Println("  4. Commit: git add .envault && git commit -m 'chore: add envault'")
}

func handleLoadEnv(envName string) {
	if err := env.Load(envName); err != nil {
		fatal("Failed to load %s environment: %v", envName, err)
	}

	targets, err := env.ListTargets(envName)
	if err != nil {
		fatal("Failed to list targets: %v", err)
	}

	fmt.Printf("✓ Loaded %s secrets to:\n", envName)
	for _, target := range targets {
		fmt.Printf("  - %s\n", target)
	}
}

func handleAddKey() {
	if len(os.Args) < 3 {
		fatal("Usage: envault add-key <public-key-or-file>")
	}

	keyArg := os.Args[2]

	// Check if it's a file path
	var keyString string
	if _, err := os.Stat(keyArg); err == nil {
		data, err := os.ReadFile(keyArg)
		if err != nil {
			fatal("Failed to read key file: %v", err)
		}
		keyString = strings.TrimSpace(string(data))
	} else {
		// Treat as raw key string (allow multi-word input)
		keyString = strings.Join(os.Args[2:], " ")
	}

	if err := keys.Add(keyString); err != nil {
		fatal("Failed to add key: %v", err)
	}

	fmt.Println("✓ Added SSH public key")
	fmt.Println("\nNext steps:")
	fmt.Println("  - Encrypt/re-encrypt environments: envault encrypt <env> <file>")
	fmt.Println("  - Or re-encrypt existing: envault reencrypt <env>")
}

func handleRemoveKey() {
	if len(os.Args) < 3 {
		fatal("Usage: envault remove-key <fingerprint>")
	}

	fingerprint := os.Args[2]

	if err := keys.Remove(fingerprint); err != nil {
		fatal("Failed to remove key: %v", err)
	}

	fmt.Println("✓ Removed SSH public key")
	fmt.Println("\nIMPORTANT: Re-encrypt all environments to revoke access:")
	fmt.Println("  envault reencrypt")
}

func handleListKeys() {
	authorizedKeys, err := keys.Load()
	if err != nil {
		fatal("Failed to load keys: %v", err)
	}

	if len(authorizedKeys) == 0 {
		fmt.Println("No authorized keys found")
		fmt.Println("\nAdd keys with: envault add-key <public-key>")
		return
	}

	fmt.Printf("Authorized keys (%d):\n", len(authorizedKeys))
	for i, key := range authorizedKeys {
		fmt.Printf("  %d. %s\n", i+1, key.String())
	}
}

func handleEncrypt() {
	if len(os.Args) < 4 {
		fatal("Usage: envault encrypt <environment> <plaintext-file>")
	}

	envName := os.Args[2]
	plaintextPath := os.Args[3]

	if err := crypto.EncryptFile(envName, plaintextPath); err != nil {
		fatal("Failed to encrypt: %v", err)
	}

	fmt.Printf("✓ Encrypted %s to .envault/%s\n", plaintextPath, envName)
	fmt.Println("\nNext steps:")
	fmt.Println("  - Test decryption: envault decrypt", envName)
	fmt.Println("  - Commit: git add .envault && git commit -m 'chore: update secrets'")
}

func handleDecrypt() {
	if len(os.Args) < 3 {
		fatal("Usage: envault decrypt <environment>")
	}

	envName := os.Args[2]

	if err := crypto.DecryptToWriter(envName, os.Stdout); err != nil {
		fatal("Failed to decrypt: %v", err)
	}
}

func handleReencrypt() {
	// If no environment specified, re-encrypt all
	if len(os.Args) < 3 {
		envs, err := crypto.ReencryptAll()
		if err != nil {
			// Check if we partially succeeded
			if len(envs) > 0 {
				fmt.Printf("✓ Re-encrypted: %s\n", strings.Join(envs, ", "))
			}
			fatal("Failed to reencrypt all: %v", err)
		}

		fmt.Printf("✓ Re-encrypted all environments with current authorized_keys:\n")
		for _, env := range envs {
			fmt.Printf("  - %s\n", env)
		}
		return
	}

	// Re-encrypt specific environment
	envName := os.Args[2]

	if err := crypto.Reencrypt(envName); err != nil {
		fatal("Failed to reencrypt: %v", err)
	}

	fmt.Printf("✓ Re-encrypted %s with current authorized_keys\n", envName)
}

func handleCheck() {
	cfg, err := config.Load()
	if err != nil {
		fatal("Failed to load config: %v", err)
	}

	fmt.Println("Checking envault configuration...\n")

	// Check authorized keys
	authorizedKeys, err := keys.Load()
	if err != nil {
		fmt.Printf("✗ Failed to load authorized_keys: %v\n", err)
	} else {
		fmt.Printf("✓ Authorized keys: %d\n", len(authorizedKeys))
	}

	// Check each environment
	for envName := range cfg.Environments {
		fmt.Printf("\nEnvironment: %s\n", envName)

		// Check if encrypted file exists
		envaultDir, _ := config.EnvaultDir()
		env, _ := cfg.GetEnvironment(envName)
		encryptedPath := filepath.Join(envaultDir, env.EncryptedFile)

		if _, err := os.Stat(encryptedPath); os.IsNotExist(err) {
			fmt.Printf("  ✗ Encrypted file missing: %s\n", env.EncryptedFile)
			continue
		}
		fmt.Printf("  ✓ Encrypted file exists: %s\n", env.EncryptedFile)

		// Check if we can decrypt
		if err := crypto.CanDecrypt(envName); err != nil {
			fmt.Printf("  ✗ Cannot decrypt: %v\n", err)
		} else {
			fmt.Println("  ✓ Can decrypt with your SSH key")
		}

		// List targets
		fmt.Printf("  ✓ Targets: %d\n", len(env.Targets))
		for _, target := range env.Targets {
			fmt.Printf("    - %s\n", target.Path)
		}
	}
}

func printUsage() {
	fmt.Println("envault - Encrypted environment secrets")
	fmt.Println("\nUsage:")
	fmt.Println("  envault <command> [arguments]")
	fmt.Println("\nCommands:")
	fmt.Println("  init                          Initialize .envault directory")
	fmt.Println("  dev|staging|prod              Load environment secrets")
	fmt.Println("  add-key <public-key>          Add SSH public key")
	fmt.Println("  remove-key <fingerprint>      Remove SSH public key")
	fmt.Println("  list-keys                     List authorized keys")
	fmt.Println("  encrypt <env> <file>          Encrypt plaintext file")
	fmt.Println("  decrypt <env>                 Decrypt environment to stdout")
	fmt.Println("  reencrypt [env]               Re-encrypt with updated keys (all envs if not specified)")
	fmt.Println("  check                         Verify configuration")
	fmt.Println("  version                       Show version")
	fmt.Println("  help                          Show this help")
	fmt.Println("\nExamples:")
	fmt.Println("  envault init")
	fmt.Println("  envault add-key ~/.ssh/id_rsa.pub")
	fmt.Println("  envault encrypt dev secrets.txt")
	fmt.Println("  envault dev")
	fmt.Println("\nDocumentation: https://github.com/orchard9/envault")
}

func needsAge(command string) bool {
	cryptoCommands := []string{"encrypt", "decrypt", "reencrypt", "dev", "staging", "prod", "check"}
	for _, cmd := range cryptoCommands {
		if command == cmd {
			return true
		}
	}
	return false
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}
