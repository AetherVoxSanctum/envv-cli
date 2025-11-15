#!/bin/bash

# Verification Script for envv Repository Cleanup
# Checks that all sensitive data has been removed

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  envv Repository Cleanup Verification"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[1;34m'
NC='\033[0m'

ISSUES_FOUND=0

echo -e "${BLUE}Checking working directory...${NC}"
echo

# Check if business files still exist
echo -n "  Checking for business strategy files... "
if [ -f "BACKEND_DESIGN_PARTNERS.md" ] || \
   [ -f "DESIGN_PARTNER_BACKEND_PLAN.md" ] || \
   [ -f "DESIGN_PARTNER_READY.md" ] || \
   [ -f "DEPLOYMENT_PLAN.md" ]; then
    echo -e "${RED}✗ FOUND${NC}"
    echo "    Files still exist in working directory!"
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
else
    echo -e "${GREEN}✓ CLEAN${NC}"
fi

# Check for /Users/wdr paths
echo -n "  Checking for /Users/wdr paths... "
if grep -r "/Users/wdr" --exclude-dir=.git --exclude="verify-cleanup.sh" . 2>/dev/null | grep -v "Binary file" > /dev/null; then
    echo -e "${RED}✗ FOUND${NC}"
    echo "    Hardcoded paths still present:"
    grep -r "/Users/wdr" --exclude-dir=.git --exclude="verify-cleanup.sh" . 2>/dev/null | grep -v "Binary file" | head -5
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
else
    echo -e "${GREEN}✓ CLEAN${NC}"
fi

# Check for suspicious Stripe keys
echo -n "  Checking for real-looking Stripe keys... "
if grep -r "sk_live_4eC39HqLyjWDarjtT1zdp7dc\|sk_live_4242424242424242" --exclude-dir=.git --exclude="verify-cleanup.sh" . 2>/dev/null > /dev/null; then
    echo -e "${YELLOW}⚠ FOUND${NC}"
    echo "    Demo Stripe keys should be sanitized:"
    grep -r "sk_live_4eC39HqLyjWDarjtT1zdp7dc\|sk_live_4242424242424242" --exclude-dir=.git --exclude="verify-cleanup.sh" . 2>/dev/null | head -3
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
else
    echo -e "${GREEN}✓ CLEAN${NC}"
fi

echo
echo -e "${BLUE}Checking git history...${NC}"
echo

# Check git history for removed files
echo -n "  Checking BACKEND_DESIGN_PARTNERS.md in history... "
if git log --all --full-history --oneline -- BACKEND_DESIGN_PARTNERS.md 2>/dev/null | grep . > /dev/null; then
    echo -e "${RED}✗ FOUND${NC}"
    echo "    File still exists in git history!"
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
else
    echo -e "${GREEN}✓ REMOVED${NC}"
fi

echo -n "  Checking DESIGN_PARTNER_BACKEND_PLAN.md in history... "
if git log --all --full-history --oneline -- DESIGN_PARTNER_BACKEND_PLAN.md 2>/dev/null | grep . > /dev/null; then
    echo -e "${RED}✗ FOUND${NC}"
    echo "    File still exists in git history!"
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
else
    echo -e "${GREEN}✓ REMOVED${NC}"
fi

echo -n "  Checking DESIGN_PARTNER_READY.md in history... "
if git log --all --full-history --oneline -- DESIGN_PARTNER_READY.md 2>/dev/null | grep . > /dev/null; then
    echo -e "${RED}✗ FOUND${NC}"
    echo "    File still exists in git history!"
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
else
    echo -e "${GREEN}✓ REMOVED${NC}"
fi

echo -n "  Checking DEPLOYMENT_PLAN.md in history... "
if git log --all --full-history --oneline -- DEPLOYMENT_PLAN.md 2>/dev/null | grep . > /dev/null; then
    echo -e "${RED}✗ FOUND${NC}"
    echo "    File still exists in git history!"
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
else
    echo -e "${GREEN}✓ REMOVED${NC}"
fi

echo
echo -e "${BLUE}Additional security checks...${NC}"
echo

# Check for any other sensitive patterns
echo -n "  Checking for personal emails (robinsonwesleyd)... "
if git log --all --format='%ae %an' | grep -i "robinsonwesleyd" > /dev/null; then
    echo -e "${YELLOW}⚠ FOUND${NC}"
    echo "    Your email is in git commit history (this is normal)"
    echo "    To use GitHub noreply email in future:"
    echo "      git config user.email 'username@users.noreply.github.com'"
else
    echo -e "${GREEN}✓ CLEAN${NC}"
fi

# Repository size check
echo -n "  Checking repository size... "
REPO_SIZE=$(du -sh .git 2>/dev/null | cut -f1)
echo "$REPO_SIZE"
echo "    (Smaller is better - history rewrite should reduce size)"

echo
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ $ISSUES_FOUND -eq 0 ]; then
    echo -e "${GREEN}  ✅ ALL CHECKS PASSED!${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo
    echo "Repository is clean and ready for public release!"
    echo
    echo "Final steps:"
    echo "  1. git push --force-with-lease"
    echo "  2. Make repository public on GitHub"
    echo "  3. Delete your backup once verified"
    echo
    exit 0
else
    echo -e "${RED}  ⚠️  ISSUES FOUND: $ISSUES_FOUND${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo
    echo "Please review the issues above before making the repository public."
    echo
    echo "You may need to:"
    echo "  1. Run the cleanup script again"
    echo "  2. Manually fix remaining issues"
    echo "  3. Re-run this verification script"
    echo
    exit 1
fi
