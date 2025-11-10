# envault Quick Start

Get your team using encrypted secrets in 5 minutes.

## Prerequisites

```bash
# Install age encryption tool
brew install age

# Install envault
go install github.com/orchard9/envault/cmd/envault@latest
```

## Setup (Admin - First Time)

```bash
# 1. In your project root, initialize envault
cd ~/path/to/your/project
envault init
# Creates: .envault/config.yaml, .envault/authorized_keys

# 2. Customize .envault/config.yaml for your project
vim .envault/config.yaml
# Default writes to: .env
# For multi-service projects, add multiple targets:
#   - .env
#   - apps/api/.env
#   - apps/worker/.env
# See examples/config.yaml for more options

# 3. Create your secrets file (temporarily)
cat > .envault/dev.plaintext << 'EOF'
DATABASE_URL=postgresql://user:password@localhost:5432/myapp
REDIS_URL=redis://localhost:6379
API_KEY=your_api_key_here
SECRET_KEY=your_secret_key_here
JWT_SECRET=your_jwt_secret_min_32_chars
EMAIL_API_KEY=your_email_api_key_here
EOF

# 4. Add your SSH public key
envault add-key ~/.ssh/id_ed25519.pub

# 5. Add teammates' SSH public keys
envault add-key "ssh-ed25519 AAAAC3... teammate@example.com"

# 6. Encrypt secrets
envault encrypt dev .envault/dev.plaintext

# 7. Delete plaintext file (IMPORTANT!)
rm .envault/dev.plaintext

# 8. Verify everything works
envault check
# Should show: ✓ Authorized keys, ✓ Can decrypt, ✓ Targets listed

# 9. Test loading secrets
envault dev
# ✓ Loaded dev secrets to: .env (and other targets)

# 10. Verify secrets loaded correctly
cat .env | grep DATABASE_URL

# 11. Commit to git
git add .envault/
git commit -m "chore: add encrypted dev secrets with envault"
git push
```

## Developer Workflow

```bash
# Clone project
git clone git@github.com:yourorg/yourproject.git
cd yourproject

# Load secrets (one command!)
envault dev

# Secrets are now in .env (or wherever config.yaml specifies)
cat .env  # Verify secrets loaded

# Start developing
./bin/install
overmind start
```

## Common Admin Tasks

### Add a New Team Member

```bash
# Get their public key
# (they run: cat ~/.ssh/id_ed25519.pub)

envault add-key "ssh-ed25519 AAAAC3... newteammate@example.com"
envault reencrypt dev
git add .envault/ && git commit -m "chore: add teammate to envault"
git push
```

### Update Secrets

```bash
# Decrypt to plaintext
envault decrypt dev > .envault/dev.plaintext

# Edit the file
vim .envault/dev.plaintext

# Re-encrypt
envault encrypt dev .envault/dev.plaintext

# Clean up and commit
rm .envault/dev.plaintext
git add .envault/ && git commit -m "chore: update dev secrets"
git push
```

### Remove Team Member Access

```bash
# List keys to get fingerprint
envault list-keys

# Remove their key
envault remove-key <fingerprint>

# Re-encrypt all environments (revokes their access immediately)
envault reencrypt dev
envault reencrypt staging

git add .envault/ && git commit -m "chore: revoke teammate access"
git push
```

## Verify Everything Works

```bash
# Check configuration and access
envault check

# Should show:
# ✓ Authorized keys: 2
# ✓ Encrypted file exists: dev.age
# ✓ Can decrypt with your SSH key
# ✓ Targets: 1
#   - .env
```

## Integration with bin/install

Add this to your `./bin/install` script before other setup steps:

```bash
# Load development secrets
if command -v envault &> /dev/null; then
    print_info "Loading development secrets..."
    if envault dev; then
        print_success "Secrets loaded from envault"
    else
        print_error "Failed to load secrets"
        print_info "Contact admin to add your SSH public key: cat ~/.ssh/id_ed25519.pub"
        exit 1
    fi
else
    print_error "envault not installed"
    print_info "Install with: go install github.com/orchard9/envault/cmd/envault@latest"
    exit 1
fi
```

## Troubleshooting

### "age is not installed"

```bash
brew install age
```

### "No SSH private key found"

```bash
# Generate an SSH key if you don't have one
ssh-keygen -t ed25519 -C "your_email@example.com"
```

### "Cannot decrypt"

Your SSH public key isn't in authorized_keys yet. Ask admin to run:

```bash
envault add-key "your-ssh-public-key-here"
envault reencrypt dev
```

### "Failed to load config"

Make sure you're in a directory with `.envault/` initialized:

```bash
envault check
```

## Security Notes

- Encrypted `.age` files are safe to commit to git
- Only people with SSH keys in `authorized_keys` can decrypt
- Removing a key + reencrypting immediately revokes access
- Never commit `.plaintext` or `.plain` files (automatically gitignored)
- Keep plaintext files temporary - create, encrypt, delete

## Next Steps

- Read the full [README.md](README.md)
- See [examples/](examples/) for configuration examples
- Add envault to all your projects' onboarding docs
