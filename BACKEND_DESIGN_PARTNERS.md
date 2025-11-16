# envv Backend Strategy for Design Partners

## 🎯 Business Model Decision

**You're building a B2B SaaS tool, not just a CLI utility.**

Core insight: Teams need to share encrypted secrets securely. This requires:
- User accounts and authentication
- Team management and permissions
- Key distribution and rotation
- Audit logs and compliance
- Billing and subscription management

## 📊 Market Analysis

**Similar tools and their models:**

| Tool | Model | Pricing |
|------|--------|---------|
| 1Password Business | SaaS | $8/user/month |
| HashiCorp Vault | Self-hosted + Cloud | $2/secret/month |
| AWS Secrets Manager | Pay-per-use | $0.40/secret/month |
| Doppler | SaaS | $3/user/month |
| GitGuardian | SaaS | $25/dev/month |

**Your sweet spot**: Easier than Vault, cheaper than 1Password, developer-focused.

## 🚀 MVP Backend for Design Partners

### Architecture: CLI + Backend Service

```
┌─────────────┐    HTTPS/gRPC    ┌─────────────┐
│  envv CLI   │ ───────────────► │  envv API   │
│  (local)    │                  │  (hosted)   │
└─────────────┘                  └─────────────┘
      │                                  │
      ▼                                  ▼
┌─────────────┐                  ┌─────────────┐
│ .env.enc    │                  │ PostgreSQL  │
│ (project)   │                  │ (accounts)  │
└─────────────┘                  └─────────────┘
```

### Core Services Needed

#### 1. Authentication Service
```
POST /auth/signup
POST /auth/login
POST /auth/refresh
```

#### 2. Team Management
```
GET  /teams/{id}/members
POST /teams/{id}/invite
PUT  /teams/{id}/members/{user}/role
```

#### 3. Key Management
```
GET  /teams/{id}/keys
POST /teams/{id}/keys/rotate
GET  /teams/{id}/keys/public
```

#### 4. Audit Trail
```
GET /teams/{id}/audit
```

#### 5. Billing (Simple)
```
GET /teams/{id}/usage
POST /teams/{id}/billing
```

## 💾 Database Schema (Minimal)

```sql
-- Users and authentication
CREATE TABLE users (
  id UUID PRIMARY KEY,
  email VARCHAR UNIQUE NOT NULL,
  name VARCHAR NOT NULL,
  password_hash VARCHAR NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);

-- Teams/Organizations
CREATE TABLE teams (
  id UUID PRIMARY KEY,
  name VARCHAR NOT NULL,
  plan VARCHAR DEFAULT 'design_partner',
  created_at TIMESTAMP DEFAULT NOW()
);

-- Team membership
CREATE TABLE team_members (
  team_id UUID REFERENCES teams(id),
  user_id UUID REFERENCES users(id),
  role VARCHAR NOT NULL, -- owner, admin, member
  joined_at TIMESTAMP DEFAULT NOW(),
  PRIMARY KEY (team_id, user_id)
);

-- Age keys for teams
CREATE TABLE team_keys (
  id UUID PRIMARY KEY,
  team_id UUID REFERENCES teams(id),
  public_key TEXT NOT NULL,
  private_key_encrypted TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  rotated_at TIMESTAMP
);

-- Audit log
CREATE TABLE audit_events (
  id UUID PRIMARY KEY,
  team_id UUID REFERENCES teams(id),
  user_id UUID REFERENCES users(id),
  action VARCHAR NOT NULL, -- decrypt, rotate, invite, etc
  resource VARCHAR, -- secret name
  metadata JSONB,
  created_at TIMESTAMP DEFAULT NOW()
);

-- Usage tracking
CREATE TABLE usage_events (
  id UUID PRIMARY KEY,
  team_id UUID REFERENCES teams(id),
  event_type VARCHAR NOT NULL, -- decrypt, encrypt, rotate
  count INTEGER DEFAULT 1,
  date DATE DEFAULT CURRENT_DATE,
  PRIMARY KEY (team_id, event_type, date)
);
```

## 🛠️ Tech Stack Recommendations

