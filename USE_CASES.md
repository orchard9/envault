# envault Use Cases

Complete walkthrough of all envault workflows.

## Use Case 1: Initial Setup (Admin - First Time)

**Goal**: Admin has secret local files and wants to encrypt them for the team.

**Scenario**: Setting up envault for your project.

**Current State**:
- Project exists with .env.example
- No actual secrets in repo (gitignored)
- Admin has local .env with real API keys
- Need to share with 3-5 developers

**Workflow**:
```bash
cd ~/path/to/yourproject

# 1. Initialize envault
envault init
# Creates: .envault/config.yaml, .envault/authorized_keys

# 2. Configure targets (where secrets should be written)
vim .envault/config.yaml
```

Edit to match project structure:
```yaml
environments:
  dev:
    encrypted_file: dev.age
    targets:
      - path: .env
      - path: apps/api/.env
      - path: apps/worker/.env
      - path: apps/web/.env
```

Continue:
```bash
# 3. Create plaintext secrets file from existing .env
cp .env .envault/dev.plaintext
# Or create from scratch:
vim .envault/dev.plaintext

# 4. Add your SSH public key
envault add-key ~/.ssh/id_ed25519.pub
# ✓ Added SSH public key

# 5. Add teammates' SSH public keys
envault add-key "ssh-ed25519 AAAAC3Nza... teammate1@example.com"
envault add-key "ssh-ed25519 AAAAC3Nza... teammate2@example.com"
# ✓ Added SSH public key (x2)

# 6. Encrypt secrets
envault encrypt dev .envault/dev.plaintext
# ✓ Encrypted to .envault/dev.age

# 7. CRITICAL: Delete plaintext file
rm .envault/dev.plaintext
shred -u .envault/dev.plaintext  # Extra secure

# 8. Verify encryption worked
envault check
# ✓ Authorized keys: 3
# ✓ Encrypted file exists: dev.age
# ✓ Can decrypt with your SSH key
# ✓ Targets: 5

# 9. Test decryption
envault dev
# ✓ Loaded dev secrets to:
#   - .env
#   - apps/api/.env
#   - apps/worker/.env
#   - apps/web/.env

# 10. Commit encrypted secrets
git add .envault/
git commit -m "chore: add encrypted dev secrets with envault"
git push
```

**Result**:
- Secrets encrypted and in git
- 3 people can decrypt
- Team can clone and run `envault dev`

---

## Use Case 2: Developer Onboarding

**Goal**: New developer joins team and needs access to secrets.

**Scenario**: New hire needs to run project locally.

**Workflow**:

**Developer**:
```bash
# 1. Clone repo
git clone git@github.com:yourorg/yourproject.git
cd yourproject

# 2. Install envault
go install github.com/orchard9/envault/cmd/envault@latest

# 3. Try to load secrets
envault dev
# Error: age decryption failed: no identity matched any recipient

# 4. Get your SSH public key
cat ~/.ssh/id_ed25519.pub
# ssh-ed25519 AAAAC3Nza... newdev@example.com

# 5. Send to admin (Slack, email, etc.)
```

**Admin**:
```bash
cd yourproject

# 6. Add new developer's key
envault add-key "ssh-ed25519 AAAAC3Nza... newdev@example.com"
# ✓ Added SSH public key
# Next steps: re-encrypt environments

# 7. Re-encrypt dev environment with new key added
envault reencrypt dev
# ✓ Re-encrypted dev with current authorized_keys

# 8. Commit and push
git add .envault/
git commit -m "chore: add newdev to envault"
git push
```

**Developer**:
```bash
# 9. Pull updated encrypted file
git pull

# 10. Load secrets (now works!)
envault dev
# ✓ Loaded dev secrets to:
#   - .env
#   - apps/api/.env
#   ...

# 11. Verify secrets loaded
cat .env | grep API_KEY

# 12. Start developing
./bin/install
overmind start
```

**Result**: New developer has access and can decrypt secrets.

---

## Use Case 3: Updating Secrets (Single Environment)

