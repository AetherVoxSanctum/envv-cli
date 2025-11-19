# envv CLI Final Integration Plan
## Synthesized from Backend API Spec + CLI Architecture Analysis

**Version:** 1.0.0
**Date:** 2025-11-19
**Status:** Ready for Implementation

---

## Executive Summary

After analyzing both the backend integration guide and the existing CLI codebase, this document provides the **definitive implementation plan** that combines:

✅ **Backend's exact API expectations**
✅ **Correct SOPS integration approach** (shell out, not library)
✅ **Proper terminology** (Organizations, not Teams)
✅ **Accurate command structure**
✅ **Best architectural patterns** from CLI analysis

---

## Critical Corrections from Backend Guide

### ❌ What I Got Wrong in My Proposal

| My Proposal | Backend Reality | Fix Required |
|------------|-----------------|--------------|
| `envv team create` | `envv org create` | Use **Organizations** not teams |
| `POST /auth/register` | `POST /auth/register-enhanced` | Use correct endpoint |
| Single `~/.envv/config.json` | Two files: credentials + project config | Split configuration |
| Use SOPS as library | Shell out to `sops` binary | Execute sops CLI |
| Manual SOPS metadata | Extract from encrypted JSON | Parse sops output |
| `--project` flag | `.envv/config.yaml` in directory | Project context per directory |

### ✅ What I Got Right

| Concept | Status | Notes |
|---------|--------|-------|
| Backend API client with JWT | ✅ Correct | Keep this architecture |
| Age key generation | ✅ Correct | Shell out to `age-keygen` |
| Push/pull encryption flow | ✅ Correct | Zero-knowledge model intact |
| Testing strategy | ✅ Correct | Unit + integration + E2E |
| Error handling patterns | ✅ Correct | API errors, auth errors, etc. |
| Code organization (`pkg/`) | ✅ Correct | Good structure |

---

## Corrected Architecture

### Command Structure (Backend's Expectation)

```bash
# Authentication
envv auth register               # Generate age key + register
envv auth login                  # Login with email/password
envv auth logout                 # Clear credentials
envv auth whoami                 # Show current user

# Organizations (NOT teams!)
envv org create <name>           # Create organization
envv org list                    # List user's organizations
envv org select <slug>           # Set default org
envv org members                 # List org members
envv org invite <email> <role>   # Invite to organization

# Projects
envv init                        # Initialize project (creates .envv/)
envv project create <name>       # Create new project
envv project list                # List projects
envv project select <slug>       # Set default project

# Secrets (CORE)
envv push [--env=prod]          # Encrypt + push to backend
envv pull [--env=prod]          # Pull + decrypt from backend
envv exec -- <command>          # Execute with secrets
envv list [--env=prod]          # List secret keys (no values)
envv versions [--env=prod]      # Version history
envv rollback --version=N       # Rollback to version
envv sync                       # Re-encrypt after team changes

# Access Control
envv access grant <email> <perm> # Grant project access
envv access revoke <email>       # Revoke access
envv access list                 # List members

# Key Management
envv key show                    # Show public key
envv key rotate                  # Rotate key (requires re-encrypt)

# Utility
envv status                      # Current project/env status
envv version                     # CLI version
```

---

## Configuration Files (Backend's Expectation)

### 1. User Credentials: `~/.envv/credentials.json`

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "expires_at": "2025-01-16T10:30:00Z"
}
```

**Permissions:** `chmod 600`
**Location:** User's home directory
**Purpose:** JWT token for API authentication

---

### 2. Project Config: `.envv/config.yaml`

```yaml
organization_id: "750e8400-e29b-41d4-a716-446655440000"
organization_name: "Acme Corp"
project_id: "850e8400-e29b-41d4-a716-446655440000"
project_name: "Backend API"
default_environment: "production"
```

**Location:** Project directory (git-ignored)
**Purpose:** Links directory to envv project
**Created by:** `envv init`

---

### 3. SOPS Config: `.sops.yaml` (Generated)

```yaml
creation_rules:
  - path_regex: \.env.*$
    age: >-
      age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p,
      age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8r
