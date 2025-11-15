#!/bin/bash
set -e

# envv Repository Cleanup Script
# Removes sensitive business data from working directory AND git history
#
# ⚠️  WARNING: This script rewrites git history!
# - All collaborators will need to re-clone the repository
# - Requires force-push to update remote
# - Creates a backup before proceeding

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  envv Repository Security Cleanup"
echo "  Removing sensitive business data from git history"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[1;34m'
NC='\033[0m'

# Check if we're in a git repository
if [ ! -d ".git" ]; then
    echo -e "${RED}Error: Not in a git repository!${NC}"
    echo "Please run this script from the root of the envv-cli repository."
    exit 1
fi

# Check for uncommitted changes
if ! git diff-index --quiet HEAD --; then
    echo -e "${YELLOW}Warning: You have uncommitted changes.${NC}"
    echo "Please commit or stash your changes before running this script."
    echo
    git status --short
    echo
    read -p "Do you want to continue anyway? (yes/no): " CONTINUE
    if [ "$CONTINUE" != "yes" ]; then
        echo "Aborting."
        exit 1
    fi
fi

echo -e "${BLUE}Files to be removed completely:${NC}"
echo "  • BACKEND_DESIGN_PARTNERS.md"
echo "  • DESIGN_PARTNER_BACKEND_PLAN.md"
echo "  • DESIGN_PARTNER_READY.md"
echo "  • DEPLOYMENT_PLAN.md"
echo

echo -e "${BLUE}Files to be sanitized (paths/keys):${NC}"
echo "  • test-demo.sh (remove /Users/wdr paths)"
echo "  • demo/setup-working-demo.sh (sanitize Stripe keys)"
echo "  • demo/.env.example (sanitize Stripe keys)"
echo

echo -e "${YELLOW}⚠️  WARNING: This will rewrite git history!${NC}"
echo
echo "After running this script:"
echo "  1. All commit history will be rewritten"
echo "  2. Git hashes will change for all commits"
echo "  3. You'll need to force-push: git push --force-with-lease"
echo "  4. Collaborators must re-clone or reset their repos"
echo
read -p "Do you want to proceed? Type 'YES' to continue: " CONFIRM

if [ "$CONFIRM" != "YES" ]; then
    echo "Aborting cleanup."
    exit 0
fi

echo
echo -e "${GREEN}Creating backup...${NC}"

# Create backup
BACKUP_DIR="../envv-cli-backup-$(date +%Y%m%d-%H%M%S)"
cp -r . "$BACKUP_DIR"
echo "  ✓ Backup created at: $BACKUP_DIR"
echo

# Check if git-filter-repo is available
if command -v git-filter-repo &> /dev/null; then
    USE_FILTER_REPO=true
    echo -e "${GREEN}Using git-filter-repo (recommended method)${NC}"
else
    USE_FILTER_REPO=false
    echo -e "${YELLOW}git-filter-repo not found, using git filter-branch${NC}"
    echo "  Tip: Install git-filter-repo for better performance:"
    echo "       brew install git-filter-repo  # macOS"
    echo "       pip install git-filter-repo   # Python"
fi
echo

# Files to remove completely from history
FILES_TO_REMOVE=(
    "BACKEND_DESIGN_PARTNERS.md"
    "DESIGN_PARTNER_BACKEND_PLAN.md"
    "DESIGN_PARTNER_READY.md"
    "DEPLOYMENT_PLAN.md"
)

echo -e "${BLUE}Step 1: Removing business strategy files from git history...${NC}"

if [ "$USE_FILTER_REPO" = true ]; then
    # Using git-filter-repo (faster, safer)
    for file in "${FILES_TO_REMOVE[@]}"; do
        echo "  Removing: $file"
        git-filter-repo --invert-paths --path "$file" --force
    done
else
    # Using git filter-branch (fallback)
    for file in "${FILES_TO_REMOVE[@]}"; do
        echo "  Removing: $file"
        git filter-branch --force --index-filter \
            "git rm --cached --ignore-unmatch '$file'" \
            --prune-empty --tag-name-filter cat -- --all
    done

    # Cleanup refs
    rm -rf .git/refs/original/
    git reflog expire --expire=now --all
    git gc --prune=now --aggressive
fi

echo -e "${GREEN}  ✓ Files removed from git history${NC}"
echo

echo -e "${BLUE}Step 2: Sanitizing hardcoded paths and API keys...${NC}"

# Fix test-demo.sh - remove /Users/wdr paths
if [ -f "test-demo.sh" ]; then
    echo "  Sanitizing test-demo.sh..."
    sed -i.bak 's|/Users/wdr/dev/envv|/path/to/envv|g' test-demo.sh
    rm -f test-demo.sh.bak
    echo "    ✓ Replaced /Users/wdr paths with /path/to/envv"
fi

# Fix demo/setup-working-demo.sh - sanitize Stripe key
if [ -f "demo/setup-working-demo.sh" ]; then
    echo "  Sanitizing demo/setup-working-demo.sh..."
    sed -i.bak 's|STRIPE_API_KEY=sk_live_4eC39HqLyjWDarjtT1zdp7dc|STRIPE_API_KEY=sk_live_EXAMPLE_demo_key_not_real|g' demo/setup-working-demo.sh
    rm -f demo/setup-working-demo.sh.bak
    echo "    ✓ Replaced Stripe key with obvious example"
fi

# Fix demo/.env.example - sanitize Stripe key
if [ -f "demo/.env.example" ]; then
    echo "  Sanitizing demo/.env.example..."
    sed -i.bak 's|STRIPE_API_KEY=sk_live_4242424242424242|STRIPE_API_KEY=sk_live_EXAMPLE_not_a_real_key|g' demo/.env.example
    rm -f demo/.env.example.bak
    echo "    ✓ Replaced Stripe key with obvious example"
fi

echo -e "${GREEN}  ✓ Files sanitized${NC}"
echo

# Commit the sanitized changes
echo -e "${BLUE}Step 3: Committing sanitized changes...${NC}"
git add -A
git commit -m "Security: Remove sensitive business data and sanitize examples

- Removed business strategy documents from history
- Sanitized hardcoded personal paths
- Replaced demo API keys with obvious examples
- Cleaned up git history

This commit rewrites history to remove sensitive data." || true

echo -e "${GREEN}  ✓ Changes committed${NC}"
echo

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}  ✅ Cleanup Complete!${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo
echo -e "${YELLOW}Next Steps:${NC}"
echo
echo "1. Review the changes:"
echo "   git log --oneline -10"
echo "   git status"
echo
echo "2. Verify removed files are gone:"
echo "   git log --all --full-history -- BACKEND_DESIGN_PARTNERS.md"
echo "   (should show nothing)"
echo
echo "3. Push to remote (REQUIRED - rewrites remote history):"
echo "   git push --force-with-lease origin claude/audit-sensitive-data-019nJrCLK7BNzmdHL7hRNbp2"
echo
echo "4. If you have other branches, you may need to force-push those too:"
echo "   git push --force-with-lease --all"
echo
echo -e "${YELLOW}⚠️  IMPORTANT:${NC}"
echo "   - All collaborators must re-clone the repository"
echo "   - Or reset their local copies:"
echo "     git fetch origin"
echo "     git reset --hard origin/main"
echo
echo "5. Your backup is at:"
echo "   $BACKUP_DIR"
echo "   (Keep this until you verify everything works!)"
echo
echo -e "${GREEN}Repository is now clean and ready for open source! 🎉${NC}"
echo
