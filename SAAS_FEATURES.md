# envv SaaS Features

This document describes the new SaaS backend integration features added to envv CLI.

## Overview

envv CLI now includes full integration with the envv SaaS backend for team-based secrets management. This allows you to:

- **Collaborate** with team members on encrypted secrets
- **Zero-knowledge architecture** - backend never sees plaintext
- **Client-side encryption** using SOPS and Age
- **Multi-party encryption** - secrets encrypted for all team members
- **Version control** for secrets with rollback capability
- **Organizations and projects** for structured access control

## Quick Install

```bash
# Build and install
./build-and-install.sh

# Or use Make
make quickstart-envv
```

## Prerequisites

- **SOPS** - Encryption/decryption engine
- **Age** - Modern encryption tool for key generation

Install on macOS:
```bash
brew install sops age
```

Install on Linux:
- SOPS: https://github.com/getsops/sops/releases
- Age: https://github.com/FiloSottile/age/releases

## New Commands

### Authentication
```bash
envv auth register    # Register with age key generation
envv auth login       # Login to existing account
envv auth logout      # Logout and clear credentials
envv auth whoami      # Show current user info
```

### Organizations
```bash
envv org create       # Create organization
envv org list         # List your organizations
envv org members      # List organization members
envv org invite       # Invite team member
```

### Projects
```bash
envv project create   # Create project
envv project list     # List projects
envv project init     # Initialize current directory
envv project status   # Show project config
envv project members  # List project members
```

### Secrets
```bash
envv secrets push     # Push encrypted secrets
envv secrets pull     # Pull and decrypt secrets
envv secrets sync     # Two-way sync
envv secrets list     # List versions
envv secrets rotate   # Re-encrypt for new members
```

## Architecture

### Directory Structure

New packages added for SaaS integration:

```
envv-cli/
├── pkg/
│   ├── api/                 # Backend API clients
│   │   ├── client.go        # HTTP client with JWT auth
│   │   ├── auth.go          # User authentication
│   │   ├── organizations.go # Organization management
│   │   ├── projects.go      # Project management
│   │   └── secrets.go       # Secrets operations
│   ├── config/              # Configuration management
│   │   ├── credentials.go   # User credentials
│   │   └── project.go       # Project context
│   └── crypto/              # Cryptography operations
│       ├── sops.go          # SOPS integration
│       └── age.go           # Age key management
└── cmd/envv/
    ├── auth.go              # Auth commands
    ├── org.go               # Organization commands
    ├── project.go           # Project commands
    └── secrets.go           # Secrets commands
```

### Configuration Files

**Global user credentials:**
```
~/.envv/credentials.json
```

Contains:
- Access token (JWT)
- User ID
- Email
- Token expiration

**Per-project configuration:**
```
.envv/config.yaml
```

Contains:
- Organization ID
- Project ID
- Default environment

**Age private keys:**
```
~/.config/sops/age/keys.txt
```

Contains your private age keys for decryption.

### Workflow

```
┌──────────────────────────────────────────────────┐
│ 1. Create .env file locally                      │
└──────────────────────────────────────────────────┘
                    │
                    ▼
┌──────────────────────────────────────────────────┐
│ 2. envv secrets push                             │
│    - Fetches team member public keys from API   │
│    - Generates .sops.yaml with all recipients    │
│    - Encrypts file locally with SOPS             │
│    - Extracts SOPS metadata                      │
└──────────────────────────────────────────────────┘
                    │
                    ▼
┌──────────────────────────────────────────────────┐
│ 3. Backend API                                   │
│    - Stores encrypted data                       │
│    - Stores SOPS metadata                        │
│    - Never sees plaintext                        │
└──────────────────────────────────────────────────┘
                    │
                    ▼
┌──────────────────────────────────────────────────┐
│ 4. Team member pulls                             │
│    - Downloads encrypted data                    │
│    - Decrypts locally with their private key     │
│    - Saves to .env file                          │
└──────────────────────────────────────────────────┘
```

## Security Model

### Zero-Knowledge Architecture

1. **Registration**: Client generates age keypair locally
2. **Public key only**: Only public key sent to backend
3. **Private key**: Never leaves your machine
4. **Encryption**: Client-side only, using team member public keys
5. **Backend storage**: Encrypted data only, no plaintext access
6. **Decryption**: Client-side only, using your private key

### Key Features

- **AES-256-GCM** encryption via SOPS
- **X25519** key exchange via Age
- **JWT** authentication for API access
- **Multi-party encryption** for teams
- **Key rotation** support

## Example Workflow

### Initial Setup

```bash
# Register and generate keys
envv auth register \
  --email=you@company.com \
  --password=secure123 \
  --name="Your Name"

# Create organization
envv org create --name="ACME Inc"

# Create project
envv project create \
  --org-id=org_xxx \
  --name="Production API"

# Initialize directory
cd ~/projects/api
envv project init \
  --org-id=org_xxx \
  --project-id=proj_yyy
```

### Daily Usage

```bash
# Create secrets file
cat > .env.development <<EOF
DATABASE_URL=postgresql://localhost/myapp
API_KEY=secret_key_123
EOF

# Push to backend
envv secrets push .env.development

# Team member pulls
envv secrets pull development
```

### Team Collaboration

```bash
# Invite team member
envv org invite \
  --org-id=org_xxx \
  --email=teammate@company.com \
  --role=member

# After they register, rotate keys
envv secrets rotate development
```

## Environment Variables

- `ENVV_API_URL` - API base URL (default: `https://api.envv.sh`)
- `SOPS_AGE_KEY_FILE` - Path to age keys (default: `~/.config/sops/age/keys.txt`)

## Troubleshooting

See [QUICKSTART.md](QUICKSTART.md#troubleshooting) for common issues and solutions.

## Implementation Details

See [FINAL_INTEGRATION_PLAN.md](FINAL_INTEGRATION_PLAN.md) for comprehensive architectural documentation.

## Backend Requirements

See [BACKEND_REQUIREMENTS.md](BACKEND_REQUIREMENTS.md) for backend API specifications.

## Original SOPS Features

All original SOPS features are still available:

```bash
sops -e file.yaml        # Encrypt
sops -d file.yaml        # Decrypt
sops file.yaml           # Edit
```

The new `envv` commands add team collaboration on top of SOPS encryption.

## Support

- GitHub: https://github.com/AetherVoxSanctum/envv-cli
- Issues: https://github.com/AetherVoxSanctum/envv-cli/issues