**Goal**: One secret changed (e.g., external API key rotated).

**Scenario**: API key needs to be updated in dev environment.

**Workflow**:
```bash
cd yourproject

# 1. Decrypt current secrets to edit
envault decrypt dev > .envault/dev.plaintext

# 2. Edit the plaintext file
vim .envault/dev.plaintext
# Change: API_KEY=old_key
# To: API_KEY=new_key_here

# 3. Re-encrypt with updated content
envault encrypt dev .envault/dev.plaintext
# ✓ Encrypted to .envault/dev.age

# 4. CRITICAL: Delete plaintext file
rm .envault/dev.plaintext

# 5. Test decryption works
envault dev
# ✓ Loaded dev secrets to: ...

# 6. Verify new value
grep API_KEY .env
# API_KEY=new_key_here

# 7. Commit and push
git add .envault/dev.age
git commit -m "chore: rotate API key"
git push
```

**Teammates**:
```bash
# 8. Pull updated secrets
git pull

# 9. Reload secrets
envault dev
# ✓ Loaded dev secrets to: ...

# 10. Restart services to pick up new key
overmind restart
```

**Result**: All team members have the new API key.

---

## Use Case 4: Updating Multiple Secrets

**Goal**: Add new secrets needed by new feature.

**Scenario**: Adding Stripe integration, need STRIPE_SECRET_KEY and STRIPE_PUBLISHABLE_KEY.

**Workflow**:
```bash
cd yourproject

# 1. Decrypt current secrets
envault decrypt dev > .envault/dev.plaintext

# 2. Edit to add new secrets
vim .envault/dev.plaintext
# Add:
# STRIPE_SECRET_KEY=sk_test_...
# STRIPE_PUBLISHABLE_KEY=pk_test_...

# 3. Re-encrypt
envault encrypt dev .envault/dev.plaintext

# 4. Delete plaintext
rm .envault/dev.plaintext

# 5. Test and commit
envault dev
git add .envault/dev.age
git commit -m "chore: add Stripe API keys"
git push
```

**Result**: New secrets available to all team members.

---

## Use Case 5: Multiple Environments

**Goal**: Separate secrets for dev, staging, and production.

**Scenario**: Need different API keys for each environment.

**Workflow**:

**Setup**:
```bash
cd yourproject

# 1. Update config for multiple environments
vim .envault/config.yaml
```

```yaml
environments:
  dev:
    encrypted_file: dev.age
    targets:
      - path: .env

  staging:
    encrypted_file: staging.age
    targets:
      - path: .env.staging

  prod:
    encrypted_file: prod.age
    targets:
      - path: .env.production
```

Continue:
```bash
# 2. Create staging secrets
cat > .envault/staging.plaintext << 'EOF'
DATABASE_URL=postgresql://user@staging-db/myapp
API_KEY=staging_api_key
# ... other staging secrets
EOF

# 3. Encrypt staging
envault encrypt staging .envault/staging.plaintext
rm .envault/staging.plaintext

# 4. Create production secrets
cat > .envault/prod.plaintext << 'EOF'
DATABASE_URL=postgresql://user@prod-db/myapp
API_KEY=prod_api_key
# ... other prod secrets
EOF

# 5. Encrypt production
envault encrypt prod .envault/prod.plaintext
rm .envault/prod.plaintext

# 6. Commit all environments
git add .envault/
git commit -m "chore: add staging and prod secrets"
git push
```

**Usage**:
```bash
# Load dev secrets
envault dev

# Load staging secrets
envault staging

# Load production secrets
envault prod
```

**Result**: Separate encrypted secrets for each environment.

---

## Use Case 6: Removing Team Member Access

**Goal**: Developer leaves team, revoke their access.

**Scenario**: Developer leaves company, must revoke secret access immediately.

