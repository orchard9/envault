#!/usr/bin/env bash

# Comprehensive envault test script
# Tests all major use cases end-to-end

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Test counter
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Helper functions
test_start() {
    echo -e "\n${BLUE}TEST:${NC} $1"
    TESTS_RUN=$((TESTS_RUN + 1))
}

test_pass() {
    echo -e "${GREEN}✓ PASS${NC}: $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

test_fail() {
    echo -e "${RED}✗ FAIL${NC}: $1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
}

assert_file_exists() {
    if [ -f "$1" ]; then
        test_pass "File exists: $1"
    else
        test_fail "File missing: $1"
    fi
}

assert_file_contains() {
    if grep -q "$2" "$1"; then
        test_pass "File contains '$2': $1"
    else
        test_fail "File missing '$2': $1"
    fi
}

assert_command_success() {
    if eval "$1" > /dev/null 2>&1; then
        test_pass "Command succeeded: $1"
    else
        test_fail "Command failed: $1"
    fi
}

# Setup
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}envault Comprehensive Test Suite${NC}"
echo -e "${BLUE}========================================${NC}"

# Check prerequisites
echo -e "\n${BLUE}Checking prerequisites...${NC}"
if ! command -v age &> /dev/null; then
    echo -e "${RED}ERROR: age not installed${NC}"
    echo "Install with: brew install age"
    exit 1
fi
echo -e "${GREEN}✓${NC} age installed"

if ! command -v envault &> /dev/null; then
    echo -e "${RED}ERROR: envault not installed${NC}"
    echo "Install with: cd ~/orchard9/envault && make install"
    exit 1
fi
echo -e "${GREEN}✓${NC} envault installed"

# Create test directory
TEST_DIR=$(mktemp -d)/envault-test
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"
echo -e "${GREEN}✓${NC} Created test directory: $TEST_DIR"

# Check for user's SSH keys
echo -e "\n${BLUE}Checking for SSH keys...${NC}"
if [ ! -f ~/.ssh/id_ed25519.pub ] && [ ! -f ~/.ssh/id_rsa.pub ]; then
    echo -e "${RED}ERROR: No SSH keys found${NC}"
    echo "Generate one with: ssh-keygen -t ed25519 -C 'your_email@example.com'"
    exit 1
fi

# Use real SSH keys for testing
if [ -f ~/.ssh/id_ed25519.pub ]; then
    USER_KEY_PUB=~/.ssh/id_ed25519.pub
    echo -e "${GREEN}✓${NC} Found ed25519 SSH key"
elif [ -f ~/.ssh/id_rsa.pub ]; then
    USER_KEY_PUB=~/.ssh/id_rsa.pub
    echo -e "${GREEN}✓${NC} Found RSA SSH key"
fi

# Test 1: envault init
test_start "envault init creates correct structure"
envault init
assert_file_exists ".envault/config.yaml"
assert_file_exists ".envault/authorized_keys"
assert_file_exists ".envault/.gitignore"
assert_file_contains ".envault/.gitignore" "*.plaintext"

# Test 2: Add SSH key
test_start "envault add-key with user's SSH key"
envault add-key "$USER_KEY_PUB"
# Get a unique part of the key to verify
KEY_PART=$(cat "$USER_KEY_PUB" | awk '{print $1}')
assert_file_contains ".envault/authorized_keys" "$KEY_PART"

# Test 3: List keys
test_start "envault list-keys shows key"
OUTPUT=$(envault list-keys)
if echo "$OUTPUT" | grep -q "$KEY_PART"; then
    test_pass "list-keys shows user key"
else
    test_fail "list-keys missing user key"
fi

# Test 4: Create and encrypt secrets
test_start "envault encrypt creates encrypted file"
cat > .envault/dev.plaintext << 'EOF'
DATABASE_URL=postgresql://testuser:testpass@localhost:5432/testdb
REDIS_URL=redis://localhost:6379
API_KEY=test_api_key_12345
SECRET=super_secret_value
EOF

envault encrypt dev .envault/dev.plaintext
assert_file_exists ".envault/dev.age"

# Clean up plaintext
rm .envault/dev.plaintext

