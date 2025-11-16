#!/bin/bash
set -e

# envv Installation Script for Design Partners
# This script installs envv and sets up the environment

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  envv - Encrypted Secrets for Teams"
echo "  Installation for Design Partners"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo

# Detect OS
OS="$(uname -s)"
ARCH="$(uname -m)"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.envv}"
BIN_DIR="$INSTALL_DIR/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check prerequisites
check_prerequisites() {
    echo "🔍 Checking prerequisites..."

    if ! command -v git &> /dev/null; then
        echo -e "${RED}❌ Git is not installed${NC}"
        echo "Please install git first: https://git-scm.com"
        exit 1
    fi

    if ! command -v go &> /dev/null; then
        echo -e "${YELLOW}⚠️  Go is not installed${NC}"
        echo "Installing without Go - using pre-built binary..."
        USE_PREBUILT=true
    else
        GO_VERSION=$(go version | cut -d' ' -f3 | cut -d'.' -f2)
        if [ "$GO_VERSION" -lt "19" ]; then
            echo -e "${YELLOW}⚠️  Go version 1.19+ required${NC}"
            USE_PREBUILT=true
        fi
    fi

    echo -e "${GREEN}✓ Prerequisites checked${NC}"
}

# Install from source
install_from_source() {
    echo "📦 Building envv from source..."

    # Create temp directory
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"

    # Clone repository
    echo "  Cloning repository..."
    git clone --quiet https://github.com/AetherVoxSanctum/envv.git
    cd envv

    # Build
    echo "  Building binary..."
    make install > /dev/null 2>&1

    # Copy binary
    mkdir -p "$BIN_DIR"
    cp "$HOME/go/bin/envv" "$BIN_DIR/envv-core"

    # Cleanup
    cd /
    rm -rf "$TEMP_DIR"

    echo -e "${GREEN}✓ Built from source${NC}"
}

# Install pre-built binary (fallback)
install_prebuilt() {
    echo "📦 Downloading pre-built binary..."

    mkdir -p "$BIN_DIR"

    # For now, we'll build it since we don't have releases yet
    # In production, this would download from GitHub releases
    echo -e "${YELLOW}⚠️  Pre-built binaries not available yet${NC}"
    echo "Please install Go to build from source"
    exit 1
}

