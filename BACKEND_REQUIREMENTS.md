# envv Backend Requirements
## Comprehensive Implementation Guide for SaaS Secrets Management

**Document Version:** 1.0
**Target Audience:** Backend development agent
**Current CLI Status:** SOPS-based encryption working locally, wrapper commands implemented
**Backend Status:** ~70% complete (estimated)

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Architecture Overview](#architecture-overview)
3. [Short-term Requirements (MVP)](#short-term-requirements-mvp)
4. [Medium-term Requirements (Production)](#medium-term-requirements-production)
5. [API Specifications](#api-specifications)
6. [Database Schema](#database-schema)
7. [Security Requirements](#security-requirements)
8. [CLI Integration Points](#cli-integration-points)
9. [Testing Requirements](#testing-requirements)
10. [Operational Requirements](#operational-requirements)

---

## Executive Summary

The envv backend serves as an **access control and distribution layer** for encrypted secrets managed by teams. The critical security principle is: **the backend NEVER has access to plaintext secrets**. All encryption/decryption happens client-side using age keys (via SOPS).

### Core Responsibilities
- User authentication and authorization
- Team and project management
- Public key distribution for multi-party encryption
- Encrypted blob storage and versioning
- Audit logging for compliance
- Access control enforcement

### What the Backend Does NOT Do
- Decrypt secrets (no access to private keys)
- Store plaintext credentials
- Act as a keystore for private keys (users control these)

---

## Architecture Overview

### High-Level Flow

```
┌─────────────────┐
│   envv CLI      │ (User's machine)
│  - age keypair  │
│  - SOPS binary  │
└────────┬────────┘
         │ HTTPS + JWT
         ↓
┌─────────────────┐
│  Backend API    │
│  - Auth         │
│  - Teams        │
│  - Projects     │
│  - Permissions  │
└────────┬────────┘
         ↓
┌─────────────────┐
│   PostgreSQL    │
│  - Users        │
│  - Encrypted    │
│    secrets      │
│  - Audit logs   │
└─────────────────┘
```

### Key Principles

1. **Zero-Knowledge Architecture**: Backend cannot decrypt secrets
2. **Client-Side Encryption**: CLI encrypts with team members' public keys
3. **Public Key Distribution**: Backend provides list of authorized public keys
4. **Access Control Layer**: Backend enforces who can read/write which projects
5. **Immutable Audit Trail**: All operations logged for compliance

---

## Short-term Requirements (MVP)

**Goal:** Enable design partners to use envv for team secrets management with core functionality.

**Timeline:** Complete these first to unblock CLI integration and user testing.

### 1. Authentication System

#### 1.1 User Registration
- **Endpoint:** `POST /api/v1/auth/register`
- **Requirements:**
  - Accept: email, password, full_name, age_public_key
  - Validate email format
  - Hash password with bcrypt (cost factor: 12)
  - Validate age public key format: `age1[a-z0-9]{58}`
  - Generate user UUID
  - Create verification token
  - Send verification email (can be stubbed for MVP)
  - Return: user object (without password) + JWT token

- **Validation Rules:**
  - Email: Valid format, unique, max 255 chars
  - Password: Min 12 chars, must contain uppercase, lowercase, number, special char
  - Name: Max 255 chars, required
  - Public key: Valid age1 format, unique, required

#### 1.2 User Login
- **Endpoint:** `POST /api/v1/auth/login`
- **Requirements:**
  - Accept: email, password
  - Verify credentials with constant-time comparison
  - Generate JWT token (HS256, 24hr expiry)
  - Include in token: user_id, email, created_at
  - Return: JWT + user object + list of user's teams
  - Log successful/failed login attempts (include IP, user_agent)

#### 1.3 Token Refresh
- **Endpoint:** `POST /api/v1/auth/refresh`
- **Requirements:**
  - Accept: valid JWT (can be expired up to 7 days)
  - Issue new JWT with extended expiry
  - Invalidate old token (token blacklist or rotation ID)

#### 1.4 Session Management
- **Endpoint:** `GET /api/v1/auth/me`
- **Requirements:**
  - Return current user info from JWT
  - Include: teams, projects (summary), active_sessions count
  - Validate JWT on every request (middleware)

#### 1.5 Logout
- **Endpoint:** `POST /api/v1/auth/logout`
- **Requirements:**
  - Add JWT to blacklist (Redis or DB table)
  - Clear any server-side sessions
  - Return success status

**Implementation Notes:**
- Use existing JWT library (jsonwebtoken, jose, or equivalent)
- Store hashed passwords only (bcrypt)
- Implement rate limiting: 5 failed logins per IP per 15 min
- Consider OAuth2 providers (GitHub, Google) for medium-term

---

### 2. Team Management

#### 2.1 Create Team
- **Endpoint:** `POST /api/v1/teams`
- **Requirements:**
  - Accept: name, description (optional)
  - Validate: name unique per user, 3-50 chars, alphanumeric + spaces
  - Auto-add creator as owner
  - Return: team object with id, name, owner_id, created_at

#### 2.2 List User's Teams
- **Endpoint:** `GET /api/v1/teams`
- **Requirements:**
  - Return all teams where user is a member
  - Include: team details, user's role, member_count, project_count
  - Support pagination: limit (default 50), offset

#### 2.3 Get Team Details
- **Endpoint:** `GET /api/v1/teams/:team_id`
- **Requirements:**
  - Verify user is team member
  - Return: full team details, member list with roles, project list
  - Include each member's: id, email, name, public_age_key, role, joined_at

#### 2.4 Invite Team Member
- **Endpoint:** `POST /api/v1/teams/:team_id/invites`
- **Requirements:**
  - Verify requester is team admin or owner
  - Accept: email, role (member|admin)
  - Generate invite token (UUID, 7 day expiry)
  - Send invite email with link: `https://envv.app/invite/{token}`
  - Store pending invite in DB
  - Return: invite object

#### 2.5 Accept Team Invite
- **Endpoint:** `POST /api/v1/teams/invites/:token/accept`
- **Requirements:**
  - Verify token valid and not expired
  - Verify user has registered account (has public key)
  - Add user to team with specified role
  - Mark invite as accepted
  - Return: team object

#### 2.6 List Team Members
- **Endpoint:** `GET /api/v1/teams/:team_id/members`
- **Requirements:**
  - Verify user is team member
  - Return: array of members with roles and public keys
  - **Critical for CLI:** This endpoint provides public keys for encryption
  - Include: user_id, email, name, public_age_key, role, joined_at

#### 2.7 Update Member Role
- **Endpoint:** `PATCH /api/v1/teams/:team_id/members/:user_id`
- **Requirements:**
  - Verify requester is team owner or admin
  - Accept: role (member|admin)
  - Prevent owner from being demoted (require transfer ownership first)
  - Return: updated member object

#### 2.8 Remove Team Member
- **Endpoint:** `DELETE /api/v1/teams/:team_id/members/:user_id`
- **Requirements:**
  - Verify requester is team owner or admin
  - Cannot remove owner (require transfer ownership first)
  - Remove from team_members table
  - Revoke access to all team projects
  - **Important:** Secrets remain encrypted with their public key (cannot undo encryption)
  - Return: success status

**Implementation Notes:**
- Team roles: `owner` (1 per team), `admin` (manage members), `member` (default)
- Owner transfer: separate endpoint `POST /api/v1/teams/:id/transfer-ownership`
- Consider soft-delete for audit trail

---

### 3. Project Management

Projects are containers for encrypted secrets, scoped to a team.

#### 3.1 Create Project
- **Endpoint:** `POST /api/v1/projects`
- **Requirements:**
  - Accept: name, team_id, description (optional)
  - Verify user is team member
  - Validate: name unique within team, 3-50 chars
  - Create empty project (no secrets yet)
  - Auto-grant creator full access
  - Return: project object with id, name, team_id, created_by, created_at

#### 3.2 List Projects
- **Endpoint:** `GET /api/v1/projects?team_id={id}`
- **Requirements:**
  - Filter by team_id (required)
  - Return only projects user has access to
  - Include: project details, user's permission level, secret_count, last_updated
  - Support pagination

#### 3.3 Get Project Details
- **Endpoint:** `GET /api/v1/projects/:project_id`
- **Requirements:**
  - Verify user has access (any permission level)
  - Return: project metadata, team info, user's permissions
  - Include: list of users with access (names, roles, permissions)
  - Include: secret keys (names only, not values) if user has read access

#### 3.4 Update Project
- **Endpoint:** `PATCH /api/v1/projects/:project_id`
- **Requirements:**
  - Verify user has admin permission
  - Accept: name, description
  - Return: updated project object

#### 3.5 Delete Project
- **Endpoint:** `DELETE /api/v1/projects/:project_id`
- **Requirements:**
  - Verify user has admin permission
  - Soft delete (set deleted_at timestamp)
  - Cascade: mark all secrets as deleted
  - Retain audit logs
  - Consider: 30-day grace period before permanent deletion

#### 3.6 Grant Project Access
- **Endpoint:** `POST /api/v1/projects/:project_id/access`
- **Requirements:**
  - Verify requester has admin permission
  - Accept: user_id, permission (read|write|admin)
  - Verify target user is team member
  - Create project_access record
  - **Trigger:** CLI should re-encrypt secrets to include new user's public key
  - Return: access grant object

#### 3.7 Revoke Project Access
- **Endpoint:** `DELETE /api/v1/projects/:project_id/access/:user_id`
- **Requirements:**
  - Verify requester has admin permission
  - Remove project_access record
  - **Important:** User retains ability to decrypt old versions (cannot undo)
  - **Best Practice:** Recommend secret rotation after revocation
  - Return: success status

#### 3.8 List Project Members
- **Endpoint:** `GET /api/v1/projects/:project_id/members`
- **Requirements:**
  - Verify user has access to project
  - Return: users with access, their permission levels, public keys
  - **Critical:** CLI uses this to get public keys for encryption
  - Include: user_id, email, name, public_age_key, permission, granted_at

**Implementation Notes:**
- Permission levels: `read` (decrypt), `write` (set/update), `admin` (manage access)
- Team admins automatically get admin permission on team projects
- Consider: environment-based projects (dev, staging, prod)

---

### 4. Secrets Management

**Key Concept:** Backend stores SOPS-encrypted blobs. It never sees plaintext.

#### 4.1 Push Encrypted Secrets
- **Endpoint:** `POST /api/v1/projects/:project_id/secrets`
- **Requirements:**
  - Verify user has write permission
  - Accept: encrypted_data (SOPS-encrypted JSON/YAML string), format (json|yaml|env)
  - Validate: data contains SOPS metadata (sops.age field with encrypted keys)
  - Store encrypted blob in DB
  - Create version entry (for history)
  - Update project.last_updated timestamp
  - Log audit event: "secrets_pushed" with user, timestamp, IP
  - Return: version_id, uploaded_at

- **Validation:**
  - Max blob size: 1MB (configurable)
  - Must contain SOPS age metadata
  - Must be valid JSON/YAML/ENV format (don't validate structure, just format)

#### 4.2 Pull Encrypted Secrets
- **Endpoint:** `GET /api/v1/projects/:project_id/secrets`
- **Requirements:**
  - Verify user has read permission
  - Retrieve latest version of encrypted blob
  - Return: encrypted_data, format, version_id, updated_at
  - Log audit event: "secrets_pulled"
  - **Security:** User's CLI will decrypt locally with their private key

- **Query Parameters:**
  - `version`: Optional, retrieve specific version (default: latest)
  - `format`: Return format preference (json|yaml|env)

#### 4.3 Get Secret Metadata
- **Endpoint:** `GET /api/v1/projects/:project_id/secrets/metadata`
- **Requirements:**
  - Verify user has read permission
  - Return: secret key names only (no values)
  - **How:** Decrypt with temporary parser to extract keys (or require CLI to push metadata)
  - Include: key_name, last_updated, updated_by
  - Return: array of secret metadata

**Alternative Implementation:**
- Have CLI push metadata separately: key names, description
- Store in separate table: secret_metadata (project_id, key_name, description, encrypted_value_pointer)

#### 4.4 Secret Versioning
- **Endpoint:** `GET /api/v1/projects/:project_id/secrets/versions`
- **Requirements:**
  - Verify user has read permission
  - Return: list of versions with timestamps, who updated, change description
  - Support pagination (limit, offset)
  - Include: version_id, created_at, created_by (user email), size_bytes, format

#### 4.5 Rollback to Version
- **Endpoint:** `POST /api/v1/projects/:project_id/secrets/rollback`
- **Requirements:**
  - Verify user has write permission
  - Accept: version_id
  - Create new version that's a copy of specified version
  - Log audit event: "secrets_rolled_back"
  - Return: new version_id

**Implementation Notes:**
- Store encrypted blobs in PostgreSQL JSONB or TEXT column
- Consider: S3 storage for larger blobs (>100KB)
- Implement version pruning: keep last 50 versions per project
- Compression: gzip encrypted blobs before storage

---

### 5. Audit Logging

**Compliance Requirement:** All secret operations must be logged immutably.

#### 5.1 Log All Operations
- **Events to Log:**
  - User: login, logout, registration, password_change
  - Team: created, member_added, member_removed, role_changed
  - Project: created, deleted, access_granted, access_revoked
  - Secrets: pushed, pulled, rollback, key_added, key_deleted

- **Log Entry Fields:**
  - event_id (UUID)
  - timestamp (timestamptz, UTC)
  - user_id
  - user_email (denormalized for reports)
  - event_type (enum)
  - resource_type (user|team|project|secret)
  - resource_id
  - action (string: created, updated, deleted, accessed)
  - details (JSONB: additional context)
  - ip_address
  - user_agent
  - success (boolean)
  - error_message (if failed)

#### 5.2 Query Audit Logs
- **Endpoint:** `GET /api/v1/audit`
- **Requirements:**
  - Verify user is team admin (for team scope) or project admin
  - Query params: team_id, project_id, user_id, event_type, start_date, end_date
  - Return: paginated list of audit entries
  - Support CSV export for compliance
  - Apply retention policy: keep logs for minimum 1 year

#### 5.3 Project-Specific Audit
- **Endpoint:** `GET /api/v1/projects/:project_id/audit`
- **Requirements:**
  - Verify user has admin permission on project
  - Return: all events related to this project
  - Include: who accessed, when, from where (IP), what operation

**Implementation Notes:**
- Use separate audit_log table (never delete entries)
- Consider: append-only database like ClickHouse for scale
- Implement log archival: move logs >1 year to cold storage
- Add indexes: user_id, project_id, timestamp, event_type

---

### 6. Public Key Management

**Critical for Encryption:** CLI needs to know which public keys to encrypt for.

#### 6.1 Update User's Public Key
- **Endpoint:** `PATCH /api/v1/users/me/public-key`
- **Requirements:**
  - Accept: new_age_public_key
  - Validate key format
  - **Important:** Changing key requires re-encryption of all secrets user has access to
  - Store new key
  - Mark old key as deprecated (keep for history)
  - Return: updated user object

- **Warning to User:**
  - "Changing your public key will require all projects to re-encrypt secrets"
  - Consider: automated re-encryption process or require manual `envv sync` per project

#### 6.2 Get User's Public Keys (for Team)
- **Endpoint:** `GET /api/v1/teams/:team_id/public-keys`
- **Requirements:**
  - Verify user is team member
  - Return: array of {user_id, email, public_age_key} for all team members
  - **Used by CLI** when encrypting secrets for team projects
  - Include only active keys (not deprecated)

#### 6.3 Key Rotation Notification
- **Endpoint:** `POST /api/v1/projects/:project_id/notify-reencrypt`
- **Requirements:**
  - When team membership changes or keys update
  - Mark project as "needs_reencryption"
  - **CLI checks this flag** and prompts user to run `envv sync`
  - Return: list of projects needing re-encryption

**Implementation Notes:**
- Store key history: user_public_keys table (user_id, public_key, created_at, deprecated_at)
- Current key: users.current_public_key_id foreign key
- Add index on public_key for uniqueness check

---

### 7. Rate Limiting & Security

#### 7.1 Rate Limiting
Implement rate limits on all endpoints:

| Endpoint Pattern | Limit | Window |
|-----------------|-------|--------|
| `/api/v1/auth/login` | 5 requests | 15 min |
| `/api/v1/auth/register` | 3 requests | 1 hour |
| `/api/v1/*/secrets` (read) | 100 requests | 1 min |
| `/api/v1/*/secrets` (write) | 20 requests | 1 min |
| All other endpoints | 200 requests | 1 min |

- Use IP + user_id as rate limit key
- Return 429 Too Many Requests with Retry-After header
- Log rate limit violations (potential abuse)

#### 7.2 Input Validation
- Sanitize all user inputs (XSS prevention)
- Validate UUIDs format for all ID parameters
- Validate JSON/YAML structure before storage
- Max request body size: 2MB
- Reject requests with unexpected content types

#### 7.3 CORS Configuration
- Allow origins: `https://envv.app`, `http://localhost:*` (dev)
- Allowed methods: GET, POST, PATCH, DELETE, OPTIONS
- Allowed headers: Authorization, Content-Type
- Expose headers: X-RateLimit-*

#### 7.4 Security Headers
```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains
Content-Security-Policy: default-src 'self'
```

---

## Medium-term Requirements (Production)

**Goal:** Production-ready system with enhanced security, scale, and features.

### 8. Enhanced Authentication

#### 8.1 Multi-Factor Authentication (MFA)
- **Endpoint:** `POST /api/v1/auth/mfa/enable`
- **Requirements:**
  - Support TOTP (Google Authenticator, Authy)
  - Generate secret, return QR code
  - Store encrypted secret (AES-256)
  - Require MFA on login: `POST /api/v1/auth/mfa/verify`
  - Backup codes: generate 10, single-use
  - Support recovery: email-based or security questions

#### 8.2 OAuth2 Providers
- **Providers:** GitHub, Google, GitLab
- **Endpoints:**
  - `GET /api/v1/auth/oauth/:provider` (redirect to OAuth flow)
  - `GET /api/v1/auth/oauth/:provider/callback` (handle callback)
- **Requirements:**
  - Link OAuth account to existing envv account
  - On first login, auto-create account
  - Require public key setup after OAuth registration
  - Store: provider_id, provider_user_id, access_token (encrypted)

#### 8.3 API Keys (Machine Access)
- **Endpoint:** `POST /api/v1/auth/api-keys`
- **Requirements:**
  - Generate: scoped API keys (project-specific or team-specific)
  - Scopes: `read:secrets`, `write:secrets`, `admin:projects`
  - Format: `envv_<environment>_<random>` (e.g., `envv_prod_k3j4h5k6j7h8`)
  - Hash before storage (SHA-256)
  - Support: expiration date, revocation
  - Log: all operations with API keys
  - Return: key only once (cannot retrieve again)

#### 8.4 Session Management
- **Requirements:**
  - Track active sessions: device, IP, location (GeoIP), last_active
  - Endpoint: `GET /api/v1/auth/sessions` (list active sessions)
  - Endpoint: `DELETE /api/v1/auth/sessions/:session_id` (revoke session)
  - Auto-expire sessions after 7 days inactive
  - Notify user of new login from unknown device/location

---

### 9. Advanced Project Features

#### 9.1 Environments
- **Concept:** Projects can have multiple environments (dev, staging, prod)
- **Structure:**
  ```
  project/
    ├── dev/        (separate encrypted secrets)
    ├── staging/
    └── prod/
  ```
- **Endpoints:**
  - `POST /api/v1/projects/:id/environments`
  - `GET /api/v1/projects/:id/environments/:env/secrets`
  - `POST /api/v1/projects/:id/environments/:env/secrets`
- **Permissions:** Per-environment access control
  - Example: Juniors get `dev` access, seniors get `prod` access

#### 9.2 Secret Inheritance
- **Feature:** Inherit secrets from parent environment
- **Example:** `prod` inherits from `staging`, overrides specific keys
- **Implementation:**
  - CLI merges secrets client-side
  - Backend stores inheritance relationship
  - Endpoint: `GET /api/v1/projects/:id/environments/:env/resolved-secrets`

#### 9.3 Secret Tagging
- **Feature:** Tag secrets with metadata (service, category, compliance_level)
- **Schema:** `secret_tags` table (project_id, key_name, tag_name, tag_value)
- **Use Cases:**
  - Filter: "Show all PCI-compliant secrets"
  - Reports: "List all secrets tagged 'database'"
  - Compliance: "Ensure all prod secrets tagged with owner"

#### 9.4 Secret Expiration
- **Feature:** Set expiration dates on secrets (for rotation reminders)
- **Schema:** Add `expires_at` to secret metadata
- **Behavior:**
  - CLI warns: "SECRET_KEY expires in 7 days"
  - Email notifications
  - Endpoint: `GET /api/v1/secrets/expiring?days=30`

---

### 10. Webhooks & Integrations

#### 10.1 Webhooks
- **Events:**
  - `secret.pushed`, `secret.pulled`, `member.added`, `member.removed`, `project.created`
- **Endpoint:** `POST /api/v1/projects/:id/webhooks`
- **Requirements:**
  - Accept: url, events (array), secret (for HMAC signature)
  - Send POST to webhook URL with JSON payload
  - Sign payload: `HMAC-SHA256(payload, webhook_secret)`
  - Include header: `X-Envv-Signature: sha256=...`
  - Retry: 3 attempts with exponential backoff
  - Timeout: 10 seconds
  - Log: all webhook deliveries (success/failure)

#### 10.2 Slack Integration
- **Notifications:**
  - New team member joined
  - Secrets updated in production
  - Audit alert: unusual access pattern
- **Setup:** OAuth flow to install Slack app
- **Endpoint:** `POST /api/v1/integrations/slack/install`

#### 10.3 GitHub Integration
- **Feature:** Sync secrets to GitHub Actions secrets
- **Flow:**
  1. User connects GitHub repo to envv project
  2. CLI pushes to envv
  3. Backend optionally syncs to GitHub Secrets API
- **Permissions:** Require GitHub OAuth with `secrets:write` scope
- **Endpoint:** `POST /api/v1/projects/:id/integrations/github`

#### 10.4 CI/CD Integration
- **Feature:** CLI can run in CI/CD pipelines
- **Authentication:** Use API keys (not user passwords)
- **Example:**
  ```yaml
  # .github/workflows/deploy.yml
  - name: Load secrets
    run: |
      envv auth login-api-key ${{ secrets.ENVV_API_KEY }}
      envv pull --project=myapp --env=prod
      envv exec -- ./deploy.sh
  ```

---

### 11. Observability & Monitoring

#### 11.1 Metrics
Export Prometheus metrics:

```
envv_api_requests_total{method, endpoint, status}
envv_api_request_duration_seconds{method, endpoint}
envv_auth_attempts_total{status}
envv_secrets_operations_total{operation, project_id}
envv_projects_total
envv_users_total
envv_teams_total
```

#### 11.2 Health Checks
- **Endpoint:** `GET /api/v1/health`
- **Response:**
  ```json
  {
    "status": "healthy",
    "version": "1.0.0",
    "uptime": 86400,
    "checks": {
      "database": "healthy",
      "redis": "healthy",
      "storage": "healthy"
    }
  }
  ```

#### 11.3 Structured Logging
- Use JSON logs for all events
- Fields: timestamp, level, message, user_id, request_id, duration_ms
- Log levels: DEBUG, INFO, WARN, ERROR
- Integrate: Datadog, New Relic, or self-hosted ELK stack

#### 11.4 Error Tracking
- Integrate: Sentry or Rollbar
- Capture: unhandled exceptions, API errors, validation failures
- Include: user context, request details, stack trace

---

### 12. Performance & Scalability

#### 12.1 Caching Strategy
Use Redis for:
- User sessions (TTL: 24 hours)
- JWT blacklist (TTL: token expiry + 1 day)
- Team member lists (TTL: 5 minutes, invalidate on change)
- Project access lists (TTL: 5 minutes)
- Public keys (TTL: 1 hour)

Cache Keys Format:
```
user:session:{user_id}
team:members:{team_id}
project:access:{project_id}
user:public-key:{user_id}
jwt:blacklist:{token_hash}
```

#### 12.2 Database Optimization
- Indexes:
  ```sql
  CREATE INDEX idx_users_email ON users(email);
  CREATE INDEX idx_team_members_user ON team_members(user_id);
  CREATE INDEX idx_team_members_team ON team_members(team_id);
  CREATE INDEX idx_projects_team ON projects(team_id);
  CREATE INDEX idx_project_access_user ON project_access(user_id);
  CREATE INDEX idx_project_access_project ON project_access(project_id);
  CREATE INDEX idx_audit_log_timestamp ON audit_log(timestamp DESC);
  CREATE INDEX idx_audit_log_user ON audit_log(user_id);
  CREATE INDEX idx_audit_log_resource ON audit_log(resource_type, resource_id);
  ```

- Connection Pooling:
  - Min: 10 connections
  - Max: 50 connections
  - Idle timeout: 10 minutes

#### 12.3 Background Jobs
Use job queue (e.g., BullMQ, Sidekiq, Celery):

- **Email sending:** Registration, invites, notifications (low priority)
- **Webhook delivery:** Async, with retries (medium priority)
- **Audit log archival:** Move old logs to cold storage (low priority)
- **Metrics aggregation:** Daily stats (low priority)
- **Cleanup:** Expired tokens, old versions (low priority)

#### 12.4 Rate Limiting Implementation
- Use Redis with sliding window algorithm
- Key format: `ratelimit:{endpoint}:{user_id}:{window}`
- Algorithm: Token bucket or sliding window log
- Graceful degradation: If Redis down, allow requests (fail open)

---

## API Specifications

### Request/Response Format

#### Standard Request Headers
```http
Authorization: Bearer <jwt_token>
Content-Type: application/json
X-Request-ID: <uuid>  (optional, for tracing)
```

#### Standard Response Format

**Success Response:**
```json
{
  "success": true,
  "data": { ... },
  "meta": {
    "request_id": "uuid",
    "timestamp": "2025-01-15T10:30:00Z"
  }
}
```

**Error Response:**
```json
{
  "success": false,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid or expired token",
    "details": { ... },
    "request_id": "uuid"
  }
}
```

#### Standard Error Codes

| HTTP Status | Error Code | Description |
|------------|------------|-------------|
| 400 | `INVALID_REQUEST` | Malformed request |
| 401 | `UNAUTHORIZED` | Missing/invalid auth token |
| 403 | `FORBIDDEN` | User lacks permission |
| 404 | `NOT_FOUND` | Resource doesn't exist |
| 409 | `CONFLICT` | Resource already exists |
| 422 | `VALIDATION_ERROR` | Input validation failed |
| 429 | `RATE_LIMIT_EXCEEDED` | Too many requests |
| 500 | `INTERNAL_ERROR` | Server error |
| 503 | `SERVICE_UNAVAILABLE` | Temporary outage |

---

### Complete API Endpoint Reference

#### Authentication Endpoints

```
POST   /api/v1/auth/register
POST   /api/v1/auth/login
POST   /api/v1/auth/logout
POST   /api/v1/auth/refresh
GET    /api/v1/auth/me
POST   /api/v1/auth/forgot-password
POST   /api/v1/auth/reset-password
POST   /api/v1/auth/verify-email
POST   /api/v1/auth/resend-verification

# MFA (Medium-term)
POST   /api/v1/auth/mfa/enable
POST   /api/v1/auth/mfa/verify
POST   /api/v1/auth/mfa/disable
POST   /api/v1/auth/mfa/recovery-codes/generate

# OAuth (Medium-term)
GET    /api/v1/auth/oauth/:provider
GET    /api/v1/auth/oauth/:provider/callback

# API Keys (Medium-term)
POST   /api/v1/auth/api-keys
GET    /api/v1/auth/api-keys
DELETE /api/v1/auth/api-keys/:key_id
```

#### User Endpoints

```
GET    /api/v1/users/me
PATCH  /api/v1/users/me
PATCH  /api/v1/users/me/password
PATCH  /api/v1/users/me/public-key
DELETE /api/v1/users/me
GET    /api/v1/users/me/sessions
DELETE /api/v1/users/me/sessions/:session_id
```

#### Team Endpoints

```
POST   /api/v1/teams
GET    /api/v1/teams
GET    /api/v1/teams/:team_id
PATCH  /api/v1/teams/:team_id
DELETE /api/v1/teams/:team_id

# Team Members
GET    /api/v1/teams/:team_id/members
POST   /api/v1/teams/:team_id/invites
GET    /api/v1/teams/:team_id/invites
DELETE /api/v1/teams/:team_id/invites/:invite_id
POST   /api/v1/teams/invites/:token/accept
PATCH  /api/v1/teams/:team_id/members/:user_id
DELETE /api/v1/teams/:team_id/members/:user_id

# Team Public Keys (Critical)
GET    /api/v1/teams/:team_id/public-keys
```

#### Project Endpoints

```
POST   /api/v1/projects
GET    /api/v1/projects
GET    /api/v1/projects/:project_id
PATCH  /api/v1/projects/:project_id
DELETE /api/v1/projects/:project_id

# Project Access
POST   /api/v1/projects/:project_id/access
GET    /api/v1/projects/:project_id/members
PATCH  /api/v1/projects/:project_id/access/:user_id
DELETE /api/v1/projects/:project_id/access/:user_id

# Environments (Medium-term)
POST   /api/v1/projects/:project_id/environments
GET    /api/v1/projects/:project_id/environments
DELETE /api/v1/projects/:project_id/environments/:env_name
```

#### Secrets Endpoints (Core)

```
# Main Operations
POST   /api/v1/projects/:project_id/secrets
GET    /api/v1/projects/:project_id/secrets
GET    /api/v1/projects/:project_id/secrets/metadata

# Versioning
GET    /api/v1/projects/:project_id/secrets/versions
GET    /api/v1/projects/:project_id/secrets/versions/:version_id
POST   /api/v1/projects/:project_id/secrets/rollback

# With Environments (Medium-term)
POST   /api/v1/projects/:project_id/environments/:env/secrets
GET    /api/v1/projects/:project_id/environments/:env/secrets
GET    /api/v1/projects/:project_id/environments/:env/secrets/resolved
```

#### Audit Endpoints

```
GET    /api/v1/audit
GET    /api/v1/teams/:team_id/audit
GET    /api/v1/projects/:project_id/audit
GET    /api/v1/audit/export  (CSV download)
```

#### Integration Endpoints (Medium-term)

```
# Webhooks
POST   /api/v1/projects/:project_id/webhooks
GET    /api/v1/projects/:project_id/webhooks
PATCH  /api/v1/projects/:project_id/webhooks/:webhook_id
DELETE /api/v1/projects/:project_id/webhooks/:webhook_id
GET    /api/v1/projects/:project_id/webhooks/:webhook_id/deliveries

# Integrations
POST   /api/v1/integrations/slack/install
POST   /api/v1/integrations/github/connect
```

---

## Database Schema

### Core Tables

```sql
-- Users and Authentication
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    current_public_key_id UUID,
    email_verified BOOLEAN DEFAULT FALSE,
    email_verification_token VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE user_public_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    public_key VARCHAR(255) UNIQUE NOT NULL,
    key_type VARCHAR(50) DEFAULT 'age',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    deprecated_at TIMESTAMPTZ,
    CONSTRAINT valid_age_key CHECK (public_key ~ '^age1[a-z0-9]{58}$')
);

-- Add foreign key after table creation
ALTER TABLE users
    ADD CONSTRAINT fk_current_public_key
    FOREIGN KEY (current_public_key_id)
    REFERENCES user_public_keys(id);

-- Sessions
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    ip_address INET,
    user_agent TEXT,
    device_info JSONB,
    last_active_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- JWT Blacklist (for logout)
CREATE TABLE jwt_blacklist (
    token_hash VARCHAR(255) PRIMARY KEY,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Teams
CREATE TABLE teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    owner_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE team_members (
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('owner', 'admin', 'member')),
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (team_id, user_id)
);

CREATE TABLE team_invites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('admin', 'member')),
    token VARCHAR(255) UNIQUE NOT NULL,
    invited_by UUID NOT NULL REFERENCES users(id),
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Projects
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    created_by UUID NOT NULL REFERENCES users(id),
    needs_reencryption BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (team_id, name)
);

CREATE TABLE project_access (
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission VARCHAR(20) NOT NULL CHECK (permission IN ('read', 'write', 'admin')),
    granted_by UUID REFERENCES users(id),
    granted_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (project_id, user_id)
);

-- Secrets Storage
CREATE TABLE secrets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    environment VARCHAR(50) DEFAULT 'default',
    encrypted_data TEXT NOT NULL,
    format VARCHAR(20) DEFAULT 'env' CHECK (format IN ('json', 'yaml', 'env')),
    version INTEGER NOT NULL DEFAULT 1,
    size_bytes INTEGER NOT NULL,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (project_id, environment, version)
);

-- Secret Metadata (optional, for showing key names without decryption)
CREATE TABLE secret_metadata (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    secret_id UUID NOT NULL REFERENCES secrets(id) ON DELETE CASCADE,
    key_name VARCHAR(255) NOT NULL,
    description TEXT,
    tags JSONB,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Audit Log
CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    user_id UUID REFERENCES users(id),
    user_email VARCHAR(255),
    event_type VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id UUID,
    action VARCHAR(50) NOT NULL,
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    success BOOLEAN DEFAULT TRUE,
    error_message TEXT
);

-- Indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_email_verified ON users(email_verified);
CREATE INDEX idx_user_public_keys_user ON user_public_keys(user_id);
CREATE INDEX idx_user_public_keys_active ON user_public_keys(user_id, deprecated_at) WHERE deprecated_at IS NULL;

CREATE INDEX idx_user_sessions_user ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_expires ON user_sessions(expires_at);
CREATE INDEX idx_jwt_blacklist_expires ON jwt_blacklist(expires_at);

CREATE INDEX idx_teams_owner ON teams(owner_id);
CREATE INDEX idx_team_members_user ON team_members(user_id);
CREATE INDEX idx_team_members_team ON team_members(team_id);
CREATE INDEX idx_team_invites_email ON team_invites(email);
CREATE INDEX idx_team_invites_token ON team_invites(token);

CREATE INDEX idx_projects_team ON projects(team_id);
CREATE INDEX idx_projects_created_by ON projects(created_by);
CREATE INDEX idx_project_access_user ON project_access(user_id);
CREATE INDEX idx_project_access_project ON project_access(project_id);

CREATE INDEX idx_secrets_project ON secrets(project_id);
CREATE INDEX idx_secrets_project_env ON secrets(project_id, environment);
CREATE INDEX idx_secrets_latest ON secrets(project_id, environment, version DESC);

CREATE INDEX idx_audit_log_timestamp ON audit_log(timestamp DESC);
CREATE INDEX idx_audit_log_user ON audit_log(user_id);
CREATE INDEX idx_audit_log_resource ON audit_log(resource_type, resource_id);
CREATE INDEX idx_audit_log_event_type ON audit_log(event_type);
```

### Medium-term Tables

```sql
-- MFA
CREATE TABLE user_mfa (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    method VARCHAR(20) NOT NULL CHECK (method IN ('totp', 'sms')),
    encrypted_secret VARCHAR(255) NOT NULL,
    backup_codes TEXT[],
    enabled_at TIMESTAMPTZ DEFAULT NOW()
);

-- API Keys
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    key_hash VARCHAR(255) UNIQUE NOT NULL,
    key_prefix VARCHAR(20) NOT NULL,
    scopes TEXT[] NOT NULL,
    team_id UUID REFERENCES teams(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    last_used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

-- Environments
CREATE TABLE project_environments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    description TEXT,
    parent_environment_id UUID REFERENCES project_environments(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (project_id, name)
);

-- Webhooks
CREATE TABLE webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    events TEXT[] NOT NULL,
    secret VARCHAR(255) NOT NULL,
    active BOOLEAN DEFAULT TRUE,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    webhook_id UUID NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL,
    response_status INTEGER,
    response_body TEXT,
    duration_ms INTEGER,
    success BOOLEAN,
    attempt INTEGER DEFAULT 1,
    delivered_at TIMESTAMPTZ DEFAULT NOW()
);

-- Integrations
CREATE TABLE integrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    integration_type VARCHAR(50) NOT NULL CHECK (integration_type IN ('slack', 'github', 'gitlab')),
    config JSONB NOT NULL,
    encrypted_credentials TEXT,
    active BOOLEAN DEFAULT TRUE,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

---

## Security Requirements

### 1. Data Protection

#### 1.1 Encryption at Rest
- **Database:** Enable PostgreSQL encryption (AWS RDS encryption, etc.)
- **Backups:** Encrypted with AES-256
- **Logs:** PII redacted or encrypted
- **Credentials:** Never store plaintext passwords, API tokens, webhook secrets
  - Use: bcrypt (passwords), AES-256-GCM (tokens)

#### 1.2 Encryption in Transit
- **All API traffic:** TLS 1.3 (minimum: TLS 1.2)
- **Certificate:** Valid, auto-renewed (Let's Encrypt or managed cert)
- **HSTS:** Enforce HTTPS with Strict-Transport-Security header
- **Database connections:** Require SSL/TLS

#### 1.3 Secrets Handling (Backend)
Backend stores ONLY encrypted secrets. However, backend DOES handle:
- User passwords (hashed with bcrypt)
- JWT secrets (environment variable, rotated regularly)
- Database credentials (environment variable or secrets manager)
- Webhook secrets (encrypted in DB)
- OAuth tokens (encrypted in DB)

**Implementation:**
- Use envelope encryption for sensitive DB fields
- Store encryption key in AWS KMS, GCP Secret Manager, or HashiCorp Vault
- Never log sensitive data (passwords, tokens, secrets)

---

### 2. Authentication & Authorization

#### 2.1 JWT Security
- **Algorithm:** HS256 or RS256 (prefer RS256 for distributed systems)
- **Secret/Key:** Store in environment variable, minimum 256 bits
- **Expiry:** 24 hours (configurable)
- **Refresh:** Separate refresh token with 7-day expiry
- **Claims:** Include: user_id, email, iat, exp, jti (unique ID for revocation)
- **Validation:** Check expiry, signature, not blacklisted

#### 2.2 Password Requirements
- **Length:** Minimum 12 characters
- **Complexity:** At least 1 uppercase, 1 lowercase, 1 number, 1 special character
- **Hashing:** bcrypt with cost factor 12 (adjust based on server performance)
- **Salt:** Unique per user (bcrypt handles this automatically)
- **Storage:** Never store plaintext or reversible encryption

#### 2.3 Permission Checks
On EVERY protected endpoint:
1. Verify JWT valid and not expired
2. Verify user exists and not deleted
3. Check resource-specific permission:
   - Team member? (for team operations)
   - Project access? (for project/secrets operations)
   - Correct permission level? (read/write/admin)
4. Log authorization failures (potential breach attempts)

**Example Middleware Flow:**
```javascript
async function authorize(req, res, next) {
    // 1. Extract JWT from Authorization header
    const token = req.headers.authorization?.split(' ')[1];
    if (!token) return res.status(401).json({error: 'Missing token'});

    // 2. Verify and decode JWT
    const decoded = await verifyJWT(token);
    if (!decoded) return res.status(401).json({error: 'Invalid token'});

    // 3. Check not blacklisted
    const isBlacklisted = await isTokenBlacklisted(token);
    if (isBlacklisted) return res.status(401).json({error: 'Token revoked'});

    // 4. Load user
    const user = await db.users.findById(decoded.user_id);
    if (!user || user.deleted_at) return res.status(401).json({error: 'User not found'});

    // 5. Attach to request
    req.user = user;
    next();
}

async function requireProjectAccess(permission) {
    return async (req, res, next) => {
        const projectId = req.params.project_id;
        const access = await db.project_access.find({
            project_id: projectId,
            user_id: req.user.id
        });

        if (!access) return res.status(403).json({error: 'No access to project'});

        if (permission === 'write' && access.permission === 'read') {
            return res.status(403).json({error: 'Insufficient permissions'});
        }

        req.projectAccess = access;
        next();
    };
}
```

---

### 3. Vulnerability Protection

#### 3.1 SQL Injection
- **Prevention:** Use parameterized queries ALWAYS
- **ORM:** Prefer Sequelize, TypeORM, Prisma (handles escaping)
- **Never:** Concatenate user input into SQL strings

#### 3.2 XSS (Cross-Site Scripting)
- **Input:** Sanitize all user inputs (strip HTML tags)
- **Output:** Escape when rendering (though backend is API-only)
- **Headers:** Set Content-Type: application/json
- **CSP:** Content-Security-Policy: default-src 'self'

#### 3.3 CSRF (Cross-Site Request Forgery)
- **JWT-based API:** Not vulnerable if using Authorization header (not cookies)
- **If using cookies:** Implement CSRF tokens

#### 3.4 Timing Attacks
- **Password comparison:** Use constant-time comparison (crypto.timingSafeEqual)
- **Token validation:** Same principle
- **API key lookup:** Hash before comparison

#### 3.5 Dependency Vulnerabilities
- **Audit:** Run `npm audit` or equivalent regularly
- **Updates:** Keep dependencies up-to-date
- **Monitoring:** Use Snyk or Dependabot for alerts

---

### 4. Compliance & Privacy

#### 4.1 GDPR Compliance (if applicable)
- **Right to access:** Endpoint to export user's data
- **Right to deletion:** Soft delete + anonymization after 30 days
- **Data retention:** Clear policies on how long data is kept
- **Cookie consent:** If using cookies, implement consent banner

#### 4.2 SOC 2 / ISO 27001 Preparation
- **Audit logs:** Immutable, tamper-proof logs of all access
- **Access controls:** Documented permission model
- **Encryption:** At rest and in transit
- **Incident response:** Plan for security breaches
- **Backup & recovery:** Tested backup procedures

#### 4.3 Data Residency
- **Configuration:** Allow users to choose data region (US, EU, etc.)
- **Implementation:** Multi-region deployment with region-specific databases

---

## CLI Integration Points

The CLI needs to communicate with these backend endpoints:

### 1. CLI Authentication Flow

```bash
envv auth login
```

**CLI Actions:**
1. Prompt for email/password
2. POST to `/api/v1/auth/login`
3. Store JWT in `~/.envv/credentials.json` (0600 permissions)
4. Store user info: user_id, email, teams

**Security:**
- JWT stored locally, never sent to third parties
- File permissions: 0600 (read/write for owner only)
- Consider: Encrypted credential storage (OS keychain on macOS, Windows Credential Manager)

---

### 2. CLI Key Registration

```bash
envv auth register
```

**CLI Actions:**
1. Prompt for email, password, name
2. Generate age keypair locally: `age-keygen`
3. Store private key: `~/.config/sops/age/keys.txt`
4. POST to `/api/v1/auth/register` with public key
5. Store JWT

**Critical:** Private key NEVER leaves user's machine.

---

### 3. CLI Encryption Flow (Push Secrets)

```bash
envv push --project=myapp
```

**CLI Actions:**
1. Read local `.env` file (plaintext)
2. GET `/api/v1/projects/{id}/members` → returns team public keys
3. Generate `.sops.yaml` with all public keys:
   ```yaml
   creation_rules:
     - path_regex: \.env.*$
       age: >-
         age1alice...,
         age1bob...,
         age1charlie...
   ```
4. Encrypt with SOPS: `sops -e .env > .env.encrypted`
5. POST encrypted blob to `/api/v1/projects/{id}/secrets`
6. Delete plaintext `.env` from disk
7. Keep `.env.encrypted` for version control

**Backend Response:**
- Success: version_id, uploaded_at
- Error: validation errors, permission denied

---

### 4. CLI Decryption Flow (Pull Secrets)

```bash
envv pull --project=myapp
```

**CLI Actions:**
1. GET `/api/v1/projects/{id}/secrets`
2. Receive encrypted blob
3. Write to `.env.encrypted`
4. Decrypt with SOPS: `sops -d .env.encrypted > .env`
5. User's private key (from `~/.config/sops/age/keys.txt`) used for decryption

**Error Handling:**
- If user not in age recipients → "You don't have access to these secrets"
- If private key missing → "Run 'age-keygen' to generate keys"
- If backend denies access → "Insufficient permissions"

---

### 5. CLI Execution Flow

```bash
envv exec npm start
```

**CLI Actions:**
1. GET `/api/v1/projects/{id}/secrets` (or use cached `.env.encrypted`)
2. Decrypt to memory (never write to disk)
3. Load into environment variables
4. Execute command with env vars
5. Clear env vars after execution

**Implementation:**
Uses SOPS's built-in `exec-env`:
```bash
sops exec-env .env.encrypted 'npm start'
```

---

### 6. CLI Team Sync (After Member Added)

```bash
envv sync --project=myapp
```

**CLI Actions:**
1. GET `/api/v1/projects/{id}/members` → check for new members
2. If new members detected:
   - Decrypt existing secrets (with own key)
   - Re-encrypt with updated list of public keys
   - POST re-encrypted blob to backend
3. Log: "Re-encrypted secrets for 2 new team members"

**Automation:**
- Backend sets `needs_reencryption=true` on project
- CLI checks this flag on every pull
- Prompts: "This project needs re-encryption. Run 'envv sync'?"

---

### 7. CLI Commands Summary

| CLI Command | Backend Endpoint | Auth Required | Purpose |
|------------|------------------|---------------|---------|
| `envv auth login` | `POST /auth/login` | No | Authenticate |
| `envv auth register` | `POST /auth/register` | No | Create account |
| `envv auth logout` | `POST /auth/logout` | Yes | Invalidate token |
| `envv team create` | `POST /teams` | Yes | Create team |
| `envv team invite` | `POST /teams/:id/invites` | Yes | Invite member |
| `envv team list` | `GET /teams` | Yes | List teams |
| `envv project create` | `POST /projects` | Yes | Create project |
| `envv project list` | `GET /projects` | Yes | List projects |
| `envv push` | `POST /projects/:id/secrets` | Yes | Upload encrypted secrets |
| `envv pull` | `GET /projects/:id/secrets` | Yes | Download encrypted secrets |
| `envv list` | `GET /projects/:id/secrets/metadata` | Yes | List secret keys |
| `envv set KEY VAL` | Local + push | Yes | Update secret |
| `envv get KEY` | Local decrypt | Yes | Reveal secret value |
| `envv exec CMD` | Local decrypt + exec | Yes | Run command with secrets |
| `envv sync` | GET members + POST secrets | Yes | Re-encrypt for team |

---

## Testing Requirements

### 1. Unit Tests

**Coverage Target:** 80% minimum

#### 1.1 Authentication Tests
- User registration: valid input, duplicate email, invalid email
- Login: correct credentials, wrong password, non-existent user
- JWT generation: valid token, correct claims, proper expiry
- JWT validation: valid token, expired token, invalid signature, blacklisted
- Password hashing: bcrypt properly applied, different salts

#### 1.2 Authorization Tests
- Team member check: is member, not member, deleted team
- Project access check: has access, no access, insufficient permission
- Permission levels: read, write, admin enforcement
- Owner/admin checks: proper role hierarchy

#### 1.3 Business Logic Tests
- Team creation: owner assigned correctly
- Member invitation: token generated, expiry set
- Project creation: team association correct
- Secret storage: encryption detected, validation applied
- Audit logging: events recorded properly

### 2. Integration Tests

#### 2.1 API Endpoint Tests
Test each endpoint with:
- Valid input → success response
- Invalid input → validation error
- Missing auth → 401 Unauthorized
- Insufficient permission → 403 Forbidden
- Non-existent resource → 404 Not Found

**Example Test Flow:**
```javascript
describe('POST /api/v1/projects/:id/secrets', () => {
    it('should accept encrypted secrets with write permission', async () => {
        const user = await createTestUser();
        const team = await createTestTeam(user);
        const project = await createTestProject(team);
        await grantProjectAccess(project, user, 'write');

        const encryptedData = await encryptWithSOPS({TEST_KEY: 'value'});

        const response = await request(app)
            .post(`/api/v1/projects/${project.id}/secrets`)
            .set('Authorization', `Bearer ${user.token}`)
            .send({ encrypted_data: encryptedData })
            .expect(200);

        expect(response.body.data).toHaveProperty('version_id');
    });

    it('should reject secrets without write permission', async () => {
        // ... test with read-only user
    });
});
```

#### 2.2 End-to-End Tests
Simulate full user journeys:
- **New user signup:**
  1. Register → verify email → login → create team → invite member → create project → push secrets

- **Team collaboration:**
  1. User A pushes secrets
  2. User B pulls secrets
  3. User B updates secret
  4. User A pulls updated secrets

- **Permission changes:**
  1. Admin grants access to user
  2. User pulls secrets
  3. Admin revokes access
  4. User cannot pull secrets

### 3. Security Tests

#### 3.1 Authentication Security
- Brute force protection: verify rate limiting works
- SQL injection: test with malicious inputs
- JWT tampering: modified tokens rejected
- Token expiry: expired tokens rejected

#### 3.2 Authorization Security
- Horizontal privilege escalation: User A cannot access User B's projects
- Vertical privilege escalation: Member cannot perform admin actions
- IDOR (Insecure Direct Object Reference): UUIDs prevent guessing

#### 3.3 Data Leakage
- Error messages: don't reveal sensitive info
- Timing attacks: password comparison is constant-time
- Logs: no secrets in logs (passwords, tokens redacted)

### 4. Performance Tests

#### 4.1 Load Testing
- **Tool:** k6, Artillery, or JMeter
- **Scenarios:**
  - 100 concurrent users logging in
  - 1000 secrets pulls per minute
  - 500 secrets pushes per minute
- **Metrics:** Response time (p95, p99), error rate, throughput

#### 4.2 Database Performance
- Query performance: all queries < 100ms
- Index usage: EXPLAIN ANALYZE on critical queries
- Connection pooling: no connection exhaustion under load

#### 4.3 Caching Effectiveness
- Cache hit rate: > 80% for public keys, team members
- Cache invalidation: updates reflected within 5 seconds

### 5. Test Data Management

#### 5.1 Test Fixtures
Create reusable test data:
```javascript
const fixtures = {
    user: {
        email: 'test@example.com',
        password: 'SecureP@ssw0rd123',
        public_key: 'age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p'
    },
    team: {
        name: 'Test Team',
        description: 'A test team'
    },
    project: {
        name: 'Test Project',
        description: 'A test project'
    },
    encryptedSecret: '...' // Valid SOPS-encrypted blob
};
```

#### 5.2 Database Reset
- Before each test suite: reset database to clean state
- Use transactions: rollback after each test
- Separate test database: never test on production data

---

## Operational Requirements

### 1. Deployment

#### 1.1 Infrastructure
- **Hosting:** AWS, GCP, Azure, or DigitalOcean
- **Compute:** Kubernetes, ECS, or managed app platform (Heroku, Render, Railway)
- **Database:** Managed PostgreSQL (RDS, Cloud SQL, etc.)
- **Cache:** Managed Redis (ElastiCache, Cloud Memorystore)
- **Storage:** S3, GCS, or Azure Blob (for large encrypted blobs)

#### 1.2 Environment Variables
Required configuration:
```bash
# Application
NODE_ENV=production
PORT=8080
API_VERSION=v1

# Database
DATABASE_URL=postgresql://user:pass@host:5432/envv
DATABASE_POOL_MIN=10
DATABASE_POOL_MAX=50

# Redis
REDIS_URL=redis://host:6379
REDIS_PASSWORD=...

# Security
JWT_SECRET=<256-bit-secret>
JWT_EXPIRY=24h
BCRYPT_ROUNDS=12
ENCRYPTION_KEY=<for envelope encryption>

# Email (for invites, notifications)
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USER=apikey
SMTP_PASSWORD=...
FROM_EMAIL=noreply@envv.app

# Observability
SENTRY_DSN=https://...
LOG_LEVEL=info

# Features
ENABLE_MFA=false
ENABLE_OAUTH=false
RATE_LIMIT_ENABLED=true
```

#### 1.3 CI/CD Pipeline
```yaml
# Example: GitHub Actions
name: Deploy Backend
on:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Run tests
        run: npm test
      - name: Run security audit
        run: npm audit

  deploy:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to production
        run: |
          # Deploy to Kubernetes, ECS, etc.
```

### 2. Monitoring & Alerting

#### 2.1 Key Metrics to Monitor
- **Application:**
  - Request rate (requests/second)
  - Error rate (%)
  - Response time (p50, p95, p99)
  - Active users

- **Database:**
  - Query latency
  - Connection pool utilization
  - Slow queries (> 1 second)
  - Replication lag (if applicable)

- **Cache:**
  - Hit rate
  - Memory usage
  - Eviction rate

- **Business:**
  - New user signups
  - Teams created
  - Secrets pushed/pulled
  - Failed auth attempts

#### 2.2 Alerts
Set up alerts for:
- Error rate > 1% for 5 minutes
- Response time p99 > 2 seconds for 5 minutes
- Database connection pool > 90% for 5 minutes
- Failed auth attempts > 10 per IP per minute
- Disk space > 80%
- Memory usage > 90%

#### 2.3 Tools
- **Metrics:** Prometheus + Grafana, Datadog, New Relic
- **Logs:** ELK Stack (Elasticsearch, Logstash, Kibana), Papertrail, Datadog
- **Errors:** Sentry, Rollbar
- **Uptime:** Pingdom, UptimeRobot

### 3. Backup & Recovery

#### 3.1 Database Backups
- **Frequency:** Daily full backups, hourly incremental (if supported)
- **Retention:** 30 days of daily backups, 12 months of monthly backups
- **Storage:** Separate region/account (disaster recovery)
- **Encryption:** Encrypted at rest
- **Testing:** Monthly restore test to verify backups work

#### 3.2 Disaster Recovery Plan
- **RTO (Recovery Time Objective):** 4 hours
- **RPO (Recovery Point Objective):** 1 hour (max data loss)
- **Procedure:**
  1. Detect outage (monitoring alerts)
  2. Assess impact (database, application, infrastructure)
  3. Restore from backup (most recent valid backup)
  4. Verify data integrity
  5. Redirect traffic to restored instance
  6. Post-mortem: document what happened, how to prevent

### 4. Maintenance

#### 4.1 Regular Tasks
- **Weekly:** Review error logs, check for anomalies
- **Monthly:** Update dependencies, security patches
- **Quarterly:** Database optimization (VACUUM, reindex if needed)
- **Annually:** Security audit, penetration testing

#### 4.2 Database Migrations
- **Tool:** Knex.js, Sequelize migrations, or Prisma migrations
- **Process:**
  1. Write migration (up and down)
  2. Test on staging environment
  3. Schedule maintenance window (if downtime required)
  4. Run migration on production
  5. Verify data integrity
  6. Be prepared to rollback

**Example Migration:**
```javascript
// migrations/20250115_add_project_environments.js
exports.up = async function(knex) {
    await knex.schema.createTable('project_environments', (table) => {
        table.uuid('id').primary().defaultTo(knex.raw('gen_random_uuid()'));
        table.uuid('project_id').references('projects.id').onDelete('CASCADE');
        table.string('name', 50).notNullable();
        table.text('description');
        table.uuid('parent_environment_id').references('project_environments.id');
        table.timestamp('created_at').defaultTo(knex.fn.now());
        table.unique(['project_id', 'name']);
    });
};

exports.down = async function(knex) {
    await knex.schema.dropTable('project_environments');
};
```

---

## Implementation Priority & Timeline

### Phase 1: MVP (Weeks 1-4)
**Goal:** Core functionality for design partners

**Week 1:**
- [ ] Authentication endpoints (register, login, logout, refresh)
- [ ] User management (profile, public key storage)
- [ ] Database schema setup
- [ ] JWT middleware

**Week 2:**
- [ ] Team management (create, invite, members list)
- [ ] Team member roles (owner, admin, member)
- [ ] Team public keys endpoint (critical for CLI)

**Week 3:**
- [ ] Project management (create, list, access control)
- [ ] Project permissions (read, write, admin)
- [ ] Secrets endpoints (push, pull)
- [ ] Audit logging (basic)

**Week 4:**
- [ ] Testing (unit + integration)
- [ ] Documentation (API docs, setup guide)
- [ ] Deployment (staging environment)
- [ ] Design partner onboarding

### Phase 2: Production Readiness (Weeks 5-8)
**Goal:** Security, reliability, observability

**Week 5:**
- [ ] Rate limiting
- [ ] Security headers
- [ ] CORS configuration
- [ ] Input validation hardening

**Week 6:**
- [ ] Secret versioning
- [ ] Rollback functionality
- [ ] Secret metadata (key names without decryption)
- [ ] Re-encryption on team changes

**Week 7:**
- [ ] Monitoring (Prometheus metrics, logs)
- [ ] Alerting (PagerDuty, Slack)
- [ ] Health checks
- [ ] Database optimization (indexes, query tuning)

**Week 8:**
- [ ] Load testing
- [ ] Security testing (penetration test)
- [ ] Backup procedures
- [ ] Disaster recovery plan
- [ ] Production deployment

### Phase 3: Enhanced Features (Weeks 9-12)
**Goal:** Differentiation, scale, integrations

**Week 9:**
- [ ] MFA (TOTP)
- [ ] OAuth (GitHub, Google)
- [ ] API keys for machine access

**Week 10:**
- [ ] Environments (dev, staging, prod)
- [ ] Secret inheritance
- [ ] Secret tagging

**Week 11:**
- [ ] Webhooks
- [ ] Slack integration
- [ ] GitHub integration

**Week 12:**
- [ ] Web dashboard (basic UI for team management)
- [ ] Audit log viewer
- [ ] Usage analytics

---

## Acceptance Criteria

Before marking each phase complete, verify:

### MVP Acceptance
- [ ] CLI can register, login, logout successfully
- [ ] User can create team and invite members
- [ ] User can create project within team
- [ ] User can push encrypted secrets to backend
- [ ] User can pull encrypted secrets from backend
- [ ] User can execute commands with decrypted secrets
- [ ] Only authorized team members can access project secrets
- [ ] All secret operations are logged in audit log
- [ ] API documentation is complete and accurate
- [ ] At least 2 design partners onboarded and testing

### Production Acceptance
- [ ] All endpoints have rate limiting
- [ ] Security headers are properly configured
- [ ] SSL/TLS certificate is valid and auto-renewing
- [ ] Database backups are running daily
- [ ] Monitoring dashboards are set up
- [ ] Alerts are configured and tested
- [ ] Load test passes (100 concurrent users, < 500ms p99 response time)
- [ ] Security audit completed with no high-severity issues
- [ ] Disaster recovery procedure documented and tested

### Enhanced Features Acceptance
- [ ] MFA enrollment and verification working
- [ ] OAuth login working for at least 1 provider
- [ ] API keys can be generated and used for auth
- [ ] Environments (dev/staging/prod) can be created and managed
- [ ] Webhooks deliver successfully with retries
- [ ] At least 1 integration (Slack or GitHub) is functional
- [ ] Web dashboard allows team management without CLI

---

## Communication & Documentation

### For the Agent

Thank you for your work on this critical infrastructure. This document aims to provide you with:
- **Clear requirements:** What needs to be built
- **Context:** Why each feature matters
- **Guidance:** How to implement securely
- **Priorities:** What to build first

Please feel free to:
- Ask clarifying questions if any requirement is ambiguous
- Suggest improvements or alternative approaches
- Flag potential security concerns early
- Request code reviews at key milestones

### Documentation to Create

As you implement, please maintain:

1. **API Documentation:**
   - OpenAPI/Swagger spec for all endpoints
   - Request/response examples
   - Error codes and meanings
   - Rate limits per endpoint

2. **Architecture Docs:**
   - System architecture diagram
   - Database schema diagram (ERD)
   - Authentication flow diagrams
   - Encryption flow diagrams

3. **Runbook:**
   - How to deploy
   - How to rollback
   - How to handle common issues
   - Emergency procedures

4. **Developer Guide:**
   - Local development setup
   - Running tests
   - Database migrations
   - Code style and conventions

---

## Final Notes

This backend is the foundation for a zero-knowledge secrets management platform. The guiding principle is: **the backend facilitates collaboration but never has access to plaintext secrets**.

Key architectural decisions:
- **Client-side encryption:** All encryption/decryption happens on user's machine
- **Public key distribution:** Backend provides lists of authorized public keys
- **Access control:** Backend enforces who can read/write, but cannot decrypt
- **Audit trail:** Every operation is logged for compliance
- **Secure by default:** Rate limiting, validation, encryption at rest

Building this correctly creates:
- **User trust:** Users know their secrets are safe
- **Compliance:** Audit logs and access controls meet SOC 2 requirements
- **Scalability:** Encrypted blobs are small, caching is effective
- **Extensibility:** Webhooks and integrations add value without compromising security

Thank you for your attention to detail and commitment to security. Looking forward to seeing this system come to life.

---

**Document End**

If you need clarification on any section or would like me to expand on specific areas (e.g., a particular endpoint's implementation, a security concern, or a testing strategy), please don't hesitate to ask.
