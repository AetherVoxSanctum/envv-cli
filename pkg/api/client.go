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

// Client handles all backend API communication
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	token      string
}

// NewClient creates a new API client
func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
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

// Do makes an HTTP request to the backend
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

// APIError represents an error from the backend API
type APIError struct {
	StatusCode int
	Code       string
	Message    string
	Details    map[string]interface{}
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("API error %d (%s): %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("API error %d (%s)", e.StatusCode, e.Code)
}

// IsUnauthorized checks if error is 401 Unauthorized
func (e *APIError) IsUnauthorized() bool {
	return e.StatusCode == 401
}

// IsForbidden checks if error is 403 Forbidden
func (e *APIError) IsForbidden() bool {
	return e.StatusCode == 403
}

// IsNotFound checks if error is 404 Not Found
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == 404
}
