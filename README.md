# envault

**Encrypted environment secrets, version controlled, SSH-authenticated**

```bash
# Developer workflow (one command)
envault dev

# Admin workflow
envault add-key <ssh-public-key>
envault encrypt dev
git add .envault/ && git commit -m "chore: update dev secrets"

# Load dev secrets (one command!)
envault dev
```

## What it does

`envault` stores encrypted secrets in your git repo and decrypts them with SSH keys. No services to host, no complicated setup.

**Key features:**
- Config-driven: `.envault/config.yaml` defines where secrets go
- Multi-environment: `dev`, `staging`, `prod` encrypted separately
- Multi-user: Multiple SSH public keys can decrypt (you control the list)
- Zero-trust: Encrypted secrets sit in git, only authorized devs can decrypt
- Simple: One command loads all secrets for an environment

## Architecture

```
.envault/
├── config.yaml          # Defines target locations for secrets
├── dev.age              # Encrypted dev environment secrets
├── staging.age          # Encrypted staging secrets (optional)
├── prod.age             # Encrypted production secrets (optional)
└── authorized_keys      # SSH public keys that can decrypt
```

### Example config.yaml

```yaml
# Define where secrets get written for each environment
environments:
  dev:
    encrypted_file: dev.age
    targets:
      - path: .env                    # Root .env
      - path: apps/yourproject-api/.env     # Service-specific
      - path: apps/billing-api/.env

  staging:
    encrypted_file: staging.age
    targets:
      - path: .env.staging

  prod:
    encrypted_file: prod.age
    targets:
      - path: .env.production
```

## Installation

```bash
# Install envault
go install github.com/orchard9/envault/cmd/envault@latest

# Install age (encryption tool)
brew install age  # macOS
# or: go install filippo.io/age/cmd/...@latest
```

## Setup (First Time - Admin Only)

```bash
# 1. Initialize .envault directory
envault init

# 2. Create secrets file (plain text, temporarily)
cat > .envault/dev.plaintext << 'EOF'
DATABASE_URL=postgresql://yourproject:yourproject@localhost:20325/yourproject
REDIS_URL=redis://localhost:20326
CRIME_API_KEY=dev_crime_key_here
COMM10_API_KEY=dev_comm10_key_here
WARDEN_API_KEY=dev_warden_key_here
EOF

# 3. Add team members' SSH public keys
envault add-key ~/.ssh/id_rsa.pub          # Your key
envault add-key <teammate-public-key>       # Teammate's key

# 4. Encrypt secrets for dev environment
envault encrypt dev .envault/dev.plaintext

# 5. Clean up plaintext (important!)
rm .envault/dev.plaintext

# 6. Commit encrypted secrets
git add .envault/
git commit -m "chore: add encrypted dev secrets"
git push
```

## Developer Workflow

```bash
# Clone repo
git clone git@github.com:orchard9/yourproject.git
cd yourproject

# Load dev secrets (one command!)
envault dev

# Secrets are now decrypted and written to:
# - .env
# - apps/yourproject-api/.env
# - apps/billing-api/.env
# (as defined in .envault/config.yaml)

# Start developing
./bin/install
overmind start
```

## Admin Operations

### Add a new team member

```bash
# Get their SSH public key
# (they run: cat ~/.ssh/id_rsa.pub)

envault add-key <their-public-key>
envault reencrypt dev  # Re-encrypt with new key added
git add .envault/ && git commit -m "chore: add teammate to envault"
git push
```

### Update secrets

```bash
# 1. Decrypt to edit
envault decrypt dev > .envault/dev.plaintext

# 2. Edit the plaintext file
vim .envault/dev.plaintext

# 3. Re-encrypt
envault encrypt dev .envault/dev.plaintext

# 4. Clean up and commit
rm .envault/dev.plaintext
git add .envault/ && git commit -m "chore: update dev secrets"
git push
```

### Remove a team member

```bash
# Remove their key from authorized_keys
envault remove-key <key-fingerprint>

# Re-encrypt all environments (they can no longer decrypt)
envault reencrypt dev
envault reencrypt staging
envault reencrypt prod

git add .envault/ && git commit -m "chore: remove teammate from envault"
git push
```

## Security Model

- **Encrypted at rest**: All secrets encrypted with age (modern, audited)
- **SSH key auth**: Uses developers' existing SSH keys (no new credentials)
- **Access control**: Only SSH keys in `authorized_keys` can decrypt
- **Audit trail**: Git history shows who changed secrets and when
- **Zero-trust**: Encrypted secrets safe in public or private repos
- **Key revocation**: Remove key + reencrypt = immediate access revocation

## How it works

1. **Encryption**: Uses [age](https://age-encryption.org) with SSH keys as recipients
2. **Multi-recipient**: age encrypts once, multiple SSH keys can decrypt
3. **Config-driven**: `config.yaml` defines target file locations
4. **Atomic writes**: Secrets written atomically to prevent partial writes
5. **Gitignored targets**: Target files (`.env`, etc.) stay gitignored

## Integration with bin/install

Update `./bin/install` to load secrets automatically:

```bash
#!/usr/bin/env bash

# ... existing setup code ...

# Load development secrets
if command -v envault &> /dev/null; then
    print_info "Loading development secrets..."
    if envault dev; then
        print_success "Secrets loaded from envault"
    else
        print_error "Failed to load secrets. Ensure you have access to .envault/"
        print_info "Contact admin to add your SSH public key"
        exit 1
    fi
else
    print_warning "envault not installed. Install with: go install github.com/orchard9/envault/cmd/envault@latest"
    exit 1
fi

# ... rest of install script ...
```

## Commands Reference

```bash
envault init                    # Initialize .envault/ directory
envault dev                     # Decrypt and load dev secrets
envault staging                 # Load staging secrets
envault prod                    # Load production secrets
envault add-key <public-key>    # Add SSH public key to authorized_keys
envault remove-key <fingerprint> # Remove key from authorized_keys
envault encrypt <env> <file>    # Encrypt plaintext file for environment
envault decrypt <env>           # Decrypt environment to stdout
envault reencrypt <env>         # Re-encrypt with updated authorized_keys
envault list-keys               # Show authorized SSH keys
envault check                   # Verify you can decrypt environments
```

## Why not Google Secret Manager directly?

GSM is great for production, but for local dev:
- ❌ Requires GCP permissions setup per developer
- ❌ Requires active internet connection
- ❌ More complex IAM management
- ❌ Costs money (small, but non-zero)

`envault`:
- ✅ Works offline
- ✅ Zero infrastructure costs
- ✅ Uses SSH keys devs already have
- ✅ Version controlled (see history of changes)
- ✅ Simple admin (just manage authorized_keys)

**Best of both worlds**: Use `envault` for local dev, GSM for deployed environments (staging/prod on GKE).

## Development

```bash
# Clone and build
git clone git@github.com:orchard9/envault.git
cd envault
make build

# Run tests
make test

# Install locally
make install
```

## License

MIT
