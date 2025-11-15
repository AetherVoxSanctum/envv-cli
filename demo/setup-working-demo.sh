#!/bin/bash
set -e

# Setup script that ACTUALLY WORKS for the demo
# This creates real age keys and configures everything properly

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  envv Demo Setup (Working Version)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[1;34m'
NC='\033[0m'

# Check for age
check_age() {
    if ! command -v age &> /dev/null; then
        echo -e "${YELLOW}Installing age...${NC}"
        if [[ "$OSTYPE" == "darwin"* ]]; then
            brew install age
        else
            echo "Please install age: https://github.com/FiloSottile/age"
            exit 1
        fi
    fi
}

# Generate demo keys
generate_keys() {
    echo -e "${BLUE}Generating demo encryption keys...${NC}"

    # Create keys directory
    mkdir -p keys

    # Generate age key
    age-keygen -o keys/demo.agekey 2> keys/demo.public

    # Extract public key
    PUB_KEY=$(grep "Public key:" keys/demo.public | cut -d' ' -f3)

    echo -e "${GREEN}✓ Generated age keypair${NC}"
    echo "  Public key: $PUB_KEY"

    # Create .sops.yaml with the real public key
    cat > .sops.yaml << EOF
creation_rules:
  - path_regex: \.env.*$
    age: $PUB_KEY
EOF

    echo -e "${GREEN}✓ Created .sops.yaml${NC}"
}

# Create demo secrets
create_demo_secrets() {
    echo -e "${BLUE}Creating demo secrets...${NC}"

    cat > .env << 'EOF'
ANALYTICS_KEY_GOOGLE=GA-987654321
ANALYTICS_KEY_MIXPANEL=mix_prod_xyz789def456
STRIPE_API_KEY=sk_live_EXAMPLE_demo_key_not_real
BACKEND_SECRET_KEY=ultra-secret-backend-key-2024-prod
PORT=3000
EOF

    echo -e "${GREEN}✓ Created .env with demo secrets${NC}"
}

# Encrypt the secrets
encrypt_secrets() {
    echo -e "${BLUE}Encrypting secrets...${NC}"

    # Set the age key for SOPS
    export SOPS_AGE_KEY_FILE="$(pwd)/keys/demo.agekey"

    # Encrypt using envv/sops
    ~/go/bin/envv -e .env > .env.encrypted

    # Remove plaintext
    rm .env

    echo -e "${GREEN}✓ Encrypted secrets to .env.encrypted${NC}"
}

# Create working wrapper script
create_wrapper() {
    echo -e "${BLUE}Creating envv command wrapper...${NC}"

    cat > envv << 'WRAPPER'
#!/bin/bash
# envv wrapper for demo - makes commands user-friendly

ENVV_BIN="${ENVV_BIN:-$HOME/go/bin/envv}"
ENV_FILE="${ENV_FILE:-.env.encrypted}"
export SOPS_AGE_KEY_FILE="${SOPS_AGE_KEY_FILE:-$(pwd)/keys/demo.agekey}"

case "$1" in
    "init")
        if [ ! -f "$ENV_FILE" ]; then
            echo "# New envv project" > .env.tmp
            $ENVV_BIN -e .env.tmp > "$ENV_FILE"
            rm .env.tmp
            echo "✓ Initialized envv project"
        else
            echo "Project already initialized"
        fi
        ;;

    "set")
        KEY="$2"
        VALUE="$3"
        # Decrypt, update, re-encrypt
        if [ -f "$ENV_FILE" ]; then
            $ENVV_BIN -d "$ENV_FILE" > /tmp/env.$$ 2>/dev/null
        else
            touch /tmp/env.$$
        fi
        # Remove old value if exists
        grep -v "^${KEY}=" /tmp/env.$$ > /tmp/env2.$$ 2>/dev/null || true
        # Add new value
        echo "${KEY}=${VALUE}" >> /tmp/env2.$$
        # Re-encrypt
        $ENVV_BIN -e /tmp/env2.$$ > "$ENV_FILE"
        rm -f /tmp/env.$$ /tmp/env2.$$
        echo "✓ Set $KEY"
        ;;

    "get"|"reveal")
        KEY="$2"
        $ENVV_BIN -d "$ENV_FILE" 2>/dev/null | grep "^${KEY}=" | cut -d= -f2-
        ;;

    "list")
        echo "🔐 Encrypted secrets:"
        $ENVV_BIN -d "$ENV_FILE" 2>/dev/null | cut -d= -f1 | sed 's/^/  • /'
        ;;

    "exec")
        shift
        $ENVV_BIN exec-env "$ENV_FILE" "$@"
        ;;

    "decrypt")
        $ENVV_BIN -d "$ENV_FILE"
        ;;

    *)
        echo "Usage: ./envv [init|set|get|list|exec|decrypt]"
        ;;