**Workflow**:
```bash
cd yourproject

# 1. List current authorized keys
envault list-keys
# Authorized keys (4):
#   1. a1b2c3d4 (ssh-ed25519) - admin@example.com
#   2. e5f6g7h8 (ssh-ed25519) - dev1@example.com
#   3. i9j0k1l2 (ssh-ed25519) - dev2@example.com
#   4. m3n4o5p6 (ssh-ed25519) - leaving-dev@example.com

# 2. Remove leaving developer's key
envault remove-key m3n4o5p6
# ✓ Removed SSH public key
# IMPORTANT: Re-encrypt all environments to revoke access

# 3. Re-encrypt ALL environments (critical!)
envault reencrypt dev
# ✓ Re-encrypted dev with current authorized_keys

envault reencrypt staging
# ✓ Re-encrypted staging with current authorized_keys

envault reencrypt prod
# ✓ Re-encrypted prod with current authorized_keys

# 4. Commit and push
git add .envault/
git commit -m "chore: revoke access for leaving-dev"
git push
```

**What happens**:
- Leaving developer can still decrypt OLD commits (before revocation)
- Leaving developer CANNOT decrypt NEW commits (after re-encryption)
- This is immediate - no waiting period

**Result**: Access revoked, only 3 people can decrypt going forward.

---

## Use Case 7: Customizing Configuration

**Goal**: Customize where secrets are written per service.

**Scenario**: Different services need secrets in different locations.

**Workflow**:
```bash
cd yourproject

# 1. Edit config
vim .envault/config.yaml
```

Example configurations:

**Simple (one target)**:
```yaml
environments:
  dev:
    encrypted_file: dev.age
    targets:
      - path: .env
```

**Multiple targets (same content)**:
```yaml
environments:
  dev:
    encrypted_file: dev.age
    targets:
      - path: .env                    # Root
      - path: apps/api/.env           # API service
      - path: apps/worker/.env        # Worker service
```

**Multiple environments**:
```yaml
environments:
  dev:
    encrypted_file: dev.age
    targets:
      - path: .env

  staging:
    encrypted_file: staging.age
    targets:
      - path: .env.staging

  prod:
    encrypted_file: prod.age
    targets:
      - path: .env.production
```

**Docker-compose setup**:
```yaml
environments:
  dev:
    encrypted_file: dev.age
    targets:
      - path: .env
      - path: docker/.env
```

**Result**: Flexible configuration for any project structure.

---

## Use Case 8: Verifying Configuration

**Goal**: Check that envault is set up correctly.

**Workflow**:
```bash
cd yourproject

# Run diagnostics
envault check
```

**Successful output**:
```
Checking envault configuration...

✓ Authorized keys: 3

Environment: dev
  ✓ Encrypted file exists: dev.age
  ✓ Can decrypt with your SSH key
  ✓ Targets: 5
    - .env
    - apps/yourproject-api/.env
    - apps/billing-api/.env
    - apps/yourproject-web/.env
    - apps/yourproject-admin/.env
```

**Error output**:
```
Checking envault configuration...

✓ Authorized keys: 3

Environment: dev
  ✗ Encrypted file missing: dev.age
```

**Result**: Know immediately if configuration is correct.

---

## Key Principles

1. **One plaintext file per environment**: dev.age contains ALL dev secrets
2. **Multiple targets**: Same secrets written to multiple locations for convenience
3. **Multi-recipient**: Multiple SSH keys can decrypt the same file
4. **Immediate revocation**: Remove key + reencrypt = instant access revocation
5. **Git-safe**: Encrypted .age files are safe to commit
6. **Temporary plaintext**: Never commit .plaintext files, delete immediately after encryption

## Security Model

**What's protected**:
- API keys, passwords, tokens
- Database connection strings
- Service credentials

**How it works**:
- Encrypted with age (modern, audited encryption)
- Only authorized SSH keys can decrypt
- Removing key + re-encrypting revokes access

**What's NOT protected**:
- Historical commits (old secrets before re-encryption)
- Metadata (which environments exist, how many targets)
- Config structure (.envault/config.yaml is plaintext)

**Best practices**:
- Rotate secrets when removing team members
- Use separate environments (dev/staging/prod)
- Never commit .plaintext files
- Keep authorized_keys minimal (only current team)
