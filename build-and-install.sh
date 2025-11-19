#!/bin/bash
# Simple build and install script for envv CLI with SaaS integration

set -e

BINARY_NAME="envv"
INSTALL_DIR="$HOME/.local/bin"
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "  envv CLI Build & Install"
echo "=========================================="
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}✗ Go is not installed${NC}"
    echo "  Install Go from: https://golang.org/dl/"
    exit 1
fi
echo -e "${GREEN}✓ Go is installed: $(go version)${NC}"

# Check dependencies
echo ""
echo "Checking dependencies..."

if command -v sops &> /dev/null; then
    echo -e "${GREEN}✓ sops is installed${NC}"
else
    echo -e "${YELLOW}⚠ sops is NOT installed${NC}"
    echo "  Install from: https://github.com/getsops/sops#install"
    echo "  macOS: brew install sops"
    echo "  Linux: Download from releases"
    MISSING_DEPS=1
fi

if command -v age-keygen &> /dev/null; then
    echo -e "${GREEN}✓ age is installed${NC}"
else
    echo -e "${YELLOW}⚠ age is NOT installed${NC}"
    echo "  Install from: https://github.com/FiloSottile/age#installation"
    echo "  macOS: brew install age"
    echo "  Linux: Download from releases"
    MISSING_DEPS=1
fi

if [ "$MISSING_DEPS" = "1" ]; then
    echo ""
    echo -e "${YELLOW}Warning: Missing dependencies. envv SaaS features require sops and age.${NC}"
    echo "The CLI will build, but you'll need to install these before using SaaS features."
    echo ""
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Build the binary
echo ""
echo "Building envv CLI..."
if go build -o "$BINARY_NAME" ./cmd/envv; then
    echo -e "${GREEN}✓ Build successful${NC}"
else
    echo -e "${RED}✗ Build failed${NC}"
    exit 1
fi

# Create install directory
mkdir -p "$INSTALL_DIR"

# Install binary
echo ""
echo "Installing to $INSTALL_DIR..."
cp "$BINARY_NAME" "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/$BINARY_NAME"
echo -e "${GREEN}✓ Installed: $INSTALL_DIR/$BINARY_NAME${NC}"

# Check if install dir is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo ""
    echo -e "${YELLOW}⚠ $INSTALL_DIR is not in your PATH${NC}"
    echo ""
    echo "Add this to your ~/.bashrc or ~/.zshrc:"
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
    echo ""
    echo "Then reload your shell:"
    echo "  source ~/.bashrc  # or source ~/.zshrc"
fi

# Verify installation
echo ""
if "$INSTALL_DIR/$BINARY_NAME" --version &> /dev/null; then
    echo -e "${GREEN}✓ Installation verified${NC}"
else
    echo -e "${YELLOW}⚠ Could not verify installation${NC}"
fi

echo ""
echo "=========================================="
echo -e "${GREEN}  envv CLI installed successfully!${NC}"
echo "=========================================="
echo ""
echo "Next steps:"
echo "  1. Make sure $INSTALL_DIR is in your PATH"
echo "  2. Run: envv auth register"
echo "  3. See QUICKSTART.md for detailed usage"
echo ""
echo "Quick start:"
echo "  envv auth register     # Register and generate keys"
echo "  envv org create        # Create an organization"
echo "  envv project create    # Create a project"
echo "  envv project init      # Initialize current directory"
echo "  envv secrets push      # Push secrets to backend"
echo ""
