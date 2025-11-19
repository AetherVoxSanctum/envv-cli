# envv CLI Production Readiness Proposal
## Making the CLI Backend-Ready for Team Secrets Management

**Status:** Proposal for Review
**Target:** Production-ready CLI that integrates with envv backend
**Current State:** SOPS fork with demo wrapper scripts (70% complete)
**Goal:** Native CLI with backend integration (100% production-ready)

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Current State Analysis](#current-state-analysis)
3. [Gap Analysis](#gap-analysis)
4. [Proposed Architecture](#proposed-architecture)
5. [Implementation Plan](#implementation-plan)
6. [New Command Structure](#new-command-structure)
7. [Code Organization](#code-organization)
8. [Integration with Existing SOPS](#integration-with-existing-sops)
9. [Configuration Management](#configuration-management)
10. [Testing Strategy](#testing-strategy)
11. [Security Considerations](#security-considerations)
12. [Documentation Plan](#documentation-plan)
13. [Release Strategy](#release-strategy)
14. [Timeline & Milestones](#timeline--milestones)
15. [Success Criteria](#success-criteria)

---

## Executive Summary

### Current Situation

The envv-cli repository is a fork of mozilla/sops with:
- ✅ Core SOPS encryption/decryption working (age, PGP, KMS)
- ✅ Demo wrapper scripts in `demo/` directory
- ✅ Local-only secret management
- ❌ No backend API integration
- ❌ No native team/project commands
- ❌ No authentication system
- ❌ No configuration management for SaaS

### Proposed Transformation

Transform envv-cli into a **production-ready SaaS CLI** that:
1. **Integrates with backend API** for team collaboration
2. **Maintains zero-knowledge encryption** (client-side only)
3. **Provides native commands** for auth, teams, projects
4. **Manages configuration** (API endpoint, sessions, context)
5. **Handles multi-user encryption** (team public key distribution)
6. **Includes comprehensive testing** and documentation

### Key Principle

**Backend Integration WITHOUT Compromising Security:**
- Secrets ALWAYS encrypted client-side
- Backend NEVER sees plaintext
- Backend provides: authentication, access control, public key distribution, encrypted blob storage
- CLI handles: encryption, decryption, key management

---

## Current State Analysis

### Existing Codebase Structure

```
envv-cli/
├── cmd/envv/
│   ├── main.go                    # CLI entry point (urfave/cli)
│   ├── encrypt.go, decrypt.go     # Core operations
│   ├── edit.go, set.go, rotate.go # File manipulation
│   ├── subcommand/
│   │   ├── exec/                  # Execute with secrets
│   │   ├── publish/               # Publish to S3/GCS/Vault
│   │   ├── keyservice/            # gRPC key service
│   │   └── groups/                # Key group management
│   └── common/                    # Shared utilities
├── sops.go                        # Core SOPS library
├── age/, pgp/, kms/, gcpkms/, azkv/, hcvault/  # Key providers
├── stores/                        # Format handlers (yaml, json, env)
├── config/                        # .sops.yaml configuration
├── demo/                          # Demo wrapper scripts
└── version/                       # Version management
```

### What Works Today

**Core Functionality:**
- ✅ Encrypt/decrypt files with age, PGP, AWS KMS, GCP KMS, Azure KeyVault
- ✅ Multiple file formats: YAML, JSON, ENV, INI, binary
- ✅ Key groups and Shamir secret sharing
- ✅ Edit encrypted files in $EDITOR
- ✅ Execute commands with decrypted environment
- ✅ Publish encrypted files to S3/GCS/Vault

**What's Missing for SaaS:**
- ❌ Backend authentication (login, logout, register)
- ❌ Team management commands
- ❌ Project management commands
- ❌ Push/pull secrets to/from backend
- ❌ Multi-user encryption coordination
- ❌ Configuration file for API endpoint
- ❌ Session token storage
- ❌ Public key distribution from backend
- ❌ Sync command (re-encrypt for team changes)

---

## Gap Analysis

### Critical Gaps (Must-Have for MVP)

| Gap | Current State | Required State | Complexity |
|-----|--------------|----------------|-----------|
| **Backend API Client** | None | HTTP client with JWT auth | High |
| **Auth Commands** | None | login, register, logout | Medium |
| **Team Commands** | None | create, list, invite | Medium |
| **Project Commands** | None | create, list, select | Medium |
| **Push Command** | None | Encrypt + upload to backend | High |
| **Pull Command** | None | Download + decrypt from backend | High |
| **Sync Command** | None | Re-encrypt for team changes | High |
| **Config Management** | .sops.yaml only | ~/.envv/config.json | Medium |
| **Session Storage** | None | JWT token storage | Low |
| **Error Handling** | Local only | Network + auth errors | Medium |

### Secondary Gaps (Nice-to-Have)

| Gap | Priority | Complexity |
|-----|---------|-----------|
| MFA support | Medium | Medium |
| OAuth login | Low | High |
| Offline mode | Medium | Medium |
| Web dashboard link | Low | Low |
| Shell completion for new commands | Medium | Low |
| Progress indicators for network ops | Low | Low |

---

## Proposed Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                       envv CLI                               │
│                                                              │
│  ┌────────────────┐  ┌────────────────┐  ┌──────────────┐  │
│  │ User Commands  │  │ SOPS Core      │  │ Config Mgr   │  │
│  │ - auth         │  │ - encrypt      │  │ - API URL    │  │
│  │ - team         │  │ - decrypt      │  │ - JWT token  │  │
│  │ - project      │  │ - age keys     │  │ - context    │  │
│  │ - push/pull    │  │ - formats      │  │              │  │
│  └────────┬───────┘  └────────┬───────┘  └──────┬───────┘  │
│           │                   │                   │          │
│           ├───────────────────┴───────────────────┤          │
│           │                                       │          │
│  ┌────────▼───────────────────────────────────────▼───────┐ │
│  │              Backend API Client                        │ │
│  │  - HTTP requests with JWT                              │ │
│  │  - Error handling & retries                            │ │
│  │  - Response parsing                                    │ │
│  └────────────────────────┬───────────────────────────────┘ │
└───────────────────────────┼─────────────────────────────────┘
                            │ HTTPS
                            ▼
                   ┌────────────────┐
                   │  envv Backend  │
                   │  - Auth        │
                   │  - Teams       │
                   │  - Projects    │
                   │  - Secrets     │
                   └────────────────┘
```

### Component Interactions

**Authentication Flow:**
```
User → envv auth login
  ↓
CLI prompts email/password
  ↓
CLI → POST /api/v1/auth/login → Backend
  ↓
Backend validates & returns JWT
  ↓
CLI stores JWT in ~/.envv/config.json
  ↓
CLI confirms: "Logged in as user@example.com"
```

**Push Secrets Flow:**
```
User → envv push --project=myapp
  ↓
CLI reads .env (plaintext)
  ↓
CLI → GET /api/v1/projects/{id}/members → Backend
  ↓
Backend returns: [{user_id, email, public_age_key}, ...]
  ↓
CLI generates .sops.yaml with all public keys
  ↓
CLI encrypts with SOPS (multi-recipient age encryption)
  ↓
CLI → POST /api/v1/projects/{id}/secrets (encrypted blob) → Backend
  ↓
Backend stores encrypted data (cannot decrypt)
  ↓
CLI deletes plaintext .env
  ↓
CLI confirms: "Pushed secrets to myapp (encrypted for 3 team members)"
```

**Pull Secrets Flow:**
```
User → envv pull --project=myapp
  ↓
CLI → GET /api/v1/projects/{id}/secrets → Backend
  ↓
Backend checks user has read access
  ↓
Backend returns encrypted blob
  ↓
CLI writes to .env.encrypted
  ↓
CLI decrypts with user's private age key
  ↓
CLI writes to .env (plaintext, local only)
  ↓
CLI confirms: "Pulled secrets from myapp (4 keys)"
```

---

## Implementation Plan

### Phase 1: Foundation (Week 1)

**Goal:** Core infrastructure for backend communication

#### 1.1 Backend API Client Package
**New file:** `pkg/api/client.go`

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

// Client handles all backend API communication
type Client struct {
    baseURL    string
    httpClient *http.Client
    token      string
}

// NewClient creates a new API client
func NewClient(baseURL string) *Client {
    return &Client{
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
            Transport: &http.Transport{
                MaxIdleConns:        10,
                IdleConnTimeout:     30 * time.Second,
                TLSHandshakeTimeout: 10 * time.Second,
            },
        },
    }
}

// SetToken sets the JWT token for authenticated requests
func (c *Client) SetToken(token string) {
    c.token = token
}

// Request makes an HTTP request to the backend
func (c *Client) Request(ctx context.Context, method, path string, body interface{}, result interface{}) error {
    var reqBody io.Reader
    if body != nil {
        jsonData, err := json.Marshal(body)
        if err != nil {
            return fmt.Errorf("failed to marshal request: %w", err)
        }
        reqBody = bytes.NewReader(jsonData)
    }

    req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }

    req.Header.Set("Content-Type", "application/json")
    if c.token != "" {
        req.Header.Set("Authorization", "Bearer "+c.token)
    }

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        return c.handleErrorResponse(resp)
    }

    if result != nil {
        if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
            return fmt.Errorf("failed to decode response: %w", err)
        }
    }

    return nil
}

// handleErrorResponse parses error responses from the backend
func (c *Client) handleErrorResponse(resp *http.Response) error {
    var errResp struct {
        Error struct {
            Code    string `json:"code"`
            Message string `json:"message"`
        } `json:"error"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
        return &APIError{
            StatusCode: resp.StatusCode,
            Code:       errResp.Error.Code,
            Message:    errResp.Error.Message,
        }
    }

    return &APIError{
        StatusCode: resp.StatusCode,
        Code:       "UNKNOWN",
        Message:    resp.Status,
    }
}

// APIError represents an error from the backend API
type APIError struct {
    StatusCode int
    Code       string
    Message    string
}

func (e *APIError) Error() string {
    return fmt.Sprintf("API error %d (%s): %s", e.StatusCode, e.Code, e.Message)
}

// IsUnauthorized checks if error is 401 Unauthorized
func (e *APIError) IsUnauthorized() bool {
    return e.StatusCode == 401
}

// IsForbidden checks if error is 403 Forbidden
func (e *APIError) IsForbidden() bool {
    return e.StatusCode == 403
}
```

#### 1.2 Configuration Management
**New file:** `pkg/config/config.go`

```go
package config

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
)

const (
    DefaultConfigFileName = "config.json"
    DefaultAPIURL         = "https://api.envv.app"
)

// Config represents the CLI configuration
type Config struct {
    APIEndpoint    string `json:"api_endpoint"`
    AuthToken      string `json:"auth_token"`
    CurrentTeamID  string `json:"current_team_id"`
    CurrentProject string `json:"current_project"`
    UserEmail      string `json:"user_email"`
    UserID         string `json:"user_id"`
}

// Manager handles configuration persistence
type Manager struct {
    configPath string
}

// NewManager creates a new configuration manager
func NewManager() (*Manager, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return nil, fmt.Errorf("failed to get home directory: %w", err)
    }

    configDir := filepath.Join(homeDir, ".envv")
    if err := os.MkdirAll(configDir, 0700); err != nil {
        return nil, fmt.Errorf("failed to create config directory: %w", err)
    }

    return &Manager{
        configPath: filepath.Join(configDir, DefaultConfigFileName),
    }, nil
}

// Load reads configuration from disk
func (m *Manager) Load() (*Config, error) {
    data, err := os.ReadFile(m.configPath)
    if err != nil {
        if os.IsNotExist(err) {
            // Return default config if file doesn't exist
            return &Config{
                APIEndpoint: DefaultAPIURL,
            }, nil
        }
        return nil, fmt.Errorf("failed to read config: %w", err)
    }

    var config Config
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("failed to parse config: %w", err)
    }

    return &config, nil
}

// Save writes configuration to disk
func (m *Manager) Save(config *Config) error {
    data, err := json.MarshalIndent(config, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal config: %w", err)
    }

    if err := os.WriteFile(m.configPath, data, 0600); err != nil {
        return fmt.Errorf("failed to write config: %w", err)
    }

    return nil
}

// Clear removes the configuration file
func (m *Manager) Clear() error {
    if err := os.Remove(m.configPath); err != nil && !os.IsNotExist(err) {
        return fmt.Errorf("failed to remove config: %w", err)
    }
    return nil
}
```

#### 1.3 Age Key Management Helper
**New file:** `pkg/keys/age.go`

```go
package keys

import (
    "crypto/rand"
    "fmt"
    "os"
    "path/filepath"

    "filippo.io/age"
)

const (
    DefaultAgeKeyFileName = "keys.txt"
)

// GenerateAgeKey generates a new age keypair
func GenerateAgeKey() (publicKey, privateKey string, err error) {
    identity, err := age.GenerateX25519Identity()
    if err != nil {
        return "", "", fmt.Errorf("failed to generate age key: %w", err)
    }

    return identity.Recipient().String(), identity.String(), nil
}

// GetDefaultAgeKeyPath returns the default age key file path
func GetDefaultAgeKeyPath() (string, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return "", fmt.Errorf("failed to get home directory: %w", err)
    }

    // Use standard SOPS age key location
    return filepath.Join(homeDir, ".config", "sops", "age", DefaultAgeKeyFileName), nil
}

// SaveAgeKey saves the private age key to disk
func SaveAgeKey(privateKey string) error {
    keyPath, err := GetDefaultAgeKeyPath()
    if err != nil {
        return err
    }

    keyDir := filepath.Dir(keyPath)
    if err := os.MkdirAll(keyDir, 0700); err != nil {
        return fmt.Errorf("failed to create age key directory: %w", err)
    }

    // Append to key file (SOPS supports multiple keys)
    f, err := os.OpenFile(keyPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
    if err != nil {
        return fmt.Errorf("failed to open key file: %w", err)
    }
    defer f.Close()

    if _, err := f.WriteString(fmt.Sprintf("%s\n", privateKey)); err != nil {
        return fmt.Errorf("failed to write key: %w", err)
    }

    return nil
}

// LoadAgePublicKey loads the user's age public key
func LoadAgePublicKey() (string, error) {
    keyPath, err := GetDefaultAgeKeyPath()
    if err != nil {
        return "", err
    }

    data, err := os.ReadFile(keyPath)
    if err != nil {
        return "", fmt.Errorf("failed to read age key: %w (run 'envv auth register' first)", err)
    }

    // Parse first identity to get recipient (public key)
    identities, err := age.ParseIdentities(bytes.NewReader(data))
    if err != nil || len(identities) == 0 {
        return "", fmt.Errorf("failed to parse age key: %w", err)
    }

    return identities[0].Recipient().String(), nil
}
```

---

### Phase 2: Authentication Commands (Week 1)

**Goal:** Implement auth subcommands

#### 2.1 Auth Package
**New file:** `pkg/api/auth.go`

```go
package api

import (
    "context"
    "fmt"
)

// AuthService handles authentication operations
type AuthService struct {
    client *Client
}

// NewAuthService creates a new auth service
func NewAuthService(client *Client) *AuthService {
    return &AuthService{client: client}
}

// LoginRequest represents login credentials
type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

// LoginResponse represents login result
type LoginResponse struct {
    Success bool `json:"success"`
    Data    struct {
        Token string `json:"token"`
        User  struct {
            ID        string `json:"id"`
            Email     string `json:"email"`
            Name      string `json:"full_name"`
            PublicKey string `json:"public_key"`
        } `json:"user"`
    } `json:"data"`
}

// Login authenticates the user
func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
    req := LoginRequest{
        Email:    email,
        Password: password,
    }

    var resp LoginResponse
    if err := s.client.Request(ctx, "POST", "/api/v1/auth/login", req, &resp); err != nil {
        return nil, err
    }

    return &resp, nil
}

// RegisterRequest represents registration details
type RegisterRequest struct {
    Email     string `json:"email"`
    Password  string `json:"password"`
    Name      string `json:"full_name"`
    PublicKey string `json:"age_public_key"`
}

// RegisterResponse represents registration result
type RegisterResponse struct {
    Success bool `json:"success"`
    Data    struct {
        Token string `json:"token"`
        User  struct {
            ID        string `json:"id"`
            Email     string `json:"email"`
            Name      string `json:"full_name"`
            PublicKey string `json:"public_key"`
        } `json:"user"`
    } `json:"data"`
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, email, password, name, publicKey string) (*RegisterResponse, error) {
    req := RegisterRequest{
        Email:     email,
        Password:  password,
        Name:      name,
        PublicKey: publicKey,
    }

    var resp RegisterResponse
    if err := s.client.Request(ctx, "POST", "/api/v1/auth/register", req, &resp); err != nil {
        return nil, err
    }

    return &resp, nil
}

// Logout invalidates the current session
func (s *AuthService) Logout(ctx context.Context) error {
    return s.client.Request(ctx, "POST", "/api/v1/auth/logout", nil, nil)
}

// GetCurrentUser retrieves current user info
func (s *AuthService) GetCurrentUser(ctx context.Context) (*LoginResponse, error) {
    var resp LoginResponse
    if err := s.client.Request(ctx, "GET", "/api/v1/auth/me", nil, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}
```

#### 2.2 Auth Command Implementation
**New file:** `cmd/envv/subcommand/auth/auth.go`

```go
package auth

import (
    "context"
    "fmt"
    "os"
    "syscall"

    "github.com/AetherVoxSanctum/envv-cli/v3/pkg/api"
    "github.com/AetherVoxSanctum/envv-cli/v3/pkg/config"
    "github.com/AetherVoxSanctum/envv-cli/v3/pkg/keys"
    "github.com/sirupsen/logrus"
    "github.com/urfave/cli"
    "golang.org/x/term"
)

var log *logrus.Logger

func init() {
    log = logrus.New()
}

// RegisterCommand creates the auth command
func RegisterCommand() cli.Command {
    return cli.Command{
        Name:  "auth",
        Usage: "Manage authentication with envv backend",
        Subcommands: []cli.Command{
            loginCommand(),
            registerCommand(),
            logoutCommand(),
            statusCommand(),
        },
    }
}

func loginCommand() cli.Command {
    return cli.Command{
        Name:      "login",
        Usage:     "Login to envv backend",
        ArgsUsage: " ",
        Action: func(c *cli.Context) error {
            cfg, err := loadConfig()
            if err != nil {
                return err
            }

            // Prompt for email
            fmt.Print("Email: ")
            var email string
            fmt.Scanln(&email)

            // Prompt for password (hidden)
            fmt.Print("Password: ")
            passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
            fmt.Println()
            if err != nil {
                return fmt.Errorf("failed to read password: %w", err)
            }
            password := string(passwordBytes)

            // Authenticate with backend
            client := api.NewClient(cfg.APIEndpoint)
            authService := api.NewAuthService(client)

            resp, err := authService.Login(context.Background(), email, password)
            if err != nil {
                return fmt.Errorf("login failed: %w", err)
            }

            // Save credentials
            cfg.AuthToken = resp.Data.Token
            cfg.UserEmail = resp.Data.User.Email
            cfg.UserID = resp.Data.User.ID

            if err := saveConfig(cfg); err != nil {
                return err
            }

            log.Infof("✓ Logged in as %s", resp.Data.User.Email)
            return nil
        },
    }
}

func registerCommand() cli.Command {
    return cli.Command{
        Name:      "register",
        Usage:     "Register a new account",
        ArgsUsage: " ",
        Action: func(c *cli.Context) error {
            cfg, err := loadConfig()
            if err != nil {
                return err
            }

            // Prompt for details
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
                return fmt.Errorf("failed to read password: %w", err)
            }
            password := string(passwordBytes)

            // Generate age key
            log.Info("Generating encryption keypair...")
            publicKey, privateKey, err := keys.GenerateAgeKey()
            if err != nil {
                return err
            }

            // Save private key
            if err := keys.SaveAgeKey(privateKey); err != nil {
                return fmt.Errorf("failed to save key: %w", err)
            }
            log.Infof("✓ Saved private key to %s", keys.DefaultAgeKeyPath)

            // Register with backend
            client := api.NewClient(cfg.APIEndpoint)
            authService := api.NewAuthService(client)

            resp, err := authService.Register(context.Background(), email, password, name, publicKey)
            if err != nil {
                return fmt.Errorf("registration failed: %w", err)
            }

            // Save credentials
            cfg.AuthToken = resp.Data.Token
            cfg.UserEmail = resp.Data.User.Email
            cfg.UserID = resp.Data.User.ID

            if err := saveConfig(cfg); err != nil {
                return err
            }

            log.Infof("✓ Registered and logged in as %s", resp.Data.User.Email)
            log.Info("Your age public key: " + publicKey)
            return nil
        },
    }
}

func logoutCommand() cli.Command {
    return cli.Command{
        Name:      "logout",
        Usage:     "Logout from envv backend",
        ArgsUsage: " ",
        Action: func(c *cli.Context) error {
            cfg, err := loadConfig()
            if err != nil {
                return err
            }

            if cfg.AuthToken == "" {
                log.Info("Not logged in")
                return nil
            }

            // Call backend logout
            client := api.NewClient(cfg.APIEndpoint)
            client.SetToken(cfg.AuthToken)
            authService := api.NewAuthService(client)

            if err := authService.Logout(context.Background()); err != nil {
                log.Warnf("Backend logout failed: %v", err)
            }

            // Clear local config
            cfgMgr, _ := config.NewManager()
            if err := cfgMgr.Clear(); err != nil {
                return err
            }

            log.Info("✓ Logged out")
            return nil
        },
    }
}

func statusCommand() cli.Command {
    return cli.Command{
        Name:      "status",
        Usage:     "Show current authentication status",
        ArgsUsage: " ",
        Action: func(c *cli.Context) error {
            cfg, err := loadConfig()
            if err != nil {
                return err
            }

            if cfg.AuthToken == "" {
                log.Info("Not logged in")
                log.Info("Run 'envv auth login' to authenticate")
                return nil
            }

            // Verify token with backend
            client := api.NewClient(cfg.APIEndpoint)
            client.SetToken(cfg.AuthToken)
            authService := api.NewAuthService(client)

            resp, err := authService.GetCurrentUser(context.Background())
            if err != nil {
                log.Errorf("Session expired or invalid: %v", err)
                log.Info("Run 'envv auth login' to re-authenticate")
                return nil
            }

            log.Info("Logged in as: " + resp.Data.User.Email)
            log.Info("User ID: " + resp.Data.User.ID)
            log.Info("API endpoint: " + cfg.APIEndpoint)
            return nil
        },
    }
}

// Helper functions
func loadConfig() (*config.Config, error) {
    mgr, err := config.NewManager()
    if err != nil {
        return nil, err
    }
    return mgr.Load()
}

func saveConfig(cfg *config.Config) error {
    mgr, err := config.NewManager()
    if err != nil {
        return err
    }
    return mgr.Save(cfg)
}
```

---

### Phase 3: Team & Project Commands (Week 2)

#### 3.1 Team Commands
**New file:** `cmd/envv/subcommand/team/team.go`

```go
package team

import (
    "context"
    "fmt"
    "text/tabwriter"
    "os"

    "github.com/AetherVoxSanctum/envv-cli/v3/pkg/api"
    "github.com/AetherVoxSanctum/envv-cli/v3/pkg/config"
    "github.com/urfave/cli"
)

// RegisterCommand creates the team command
func RegisterCommand() cli.Command {
    return cli.Command{
        Name:  "team",
        Usage: "Manage teams",
        Subcommands: []cli.Command{
            createCommand(),
            listCommand(),
            inviteCommand(),
            membersCommand(),
        },
    }
}

func createCommand() cli.Command {
    return cli.Command{
        Name:      "create",
        Usage:     "Create a new team",
        ArgsUsage: "<team-name>",
        Action: func(c *cli.Context) error {
            if c.NArg() < 1 {
                return fmt.Errorf("team name required")
            }

            teamName := c.Args().Get(0)

            client, err := getAuthenticatedClient()
            if err != nil {
                return err
            }

            teamService := api.NewTeamService(client)
            team, err := teamService.Create(context.Background(), teamName, "")
            if err != nil {
                return fmt.Errorf("failed to create team: %w", err)
            }

            fmt.Printf("✓ Created team '%s' (ID: %s)\n", team.Name, team.ID)
            return nil
        },
    }
}

func listCommand() cli.Command {
    return cli.Command{
        Name:      "list",
        Aliases:   []string{"ls"},
        Usage:     "List your teams",
        ArgsUsage: " ",
        Action: func(c *cli.Context) error {
            client, err := getAuthenticatedClient()
            if err != nil {
                return err
            }

            teamService := api.NewTeamService(client)
            teams, err := teamService.List(context.Background())
            if err != nil {
                return fmt.Errorf("failed to list teams: %w", err)
            }

            if len(teams) == 0 {
                fmt.Println("No teams yet. Create one with 'envv team create <name>'")
                return nil
            }

            w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
            fmt.Fprintln(w, "NAME\tID\tROLE\tMEMBERS\tPROJECTS")
            for _, team := range teams {
                fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\n",
                    team.Name, team.ID, team.Role, team.MemberCount, team.ProjectCount)
            }
            w.Flush()

            return nil
        },
    }
}

func inviteCommand() cli.Command {
    return cli.Command{
        Name:      "invite",
        Usage:     "Invite member to team",
        ArgsUsage: "<team-id> <email>",
        Flags: []cli.Flag{
            cli.StringFlag{
                Name:  "role",
                Value: "member",
                Usage: "Role: member or admin",
            },
        },
        Action: func(c *cli.Context) error {
            if c.NArg() < 2 {
                return fmt.Errorf("team ID and email required")
            }

            teamID := c.Args().Get(0)
            email := c.Args().Get(1)
            role := c.String("role")

            client, err := getAuthenticatedClient()
            if err != nil {
                return err
            }

            teamService := api.NewTeamService(client)
            invite, err := teamService.Invite(context.Background(), teamID, email, role)
            if err != nil {
                return fmt.Errorf("failed to invite: %w", err)
            }

            fmt.Printf("✓ Invited %s to team (invite ID: %s)\n", email, invite.ID)
            fmt.Printf("They will receive an email with instructions\n")
            return nil
        },
    }
}

func membersCommand() cli.Command {
    return cli.Command{
        Name:      "members",
        Usage:     "List team members",
        ArgsUsage: "<team-id>",
        Action: func(c *cli.Context) error {
            if c.NArg() < 1 {
                return fmt.Errorf("team ID required")
            }

            teamID := c.Args().Get(0)

            client, err := getAuthenticatedClient()
            if err != nil {
                return err
            }

            teamService := api.NewTeamService(client)
            members, err := teamService.ListMembers(context.Background(), teamID)
            if err != nil {
                return fmt.Errorf("failed to list members: %w", err)
            }

            w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
            fmt.Fprintln(w, "EMAIL\tNAME\tROLE\tPUBLIC KEY")
            for _, member := range members {
                fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
                    member.Email, member.Name, member.Role, member.PublicKey[:16]+"...")
            }
            w.Flush()

            return nil
        },
    }
}

// Helper to get authenticated API client
func getAuthenticatedClient() (*api.Client, error) {
    mgr, err := config.NewManager()
    if err != nil {
        return nil, err
    }

    cfg, err := mgr.Load()
    if err != nil {
        return nil, err
    }

    if cfg.AuthToken == "" {
        return nil, fmt.Errorf("not logged in. Run 'envv auth login'")
    }

    client := api.NewClient(cfg.APIEndpoint)
    client.SetToken(cfg.AuthToken)
    return client, nil
}
```

#### 3.2 Project Commands
**New file:** `cmd/envv/subcommand/project/project.go`

Similar structure to team commands, implementing:
- `create <name> --team=<id>`
- `list --team=<id>`
- `select <project-id>` (sets current project in config)
- `access grant <user-email> --permission=read|write|admin`
- `access revoke <user-email>`
- `members` (list who has access)

---

### Phase 4: Push/Pull Commands (Week 2-3)

#### 4.1 Push Command
**New file:** `cmd/envv/subcommand/secrets/push.go`

```go
package secrets

import (
    "context"
    "fmt"
    "os"
    "path/filepath"

    "github.com/AetherVoxSanctum/envv-cli/v3"
    "github.com/AetherVoxSanctum/envv-cli/v3/pkg/api"
    "github.com/AetherVoxSanctum/envv-cli/v3/pkg/config"
    "github.com/urfave/cli"
)

func pushCommand() cli.Command {
    return cli.Command{
        Name:      "push",
        Usage:     "Encrypt and push secrets to backend",
        ArgsUsage: "[file]",
        Flags: []cli.Flag{
            cli.StringFlag{
                Name:  "project",
                Usage: "Project ID or name",
            },
            cli.StringFlag{
                Name:  "env",
                Value: "default",
                Usage: "Environment (default, dev, staging, prod)",
            },
        },
        Action: func(c *cli.Context) error {
            projectID := c.String("project")
            if projectID == "" {
                // Try to get from config
                cfg, err := loadConfig()
                if err != nil {
                    return err
                }
                projectID = cfg.CurrentProject
                if projectID == "" {
                    return fmt.Errorf("no project specified. Use --project or 'envv project select'")
                }
            }

            // Determine input file
            inputFile := ".env"
            if c.NArg() > 0 {
                inputFile = c.Args().Get(0)
            }

            // Verify file exists
            if _, err := os.Stat(inputFile); err != nil {
                return fmt.Errorf("file not found: %s", inputFile)
            }

            // Get authenticated client
            client, err := getAuthenticatedClient()
            if err != nil {
                return err
            }

            // Get project members (for encryption)
            projectService := api.NewProjectService(client)
            members, err := projectService.ListMembers(context.Background(), projectID)
            if err != nil {
                return fmt.Errorf("failed to get project members: %w", err)
            }

            if len(members) == 0 {
                return fmt.Errorf("project has no members")
            }

            fmt.Printf("Encrypting for %d team member(s)...\n", len(members))

            // Generate .sops.yaml with all member public keys
            publicKeys := make([]string, len(members))
            for i, member := range members {
                publicKeys[i] = member.PublicKey
            }

            if err := generateSOPSConfig(publicKeys); err != nil {
                return fmt.Errorf("failed to generate SOPS config: %w", err)
            }

            // Encrypt with SOPS
            encryptedFile := inputFile + ".encrypted"
            if err := encryptFile(inputFile, encryptedFile); err != nil {
                return fmt.Errorf("encryption failed: %w", err)
            }
            defer os.Remove(encryptedFile) // Clean up temp file

            // Read encrypted content
            encryptedData, err := os.ReadFile(encryptedFile)
            if err != nil {
                return fmt.Errorf("failed to read encrypted file: %w", err)
            }

            // Push to backend
            secretsService := api.NewSecretsService(client)
            version, err := secretsService.Push(context.Background(), projectID, c.String("env"), string(encryptedData), "env")
            if err != nil {
                return fmt.Errorf("failed to push secrets: %w", err)
            }

            fmt.Printf("✓ Pushed secrets to project %s (version %s)\n", projectID, version.ID)

            // Optionally remove plaintext file
            fmt.Printf("Remove plaintext file %s? (y/N): ", inputFile)
            var response string
            fmt.Scanln(&response)
            if response == "y" || response == "Y" {
                os.Remove(inputFile)
                fmt.Println("✓ Removed plaintext file")
            }

            return nil
        },
    }
}

// encryptFile uses SOPS to encrypt a file
func encryptFile(input, output string) error {
    // Use existing SOPS encryption logic
    tree, err := sops.LoadEncryptionConfigFromFile(".sops.yaml")
    if err != nil {
        return err
    }

    // ... actual SOPS encryption implementation
    // This integrates with existing sops.go encryption logic

    return nil
}

// generateSOPSConfig creates a .sops.yaml with team member public keys
func generateSOPSConfig(publicKeys []string) error {
    sopsConfig := fmt.Sprintf(`creation_rules:
  - path_regex: \.env.*$
    age: >-
      %s
`, strings.Join(publicKeys, ",\n      "))

    return os.WriteFile(".sops.yaml", []byte(sopsConfig), 0644)
}
```

#### 4.2 Pull Command
**New file:** `cmd/envv/subcommand/secrets/pull.go`

```go
package secrets

import (
    "context"
    "fmt"
    "os"

    "github.com/AetherVoxSanctum/envv-cli/v3/pkg/api"
    "github.com/urfave/cli"
)

func pullCommand() cli.Command {
    return cli.Command{
        Name:      "pull",
        Usage:     "Pull and decrypt secrets from backend",
        ArgsUsage: " ",
        Flags: []cli.Flag{
            cli.StringFlag{
                Name:  "project",
                Usage: "Project ID or name",
            },
            cli.StringFlag{
                Name:  "env",
                Value: "default",
                Usage: "Environment",
            },
            cli.StringFlag{
                Name:  "output",
                Value: ".env",
                Usage: "Output file",
            },
        },
        Action: func(c *cli.Context) error {
            projectID := c.String("project")
            if projectID == "" {
                cfg, err := loadConfig()
                if err != nil {
                    return err
                }
                projectID = cfg.CurrentProject
                if projectID == "" {
                    return fmt.Errorf("no project specified")
                }
            }

            client, err := getAuthenticatedClient()
            if err != nil {
                return err
            }

            // Pull encrypted secrets from backend
            secretsService := api.NewSecretsService(client)
            secrets, err := secretsService.Pull(context.Background(), projectID, c.String("env"))
            if err != nil {
                return fmt.Errorf("failed to pull secrets: %w", err)
            }

            // Write encrypted content to temp file
            encryptedFile := ".env.encrypted"
            if err := os.WriteFile(encryptedFile, []byte(secrets.Data), 0600); err != nil {
                return fmt.Errorf("failed to write encrypted file: %w", err)
            }

            fmt.Println("Decrypting with your private key...")

            // Decrypt with SOPS
            outputFile := c.String("output")
            if err := decryptFile(encryptedFile, outputFile); err != nil {
                return fmt.Errorf("decryption failed: %w (do you have access to this project?)", err)
            }

            fmt.Printf("✓ Pulled and decrypted secrets to %s\n", outputFile)
            return nil
        },
    }
}

// decryptFile uses SOPS to decrypt a file
func decryptFile(input, output string) error {
    // Use existing SOPS decryption logic
    // Reads user's private key from ~/.config/sops/age/keys.txt
    // Decrypts and writes to output file

    return nil
}
```

---

### Phase 5: Integration & Testing (Week 3-4)

#### 5.1 Update Main CLI Entry Point
**Modify:** `cmd/envv/main.go`

```go
// Add imports
import (
    "github.com/AetherVoxSanctum/envv-cli/v3/cmd/envv/subcommand/auth"
    "github.com/AetherVoxSanctum/envv-cli/v3/cmd/envv/subcommand/team"
    "github.com/AetherVoxSanctum/envv-cli/v3/cmd/envv/subcommand/project"
    "github.com/AetherVoxSanctum/envv-cli/v3/cmd/envv/subcommand/secrets"
)

func main() {
    // ... existing code ...

    // Add new commands
    app.Commands = append(app.Commands,
        auth.RegisterCommand(),
        team.RegisterCommand(),
        project.RegisterCommand(),
        secrets.RegisterCommand(),
    )

    // ... rest of main ...
}
```

#### 5.2 Testing Infrastructure

**Unit Tests:**
- `pkg/api/client_test.go` - API client tests with mock server
- `pkg/config/config_test.go` - Config management tests
- `pkg/keys/age_test.go` - Key generation tests

**Integration Tests:**
- `tests/integration/auth_test.go` - Auth flow tests
- `tests/integration/push_pull_test.go` - Secret sync tests
- `tests/integration/team_test.go` - Team management tests

**E2E Tests:**
- `tests/e2e/full_workflow_test.go` - Complete user journey

---

## New Command Structure

```
envv
├── auth
│   ├── login              # Authenticate with backend
│   ├── register           # Create new account
│   ├── logout             # Clear session
│   └── status             # Show current user
├── team
│   ├── create <name>      # Create team
│   ├── list               # List your teams
│   ├── invite <email>     # Invite member
│   └── members <team-id>  # List members
├── project
│   ├── create <name>      # Create project
│   ├── list               # List projects
│   ├── select <id>        # Set current project
│   ├── access
│   │   ├── grant <email>  # Grant access
│   │   └── revoke <email> # Revoke access
│   └── members            # List members
├── push [file]            # Encrypt & upload secrets
├── pull [--output file]   # Download & decrypt secrets
├── sync                   # Re-encrypt for team changes
├── list                   # List secret keys (no values)
├── exec <command>         # Run command with secrets
│
├── encrypt [file]         # (existing SOPS)
├── decrypt [file]         # (existing SOPS)
├── edit [file]            # (existing SOPS)
└── ... (other SOPS commands)
```

---

## Code Organization

```
envv-cli/
├── cmd/envv/
│   ├── main.go                        # Entry point (updated)
│   ├── subcommand/
│   │   ├── auth/                      # NEW: Auth commands
│   │   │   └── auth.go
│   │   ├── team/                      # NEW: Team commands
│   │   │   └── team.go
│   │   ├── project/                   # NEW: Project commands
│   │   │   └── project.go
│   │   ├── secrets/                   # NEW: Push/pull/sync
│   │   │   ├── push.go
│   │   │   ├── pull.go
│   │   │   └── sync.go
│   │   └── ... (existing SOPS subcommands)
│   └── ... (existing files)
├── pkg/                               # NEW: Shared packages
│   ├── api/                           # Backend API client
│   │   ├── client.go                  # HTTP client
│   │   ├── auth.go                    # Auth service
│   │   ├── team.go                    # Team service
│   │   ├── project.go                 # Project service
│   │   └── secrets.go                 # Secrets service
│   ├── config/                        # Configuration management
│   │   └── config.go
│   └── keys/                          # Key management
│       └── age.go
├── tests/                             # NEW: Test suite
│   ├── unit/
│   ├── integration/
│   └── e2e/
├── docs/                              # NEW: User documentation
│   ├── getting-started.md
│   ├── commands.md
│   └── architecture.md
└── ... (existing SOPS files)
```

---

## Integration with Existing SOPS

### Key Integration Points

1. **Encryption/Decryption:**
   - New commands use existing `sops.Encrypt()` and `sops.Decrypt()` functions
   - No changes to core SOPS logic
   - Add wrappers in `pkg/sopsutil/` for convenience

2. **Key Management:**
   - Continue using SOPS's age key handling
   - Store keys in standard location: `~/.config/sops/age/keys.txt`
   - Backend receives only public keys

3. **File Formats:**
   - Leverage existing store implementations (yaml, json, env, ini)
   - No changes needed

4. **Configuration:**
   - Dynamically generate `.sops.yaml` based on team members
   - Override with CLI flags if needed

---

## Configuration Management

### Config File Location

`~/.envv/config.json` (0600 permissions)

### Config Schema

```json
{
  "api_endpoint": "https://api.envv.app",
  "auth_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user_id": "uuid",
  "user_email": "user@example.com",
  "current_team_id": "team-uuid",
  "current_project": "project-uuid",
  "preferences": {
    "auto_remove_plaintext": false,
    "default_environment": "default"
  }
}
```

### Environment Variables

Support configuration via environment variables:
- `ENVV_API_URL` - Override API endpoint
- `ENVV_TOKEN` - Override auth token (for CI/CD)
- `ENVV_PROJECT` - Override current project
- `SOPS_AGE_KEY_FILE` - Age key location (existing SOPS)

---

## Testing Strategy

### Unit Tests (80% coverage target)

**API Client Tests:**
```go
// pkg/api/client_test.go
func TestClient_Request_Success(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "Bearer token123", r.Header.Get("Authorization"))
        w.WriteHeader(200)
        json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
    }))
    defer server.Close()

    client := api.NewClient(server.URL)
    client.SetToken("token123")

    var result map[string]string
    err := client.Request(context.Background(), "GET", "/test", nil, &result)

    assert.NoError(t, err)
    assert.Equal(t, "ok", result["status"])
}

func TestClient_Request_Unauthorized(t *testing.T) {
    // Test 401 handling
}

func TestClient_Request_NetworkError(t *testing.T) {
    // Test network failure
}
```

**Config Tests:**
```go
// pkg/config/config_test.go
func TestConfig_SaveAndLoad(t *testing.T) {
    // Test round-trip save/load
}

func TestConfig_SecurePermissions(t *testing.T) {
    // Verify file permissions are 0600
}
```

### Integration Tests

**With Mock Backend:**
```go
// tests/integration/auth_test.go
func TestAuthFlow(t *testing.T) {
    mockBackend := setupMockBackend()
    defer mockBackend.Close()

    // Test: register → login → status → logout
    // Verify config is updated correctly at each step
}
```

**With Real SOPS:**
```go
// tests/integration/push_pull_test.go
func TestPushPullFlow(t *testing.T) {
    // Generate test age keys
    // Create .env with secrets
    // Push to mock backend
    // Pull from mock backend
    // Verify secrets match
}
```

### E2E Tests

**Complete User Journey:**
```bash
#!/bin/bash
# tests/e2e/full_workflow.sh

# Setup
export ENVV_API_URL="http://localhost:8080"
./envv auth register << EOF
Test User
test@example.com
SecurePassword123!
EOF

# Create team
./envv team create "Test Team"
TEAM_ID=$(./envv team list --format=json | jq -r '.[0].id')

# Create project
./envv project create "Test Project" --team=$TEAM_ID
PROJECT_ID=$(./envv project list --format=json | jq -r '.[0].id')

# Push secrets
echo "DATABASE_URL=postgres://localhost" > .env
./envv push --project=$PROJECT_ID

# Pull secrets
rm .env
./envv pull --project=$PROJECT_ID
grep -q "DATABASE_URL=postgres://localhost" .env && echo "✓ E2E test passed"
```

---

## Security Considerations

### Credential Storage

1. **JWT Token:**
   - Stored in `~/.envv/config.json` with 0600 permissions
   - Never logged or printed
   - Cleared on logout

2. **Private Age Keys:**
   - Stored in `~/.config/sops/age/keys.txt` with 0600 permissions
   - Never sent to backend
   - Never logged

3. **Plaintext Secrets:**
   - Only exist temporarily during push/pull
   - Prompt to remove after push
   - Never committed to git (via .gitignore)

### Network Security

1. **HTTPS Only:**
   - All API calls over TLS
   - Certificate validation enforced
   - No insecure fallback

2. **Token Expiry:**
   - Handle 401 Unauthorized gracefully
   - Prompt user to re-login
   - Clear invalid tokens

3. **Rate Limiting:**
   - Respect backend rate limits
   - Implement exponential backoff for retries

### Input Validation

1. **File Paths:**
   - Prevent path traversal
   - Validate file extensions
   - Check file permissions before reading

2. **API Inputs:**
   - Validate email format
   - Sanitize team/project names
   - Limit request sizes

---

## Documentation Plan

### User Documentation

1. **Getting Started Guide** (`docs/getting-started.md`)
   - Installation
   - Registration
   - First project setup
   - Push/pull workflow

2. **Command Reference** (`docs/commands.md`)
   - Complete command list with examples
   - Common use cases
   - Troubleshooting

3. **Team Guide** (`docs/teams.md`)
   - Team management
   - Inviting members
   - Access control

4. **Integration Guide** (`docs/integrations.md`)
   - CI/CD pipelines
   - GitHub Actions
   - GitLab CI
   - Docker usage

### Developer Documentation

1. **Architecture** (`docs/architecture.md`)
   - System design
   - Component interactions
   - Security model

2. **Contributing** (`CONTRIBUTING.md`)
   - Development setup
   - Code style
   - Pull request process

3. **API Integration** (`docs/api.md`)
   - Backend API usage
   - Request/response formats
   - Error handling

---

## Release Strategy

### Versioning

**Semantic Versioning:** `v3.x.x`
- Major: Breaking changes
- Minor: New features (backward compatible)
- Patch: Bug fixes

### Build Process

```makefile
# Makefile additions

.PHONY: build-all
build-all: build-linux build-darwin build-windows

build-linux:
	GOOS=linux GOARCH=amd64 go build -o dist/envv-linux-amd64 ./cmd/envv
	GOOS=linux GOARCH=arm64 go build -o dist/envv-linux-arm64 ./cmd/envv

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o dist/envv-darwin-amd64 ./cmd/envv
	GOOS=darwin GOARCH=arm64 go build -o dist/envv-darwin-arm64 ./cmd/envv

build-windows:
	GOOS=windows GOARCH=amd64 go build -o dist/envv-windows-amd64.exe ./cmd/envv

.PHONY: release
release: test build-all
	@echo "Creating release packages..."
	cd dist && tar -czf envv-$(VERSION)-linux-amd64.tar.gz envv-linux-amd64
	cd dist && tar -czf envv-$(VERSION)-darwin-amd64.tar.gz envv-darwin-amd64
	# ... more packaging
```

### Release Checklist

- [ ] All tests passing
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Version bumped in code
- [ ] Git tag created
- [ ] Binaries built for all platforms
- [ ] GitHub release created with binaries
- [ ] Homebrew formula updated (macOS)
- [ [ ] Snap/deb packages created (Linux)

### Installation Methods

1. **Direct Download:**
   ```bash
   curl -L https://github.com/AetherVoxSanctum/envv-cli/releases/download/v3.x.x/envv-linux-amd64 -o envv
   chmod +x envv
   sudo mv envv /usr/local/bin/
   ```

2. **Homebrew (macOS):**
   ```bash
   brew install AetherVoxSanctum/tap/envv
   ```

3. **Install Script:**
   ```bash
   curl -sSL https://get.envv.app | bash
   ```

4. **Go Install:**
   ```bash
   go install github.com/AetherVoxSanctum/envv-cli/v3/cmd/envv@latest
   ```

---

## Timeline & Milestones

### Week 1: Foundation
- [x] Backend API client package
- [x] Configuration management
- [x] Age key management
- [ ] Auth commands (login, register, logout, status)
- [ ] Unit tests for API client & config

### Week 2: Team & Project Management
- [ ] Team commands (create, list, invite, members)
- [ ] Project commands (create, list, select, access)
- [ ] API services for teams & projects
- [ ] Integration tests

### Week 3: Secrets Management
- [ ] Push command (encrypt + upload)
- [ ] Pull command (download + decrypt)
- [ ] Sync command (re-encrypt for team changes)
- [ ] List command (show keys without values)
- [ ] Integration with SOPS encryption

### Week 4: Polish & Testing
- [ ] Comprehensive testing (unit, integration, E2E)
- [ ] Error handling improvements
- [ ] User documentation
- [ ] Demo video/GIF
- [ ] Release preparation

### Week 5+: Enhancements
- [ ] MFA support
- [ ] OAuth login (GitHub, Google)
- [ ] Offline mode
- [ ] Shell completion
- [ ] Progress indicators
- [ ] Audit log viewer

---

## Success Criteria

### MVP Success (Week 4)

✅ **Functional Requirements:**
- [ ] User can register and login
- [ ] User can create team and invite members
- [ ] User can create project within team
- [ ] User can push secrets (encrypted for team)
- [ ] User can pull secrets (decrypt with own key)
- [ ] User can execute commands with secrets
- [ ] All secrets remain client-side encrypted

✅ **Technical Requirements:**
- [ ] 80%+ test coverage
- [ ] Zero known security vulnerabilities
- [ ] Documentation complete
- [ ] Binaries for Linux, macOS, Windows
- [ ] Installation script working

✅ **User Experience:**
- [ ] Clear error messages
- [ ] Intuitive command structure
- [ ] Fast operations (< 2s for most commands)
- [ ] Helpful prompts and confirmations

### Production Ready (Week 8)

✅ **Additional Requirements:**
- [ ] MFA support
- [ ] OAuth integration
- [ ] Comprehensive audit logging
- [ ] 95%+ test coverage
- [ ] Load tested (1000+ concurrent users)
- [ ] Security audit completed
- [ ] 10+ design partner deployments
- [ ] Monitoring & error tracking integrated

---

## Risk Assessment & Mitigation

### Technical Risks

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|-----------|
| SOPS integration issues | Medium | High | Thorough testing, maintain backward compatibility |
| Age key management complexity | Low | Medium | Clear documentation, automated key generation |
| Backend API changes | Low | High | Version API, maintain backward compatibility |
| Cross-platform compatibility | Medium | Medium | Test on all platforms, use CI/CD |

### Security Risks

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|-----------|
| Token theft | Low | High | Secure storage (0600), short expiry, MFA |
| Private key exposure | Low | Critical | File permissions, never send to backend |
| MITM attacks | Very Low | High | TLS only, certificate pinning (future) |
| Dependency vulnerabilities | Medium | Medium | Regular audits, auto-updates |

### User Experience Risks

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|-----------|
| Confusing command structure | Medium | Medium | User testing, clear documentation |
| Lost access to secrets | Low | Critical | Backup keys, team redundancy |
| Slow operations | Low | Medium | Optimize API calls, local caching |

---

## Open Questions

1. **Should we support multiple backends?**
   - Self-hosted envv backend?
   - Enterprise on-premise?
   - Decision: Start with single backend, add later

2. **Offline mode?**
   - Work without internet?
   - Cache encrypted secrets locally?
   - Decision: Phase 2 feature

3. **Migration from SOPS?**
   - Import existing SOPS files?
   - Convert to envv format?
   - Decision: Provide migration guide, maintain compatibility

4. **Web UI integration?**
   - Link to web dashboard from CLI?
   - Open browser for invite accept?
   - Decision: Nice-to-have, not MVP

---

## Conclusion

This proposal outlines a comprehensive plan to transform the envv-cli from a SOPS fork with demo scripts into a production-ready SaaS CLI that integrates seamlessly with the envv backend.

### Key Principles Maintained:
- ✅ **Zero-knowledge encryption**: Backend never sees plaintext
- ✅ **Client-side operations**: All encryption/decryption local
- ✅ **SOPS compatibility**: Leverage existing, battle-tested crypto
- ✅ **Security first**: Secure defaults, clear warnings

### Deliverables:
- Native CLI commands for auth, teams, projects, secrets
- Backend API integration
- Comprehensive testing
- User & developer documentation
- Cross-platform releases

### Timeline:
- **Week 1-2**: Foundation & team management
- **Week 3-4**: Secrets sync & testing
- **Week 5+**: Enhancements & production hardening

**Ready to proceed with implementation?**

---

**Next Steps:**
1. Review and approve this proposal
2. Set up development environment
3. Begin Phase 1 implementation (Backend API client)
4. Daily standups to track progress
5. Weekly demos to stakeholders

**Questions or concerns? Let's discuss before proceeding.**