```

**Generated dynamically** from project members' public keys
**Not committed to git** (or can be if desired)
**Updated on:** push, sync

---

## Backend API Endpoints (Actual)

### Authentication

| Endpoint | Method | Purpose | Request | Response |
|----------|--------|---------|---------|----------|
| `/api/v1/auth/register-enhanced` | POST | Register + age key | email, password, name, **age_public_key** | JWT token, user |
| `/api/v1/auth/login` | POST | Login | email, password | JWT token, user |
| `/api/v1/auth/logout` | POST | Logout | - | success |
| `/api/v1/auth/refresh` | POST | Refresh token | token | new token |
| `/api/v1/auth/me` | GET | Current user | - | user info |

---

### Organizations (NOT Teams!)

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/v1/organizations` | GET | List user's orgs |
| `/api/v1/organizations` | POST | Create organization |
| `/api/v1/organizations/:id` | GET | Get org details |
| `/api/v1/organizations/:id/members` | GET | List org members |
| `/api/v1/organizations/:id/invites` | POST | Invite to org |
| `/api/v1/organizations/:id/members/keys` | GET | **Get all member public keys** |

---

### Projects

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/v1/organizations/:orgId/projects` | POST | Create project in org |
| `/api/v1/projects` | GET | List projects (filtered by org) |
| `/api/v1/projects/:id` | GET | Get project (includes `needs_reencryption` flag) |
| `/api/v1/projects/:id/members` | GET | **Get project members with public keys** |
| `/api/v1/projects/:id/access` | POST | Grant project access |
| `/api/v1/projects/:id/access/:userId` | DELETE | Revoke access |

---

### Secrets (CORE)

| Endpoint | Method | Purpose | Request | Response |
|----------|--------|---------|---------|----------|
| `/api/v1/projects/:id/secrets` | POST | Push encrypted secrets | **encrypted_data**, format, environment, **sops_metadata** | version_id, version |
| `/api/v1/projects/:id/secrets` | GET | Pull encrypted secrets | `?environment=prod` | **encrypted_data**, sops_metadata, version |
| `/api/v1/projects/:id/secrets/versions` | GET | List versions | `?environment=prod` | versions array |
| `/api/v1/projects/:id/secrets/rollback` | POST | Rollback to version | version_id | new version |

---

### Key Management

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/v1/users/me/public-key` | GET | Get user's public key |
| `/api/v1/users/me/public-key` | PATCH | Update/rotate public key |

---

## Corrected Implementation Plan

### Phase 1: Foundation (Week 1)

#### 1.1 Project Structure

```
envv-cli/
├── cmd/
│   ├── root.go              # Root command + global flags
│   ├── auth.go              # auth subcommands
│   ├── org.go               # org subcommands (NOT team!)
│   ├── project.go           # project subcommands
│   ├── secrets.go           # push/pull/sync/exec (CORE)
│   ├── access.go            # access control
│   ├── key.go               # key management
│   └── version.go           # version command
│
├── pkg/
│   ├── api/
│   │   ├── client.go        # HTTP client with JWT auth
│   │   ├── auth.go          # Auth endpoints
│   │   ├── organizations.go # Org endpoints (NOT teams!)
│   │   ├── projects.go      # Project endpoints
│   │   ├── secrets.go       # Secrets endpoints
│   │   └── models.go        # Request/response structs
│   │
│   ├── crypto/
│   │   ├── age.go           # Shell out to age-keygen
│   │   ├── sops.go          # Shell out to sops binary
│   │   └── metadata.go      # Extract SOPS metadata from JSON
│   │
│   ├── config/
│   │   ├── credentials.go   # ~/.envv/credentials.json
│   │   ├── project.go       # .envv/config.yaml
│   │   └── manager.go       # Config management
│   │
│   └── ui/
│       ├── prompt.go        # User prompts
│       ├── output.go        # Formatted output
│       └── spinner.go       # Progress indicators
│
├── internal/
│   └── version/
│       └── version.go       # Version constant
│
├── go.mod
├── go.sum
├── main.go
└── README.md
```

---

#### 1.2 API Client (Corrected)

**File:** `pkg/api/client.go`

