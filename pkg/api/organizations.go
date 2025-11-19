package api

import (
	"context"
	"fmt"
)

// Organization represents an organization
type Organization struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Role        string `json:"role"` // owner, admin, member
	MemberCount int    `json:"member_count"`
	CreatedAt   string `json:"created_at"`
}

// CreateOrgRequest represents organization creation request
type CreateOrgRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// CreateOrganization creates a new organization
func (c *Client) CreateOrganization(ctx context.Context, req CreateOrgRequest) (*Organization, error) {
	var org Organization
	err := c.Do(ctx, "POST", "/api/v1/organizations", req, &org)
	return &org, err
}

// ListOrganizations returns all organizations the user belongs to
func (c *Client) ListOrganizations(ctx context.Context) ([]Organization, error) {
	var orgs []Organization
	err := c.Do(ctx, "GET", "/api/v1/organizations", nil, &orgs)
	return orgs, err
}

// GetOrganization retrieves organization details
func (c *Client) GetOrganization(ctx context.Context, orgID string) (*Organization, error) {
	var org Organization
	path := fmt.Sprintf("/api/v1/organizations/%s", orgID)
	err := c.Do(ctx, "GET", path, nil, &org)
	return &org, err
}

// OrgMember represents an organization member
type OrgMember struct {
	UserID       string `json:"user_id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	AgePublicKey string `json:"age_public_key"` // CRITICAL for encryption!
	Role         string `json:"role"`
}

// GetOrganizationMemberKeys retrieves all member public keys for encryption
func (c *Client) GetOrganizationMemberKeys(ctx context.Context, orgID string) ([]OrgMember, error) {
	var members []OrgMember
	path := fmt.Sprintf("/api/v1/organizations/%s/members/keys", orgID)
	err := c.Do(ctx, "GET", path, nil, &members)
	return members, err
}

// InviteMemberRequest represents member invitation request
type InviteMemberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"` // admin, member
}

// InviteMember invites a user to the organization
func (c *Client) InviteMember(ctx context.Context, orgID string, req InviteMemberRequest) error {
	path := fmt.Sprintf("/api/v1/organizations/%s/invites", orgID)
	return c.Do(ctx, "POST", path, req, nil)
}