# Create wrapper script
create_wrapper() {
    echo "🔧 Creating envv command wrapper..."

    cat > "$BIN_DIR/envv" << 'WRAPPER_SCRIPT'
#!/bin/bash
# envv wrapper script - provides user-friendly commands

ENVV_CORE="${ENVV_CORE:-$HOME/.envv/bin/envv-core}"
ENV_FILE="${ENV_FILE:-.env.encrypted}"

# Helper functions
print_success() { echo -e "\033[0;32m✓\033[0m $1"; }
print_error() { echo -e "\033[0;31m✗\033[0m $1"; }
print_info() { echo -e "\033[1;34mℹ\033[0m $1"; }

# Ensure .sops.yaml exists
ensure_config() {
    if [ ! -f ".sops.yaml" ]; then
        echo "Creating encryption config..."
        cat > .sops.yaml << 'EOF'
creation_rules:
  - path_regex: \.env.*$
    age: age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p
EOF
        print_success "Created .sops.yaml"
    fi
}

case "$1" in
    "init")
        ensure_config
        if [ ! -f "$ENV_FILE" ]; then
            echo "# Encrypted environment variables" > .env
            $ENVV_CORE -e .env > "$ENV_FILE"
            rm .env
            print_success "Initialized envv project"
        else
            print_info "Project already initialized"
        fi
        ;;

    "set")
        ensure_config
        KEY="$2"
        VALUE="$3"
        if [ -z "$KEY" ] || [ -z "$VALUE" ]; then
            print_error "Usage: envv set KEY VALUE"
            exit 1
        fi

        # Decrypt existing or create new
        if [ -f "$ENV_FILE" ]; then
            $ENVV_CORE -d "$ENV_FILE" > /tmp/envv_temp_$$ 2>/dev/null
        else
            touch /tmp/envv_temp_$$
        fi

        # Update or add the key
        if grep -q "^${KEY}=" /tmp/envv_temp_$$; then
            sed -i.bak "/^${KEY}=/d" /tmp/envv_temp_$$
        fi
        echo "${KEY}=${VALUE}" >> /tmp/envv_temp_$$

        # Re-encrypt
        $ENVV_CORE -e /tmp/envv_temp_$$ > "$ENV_FILE"
        rm -f /tmp/envv_temp_$$ /tmp/envv_temp_$$.bak

        print_success "Set $KEY"
        ;;

    "get"|"reveal")
        KEY="$2"
        if [ -z "$KEY" ]; then
            print_error "Usage: envv get KEY"
            exit 1
        fi
        VALUE=$($ENVV_CORE -d "$ENV_FILE" 2>/dev/null | grep "^${KEY}=" | cut -d= -f2-)
        if [ -z "$VALUE" ]; then
            print_error "Key '$KEY' not found"
            exit 1
        fi
        echo "$VALUE"
        ;;

    "list")
        if [ ! -f "$ENV_FILE" ]; then
            print_error "No encrypted file found. Run 'envv init' first."
            exit 1
        fi
        echo "🔐 Encrypted secrets:"
        $ENVV_CORE -d "$ENV_FILE" 2>/dev/null | cut -d= -f1 | sed 's/^/  • /'
        ;;

    "delete"|"unset")
        KEY="$2"
        if [ -z "$KEY" ]; then
            print_error "Usage: envv delete KEY"
            exit 1
        fi

        $ENVV_CORE -d "$ENV_FILE" > /tmp/envv_temp_$$ 2>/dev/null
        grep -v "^${KEY}=" /tmp/envv_temp_$$ > /tmp/envv_temp2_$$
        $ENVV_CORE -e /tmp/envv_temp2_$$ > "$ENV_FILE"
        rm -f /tmp/envv_temp_$$ /tmp/envv_temp2_$$

        print_success "Deleted $KEY"
        ;;

    "exec")
        shift
        if [ ! -f "$ENV_FILE" ]; then
            print_error "No encrypted file found. Run 'envv init' first."
            exit 1
        fi
        $ENVV_CORE exec-env "$ENV_FILE" "$@"
        ;;

    "edit")
        $ENVV_CORE edit "${2:-$ENV_FILE}"
        ;;

    "encrypt")
        INPUT="${2:-.env}"
        OUTPUT="${3:-$ENV_FILE}"
        ensure_config
        $ENVV_CORE -e "$INPUT" > "$OUTPUT"
        print_success "Encrypted $INPUT → $OUTPUT"
        ;;

    "decrypt")
        INPUT="${2:-$ENV_FILE}"
        $ENVV_CORE -d "$INPUT"
        ;;

    "help"|"--help"|"-h")
        cat << 'HELP'
envv - Encrypted Secrets for Teams

Commands:
  init                Initialize envv in current directory
  set KEY VALUE       Set an encrypted environment variable
  get KEY            Get value of a specific key
  list               List all available keys
  delete KEY         Remove a key
  exec COMMAND       Run command with decrypted environment
  edit [FILE]        Edit encrypted file in $EDITOR
  encrypt [IN] [OUT] Encrypt a plaintext file
  decrypt [FILE]     Decrypt and display file contents

Examples:
  envv init
  envv set DATABASE_URL "postgres://localhost/mydb"
  envv exec npm start
  envv list

Environment:
  ENV_FILE    Encrypted file to use (default: .env.encrypted)
  ENVV_CORE   Path to core binary (default: ~/.envv/bin/envv-core)
HELP
        ;;

    "version"|"--version")
        echo "envv 0.1.0-alpha (Design Partner Edition)"
        $ENVV_CORE --version 2>/dev/null | head -1
        ;;

    *)
        echo "Usage: envv [command] [arguments]"
        echo "Run 'envv help' for more information"
        exit 1
        ;;