esac
WRAPPER

    chmod +x envv

    echo -e "${GREEN}✓ Created ./envv wrapper${NC}"
}

# Test the setup
test_setup() {
    echo -e "${BLUE}Testing setup...${NC}"

    # Set key file
    export SOPS_AGE_KEY_FILE="$(pwd)/keys/demo.agekey"

    # Test decryption
    if ~/go/bin/envv -d .env.encrypted 2>/dev/null | grep -q "STRIPE_API_KEY"; then
        echo -e "${GREEN}✓ Decryption works${NC}"
    else
        echo "✗ Decryption failed"
        return 1
    fi

    # Test wrapper
    if ./envv list 2>/dev/null | grep -q "STRIPE_API_KEY"; then
        echo -e "${GREEN}✓ Wrapper works${NC}"
    else
        echo "✗ Wrapper failed"
        return 1
    fi

    # Test execution
    echo 'echo "STRIPE=$STRIPE_API_KEY"' > test.sh
    chmod +x test.sh
    if ./envv exec ./test.sh 2>/dev/null | grep -q "sk_live"; then
        echo -e "${GREEN}✓ Execution works${NC}"
        rm test.sh
    else
        echo "✗ Execution failed"
        rm test.sh
        return 1
    fi

    return 0
}

# Create demo instructions
create_instructions() {
    cat > DEMO_INSTRUCTIONS.md << 'EOF'
# Demo Instructions

## Quick Demo Script

```bash
# Show the problem - plaintext secrets
echo "Traditional approach - secrets are exposed:"
cat .env.example

# Show the solution - encrypted secrets
echo "With envv - secrets are encrypted:"
cat .env.encrypted

# Show that secrets are actually encrypted
echo "Trying to grep for secrets:"
grep -r "sk_live" . 2>/dev/null || echo "✓ No secrets found in plaintext!"

# Run the app with encrypted secrets
echo "Running app with encrypted secrets:"
./envv exec npm start

# List available secrets (without showing values)
./envv list

# Set a new secret
./envv set NEW_API_KEY "abc123xyz789"

# Use the new secret
./envv exec printenv | grep NEW_API_KEY
```

## Key Files

- `keys/demo.agekey` - Private key for decryption (keep secret!)
- `keys/demo.public` - Public key for encryption (safe to share)
- `.env.encrypted` - Your encrypted secrets
- `.sops.yaml` - Configuration for encryption
- `envv` - Wrapper script for easy commands

## Team Simulation

To simulate multiple team members:

1. Alice (Team Lead) - has the private key
2. Bob (Developer) - gets the private key securely
3. Charlie (New Joiner) - needs access granted

Each would have their own age key in production.
EOF

    echo -e "${GREEN}✓ Created DEMO_INSTRUCTIONS.md${NC}"
}

# Main setup
main() {
    echo "Setting up in: $(pwd)"
    echo

    # Check prerequisites
    check_age

    # Check for envv binary
    if [ ! -f "$HOME/go/bin/envv" ]; then
        echo "envv not found. Building it now..."
        cd ..
        make install
        cd demo
    fi

    # Run setup steps
    generate_keys
    create_demo_secrets
    encrypt_secrets
    create_wrapper
    create_instructions

    echo
    if test_setup; then
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo -e "${GREEN}  ✨ Demo setup complete!${NC}"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo
        echo "  Test it now:"
        echo "    ./envv list"
        echo "    ./envv exec npm start"
        echo "    ./envv set MY_KEY 'my_value'"
        echo
        echo "  See DEMO_INSTRUCTIONS.md for full demo script"
        echo
    else
        echo "Setup completed with errors. Please check above."
        exit 1
    fi
}

# Run if executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    main "$@"
fi