```go
package api

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

const (
    DefaultBaseURL = "https://api.envv.app"
    APIVersion     = "v1"
)

type Client struct {
    BaseURL    string
    HTTPClient *http.Client
    token      string
}

func NewClient(baseURL string) *Client {
    if baseURL == "" {
        baseURL = DefaultBaseURL
    }

    return &Client{
        BaseURL: baseURL,
        HTTPClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

func (c *Client) SetToken(token string) {
    c.token = token
}

func (c *Client) Do(ctx context.Context, method, path string, body, result interface{}) error {
    var reqBody io.Reader
    if body != nil {
        jsonData, err := json.Marshal(body)
        if err != nil {
            return fmt.Errorf("failed to marshal request: %w", err)
        }
        reqBody = bytes.NewReader(jsonData)
    }

    url := c.BaseURL + path
    req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }

    req.Header.Set("Content-Type", "application/json")
    if c.token != "" {
        req.Header.Set("Authorization", "Bearer "+c.token)
    }

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    // Handle error responses
    if resp.StatusCode >= 400 {
        return c.parseError(resp)
    }

    // Decode success response
    if result != nil {
        if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
            return fmt.Errorf("failed to decode response: %w", err)
        }
    }

    return nil
}

func (c *Client) parseError(resp *http.Response) error {
    var errResp struct {
        Error   string                 `json:"error"`
        Code    string                 `json:"code"`
        Details map[string]interface{} `json:"details"`
    }

    json.NewDecoder(resp.Body).Decode(&errResp)

    return &APIError{
        StatusCode: resp.StatusCode,
        Code:       errResp.Code,
        Message:    errResp.Error,
        Details:    errResp.Details,
    }
}

type APIError struct {
    StatusCode int
    Code       string
    Message    string
    Details    map[string]interface{}
}

func (e *APIError) Error() string {
    return fmt.Sprintf("API error %d (%s): %s", e.StatusCode, e.Code, e.Message)
}

func (e *APIError) IsUnauthorized() bool {
    return e.StatusCode == 401
}

func (e *APIError) IsForbidden() bool {
    return e.StatusCode == 403
}

func (e *APIError) IsNotFound() bool {
    return e.StatusCode == 404
}
```

---

#### 1.3 Authentication API (Corrected Endpoint)

**File:** `pkg/api/auth.go`

```go
package api

import "context"

// RegisterEnhancedRequest matches backend expectation
type RegisterEnhancedRequest struct {
    Email        string `json:"email"`
    Password     string `json:"password"`
    Name         string `json:"name"`
    AgePublicKey string `json:"age_public_key"` // CRITICAL: age key required!
}

type AuthResponse struct {
    AccessToken string `json:"access_token"`
    User        struct {
        ID           string `json:"id"`
        Email        string `json:"email"`
        Name         string `json:"name"`
        AgePublicKey string `json:"age_public_key"`
    } `json:"user"`
    ExpiresAt string `json:"expires_at"`
}

// RegisterEnhanced - Use the correct endpoint!
func (c *Client) RegisterEnhanced(ctx context.Context, req RegisterEnhancedRequest) (*AuthResponse, error) {
    var resp AuthResponse
    err := c.Do(ctx, "POST", "/api/v1/auth/register-enhanced", req, &resp)
    return &resp, err
}

type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

func (c *Client) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
    var resp AuthResponse
    err := c.Do(ctx, "POST", "/api/v1/auth/login", req, &resp)
    return &resp, err
}

func (c *Client) Logout(ctx context.Context) error {
    return c.Do(ctx, "POST", "/api/v1/auth/logout", nil, nil)
}

type RefreshRequest struct {
    Token string `json:"token"`
}

func (c *Client) RefreshToken(ctx context.Context, req RefreshRequest) (*AuthResponse, error) {
    var resp AuthResponse
    err := c.Do(ctx, "POST", "/api/v1/auth/refresh", req, &resp)
    return &resp, err
}

func (c *Client) GetCurrentUser(ctx context.Context) (*AuthResponse, error) {
    var resp AuthResponse
    err := c.Do(ctx, "GET", "/api/v1/auth/me", nil, &resp)
    return &resp, err
}
```

---

#### 1.4 Organizations API (Not Teams!)

**File:** `pkg/api/organizations.go`

