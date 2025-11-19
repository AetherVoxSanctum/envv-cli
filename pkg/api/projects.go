package api

import (
	"context"
	"fmt"
)

// Project represents a project
type Project struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Slug              string `json:"slug"`
	OrganizationID    string `json:"organization_id"`
	NeedsReencryption bool   `json:"needs_reencryption"`
	CreatedAt         string `json:"created_at"`
}

// CreateProjectRequest represents project creation request
type CreateProjectRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug,omitempty"`
	Description string `json:"description,omitempty"`
}

// CreateProject creates a new project in an organization
func (c *Client) CreateProject(ctx context.Context, orgID string, req CreateProjectRequest) (*Project, error) {
	var project Project
	path := fmt.Sprintf("/api/v1/organizations/%s/projects", orgID)
	err := c.Do(ctx, "POST", path, req, &project)
	return &project, err
}

// ListProjects returns all projects in an organization
func (c *Client) ListProjects(ctx context.Context, orgID string) ([]Project, error) {
	var projects []Project
	path := fmt.Sprintf("/api/v1/projects?organization_id=%s", orgID)
	err := c.Do(ctx, "GET", path, nil, &projects)
	return projects, err
}

// GetProject retrieves project details
func (c *Client) GetProject(ctx context.Context, projectID string) (*Project, error) {
	var project Project
	path := fmt.Sprintf("/api/v1/projects/%s", projectID)
	err := c.Do(ctx, "GET", path, nil, &project)
	return &project, err
}

// ProjectMember represents a project member with encryption key
type ProjectMember struct {
	UserID       string `json:"user_id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	AgePublicKey string `json:"age_public_key"` // CRITICAL for encryption!
	Permission   string `json:"permission"`     // read, write, admin
}

// ProjectMembersResponse represents project members response
type ProjectMembersResponse struct {
	ProjectID string          `json:"project_id"`
	Members   []ProjectMember `json:"members"`
	Total     int             `json:"total"`
}

// GetProjectMembers retrieves project members with their public keys
func (c *Client) GetProjectMembers(ctx context.Context, projectID string) (*ProjectMembersResponse, error) {
	var resp ProjectMembersResponse
	path := fmt.Sprintf("/api/v1/projects/%s/members", projectID)
	err := c.Do(ctx, "GET", path, nil, &resp)
	return &resp, err
}

// GrantAccessRequest represents access grant request
type GrantAccessRequest struct {
	Email      string `json:"email"`
	Permission string `json:"permission"` // read, write, admin
}

// GrantAccess grants project access to a user
func (c *Client) GrantAccess(ctx context.Context, projectID string, req GrantAccessRequest) error {
	path := fmt.Sprintf("/api/v1/projects/%s/access", projectID)
	return c.Do(ctx, "POST", path, req, nil)
}

// RevokeAccess revokes project access from a user
func (c *Client) RevokeAccess(ctx context.Context, projectID, userID string) error {
	path := fmt.Sprintf("/api/v1/projects/%s/access/%s", projectID, userID)
	return c.Do(ctx, "DELETE", path, nil, nil)
}
