# Personal Testing Guide

This guide will help you test the newly implemented envv SaaS integration.

## Prerequisites Check

Before testing, verify you have the required tools:

```bash
# Check Go
go version  # Should be 1.21+

# Check SOPS
sops --version

# Check Age
age-keygen --version
```

If missing, install them:

**macOS:**
```bash
brew install sops age
```

**Linux:**
- SOPS: https://github.com/getsops/sops/releases
- Age: https://github.com/FiloSottile/age/releases

## Quick Setup

### Option 1: One-Command Install (Recommended)

```bash
cd /home/user/envv-cli
make quickstart-envv
```

This will:
- Build the CLI
- Install to ~/.local/bin
- Check dependencies
- Show next steps

### Option 2: Manual Build

```bash
cd /home/user/envv-cli
./build-and-install.sh
```

### Option 3: Just Build (No Install)

```bash
cd /home/user/envv-cli
make build-envv
./envv --help
```

## Test Plan

### 1. Verify Build

```bash
# Check version
envv --version

# Check new commands exist
envv --help | grep -A 5 "auth\|org\|project\|secrets"

# Test help for each command
envv auth --help
envv org --help
envv project --help
envv secrets --help
```

### 2. Set Backend URL (if testing against local backend)

```bash
export ENVV_API_URL="http://localhost:3000"  # Or your backend URL
```

For production testing:
```bash
export ENVV_API_URL="https://api.envv.sh"
```

### 3. Test Registration Flow

```bash
# Register (will generate age keypair automatically)
envv auth register

# Follow prompts:
# - Enter email
# - Enter password
# - Enter name

# Verify files created
ls -la ~/.envv/credentials.json
ls -la ~/.config/sops/age/keys.txt

# Check authentication
envv auth whoami
```

### 4. Test Organization Management

```bash
# Create organization
envv org create --name="Test Org"

# List organizations
envv org list

# Note the org ID for next steps
export ORG_ID="org_xxx"  # Replace with actual ID

# View members
envv org members --org-id=$ORG_ID
```

### 5. Test Project Management

```bash
# Create project
envv project create --org-id=$ORG_ID --name="Test Project"

# List projects
envv project list --org-id=$ORG_ID

# Note the project ID
export PROJECT_ID="proj_xxx"  # Replace with actual ID

# Initialize current directory
mkdir -p ~/test-envv-project
cd ~/test-envv-project
envv project init --org-id=$ORG_ID --project-id=$PROJECT_ID

# Verify config created
cat .envv/config.yaml

# Check status
envv project status
```

### 6. Test Secrets Management

```bash
cd ~/test-envv-project

# Create a test secrets file
cat > .env.development <<EOF
DATABASE_URL=postgresql://localhost/testdb
API_KEY=test_key_12345
SECRET_TOKEN=super_secret_token
STRIPE_KEY=sk_test_xxxxx
EOF

# Push secrets (this will encrypt and upload)
envv secrets push .env.development

# Move original file
mv .env.development .env.development.backup

# Pull secrets back
envv secrets pull development

# Verify they match
diff .env.development .env.development.backup

# List versions
envv secrets list --env=development

# Test with other environments
echo "STAGING_VAR=staging_value" > .env.staging
envv secrets push .env.staging --env=staging
envv secrets pull staging
```

### 7. Test Team Collaboration (if you have another account)

```bash
# Invite another user
envv org invite --org-id=$ORG_ID --email=other@example.com --role=member

# After they register and join, rotate keys
envv secrets rotate development

# List project members to verify
envv project members
```

### 8. Test Edge Cases

```bash
# Try pulling non-existent environment
envv secrets pull nonexistent  # Should error gracefully

# Try without project init
cd /tmp
envv secrets pull development  # Should error about missing config

# Try without authentication
envv auth logout
envv org list  # Should error about authentication
envv auth login  # Re-login
```

## What to Look For

### Success Indicators

✅ **Build succeeds** without errors
✅ **All commands** show up in help
✅ **Registration** generates age keypair
✅ **Credentials saved** to ~/.envv/credentials.json
✅ **Age keys saved** to ~/.config/sops/age/keys.txt
✅ **Organizations** can be created and listed
✅ **Projects** can be created and initialized
✅ **Secrets encrypted** locally before upload
✅ **Secrets decrypted** successfully after pull
✅ **SOPS metadata** included in push requests
✅ **Multi-member encryption** works with team members

### Potential Issues

⚠️ **Network errors** - Check ENVV_API_URL is correct
⚠️ **Build errors** - Check Go version and dependencies
⚠️ **sops not found** - Install SOPS
⚠️ **age-keygen not found** - Install Age
⚠️ **Decryption fails** - Check private key exists
⚠️ **401 errors** - Check authentication token

## Manual Testing Checklist

- [ ] CLI builds successfully
- [ ] `envv --version` works
- [ ] `envv auth register` creates account
- [ ] Age keypair generated and saved
- [ ] `envv auth whoami` shows user info
- [ ] `envv org create` creates organization
- [ ] `envv org list` shows organizations
- [ ] `envv project create` creates project
- [ ] `envv project init` creates .envv/config.yaml
- [ ] `envv project status` shows config
- [ ] `envv secrets push` encrypts and uploads
- [ ] `envv secrets pull` downloads and decrypts
- [ ] Pulled secrets match original
- [ ] `envv secrets list` shows versions
- [ ] `envv auth logout` clears credentials

## Debugging

### Enable Verbose Logging

```bash
# Set log level (if implemented)
export LOG_LEVEL=debug
envv secrets push .env.development
```

### Check API Requests

```bash
# Watch for HTTP requests
export ENVV_DEBUG=1  # If implemented
envv secrets push .env.development
```

### Inspect Files

```bash
# Check credentials
cat ~/.envv/credentials.json | jq .

# Check project config
cat .envv/config.yaml

# Check age keys
cat ~/.config/sops/age/keys.txt
```

### Test SOPS Directly

```bash
# Generate test config
cat > .sops.yaml <<EOF
creation_rules:
  - age: age1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6l5pzq
EOF

# Test encryption
echo "TEST_VAR=test_value" > test.env
sops -e test.env

# Test decryption
sops -d test.env
```

## Next Steps After Testing

1. **Report Issues**: Note any errors or unexpected behavior
2. **Test Different Scenarios**: Try with real secrets, multiple team members
3. **Performance Testing**: Test with large secret files
4. **Integration Testing**: Test with actual backend API
5. **Documentation Review**: Update docs based on real usage

## Useful Commands

```bash
# Clean up test data
rm -rf ~/test-envv-project
rm ~/.envv/credentials.json
envv auth logout

# Start fresh
envv auth register
# ... repeat test flow
```

## Getting Help

If you encounter issues:

1. Check logs and error messages
2. Verify prerequisites are installed
3. Check ENVV_API_URL is correct
4. Review QUICKSTART.md for common issues
5. Check FINAL_INTEGRATION_PLAN.md for architecture details

## Test Results Template

```markdown
## Test Results

Date: YYYY-MM-DD
Tester: Your Name
Backend: [local/staging/production]

### Build
- [ ] Success / [ ] Failed
- Notes:

### Registration
- [ ] Success / [ ] Failed
- Notes:

### Organizations
- [ ] Success / [ ] Failed
- Notes:

### Projects
- [ ] Success / [ ] Failed
- Notes:

### Secrets Push/Pull
- [ ] Success / [ ] Failed
- Notes:

### Issues Found
1.
2.
3.

### Suggestions
1.
2.
3.
```

---

Happy testing! 🚀
