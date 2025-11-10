# Design Partner Backend - 1 Week Sprint Plan

## 🎯 The Strategic Decision

**You're right** - you need accounts for design partners because:
1. **Team sharing** is your core value prop
2. **Audit trails** are essential for enterprise buyers
3. **Billing/metering** enables sustainable business
4. **Key management** needs orchestration

## ⚡ The "Fake It Till You Make It" Approach

### Option A: Manual Backend (Recommended for Speed)

**Week 1**: Build appearance of SaaS without full automation

```bash
# Design partner onboarding
envv auth signup      # → Emails you, you manually create account
envv team create      # → You manually provision team keys
envv init             # → Downloads pre-generated keys
```

**Behind the scenes:**
- Simple landing page with signup form
- Email notifications to you for each signup
- You manually create age keys and email them
- SQLite database for tracking users/teams

**Pros:**
- Ships this week
- Validates demand
- No hosting complexity
- Personal touch with design partners

**Cons:**
- Doesn't scale past 10 teams
- Manual work per signup

### Option B: Full Backend (More Work, More Professional)

**Week 1-2**: Build real SaaS infrastructure

```bash
# Automated flow
envv auth signup      # → Creates account immediately
envv team create      # → Auto-generates team keys
envv init             # → Downloads keys via API
```

**Infrastructure:**
- Go API server on Railway/Render
- PostgreSQL database
- JWT authentication
- Automated key generation

**Pros:**
- Professional appearance
- Scales to 100+ teams
- Sets up long-term architecture

**Cons:**
- 2 weeks minimum
- Hosting costs ($20-50/month)
- More complexity to debug

## 🚀 Recommended: Hybrid Approach

### This Week: Fake It Professional
1. **Landing page**: envv.dev with "Request Access" form
2. **CLI modifications**: Add auth commands that email you
3. **Manual fulfillment**: You create keys and email setup
4. **SQLite tracking**: Simple database of signups

### Next Week: Automate Critical Path
1. **Real API**: Automate the key generation
2. **Self-service**: Users can create teams instantly
3. **Billing prep**: Add usage tracking

## 🛠️ This Week's Implementation

### Day 1: Landing Page + Signup
```html
<!-- envv.dev -->
<h1>envv - Encrypted Secrets for Teams</h1>
<p>Never store secrets in plaintext again</p>
<form action="/signup" method="post">
  <input name="email" placeholder="you@company.com" required>
  <input name="company" placeholder="Your Company">
  <button>Request Design Partner Access</button>
</form>
```

### Day 2: CLI Auth Commands
```go
// cmd/envv/auth.go
func authSignup(email string) {
  // Send POST to envv.dev/api/signup
  // Display: "Account request sent! Check your email."
}

func authLogin(token string) {
  // Save token to ~/.envv/auth.json
  // Download team keys to ~/.envv/keys/
}
```

### Day 3: Backend Skeleton
```go
// Simple Go server
type User struct {
  Email   string
  Company string
  Status  string // pending, active
}

func handleSignup(w http.ResponseWriter, r *http.Request) {
  // Save to SQLite
  // Email you: "New signup: alice@acme.com"
  // Email them: "Thanks! We'll set up your account soon."
}
```

### Day 4: Manual Provisioning Script
```bash
#!/bin/bash
# provision-team.sh alice@acme.com "Acme Corp"

EMAIL=$1
COMPANY=$2

# Generate age key
age-keygen > teams/$EMAIL.key

# Create auth token
TOKEN=$(openssl rand -hex 32)

# Email setup instructions
echo "envv auth login $TOKEN" | mail -s "Your envv team is ready!" $EMAIL
```

### Day 5: Polish & Test
- Error handling
- Email templates
- Test with a few beta users

## 💰 Design Partner Pricing Strategy

### The Offer
```
"Free Pro plan for 6 months while we build this together.

In exchange, we need:
- Weekly 15-min feedback calls
- Permission to use your logo/quote
- Patience with early bugs
- Input on which features to build next

After 6 months: $9/user/month or we'll work out a special rate."
```

### What They Get
- Encrypted secrets for their team
- Audit trail of who accessed what
- Easy secret rotation
- Direct line to the founders
- Influence on product roadmap

## 📞 Design Partner Outreach Script

```
Hi [Name],

I saw your team uses [Docker/Kubernetes/etc]. Quick question:

How does your team currently share API keys and secrets?

Most teams we talk to are either:
1. Putting them in plaintext .env files (scary)
2. Using complex enterprise tools (overkill)

We built something that encrypts secrets at rest but keeps
the developer experience simple.

Want to see a 5-minute demo?

Best,
[Your name]

P.S. We're looking for 5 design partners to help shape
the product. Free access + direct founder input.
```

## 📋 This Week's Checklist

**Monday:**
- [ ] Register envv.dev domain
- [ ] Deploy simple landing page
- [ ] Set up email forwarding

**Tuesday:**
- [ ] Add auth commands to CLI
- [ ] Test signup flow end-to-end
- [ ] Create provisioning script

**Wednesday:**
- [ ] Polish signup experience
- [ ] Write email templates
- [ ] Test with 1-2 friendly users

**Thursday:**
- [ ] Create design partner deck
- [ ] Draft outreach emails
- [ ] Set up support Slack/Discord

**Friday:**
- [ ] Launch! Send to first 10 prospects
- [ ] Schedule demo calls for next week

## 🎯 Success Metrics

**This Week:**
- [ ] 5 signups from design partner outreach
- [ ] 2 demo calls scheduled
- [ ] 1 team actually using it

**Next Week:**
- [ ] 10 total signups
- [ ] 5 active teams
- [ ] $0 MRR but clear path to revenue

## 🚨 The Reality Check

**This is more work than just the CLI**, but it's the right move because:

1. **Team sharing** is your differentiation vs just "use SOPS"
2. **SaaS model** enables real business vs one-time sales
3. **Design partners** want to see the full vision
4. **Early revenue** validates market demand

**You can still demo the CLI this week** while building the backend. Just add "and here's how teams will share these secrets" to the pitch.

## 🤔 The Decision Point

**Option 1: Ship CLI demo this week, build backend next**
- Faster to first design partner feedback
- Risk: Looks like a dev tool, not a business

**Option 2: Build full backend first, demo next week**
- More impressive, professional appearance
- Risk: Delays validation by a week

**Option 3: Hybrid (my recommendation)**
- Demo CLI this week with "team features coming"
- Build manual backend in parallel
- Ship team features to early adopters next week

What feels right for your timeline and risk tolerance?