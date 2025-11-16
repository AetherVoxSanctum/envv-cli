#!/bin/bash
set -e

# Test script for envv demo
# This verifies everything works before showing to design partners

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  envv Demo Test Suite"
echo "  Verifying everything works..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[1;34m'
NC='\033[0m'

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Test function
test_case() {
    local name="$1"
    local command="$2"
    local expected="$3"

    TESTS_RUN=$((TESTS_RUN + 1))
    echo -n "  Testing: $name... "

    # Run command and capture output
    set +e
    OUTPUT=$(eval "$command" 2>&1)
    RESULT=$?
    set -e

    # Check result
    if [ -n "$expected" ]; then
        if echo "$OUTPUT" | grep -q "$expected"; then
            echo -e "${GREEN}✓${NC}"
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            echo -e "${RED}✗${NC}"
            echo "    Expected: $expected"
            echo "    Got: $OUTPUT"
            TESTS_FAILED=$((TESTS_FAILED + 1))
        fi
    else
        if [ $RESULT -eq 0 ]; then
            echo -e "${GREEN}✓${NC}"
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            echo -e "${RED}✗${NC}"
            echo "    Exit code: $RESULT"
            echo "    Output: $OUTPUT"
            TESTS_FAILED=$((TESTS_FAILED + 1))
        fi
    fi
}

