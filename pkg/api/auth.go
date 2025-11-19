package api

import "context"

// RegisterEnhancedRequest matches backend expectation
type RegisterEnhancedRequest struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	Name         string `json:"name"`
	AgePublicKey string `json:"age_public_key"` // CRITICAL: age key required!
}

// AuthResponse represents authentication response
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

// LoginRequest represents login credentials
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login authenticates the user
func (c *Client) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	var resp AuthResponse
	err := c.Do(ctx, "POST", "/api/v1/auth/login", req, &resp)
	return &resp, err
}

// Logout invalidates the current session
func (c *Client) Logout(ctx context.Context) error {
	return c.Do(ctx, "POST", "/api/v1/auth/logout", nil, nil)
}

// RefreshRequest represents token refresh request
type RefreshRequest struct {
	Token string `json:"token"`
}

// RefreshToken refreshes the JWT token
func (c *Client) RefreshToken(ctx context.Context, req RefreshRequest) (*AuthResponse, error) {
	var resp AuthResponse
	err := c.Do(ctx, "POST", "/api/v1/auth/refresh", req, &resp)
	return &resp, err
}

// GetCurrentUser retrieves current user info
func (c *Client) GetCurrentUser(ctx context.Context) (*AuthResponse, error) {
	var resp AuthResponse
	err := c.Do(ctx, "GET", "/api/v1/auth/me", nil, &resp)
	return &resp, err
}
