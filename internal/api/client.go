package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/seth4242/snet/internal/config"
)

// Client is the API client for seth4242.net
type Client struct {
	baseURL    string
	token      string
	accountID  string
	httpClient *http.Client
}

// NewClient creates a new API client from config
func NewClient(cfg *config.Config) *Client {
	return &Client{
		baseURL:   cfg.APIBase,
		token:     cfg.APIToken,
		accountID: cfg.AccountID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewClientWithToken creates a client with just a token (for login verification)
func NewClientWithToken(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Tunnel represents a tunnel from the API
type Tunnel struct {
	ID                    string    `json:"id"`
	Slug                  string    `json:"slug"`
	Name                  string    `json:"name"`
	URL                   string    `json:"url"`
	WildcardURL           string    `json:"wildcard_url,omitempty"`
	Port                  int       `json:"port,omitempty"`
	Wildcard              bool      `json:"wildcard"`
	Persistent            bool      `json:"persistent"`
	Status                string    `json:"status"`
	LastHeartbeatAt       *string   `json:"last_heartbeat_at,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
	CloudflareTunnelToken string    `json:"cloudflare_tunnel_token,omitempty"`
	CloudflareTunnelID    string    `json:"cloudflare_tunnel_id,omitempty"`
	Account               Account   `json:"account"`
}

// Account represents an account from the API
type Account struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

// CreateTunnelRequest is the request body for creating a tunnel
type CreateTunnelRequest struct {
	Name       string `json:"name,omitempty"`
	Wildcard   bool   `json:"wildcard"`
	Persistent bool   `json:"persistent"`
}

// ConnectRequest is the request body for reporting a connection
type ConnectRequest struct {
	CloudflaredVersion string                 `json:"cloudflared_version,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// ConnectResponse is the response from the connect endpoint
type ConnectResponse struct {
	ConnectionID int    `json:"connection_id"`
	Status       string `json:"status"`
}

// DisconnectRequest is the request body for reporting a disconnection
type DisconnectRequest struct {
	ConnectionID int   `json:"connection_id"`
	BytesIn      int64 `json:"bytes_in,omitempty"`
	BytesOut     int64 `json:"bytes_out,omitempty"`
}

// HeartbeatResponse is the response from the heartbeat endpoint
type HeartbeatResponse struct {
	Status          string `json:"status"`
	LastHeartbeatAt string `json:"last_heartbeat_at"`
}

// APIError represents an error from the API
type APIError struct {
	Error  string   `json:"error,omitempty"`
	Errors []string `json:"errors,omitempty"`
}

func (e *APIError) String() string {
	if e.Error != "" {
		return e.Error
	}
	if len(e.Errors) > 0 {
		return e.Errors[0]
	}
	return "unknown error"
}

// do makes an HTTP request and handles the response
func (c *Client) do(method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	url := fmt.Sprintf("%s%s", c.baseURL, path)
	if c.accountID != "" {
		if bytes.Contains([]byte(path), []byte("?")) {
			url += "&account_id=" + c.accountID
		} else {
			url += "?account_id=" + c.accountID
		}
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var apiErr APIError
		if err := json.Unmarshal(respBody, &apiErr); err == nil {
			return fmt.Errorf("API error (%d): %s", resp.StatusCode, apiErr.String())
		}
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

// ListAccounts returns all accounts for the authenticated user
func (c *Client) ListAccounts() ([]Account, error) {
	var accounts []Account
	// Don't include account_id for this request
	originalAccountID := c.accountID
	c.accountID = ""
	err := c.do("GET", "/accounts", nil, &accounts)
	c.accountID = originalAccountID
	return accounts, err
}

// ListTunnels returns all tunnels for the current account
func (c *Client) ListTunnels() ([]Tunnel, error) {
	var tunnels []Tunnel
	err := c.do("GET", "/tunnels", nil, &tunnels)
	return tunnels, err
}

// GetTunnel returns a specific tunnel with its cloudflared token
func (c *Client) GetTunnel(id string) (*Tunnel, error) {
	var tunnel Tunnel
	err := c.do("GET", "/tunnels/"+id, nil, &tunnel)
	return &tunnel, err
}

// CreateTunnel creates a new tunnel
func (c *Client) CreateTunnel(req *CreateTunnelRequest) (*Tunnel, error) {
	var tunnel Tunnel
	err := c.do("POST", "/tunnels", req, &tunnel)
	return &tunnel, err
}

// DeleteTunnel deletes a tunnel
func (c *Client) DeleteTunnel(id string) error {
	return c.do("DELETE", "/tunnels/"+id, nil, nil)
}

// Heartbeat sends a heartbeat for a tunnel
func (c *Client) Heartbeat(id string) (*HeartbeatResponse, error) {
	var resp HeartbeatResponse
	err := c.do("POST", "/tunnels/"+id+"/heartbeat", nil, &resp)
	return &resp, err
}

// Connect reports a new connection
func (c *Client) Connect(id string, req *ConnectRequest) (*ConnectResponse, error) {
	var resp ConnectResponse
	err := c.do("POST", "/tunnels/"+id+"/connect", req, &resp)
	return &resp, err
}

// Disconnect reports a disconnection
func (c *Client) Disconnect(id string, req *DisconnectRequest) error {
	return c.do("POST", "/tunnels/"+id+"/disconnect", req, nil)
}