# Test 5: Decrypt to stdout
test_start "envault decrypt outputs to stdout"
OUTPUT=$(envault decrypt dev)
if echo "$OUTPUT" | grep -q "DATABASE_URL=postgresql://testuser:testpass@localhost:5432/testdb"; then
    test_pass "decrypt outputs correct content"
else
    test_fail "decrypt output incorrect"
fi

# Test 6: Load environment
test_start "envault dev writes to target files"
envault dev
assert_file_exists ".env"
assert_file_contains ".env" "DATABASE_URL=postgresql://testuser:testpass@localhost:5432/testdb"
assert_file_contains ".env" "API_KEY=test_api_key_12345"

# Test 7: Multi-target configuration
test_start "envault with multiple targets"
cat > .envault/config.yaml << 'EOF'
environments:
  dev:
    encrypted_file: dev.age
    targets:
      - path: .env
      - path: apps/api/.env
      - path: apps/web/.env
EOF

mkdir -p apps/api apps/web
envault dev
assert_file_exists ".env"
assert_file_exists "apps/api/.env"
assert_file_exists "apps/web/.env"
assert_file_contains "apps/api/.env" "DATABASE_URL"
assert_file_contains "apps/web/.env" "DATABASE_URL"

# Test 8: Update secrets
test_start "Update and re-encrypt secrets"
envault decrypt dev > .envault/dev.plaintext
echo "NEW_SECRET=new_value_here" >> .envault/dev.plaintext
envault encrypt dev .envault/dev.plaintext
rm .envault/dev.plaintext
envault dev
assert_file_contains ".env" "NEW_SECRET=new_value_here"

# Test 9: Reencrypt
test_start "envault reencrypt works"
envault reencrypt dev
# Verify we can still decrypt
assert_command_success "envault decrypt dev > /dev/null"

# Test 10: Check command
test_start "envault check validates configuration"
OUTPUT=$(envault check)
if echo "$OUTPUT" | grep -q "✓ Authorized keys: 1"; then
    test_pass "check shows correct key count"
else
    test_fail "check shows wrong key count"
fi

if echo "$OUTPUT" | grep -q "✓ Can decrypt with your SSH key"; then
    test_pass "check confirms decryption works"
else
    test_fail "check doesn't confirm decryption"
fi

# Test 12: Multiple environments
test_start "Multiple environments (dev, staging, prod)"
cat > .envault/config.yaml << 'EOF'
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
EOF

# Create staging secrets
cat > .envault/staging.plaintext << 'EOF'
DATABASE_URL=postgresql://staging@db/staging
API_KEY=staging_api_key
EOF
envault encrypt staging .envault/staging.plaintext
rm .envault/staging.plaintext

# Create prod secrets
cat > .envault/prod.plaintext << 'EOF'
DATABASE_URL=postgresql://prod@db/prod
API_KEY=prod_api_key
EOF
envault encrypt prod .envault/prod.plaintext
rm .envault/prod.plaintext

# Test loading each environment
envault dev
assert_file_contains ".env" "testdb"

envault staging
assert_file_contains ".env.staging" "staging"

envault prod
assert_file_contains ".env.production" "prod"

# Test 13: Verify gitignore
test_start ".gitignore excludes plaintext files"
assert_file_contains ".envault/.gitignore" "*.plaintext"
assert_file_contains ".envault/.gitignore" "*.plain"
assert_file_contains ".envault/.gitignore" "*.decrypted"

# Summary
echo -e "\n${BLUE}========================================${NC}"
echo -e "${BLUE}Test Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "Tests run:    ${TESTS_RUN}"
echo -e "Tests passed: ${GREEN}${TESTS_PASSED}${NC}"
echo -e "Tests failed: ${RED}${TESTS_FAILED}${NC}"

if [ "$TESTS_FAILED" -eq 0 ]; then
    echo -e "\n${GREEN}✓ All tests passed!${NC}"
    # Clean up
    echo -e "\nCleaning up test directory: $TEST_DIR"
    cd /
    rm -rf "$TEST_DIR"
    exit 0
else
    echo -e "\n${RED}✗ Some tests failed${NC}"
    echo -e "Test directory preserved for inspection: $TEST_DIR"
    exit 1
fi