### Option 1: Fast & Simple (Recommended)
- **Backend**: Go (you already know it) + Gin/Echo
- **Database**: PostgreSQL (Supabase for managed)
- **Auth**: JWT + email/password
- **Hosting**: Railway/Render (cheaper than AWS for MVP)
- **Domain**: envv.dev or envv.io

### Option 2: Modern SaaS Stack
- **Backend**: Node.js + tRPC
- **Database**: PlanetScale (MySQL)
- **Auth**: Clerk or Auth0
- **Hosting**: Vercel + serverless
- **Payments**: Stripe

### Option 3: All-in-One
- **Platform**: Supabase (database + auth + API)
- **Frontend**: Next.js dashboard
- **Hosting**: Vercel
- **CLI**: Go (talks to Supabase API)

## 📅 Development Timeline

### Week 1: Core Backend
- [ ] Set up database and basic API
- [ ] User signup/login endpoints
- [ ] Team creation and management
- [ ] Basic CLI authentication

### Week 2: Key Management
- [ ] Age key generation for teams
- [ ] Key rotation endpoints
- [ ] Secure key distribution to CLI

### Week 3: Audit & Polish
- [ ] Audit logging
- [ ] Usage tracking
- [ ] Error handling and logging
- [ ] Basic admin dashboard

### Week 4: Design Partner Ready
- [ ] Onboarding flow
- [ ] Billing skeleton (manual for now)
- [ ] Documentation and support

## 💰 Pricing Strategy for Design Partners

### Free Tier (Design Partners)
- Up to 5 team members
- 50 secrets per team
- Basic audit logs (30 days)
- Email support

### Paid Plans (Future)
- **Starter**: $9/user/month - Up to 20 users, unlimited secrets
- **Pro**: $19/user/month - Advanced audit, compliance features
- **Enterprise**: Custom - SSO, on-premise, custom contracts

### Design Partner Deal
```
"Free Pro plan for 6 months in exchange for:
- Weekly feedback calls
- Case study/testimonial
- Logo on website
- Feature input priority"
```

## 🔧 CLI Changes Needed

### Authentication Commands
```bash
envv auth login
envv auth logout
envv auth whoami
```

### Team Commands
```bash
envv team create "Acme Corp"
envv team invite alice@acme.com
envv team list
envv team switch acme-corp
```

### Modified Existing Commands
```bash
envv init    # Creates team project, fetches keys
envv set     # Logs to audit trail
envv exec    # Tracks usage
envv rotate  # Coordinates with backend
```

## 🚨 What This Means for Your Demo

### Good News
- Your current demo still works perfectly
- Just adds "sign up for team account" step
- Shows the full product vision

### Demo Flow Changes
```bash
# Old flow
envv init
envv set KEY value

# New flow
envv auth signup
envv team create "Demo Corp"
envv init  # Now fetches team keys
envv set KEY value  # Now tracked and auditable
```

## 🎯 Go/No-Go Decision

### Reasons to Build Backend First
- **Clear monetization path**
- **Validates team management value prop**
- **Enables audit/compliance features**
- **Creates customer relationship from day 1**

### Reasons to Wait
- **Adds complexity and hosting costs**
- **Delays design partner feedback**
- **Risk of over-engineering**

## 💡 My Recommendation: Hybrid Approach

### Phase 1: Design Partners (This Month)
- Keep current CLI-only demo working
- Build minimal backend (auth + teams + keys)
- Offer "Design Partner Edition" with manual setup
- Get 5-10 teams using it

### Phase 2: SaaS Launch (Next Month)
- Polish backend with billing
- Self-service onboarding
- Full audit/compliance features
- Public launch

## 🚀 Next Steps (This Week)

1. **Decide**: Backend-first or CLI-first for design partners?
2. **If backend**: Spin up database + basic auth API
3. **If CLI-first**: Add "contact us for team setup" to demo
4. **Either way**: Start reaching out to design partners

## 💭 The Strategic Question

**Are you selling a CLI tool or a team collaboration platform?**

Based on your value prop (encrypted secrets for TEAMS), I think you're selling the platform. Which means you need the backend.

But you could start with "Contact us to set up your team" for the first 10 customers and build it behind the scenes.

What feels right to you?