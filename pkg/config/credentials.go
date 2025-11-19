package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const CredentialsFile = "~/.envv/credentials.json"

// Credentials represents user authentication credentials
type Credentials struct {
	AccessToken string `json:"access_token"`
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	ExpiresAt   string `json:"expires_at"`
}

// LoadCredentials reads credentials from disk
func LoadCredentials() (*Credentials, error) {
	path := expandPath(CredentialsFile)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("not logged in. Run 'envv auth login'")
		}
		return nil, fmt.Errorf("failed to read credentials: %w", err)
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("corrupted credentials file: %w", err)
	}

	return &creds, nil
}

// SaveCredentials writes credentials to disk
func SaveCredentials(creds *Credentials) error {
	path := expandPath(CredentialsFile)

	// Create directory
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write credentials
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	// Write with restricted permissions
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write credentials: %w", err)
	}

	return nil
}

// ClearCredentials removes the credentials file
func ClearCredentials() error {
	path := expandPath(CredentialsFile)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove credentials: %w", err)
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