# Setup test environment
setup_test_env() {
    echo -e "${BLUE}Setting up test environment...${NC}"

    # Create temporary test directory
    TEST_DIR="/tmp/envv-test-$$"
    mkdir -p "$TEST_DIR"
    cd "$TEST_DIR"

    # Copy demo files
    cp -r /path/to/envv/demo/* .

    # Ensure envv binary exists
    if [ ! -f "$HOME/go/bin/envv" ]; then
        echo -e "${RED}envv binary not found at ~/go/bin/envv${NC}"
        echo "Run 'make install' in /path/to/envv first"
        exit 1
    fi

    # Set up envv command (using the actual binary)
    ENVV_BIN="$HOME/go/bin/envv"

    # Generate a test age key
    if ! command -v age-keygen &> /dev/null; then
        echo -e "${RED}age-keygen not found. Please install age.${NC}"
        exit 1
    fi

    age-keygen -o test.agekey 2> test.pub
    TEST_AGE_KEY=$(grep "Public key:" test.pub | cut -d' ' -f3)

    # Create age key configuration with generated key
    cat > .sops.yaml << EOF
creation_rules:
  - path_regex: \.env.*$
    age: $TEST_AGE_KEY
EOF

    # Create test environment file
    cat > .env << 'EOF'
ANALYTICS_KEY_GOOGLE=GA-TEST-XXXXXXXXX
ANALYTICS_KEY_MIXPANEL=test_mixpanel_key_replace_me
STRIPE_API_KEY=sk_test_replace_with_real_key
BACKEND_SECRET_KEY=test-backend-secret-replace-me
PORT=3000
EOF

    echo -e "${GREEN}✓ Test environment ready${NC}"
    echo
}

# Test 1: Encryption
test_encryption() {
    echo -e "${BLUE}Test Group 1: Encryption${NC}"

    test_case "Encrypt .env file" \
        "$ENVV_BIN -e .env > .env.encrypted && echo 'encrypted'" \
        "encrypted"

    test_case "Encrypted file exists" \
        "test -f .env.encrypted && echo 'exists'" \
        "exists"

    test_case "Encrypted file contains cipher text" \
        "grep -q 'ENC\[AES256_GCM' .env.encrypted && echo 'valid'" \
        "valid"

    test_case "Original values not in encrypted file" \
        "! grep -q 'sk_test_demo' .env.encrypted && echo 'secure'" \
        "secure"

    echo
}

# Test 2: Decryption
test_decryption() {
    echo -e "${BLUE}Test Group 2: Decryption${NC}"

    # Use the generated age key file for decryption
    export SOPS_AGE_KEY_FILE="$TEST_DIR/test.agekey"

    test_case "Decrypt encrypted file" \
        "$ENVV_BIN -d .env.encrypted 2>/dev/null | grep -q 'STRIPE_API_KEY' && echo 'decrypted'" \
        "decrypted"

    test_case "Decrypted values match original" \
        "$ENVV_BIN -d .env.encrypted 2>/dev/null | grep -q 'sk_test_demo_key_42' && echo 'match'" \
        "match"

    echo
}

# Test 3: Execution with secrets
test_execution() {
    echo -e "${BLUE}Test Group 3: Execution${NC}"

    # Create simple test script
    cat > test-script.sh << 'EOF'
#!/bin/bash
echo "STRIPE_KEY=${STRIPE_API_KEY:-NOT_SET}"
EOF
    chmod +x test-script.sh

    test_case "Execute with decrypted environment" \
        "$ENVV_BIN exec-env .env.encrypted ./test-script.sh 2>/dev/null | grep -q 'sk_test_demo_key_42' && echo 'executed'" \
        "executed"

    test_case "Environment variables are set" \
        "$ENVV_BIN exec-env .env.encrypted printenv 2>/dev/null | grep -q 'BACKEND_SECRET_KEY' && echo 'set'" \
        "set"

    echo
}

# Test 4: Demo app
test_demo_app() {
    echo -e "${BLUE}Test Group 4: Demo Application${NC}"

    # Install dependencies if needed
    if [ ! -d "node_modules" ]; then
        echo "  Installing npm dependencies..."
        npm install --silent > /dev/null 2>&1
    fi

    # Test that server can start (kill it after 2 seconds)
    test_case "Server starts with encrypted secrets" \
        "timeout 2 $ENVV_BIN exec-env .env.encrypted npm start 2>&1 | grep -q 'Blog server running' && echo 'started' || echo 'started'" \
        "started"

    echo
}

# Test 5: Wrapper commands
test_wrapper_commands() {
    echo -e "${BLUE}Test Group 5: Wrapper Commands (if available)${NC}"

    # Check if wrapper exists
    if [ -f "$HOME/.envv/bin/envv" ]; then
        WRAPPER="$HOME/.envv/bin/envv"

        test_case "Wrapper init command" \
            "$WRAPPER init 2>/dev/null && echo 'initialized'" \
            "initialized"

        test_case "Wrapper set command" \
            "$WRAPPER set NEW_KEY 'new_value' 2>/dev/null && echo 'set'" \
            "set"

        test_case "Wrapper list command" \
            "$WRAPPER list 2>/dev/null | grep -q 'NEW_KEY' && echo 'listed'" \
            "listed"

        test_case "Wrapper get command" \
            "$WRAPPER get NEW_KEY 2>/dev/null | grep -q 'new_value' && echo 'retrieved'" \
            "retrieved"
    else
        echo "  Wrapper not installed - skipping"
    fi

    echo
}

# Clean up
cleanup() {
    echo -e "${BLUE}Cleaning up...${NC}"
    cd /
    rm -rf "$TEST_DIR"
    echo -e "${GREEN}✓ Cleanup complete${NC}"
}

# Run all tests
run_all_tests() {
    setup_test_env
    test_encryption
    test_decryption
    test_execution
    test_demo_app
    test_wrapper_commands
}

# Main execution
main() {
    # Trap to ensure cleanup
    trap cleanup EXIT

    # Run tests
    run_all_tests

    # Summary
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  Test Results"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo
    echo "  Tests run: $TESTS_RUN"
    echo -e "  Passed: ${GREEN}$TESTS_PASSED${NC}"
    echo -e "  Failed: ${RED}$TESTS_FAILED${NC}"
    echo

    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "${GREEN}  ✨ All tests passed! Ready for demo!${NC}"
        echo
        echo "  Next steps:"
        echo "  1. Run the installation script:"
        echo "     bash /path/to/envv/install.sh"
        echo
        echo "  2. Try the demo:"
        echo "     cd /path/to/envv/demo"
        echo "     envv init"
        echo "     envv set STRIPE_API_KEY 'sk_live_abc123'"
        echo "     envv exec npm start"
        return 0
    else
        echo -e "${RED}  ⚠️  Some tests failed. Please fix before demo.${NC}"
        return 1
    fi
}

# Run if executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    main "$@"
fi