```go
package api

import (
    "context"
    "fmt"
)

type Organization struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Slug        string `json:"slug"`
    Role        string `json:"role"` // owner, admin, member
    MemberCount int    `json:"member_count"`
    CreatedAt   string `json:"created_at"`
}

type CreateOrgRequest struct {
    Name        string `json:"name"`
    Description string `json:"description,omitempty"`
}

func (c *Client) CreateOrganization(ctx context.Context, req CreateOrgRequest) (*Organization, error) {
    var org Organization
    err := c.Do(ctx, "POST", "/api/v1/organizations", req, &org)
    return &org, err
}

func (c *Client) ListOrganizations(ctx context.Context) ([]Organization, error) {
    var orgs []Organization
    err := c.Do(ctx, "GET", "/api/v1/organizations", nil, &orgs)
    return orgs, err
}

func (c *Client) GetOrganization(ctx context.Context, orgID string) (*Organization, error) {
    var org Organization
    path := fmt.Sprintf("/api/v1/organizations/%s", orgID)
    err := c.Do(ctx, "GET", path, nil, &org)
    return &org, err
}

type OrgMember struct {
    UserID       string `json:"user_id"`
    Email        string `json:"email"`
    Name         string `json:"name"`
    AgePublicKey string `json:"age_public_key"` // CRITICAL for encryption!
    Role         string `json:"role"`
}

// GetOrganizationMemberKeys - Get all member public keys for encryption
func (c *Client) GetOrganizationMemberKeys(ctx context.Context, orgID string) ([]OrgMember, error) {
    var members []OrgMember
    path := fmt.Sprintf("/api/v1/organizations/%s/members/keys", orgID)
    err := c.Do(ctx, "GET", path, nil, &members)
    return members, err
}
```

---

#### 1.5 SOPS Integration (Shell Out, Not Library!)

**File:** `pkg/crypto/sops.go`

```go
package crypto

import (
    "encoding/json"
    "fmt"
    "os"
    "os/exec"
)

// EncryptWithSOPS shells out to sops binary
func EncryptWithSOPS(inputPath, outputPath string) error {
    cmd := exec.Command("sops", "-e", inputPath)

    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("sops encryption failed: %w\nOutput: %s", err, string(output))
    }

    if err := os.WriteFile(outputPath, output, 0600); err != nil {
        return fmt.Errorf("failed to write encrypted file: %w", err)
    }

    return nil
}

// DecryptWithSOPS shells out to sops binary
func DecryptWithSOPS(inputPath, outputPath string) error {
    cmd := exec.Command("sops", "-d", inputPath)

    output, err := cmd.CombinedOutput()
    if err != nil {
        // Check if it's a permission error
        if contains(string(output), "no valid decryption key") {
            return fmt.Errorf("you don't have permission to decrypt these secrets")
        }
        return fmt.Errorf("sops decryption failed: %w\nOutput: %s", err, string(output))
    }

    if err := os.WriteFile(outputPath, output, 0600); err != nil {
        return fmt.Errorf("failed to write decrypted file: %w", err)
    }

    return nil
}

// DecryptToMemory decrypts without writing to disk
func DecryptToMemory(inputPath string) ([]byte, error) {
    cmd := exec.Command("sops", "-d", inputPath)
    output, err := cmd.CombinedOutput()
    if err != nil {
        return nil, fmt.Errorf("sops decryption failed: %w", err)
    }
    return output, nil
}

// GenerateSOPSConfig creates .sops.yaml with team public keys
func GenerateSOPSConfig(publicKeys []string) string {
    config := "creation_rules:\n"
    config += "  - path_regex: \\.env.*$\n"
    config += "    age: >-\n"

    for i, key := range publicKeys {
        if i > 0 {
            config += ","
        }
        config += "\n      " + key
    }
    config += "\n"

    return config
}

// ExtractSOPSMetadata parses encrypted JSON and extracts sops metadata
func ExtractSOPSMetadata(encryptedData []byte) (map[string]interface{}, error) {
    var data map[string]interface{}
    if err := json.Unmarshal(encryptedData, &data); err != nil {
        return nil, fmt.Errorf("failed to parse encrypted data: %w", err)
    }

    sopsData, ok := data["sops"]
    if !ok {
        return nil, fmt.Errorf("no sops metadata found in encrypted file")
    }

    metadata, ok := sopsData.(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("invalid sops metadata format")
    }

    return metadata, nil
}

func contains(s, substr string) bool {
    return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) &&
        (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
        len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
    for i := 0; i <= len(s)-len(substr); i++ {
        if s[i:i+len(substr)] == substr {
            return true
        }
    }
    return false
}
```

