# envv Real Architecture - Where Data Goes

## 🤔 The Core Questions You Asked

1. **When users run `envv push`, where does data go?**
2. **How does Alice grant permissions to Bob and Charlie?**
3. **When Bob runs `envv auth login`, how does he get the right permissions?**

## 🏗️ The Real Architecture (Not Just Demo Fluff)

### Current State: Pure Local (SOPS)
```
Alice's Machine:                    Bob's Machine:
├── .env.encrypted (in repo)       ├── .env.encrypted (from git)
├── alice.agekey (private)          ├── bob.agekey (private)
└── .sops.yaml (age public keys)    └── .sops.yaml (same config)
```

**Problem**: Alice has to manually add Bob's public key to `.sops.yaml` and re-encrypt everything.

### Proposed Architecture: Hybrid Local + Remote

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│ Alice's CLI │    │ envv Backend│    │  Bob's CLI  │
│             │    │   (yours)   │    │             │
├─ Team Admin │◄──►│ Permissions │◄──►├─ Team Member│
├─ .env.local │    │ Key Registry│    ├─ .env.local │
└─ alice.key  │    │ Audit Log  │    └─ bob.key    │
              │    │ Team Config │    │
              │    └─────────────┘    │
              │                       │
              ▼                       ▼
         ┌─────────────────────────────────┐
         │        Git Repository           │
         │  ├── .env.encrypted             │
         │  ├── .sops.yaml (team config)   │
         │  └── envv.config.json           │
         └─────────────────────────────────┘
```

## 🔑 Key Distribution Strategy

### Option A: Backend-Managed Keys (Recommended)
```bash
# Alice sets up the team
envv team create "Acme Corp"
# → Backend generates master age key for team
# → Alice gets admin role automatically

# Alice adds Bob to team
envv team add bob@acme.com --role developer
# → Backend generates age keypair for Bob
# → Backend emails Bob his private key securely
# → Backend updates .sops.yaml with Bob's public key

# Bob joins
envv auth login  # Authenticates with backend
envv team join   # Downloads his private key + team config
envv init        # Sets up local encryption with team keys
```

### Option B: User-Generated Keys (More Secure)
```bash
# Bob generates his own key locally
envv auth signup
# → Generates age keypair on Bob's machine
# → Uploads public key to backend
# → Keeps private key local

# Alice approves Bob
envv team approve bob@acme.com --role developer
# → Backend adds Bob's public key to team
# → Re-encrypts team secrets with Bob's key included
```

## 🗄️ Where Data Actually Goes

### Local Storage (Per Developer)
```
~/.envv/
├── auth.json           # JWT token, user info
├── keys/
│   ├── bob.agekey      # Bob's private key
│   └── teams/
│       └── acme-corp/  # Team-specific keys
└── config/
    └── teams.json      # Team memberships
```

### Backend Storage (Your Database)
```sql
-- Team secrets are NOT stored in your backend!
-- Only metadata and keys

teams {
  id: uuid
  name: "Acme Corp"
  master_key_encrypted: "..." -- Encrypted with team admin keys
}

team_members {
  team_id: uuid
  user_id: uuid
  public_key: "age1ql3z..." -- For encrypting secrets
  role: "admin" | "developer" | "readonly"
}

team_secrets_metadata {
  team_id: uuid
  secret_name: "STRIPE_API_KEY"
  last_updated: timestamp
  updated_by: user_id
  -- NO SECRET VALUES STORED
}
```

### Git Repository (Shared)
```
project/
├── .env.encrypted      # SOPS-encrypted secrets
├── .sops.yaml          # Team public keys + config
└── envv.config.json    # Team metadata (optional)
```

## 🔄 The Real Workflow

### 1. Alice Sets Up Team
```bash
envv team create "Acme Corp"
# Backend creates team, generates master keys
# Alice becomes admin automatically

envv init
# Downloads team config
# Sets up .sops.yaml with team public keys

envv set STRIPE_API_KEY "sk_live_123"
# Encrypts with team keys
# Creates .env.encrypted in project

git add .env.encrypted .sops.yaml
git commit -m "Add team secrets"
git push
```

### 2. Alice Adds Bob
```bash
envv team invite bob@acme.com --role developer
# Backend sends invite email to Bob
# Bob gets signup link with team invitation
```

### 3. Bob Joins Team
```bash
envv auth signup --invite-token abc123
# Creates account linked to team invitation
# Generates age keypair for Bob

envv auth login
# Downloads team membership info

cd project  # Clone the repo
envv init
# Backend updates .sops.yaml with Bob's public key
# Re-encrypts .env.encrypted to include Bob
# Bob can now decrypt secrets
```

### 4. Secret Updates
```bash
# Alice updates a secret
envv set DATABASE_URL "postgres://new-url"
# Re-encrypts with all team member keys
# Optionally pushes to git

# Bob gets the update
git pull
envv exec npm start  # Works with new secret
```

## 💾 The "envv push" Implementation

### Option 1: Push Metadata Only
```bash
envv push
# Uploads secret metadata to backend:
# - Which secrets exist
# - Who last updated them
# - Audit trail
# Actual secret values stay in git
```

### Option 2: Central Secret Store (Enterprise)
```bash
envv push --remote
# Uploads encrypted secrets to backend
# Team members can envv pull instead of git
# More complex but enables better access control
```

### Option 3: Hybrid (Recommended)
```bash
# Default: git-based
envv set KEY "value"    # Updates .env.encrypted locally
git add .env.encrypted
git commit && git push

# Optional: metadata sync
envv sync              # Syncs metadata with backend
                      # Enables audit, notifications, etc.
```

## 🔐 Security Model

### What Backend Knows
- Team membership and roles
- Public keys for encryption
- Metadata about secrets (names, update times)
- Audit trail of who accessed when

### What Backend DOESN'T Know
- Actual secret values
- Private keys (user-generated)
- Decrypted data

### Trust Model
- **Backend**: Manages access control, audit, team membership
- **Git**: Stores encrypted secrets
- **Local**: Decryption happens on developer machines only

## 🚀 Implementation Plan

### Week 1: Basic Backend
- [x] User signup/auth
- [x] Team creation
- [ ] Age key generation API
- [ ] Team member management

### Week 2: CLI Integration
- [ ] `envv auth` commands
- [ ] `envv team` commands
- [ ] Automatic .sops.yaml management
- [ ] Backend key distribution

### Week 3: Advanced Features
- [ ] Role-based permissions
- [ ] Secret metadata sync
- [ ] Audit trail
- [ ] Key rotation

## 🤔 The Design Decisions You Need to Make

### 1. Key Generation
- **Backend-generated**: Easier UX, less secure
- **User-generated**: More secure, harder UX

### 2. Secret Storage
- **Git-only**: Simple, distributed, works offline
- **Backend-assisted**: Better audit/control, requires internet

### 3. Permission Model
- **Simple**: Admin/member only
- **Granular**: Per-secret permissions

### 4. Deployment Model
- **Git-based**: Secrets in repo, backend for metadata
- **Central**: Secrets in backend, git for code only

## 💡 My Recommendation

**Start with Git-based + metadata backend:**

1. Secrets stay in `.env.encrypted` in git repos
2. Backend manages team membership and public keys
3. CLI talks to backend for team management
4. Actual encryption/decryption uses SOPS locally

This gives you:
- ✅ Works offline
- ✅ Familiar git workflow
- ✅ Team management via backend
- ✅ Audit trail
- ✅ Scales to enterprise later

**The key insight**: Your backend doesn't store secrets, it stores **who can decrypt secrets**.