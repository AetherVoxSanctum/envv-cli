# envv SaaS Quick Start Guide

This guide will help you get started with envv's SaaS backend integration for team-based secrets management.

## Prerequisites

Before you begin, make sure you have these tools installed:

1. **Go** (1.21+) - for building the CLI
2. **SOPS** - for encryption/decryption
   ```bash
   # macOS
   brew install sops

   # Linux
   # Download from https://github.com/getsops/sops/releases
   ```

3. **Age** - for key generation
   ```bash
   # macOS
   brew install age

   # Linux
   # Download from https://github.com/FiloSottile/age/releases
   ```

## Installation

### Option 1: Install to user directory (recommended)

```bash
make install-envv-user
export PATH="$HOME/.local/bin:$PATH"  # Add to ~/.bashrc or ~/.zshrc
```

### Option 2: Install system-wide

```bash
make install-envv  # Requires sudo
```

### Option 3: Build only

```bash
make build-envv
./envv --help
```

### Quick install with dependency check

```bash
make quickstart-envv
```

## Configuration

### Environment Variables

- `ENVV_API_URL` - API base URL (default: `https://api.envv.sh`)
- `SOPS_AGE_KEY_FILE` - Path to age private keys (default: `~/.config/sops/age/keys.txt`)

Example:
```bash
export ENVV_API_URL="https://api.envv.sh"
```

## Getting Started

### 1. Register an Account

```bash
envv auth register
```

This will:
- Prompt for your email, password, and name
- Generate an age keypair automatically
- Save your private key to `~/.config/sops/age/keys.txt`
- Save your credentials to `~/.envv/credentials.json`

Example:
```
$ envv auth register
Email: you@example.com
Password: ••••••••
Full name: Your Name

🔑 Generating age encryption keypair...
✓ Age public key: age1xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
✓ Private key saved to /home/you/.config/sops/age/keys.txt

🚀 Registering account...

✅ Successfully registered and logged in as you@example.com
   User ID: usr_abc123
   Age Public Key: age1xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

### 2. Login (if you already have an account)

```bash
envv auth login
```

### 3. Check Your Authentication

```bash
envv auth whoami
```

### 4. Create an Organization

```bash
envv org create --name="My Company"
```

Example output:
```
✅ Organization created successfully!
   ID:   org_xyz789
   Name: My Company
   Slug: my-company
   Role: admin
```

### 5. List Your Organizations

```bash
envv org list
```

### 6. Create a Project

```bash
envv project create --org-id=org_xyz789 --name="Production App"
```

Example output:
```
✅ Project created successfully!
   ID:   proj_abc456
   Name: Production App
   Slug: production-app

To initialize this directory, run:
   envv project init --org-id=org_xyz789 --project-id=proj_abc456
```

### 7. Initialize Your Project Directory

```bash
cd /path/to/your/project
envv project init --org-id=org_xyz789 --project-id=proj_abc456
```

This creates `.envv/config.yaml` in your current directory:
```yaml
organization_id: org_xyz789
organization_name: My Company
project_id: proj_abc456
project_name: Production App
default_environment: development
```

### 8. Create a Secrets File

```bash
# Create your environment file
cat > .env.development <<EOF
DATABASE_URL=postgresql://localhost/myapp_dev
API_KEY=dev_secret_key_123
STRIPE_KEY=sk_test_xxxxx
EOF
```

### 9. Push Secrets to Backend

```bash
envv secrets push .env.development
```

This will:
- Fetch all project member public keys
- Encrypt the file locally with SOPS for all team members
- Extract SOPS metadata
- Upload encrypted data to backend

Example output:
```
🔐 Pushing secrets for environment: development
   Encrypting for 3 team members...
   Uploading to backend...

✅ Secrets pushed successfully!
   Environment: development
   Version:     1
   Size:        1234 bytes
```

### 10. Pull Secrets from Backend

```bash
envv secrets pull development
```

This will:
- Download encrypted data from backend
- Decrypt locally using your private key
- Save to `.env.development`

Example output:
```
🔓 Pulling secrets for environment: development
   Version: 1
   Decrypting...

✅ Secrets pulled and decrypted!
   Output:  .env.development
   Version: 1
   Updated: 2025-01-15T10:30:00Z