---

#### 1.6 Age Key Generation (Shell Out)

**File:** `pkg/crypto/age.go`

```go
package crypto

import (
    "bufio"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
)

const (
    DefaultKeyPath = "~/.config/sops/age/keys.txt"
)

type AgeKeypair struct {
    PublicKey  string
    PrivateKey string
}

// GenerateAgeKeypair shells out to age-keygen
func GenerateAgeKeypair() (*AgeKeypair, error) {
    // Check if age-keygen is installed
    if _, err := exec.LookPath("age-keygen"); err != nil {
        return nil, fmt.Errorf("age-keygen not found. Install from https://github.com/FiloSottile/age")
    }

    cmd := exec.Command("age-keygen")
    output, err := cmd.CombinedOutput()
    if err != nil {
        return nil, fmt.Errorf("age-keygen failed: %w\nOutput: %s", err, string(output))
    }

    return parseAgeKeygenOutput(string(output))
}

func parseAgeKeygenOutput(output string) (*AgeKeypair, error) {
    var publicKey, privateKey string

    lines := strings.Split(output, "\n")
    for _, line := range lines {
        if strings.HasPrefix(line, "# Public key: ") {
            publicKey = strings.TrimPrefix(line, "# Public key: ")
            publicKey = strings.TrimSpace(publicKey)
        } else if strings.HasPrefix(line, "AGE-SECRET-KEY-") {
            privateKey = strings.TrimSpace(line)
        }
    }

    if publicKey == "" || privateKey == "" {
        return nil, fmt.Errorf("failed to parse age-keygen output")
    }

    return &AgeKeypair{
        PublicKey:  publicKey,
        PrivateKey: privateKey,
    }, nil
}

func SavePrivateKey(keypair *AgeKeypair, path string) error {
    // Expand ~ to home directory
    if strings.HasPrefix(path, "~") {
        home, _ := os.UserHomeDir()
        path = filepath.Join(home, path[1:])
    }

    // Create directory
    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0700); err != nil {
        return fmt.Errorf("failed to create directory: %w", err)
    }

    // Write key file
    content := fmt.Sprintf("# created: %s\n# public key: %s\n%s\n",
        time.Now().Format(time.RFC3339),
        keypair.PublicKey,
        keypair.PrivateKey)

    if err := os.WriteFile(path, []byte(content), 0600); err != nil {
        return fmt.Errorf("failed to write key file: %w", err)
    }

    return nil
}

func GetPublicKeyFromFile(path string) (string, error) {
    // Expand ~
    if strings.HasPrefix(path, "~") {
        home, _ := os.UserHomeDir()
        path = filepath.Join(home, path[1:])
    }

    file, err := os.Open(path)
    if err != nil {
        return "", fmt.Errorf("age key file not found at %s", path)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        if strings.HasPrefix(line, "# public key: ") {
            return strings.TrimPrefix(line, "# public key: "), nil
        }
    }

    return "", fmt.Errorf("no public key found in %s", path)
}
```

---

#### 1.7 Configuration Management (Two Files!)

**File:** `pkg/config/credentials.go`

```go
package config

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
)

const CredentialsFile = "~/.envv/credentials.json"

type Credentials struct {
    AccessToken string `json:"access_token"`
    UserID      string `json:"user_id"`
    Email       string `json:"email"`
    ExpiresAt   string `json:"expires_at"`
}

func LoadCredentials() (*Credentials, error) {
    path := expandPath(CredentialsFile)

    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, fmt.Errorf("not logged in. Run 'envv auth login'")
        }
        return nil, err
    }

    var creds Credentials
    if err := json.Unmarshal(data, &creds); err != nil {
        return nil, fmt.Errorf("corrupted credentials file: %w", err)
    }

    return &creds, nil
}

func SaveCredentials(creds *Credentials) error {
    path := expandPath(CredentialsFile)

    // Create directory
    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0700); err != nil {
        return err
    }

    // Write credentials
    data, err := json.MarshalIndent(creds, "", "  ")
    if err != nil {
        return err
    }

    return os.WriteFile(path, data, 0600)
}

func ClearCredentials() error {
    path := expandPath(CredentialsFile)
    if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
        return err
    }
    return nil
}

func expandPath(path string) string {
    if strings.HasPrefix(path, "~") {
        home, _ := os.UserHomeDir()
        return filepath.Join(home, path[1:])
    }
    return path
}
```