esac
WRAPPER_SCRIPT

    chmod +x "$BIN_DIR/envv"
    echo -e "${GREEN}✓ Wrapper script created${NC}"
}

# Create age key for demo
create_demo_key() {
    echo "🔑 Setting up demo encryption key..."

    mkdir -p "$INSTALL_DIR/keys"

    # Generate a new age key for this installation
    # In production each user would have their own key
    if command -v age-keygen &> /dev/null; then
        # Generate key and capture public key
        age-keygen -o "$INSTALL_DIR/keys/demo.agekey" 2>&1 | tee "$INSTALL_DIR/keys/demo.pub"
        echo -e "${GREEN}✓ Demo key generated${NC}"
    else
        echo -e "${YELLOW}⚠ age-keygen not found. Please install age to generate encryption keys.${NC}"
        echo "Visit: https://github.com/FiloSottile/age"
        return 1
    fi

    export SOPS_AGE_KEY_FILE="$INSTALL_DIR/keys/demo.agekey"

    echo -e "${GREEN}✓ Demo key configured${NC}"
}

# Setup shell integration
setup_shell() {
    echo "🐚 Setting up shell integration..."

    SHELL_RC=""
    if [ -n "$BASH_VERSION" ]; then
        SHELL_RC="$HOME/.bashrc"
    elif [ -n "$ZSH_VERSION" ]; then
        SHELL_RC="$HOME/.zshrc"
    fi

    if [ -n "$SHELL_RC" ]; then
        # Add to PATH
        echo "" >> "$SHELL_RC"
        echo "# envv - Encrypted Secrets for Teams" >> "$SHELL_RC"
        echo "export PATH=\"\$PATH:$BIN_DIR\"" >> "$SHELL_RC"
        echo "export SOPS_AGE_KEY_FILE=\"$INSTALL_DIR/keys/demo.agekey\"" >> "$SHELL_RC"

        echo -e "${GREEN}✓ Added to $SHELL_RC${NC}"
    fi
}

# Test installation
test_installation() {
    echo "🧪 Testing installation..."

    # Create test directory
    TEST_DIR="$INSTALL_DIR/test"
    mkdir -p "$TEST_DIR"
    cd "$TEST_DIR"

    # Test commands
    export PATH="$PATH:$BIN_DIR"
    export SOPS_AGE_KEY_FILE="$INSTALL_DIR/keys/demo.agekey"

    # Initialize
    "$BIN_DIR/envv" init > /dev/null 2>&1

    # Set a value
    "$BIN_DIR/envv" set TEST_KEY "test_value" > /dev/null 2>&1

    # Get the value
    VALUE=$("$BIN_DIR/envv" get TEST_KEY 2>/dev/null)

    if [ "$VALUE" = "test_value" ]; then
        echo -e "${GREEN}✓ All tests passed!${NC}"
        rm -rf "$TEST_DIR"
        return 0
    else
        echo -e "${RED}✗ Tests failed${NC}"
        return 1
    fi
}

# Main installation flow
main() {
    check_prerequisites

    echo
    echo "📍 Installing to: $INSTALL_DIR"
    echo

    # Install based on available tools
    if [ "${USE_PREBUILT:-false}" = "true" ]; then
        install_prebuilt
    else
        install_from_source
    fi

    create_wrapper
    create_demo_key
    setup_shell

    echo
    if test_installation; then
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo -e "${GREEN}  ✨ envv installed successfully!${NC}"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo
        echo "📚 Quick Start:"
        echo
        echo "  1. Reload your shell or run:"
        echo "     source ~/.bashrc"
        echo
        echo "  2. Try it out:"
        echo "     cd /path/to/your/project"
        echo "     envv init"
        echo "     envv set DATABASE_URL \"postgres://localhost\""
        echo "     envv exec npm start"
        echo
        echo "  3. Learn more:"
        echo "     envv help"
        echo
        echo "🚀 Happy encrypting!"
    else
        echo -e "${RED}Installation completed with errors${NC}"
        echo "Please check the output above for issues"
        exit 1
    fi
}

# Run installation
main "$@"