```

## Team Collaboration

### Invite Team Members

```bash
envv org invite --org-id=org_xyz789 --email=teammate@example.com --role=member
```

### View Project Members

```bash
envv project members
```

### Rotate Keys (Re-encrypt for new members)

When you add new team members, re-encrypt secrets so they can access them:

```bash
envv secrets rotate development
```

This will:
1. Pull and decrypt current secrets
2. Fetch updated member list
3. Re-encrypt for all current members
4. Push new version to backend

## Common Workflows

### Sync Local Changes

```bash
# Make changes to .env.development
vim .env.development

# Push updates
envv secrets push .env.development --env=development
```

### Work with Different Environments

```bash
# Development
envv secrets push .env.development --env=development
envv secrets pull development

# Staging
envv secrets push .env.staging --env=staging
envv secrets pull staging

# Production
envv secrets push .env.production --env=production
envv secrets pull production
```

### List Secret Versions

```bash
envv secrets list --env=development
```

Example output:
```
📜 Secret Versions for development (5):

VERSION  SIZE        CREATED              CREATED BY
-------  ----        -------              ----------
5        1456 bytes  2025-01-15 14:30     you@example.com
4        1234 bytes  2025-01-15 10:00     you@example.com
3        1234 bytes  2025-01-14 16:20     teammate@example.com
2        1189 bytes  2025-01-14 09:15     you@example.com
1        1156 bytes  2025-01-13 15:45     you@example.com
```

### Two-way Sync

```bash
# Pull latest + push local changes
envv secrets sync --env=development
```

## Project Status

Check your current project configuration:

```bash
envv project status
```

## Logout

```bash
envv auth logout
```

## Troubleshooting

### "sops: command not found"

Install SOPS:
```bash
# macOS
brew install sops

# Linux - download from releases
curl -LO https://github.com/getsops/sops/releases/download/v3.8.1/sops-v3.8.1.linux.amd64
sudo mv sops-v3.8.1.linux.amd64 /usr/local/bin/sops
sudo chmod +x /usr/local/bin/sops
```

### "age-keygen: command not found"

Install Age:
```bash
# macOS
brew install age

# Linux - download from releases
curl -LO https://github.com/FiloSottile/age/releases/download/v1.1.1/age-v1.1.1-linux-amd64.tar.gz
tar xzf age-v1.1.1-linux-amd64.tar.gz
sudo mv age/age* /usr/local/bin/
```

### "not logged in"

Make sure you're authenticated:
```bash
envv auth login
```

### "project not initialized"

Initialize your project directory:
```bash
envv project init --org-id=YOUR_ORG_ID --project-id=YOUR_PROJECT_ID
```

### "failed to decrypt"

Make sure:
1. Your private key is in `~/.config/sops/age/keys.txt`
2. You're a member of the project
3. The secrets were encrypted for you (if not, ask an admin to run `envv secrets rotate`)

## File Locations

- **User credentials**: `~/.envv/credentials.json` (global)
- **Project config**: `.envv/config.yaml` (per-project)
- **Age private keys**: `~/.config/sops/age/keys.txt` (global)

## Architecture Overview

```
┌─────────────────┐
│  Your Computer  │
├─────────────────┤
│                 │
│  1. Create      │
│     .env file   │
│                 │
│  2. envv        │◄─── Fetches team member public keys
│     secrets     │
│     push        │
│                 │
│  3. SOPS        │◄─── Encrypts locally for all team members
│     encrypts    │
│                 │
│  4. Upload      │───► Sends encrypted data + metadata
│     to API      │
│                 │
└─────────────────┘
        │
        ▼
┌─────────────────┐
│  envv Backend   │
├─────────────────┤
│                 │
│  • Stores only  │
│    ENCRYPTED    │
│    data         │
│                 │
│  • Never sees   │
│    plaintext    │
│                 │
│  • Zero-        │
│    knowledge    │
│    architecture │
│                 │
└─────────────────┘
```

## Security Notes

1. **Zero-knowledge**: The backend never sees your plaintext secrets
2. **Client-side encryption**: All encryption happens on your machine
3. **Multi-party encryption**: Secrets are encrypted for all team members simultaneously
4. **Key management**: Your private key never leaves your machine
5. **Age encryption**: Modern, secure encryption using X25519

## Next Steps

- Read `FINAL_INTEGRATION_PLAN.md` for architectural details
- Check `BACKEND_REQUIREMENTS.md` for backend API documentation
- Join the team and start collaborating on secrets! 🔐

## Support

- GitHub Issues: https://github.com/AetherVoxSanctum/envv-cli/issues
- Documentation: See docs/ directory