**File:** `pkg/config/project.go`

```go
package config

import (
    "fmt"
    "os"

    "gopkg.in/yaml.v3"
)

const ProjectConfigFile = ".envv/config.yaml"

type ProjectConfig struct {
    OrganizationID   string `yaml:"organization_id"`
    OrganizationName string `yaml:"organization_name"`
    ProjectID        string `yaml:"project_id"`
    ProjectName      string `yaml:"project_name"`
    DefaultEnv       string `yaml:"default_environment"`
}

func LoadProjectConfig() (*ProjectConfig, error) {
    data, err := os.ReadFile(ProjectConfigFile)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, fmt.Errorf("not in envv project. Run 'envv init'")
        }
        return nil, err
    }

    var cfg ProjectConfig
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("invalid project config: %w", err)
    }

    return &cfg, nil
}

func SaveProjectConfig(cfg *ProjectConfig) error {
    // Create .envv directory
    if err := os.MkdirAll(".envv", 0755); err != nil {
        return err
    }

    data, err := yaml.Marshal(cfg)
    if err != nil {
        return err
    }

    return os.WriteFile(ProjectConfigFile, data, 0644)
}
```

---

### Phase 2: Core Commands (Week 1-2)

#### 2.1 Auth Register (Corrected!)

**File:** `cmd/auth.go`

```go
package cmd

import (
    "context"
    "fmt"
    "syscall"

    "github.com/spf13/cobra"
    "golang.org/x/term"

    "envv-cli/pkg/api"
    "envv-cli/pkg/config"
    "envv-cli/pkg/crypto"
)

var authRegisterCmd = &cobra.Command{
    Use:   "register",
    Short: "Register new account with age keypair",
    RunE:  runAuthRegister,
}

func runAuthRegister(cmd *cobra.Command, args []string) error {
    ctx := context.Background()

    // 1. Generate age keypair
    fmt.Println("🔑 Generating age keypair...")
    keypair, err := crypto.GenerateAgeKeypair()
    if err != nil {
        return fmt.Errorf("failed to generate age key: %w", err)
    }
    fmt.Printf("✓ Generated keypair\n")
    fmt.Printf("  Public key: %s\n", keypair.PublicKey)

    // 2. Save private key
    if err := crypto.SavePrivateKey(keypair, crypto.DefaultKeyPath); err != nil {
        return fmt.Errorf("failed to save private key: %w", err)
    }
    fmt.Printf("✓ Saved private key to %s\n\n", crypto.DefaultKeyPath)

    // 3. Prompt for user details
    fmt.Print("Name: ")
    var name string
    fmt.Scanln(&name)

    fmt.Print("Email: ")
    var email string
    fmt.Scanln(&email)

    fmt.Print("Password: ")
    passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
    fmt.Println()
    if err != nil {
        return err
    }
    password := string(passwordBytes)

    // 4. Validate password
    if len(password) < 12 {
        return fmt.Errorf("password must be at least 12 characters")
    }

    // 5. Register with backend
    fmt.Println("\n📡 Registering with envv backend...")
    client := api.NewClient("")

    resp, err := client.RegisterEnhanced(ctx, api.RegisterEnhancedRequest{
        Email:        email,
        Password:     password,
        Name:         name,
        AgePublicKey: keypair.PublicKey, // CRITICAL!
    })
    if err != nil {
        return fmt.Errorf("registration failed: %w", err)
    }

    // 6. Save credentials
    creds := &config.Credentials{
        AccessToken: resp.AccessToken,
        UserID:      resp.User.ID,
        Email:       resp.User.Email,
        ExpiresAt:   resp.ExpiresAt,
    }

    if err := config.SaveCredentials(creds); err != nil {
        return fmt.Errorf("failed to save credentials: %w", err)
    }

    // 7. Success!
    fmt.Println("\n✅ Registration successful!")
    fmt.Printf("   Logged in as: %s\n", resp.User.Email)
    fmt.Printf("   User ID: %s\n", resp.User.ID)
    fmt.Printf("   Access token valid until: %s\n", resp.ExpiresAt)

    return nil
}
```

