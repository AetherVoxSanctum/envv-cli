# envv Demo - Blog Application

A working demo blog application that showcases encrypted environment variable management using SOPS/envv.

## What This Demonstrates

- **The Problem**: Traditional .env files expose secrets in plaintext
- **The Solution**: Encrypted environment files that can be safely shared in git
- **Team Workflow**: How multiple developers can share encrypted secrets

## What Actually Works

This demo uses the existing SOPS functionality with a wrapper script to provide a better UX.

### Current Commands
```bash
./setup-working-demo.sh    # Sets up working encryption demo
./envv list               # Shows encrypted secret names
./envv exec npm start     # Runs app with decrypted secrets
```

### What You'll See
1. A blog application that needs 4 environment variables:
   - `ANALYTICS_KEY_GOOGLE` - Google Analytics tracking
   - `ANALYTICS_KEY_MIXPANEL` - Mixpanel analytics
   - `STRIPE_API_KEY` - Payment processing
   - `BACKEND_SECRET_KEY` - Admin API access

2. Secrets encrypted with age keys (individual or team)
3. Application that adapts based on which secrets are available
4. No plaintext secrets anywhere on disk

## Running the Demo

### Setup
```bash
# Install dependencies
npm install

# Install age for encryption (macOS)
brew install age

# Set up encrypted secrets
./setup-working-demo.sh
```

### Usage
```bash
# List available secrets (names only)
./envv list

# Run the blog application
./envv exec npm start

# View encrypted file (unreadable)
cat .env.encrypted

# Test different scenarios by setting secrets
./envv set NEW_SECRET "test_value"
```

### Demo Flow
1. Show `.env.example` - demonstrates plaintext exposure problem
2. Run `./setup-working-demo.sh` - creates encrypted secrets
3. Show `.env.encrypted` - secrets are now unreadable
4. Run `./envv exec npm start` - app works perfectly with encrypted secrets
5. Visit http://localhost:3000 - blog shows which secrets are loaded

## Architecture

### Current Implementation
- **Encryption**: SOPS with age keys
- **Local**: Wrapper script provides friendly commands
- **Team**: Manual key distribution (for demo)

### Future Implementation (Planned)
- **Backend**: Team management and key distribution
- **CLI**: Native envv commands (`envv auth`, `envv team`)
- **Permissions**: Role-based access to different secrets

## Files

```
demo/
├── README.md                 # This file
├── setup-working-demo.sh    # Demo setup script
├── envv                     # Wrapper script for SOPS
├── package.json             # Node.js dependencies
├── server/index.js          # Express server
├── public/                  # Frontend files
├── posts/                   # Blog content
└── .env.example             # Shows the problem (plaintext)
```

## Technical Notes

### Encryption
- Uses `age` for public key cryptography
- SOPS handles the actual encryption/decryption
- Keys stored in `keys/demo.agekey` (for demo purposes)

### Team Simulation
This demo simulates team access using a single age key. In production:
- Each team member would have their own age keypair
- Backend would manage team membership and key distribution
- Secrets would be encrypted for all team members' public keys

### Security Model
- Secrets are encrypted at rest
- Decryption happens in memory only when running commands
- Private keys never leave individual machines
- Public keys can be safely shared

## Limitations (Honest Assessment)

### What's Missing
- No real team management (manual key setup)
- No backend integration (pure local demo)
- No role-based permissions
- No audit logging
- No automatic key rotation

### What's Simulated
- Team member Alice, Bob, Charlie (all use same demo key)
- Backend API calls (wrapper script only)
- Automatic team synchronization

## Next Steps for Production

1. **Backend Development**: User accounts, team management, key distribution
2. **CLI Integration**: Replace wrapper with real envv commands
3. **Key Management**: Automatic generation, rotation, revocation
4. **Permissions**: Role-based access control
5. **Audit**: Logging and compliance features

## For Design Partners

This demo shows the core value proposition:
- Developers never see plaintext secrets
- Applications work normally with encrypted files
- Teams can safely share secrets via git

The missing pieces (backend, team management) are the business opportunity - what teams would pay for to use this securely at scale.

---

**Note**: This is a working prototype demonstrating the core concept. Production implementation would require backend infrastructure for team management and key distribution.