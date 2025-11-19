package api

import (
	"context"
	"fmt"
)

// PushSecretsRequest represents secrets push request
type PushSecretsRequest struct {
	EncryptedData string                 `json:"encrypted_data"`
	Format        string                 `json:"format"`        // env, json, yaml
	Environment   string                 `json:"environment"`   // production, staging, dev
	SOPSMetadata  map[string]interface{} `json:"sops_metadata"` // CRITICAL: SOPS metadata
}

// PushSecretsResponse represents secrets push response
type PushSecretsResponse struct {
	Message     string `json:"message"`
	VersionID   string `json:"version_id"`
	Version     int    `json:"version"`
	Environment string `json:"environment"`
	SizeBytes   int    `json:"size_bytes"`
}

// PushSecrets uploads encrypted secrets to backend
func (c *Client) PushSecrets(ctx context.Context, projectID string, req PushSecretsRequest) (*PushSecretsResponse, error) {
	var resp PushSecretsResponse
	path := fmt.Sprintf("/api/v1/projects/%s/secrets", projectID)
	err := c.Do(ctx, "POST", path, req, &resp)
	return &resp, err
}

// PullSecretsResponse represents secrets pull response
type PullSecretsResponse struct {
	VersionID     string                 `json:"version_id"`
	EncryptedData string                 `json:"encrypted_data"`
	Format        string                 `json:"format"`
	Version       int                    `json:"version"`
	Environment   string                 `json:"environment"`
	SOPSMetadata  map[string]interface{} `json:"sops_metadata"`
	SizeBytes     int                    `json:"size_bytes"`
	CreatedBy     string                 `json:"created_by"`
	CreatedAt     string                 `json:"created_at"`
}

// PullSecrets downloads encrypted secrets from backend
func (c *Client) PullSecrets(ctx context.Context, projectID, environment string) (*PullSecretsResponse, error) {
	var resp PullSecretsResponse
	path := fmt.Sprintf("/api/v1/projects/%s/secrets?environment=%s", projectID, environment)
	err := c.Do(ctx, "GET", path, nil, &resp)
	return &resp, err
}

// SecretVersion represents a version in history
type SecretVersion struct {
	VersionID   string `json:"version_id"`
	Version     int    `json:"version"`
	Environment string `json:"environment"`
	CreatedBy   string `json:"created_by"`
	CreatedAt   string `json:"created_at"`
	SizeBytes   int    `json:"size_bytes"`
}

// ListVersions retrieves version history
func (c *Client) ListVersions(ctx context.Context, projectID, environment string) ([]SecretVersion, error) {
	var versions []SecretVersion
	path := fmt.Sprintf("/api/v1/projects/%s/secrets/versions?environment=%s", projectID, environment)
	err := c.Do(ctx, "GET", path, nil, &versions)
	return versions, err
}

// RollbackRequest represents rollback request
type RollbackRequest struct {
	VersionID string `json:"version_id"`
}

// Rollback rolls back secrets to a previous version
func (c *Client) Rollback(ctx context.Context, projectID string, req RollbackRequest) (*PushSecretsResponse, error) {
	var resp PushSecretsResponse
	path := fmt.Sprintf("/api/v1/projects/%s/secrets/rollback", projectID)
	err := c.Do(ctx, "POST", path, req, &resp)
	return &resp, err
}