---

#### 2.2 Push Command (Complete Implementation)

**File:** `cmd/secrets.go`

```go
package cmd

import (
    "context"
    "encoding/json"
    "fmt"
    "os"

    "github.com/spf13/cobra"

    "envv-cli/pkg/api"
    "envv-cli/pkg/config"
    "envv-cli/pkg/crypto"
)

var pushCmd = &cobra.Command{
    Use:   "push",
    Short: "Encrypt and push secrets to backend",
    RunE:  runPush,
}

func init() {
    pushCmd.Flags().String("environment", "production", "Environment name")
    rootCmd.AddCommand(pushCmd)
}

func runPush(cmd *cobra.Command, args []string) error {
    ctx := context.Background()
    environment, _ := cmd.Flags().GetString("environment")

    // 1. Load project config
    projectCfg, err := config.LoadProjectConfig()
    if err != nil {
        return err
    }

    // 2. Check .env exists
    if _, err := os.Stat(".env"); os.IsNotExist(err) {
        return fmt.Errorf("no .env file found. Create one first")
    }

    // 3. Get authenticated client
    creds, err := config.LoadCredentials()
    if err != nil {
        return err
    }

    client := api.NewClient("")
    client.SetToken(creds.AccessToken)

    // 4. Fetch project members with public keys
    fmt.Println("🔍 Fetching project members...")
    members, err := client.GetProjectMembers(ctx, projectCfg.ProjectID)
    if err != nil {
        return fmt.Errorf("failed to get project members: %w", err)
    }

    if len(members.Members) == 0 {
        return fmt.Errorf("project has no members with public keys")
    }

    fmt.Printf("✓ Found %d team members\n", len(members.Members))

    // 5. Extract public keys
    publicKeys := make([]string, 0, len(members.Members))
    for _, member := range members.Members {
        if member.AgePublicKey != "" {
            publicKeys = append(publicKeys, member.AgePublicKey)
        }
    }

    // 6. Generate .sops.yaml
    sopsConfig := crypto.GenerateSOPSConfig(publicKeys)
    if err := os.WriteFile(".sops.yaml", []byte(sopsConfig), 0644); err != nil {
        return fmt.Errorf("failed to create .sops.yaml: %w", err)
    }

    // 7. Encrypt with SOPS
    fmt.Println("🔐 Encrypting secrets...")
    if err := crypto.EncryptWithSOPS(".env", ".env.encrypted"); err != nil {
        return err
    }

    // 8. Read encrypted data
    encryptedData, err := os.ReadFile(".env.encrypted")
    if err != nil {
        return err
    }

    // 9. Extract SOPS metadata
    sopsMetadata, err := crypto.ExtractSOPSMetadata(encryptedData)
    if err != nil {
        return fmt.Errorf("failed to extract SOPS metadata: %w", err)
    }

    // 10. Push to backend
    fmt.Println("📤 Pushing to backend...")
    resp, err := client.PushSecrets(ctx, projectCfg.ProjectID, api.PushSecretsRequest{
        EncryptedData: string(encryptedData),
        Format:        "env",
        Environment:   environment,
        SOPSMetadata:  sopsMetadata, // CRITICAL!
    })
    if err != nil {
        return fmt.Errorf("push failed: %w", err)
    }

    // 11. Success!
    fmt.Println("\n✅ Secrets pushed successfully!")
    fmt.Printf("   Version: %d\n", resp.Version)
    fmt.Printf("   Version ID: %s\n", resp.VersionID)
    fmt.Printf("   Environment: %s\n", environment)
    fmt.Printf("   Size: %.1f KB\n", float64(resp.SizeBytes)/1024)
    fmt.Printf("   Encrypted for %d team members\n", len(publicKeys))

    return nil
}
```

---

### Phase 3: Complete Command Set (Week 2-3)

**All commands implemented following backend spec:**

1. ✅ `envv auth register` - age key + register-enhanced
2. ✅ `envv auth login` - JWT auth
3. ✅ `envv org create` - Organizations (not teams!)
4. ✅ `envv init` - Create .envv/config.yaml
5. ✅ `envv push` - Encrypt + upload with SOPS metadata
6. ✅ `envv pull` - Download + decrypt
7. ✅ `envv sync` - Re-encrypt after team changes
8. ✅ `envv exec` - Run with secrets
9. ✅ `envv versions` - Version history
10. ✅ `envv rollback` - Rollback to version

---

## Critical Differences Summary

### Backend Expects vs. My Original Proposal

| Feature | My Proposal | Backend Reality | Status |
|---------|-------------|-----------------|--------|
| **Team/Org Model** | Teams | **Organizations** | ✅ Fixed |
| **Register Endpoint** | `/auth/register` | `/auth/register-enhanced` | ✅ Fixed |
| **Age Key** | Optional | **Required in registration** | ✅ Fixed |
| **Config Files** | Single file | **Two files** (creds + project) | ✅ Fixed |
| **SOPS Integration** | Use as library | **Shell out to binary** | ✅ Fixed |
| **SOPS Metadata** | Optional | **Required in push** | ✅ Fixed |
| **Project Context** | Global config | **Per-directory .envv/** | ✅ Fixed |
| **Environment Support** | Manual | **Built into backend** | ✅ Fixed |

---

## Implementation Checklist

### Week 1: Core Foundation
- [ ] Create Go project structure
- [ ] Implement HTTP client with JWT auth
- [ ] Implement credentials management (`~/.envv/credentials.json`)
- [ ] Implement project config (`.envv/config.yaml`)
- [ ] Shell out to `age-keygen` for key generation
- [ ] Shell out to `sops` for encryption/decryption
- [ ] Implement SOPS metadata extraction
- [ ] **Auth commands:** register, login, logout, whoami
- [ ] Unit tests for all packages

### Week 2: Organizations & Projects
- [ ] **Org commands:** create, list, select, members, invite
- [ ] **Project commands:** create, list, select
- [ ] **Init command:** Create .envv/config.yaml
- [ ] Integration tests with mock backend

### Week 3: Secrets (CORE)
- [ ] **Push command:** Full implementation
- [ ] **Pull command:** Full implementation
- [ ] **Exec command:** Run with decrypted secrets
- [ ] **List command:** Show secret keys
- [ ] **Versions command:** Version history
- [ ] **Rollback command:** Rollback to version
- [ ] **Sync command:** Re-encrypt after team changes
- [ ] E2E tests

### Week 4: Polish & Release
- [ ] Error handling improvements
- [ ] Progress indicators
- [ ] User documentation
- [ ] CLI help text
- [ ] Cross-platform builds
- [ ] Release automation

---

## Success Criteria

✅ **User can:**
1. Register with age keypair auto-generation
2. Create organization
3. Initialize project in directory
4. Push secrets encrypted for whole team
5. Pull and decrypt secrets
6. Execute commands with secrets
7. View version history and rollback
8. Re-encrypt when team changes

✅ **Backend receives:**
1. Correct endpoint calls (`/auth/register-enhanced`)
2. Age public key in registration
3. SOPS metadata in push requests
4. Proper JWT authentication

✅ **Security maintained:**
1. Private keys never leave user's machine
2. Backend never sees plaintext secrets
3. Credentials stored securely (chmod 600)
4. Multi-party encryption working

---

## Next Steps

1. **Review this synthesized plan** - Confirm all backend expectations met
2. **Set up Go project** - Initialize with correct structure
3. **Implement Phase 1** - Foundation (Week 1)
4. **Daily testing** - Test against backend after each feature
5. **Iterate based on feedback**

---

## Questions for Backend Team

1. ✅ Is `register-enhanced` the correct endpoint? (YES)
2. ✅ Is age public key required in registration? (YES)
3. ✅ Should SOPS metadata be extracted from encrypted JSON? (YES)
4. ✅ Are organizations the correct model (not teams)? (YES)
5. ✅ Should we shell out to sops binary? (YES)

**All questions answered by backend integration guide!**

---

**This is the FINAL, CORRECT implementation plan.**

Ready to proceed with coding? 🚀
