package openclaw

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Client struct {
	gatewayURL   string
	gatewayToken string
	httpClient   *http.Client
}

type Config struct {
	GatewayURL   string
	GatewayToken string
}

// NewClient creates a new OpenClaw Gateway client
func NewClient(cfg *Config) *Client {
	return &Client{
		gatewayURL:   cfg.GatewayURL,
		gatewayToken: cfg.GatewayToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewClientFromEnv creates client from environment variables
func NewClientFromEnv() (*Client, error) {
	url := os.Getenv("OPENCLAW_GATEWAY_URL")
	token := os.Getenv("OPENCLAW_GATEWAY_TOKEN")

	if url == "" {
		url = "ws://127.0.0.1:18789"
	}

	// Try to load from openclaw config if token not set
	if token == "" {
		token, _ = loadTokenFromConfig()
	}

	return &Client{
		gatewayURL:   url,
		gatewayToken: token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func loadTokenFromConfig() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configPath := filepath.Join(home, ".openclaw", "openclaw.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", err
	}

	var config struct {
		Gateway struct {
			Auth struct {
				Token string `json:"token"`
			} `json:"auth"`
		} `json:"gateway"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return "", err
	}

	return config.Gateway.Auth.Token, nil
}

// SpawnRequest represents a request to spawn a sub-agent session
type SpawnRequest struct {
	Task           string `json:"task"`
	AgentID        string `json:"agentId,omitempty"`
	Label          string `json:"label,omitempty"`
	Model          string `json:"model,omitempty"`
	Cleanup        string `json:"cleanup,omitempty"` // "delete" or "keep"
	TimeoutSeconds int    `json:"runTimeoutSeconds,omitempty"`
}

// SpawnResponse represents the response from spawning a session
type SpawnResponse struct {
	Status          string `json:"status"`
	ChildSessionKey string `json:"childSessionKey"`
	RunID           string `json:"runId"`
}

// ToolInvokeRequest is the request format for /tools/invoke
type ToolInvokeRequest struct {
	Tool       string      `json:"tool"`
	Args       interface{} `json:"args"`
	SessionKey string      `json:"sessionKey,omitempty"`
}

// ToolInvokeResponse is the response format from /tools/invoke
type ToolInvokeResponse struct {
	OK     bool            `json:"ok"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Spawn creates a new sub-agent session using the /tools/invoke endpoint
func (c *Client) Spawn(ctx context.Context, req *SpawnRequest) (*SpawnResponse, error) {
	// OpenClaw uses HTTP API at the gateway URL (convert ws:// to http://)
	baseURL := c.gatewayURL
	if len(baseURL) > 5 && baseURL[:5] == "ws://" {
		baseURL = "http://" + baseURL[5:]
	} else if len(baseURL) > 6 && baseURL[:6] == "wss://" {
		baseURL = "https://" + baseURL[6:]
	}

	// Use /tools/invoke with sessions_spawn tool
	url := fmt.Sprintf("%s/tools/invoke", baseURL)

	invokeReq := ToolInvokeRequest{
		Tool: "sessions_spawn",
		Args: req,
	}

	body, err := json.Marshal(invokeReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.gatewayToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.gatewayToken)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("spawn failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the tool invoke response
	var invokeResp ToolInvokeResponse
	if err := json.NewDecoder(resp.Body).Decode(&invokeResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !invokeResp.OK {
		errMsg := "unknown error"
		if invokeResp.Error != nil {
			errMsg = invokeResp.Error.Message
		}
		return nil, fmt.Errorf("spawn failed: %s", errMsg)
	}

	// Parse the result into SpawnResponse
	var spawnResp SpawnResponse
	if err := json.Unmarshal(invokeResp.Result, &spawnResp); err != nil {
		return nil, fmt.Errorf("failed to decode spawn result: %w", err)
	}

	return &spawnResp, nil
}

// SendMessage sends a message to an existing session using /tools/invoke
func (c *Client) SendMessage(ctx context.Context, sessionKey, message string) error {
	baseURL := c.gatewayURL
	if len(baseURL) > 5 && baseURL[:5] == "ws://" {
		baseURL = "http://" + baseURL[5:]
	}

	url := fmt.Sprintf("%s/tools/invoke", baseURL)

	invokeReq := ToolInvokeRequest{
		Tool: "sessions_send",
		Args: map[string]interface{}{
			"sessionKey": sessionKey,
			"message":    message,
		},
	}

	body, err := json.Marshal(invokeReq)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.gatewayToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.gatewayToken)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("send failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response to check for errors
	var invokeResp ToolInvokeResponse
	if err := json.NewDecoder(resp.Body).Decode(&invokeResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !invokeResp.OK {
		errMsg := "unknown error"
		if invokeResp.Error != nil {
			errMsg = invokeResp.Error.Message
		}
		return fmt.Errorf("send failed: %s", errMsg)
	}

	return nil
}

// SessionMessage represents a message from session history
type SessionMessage struct {
	Role      string `json:"role"`      // "user" or "assistant"
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

// SessionHistoryResponse represents the response from sessions_history
type SessionHistoryResponse struct {
	SessionKey string           `json:"sessionKey"`
	Messages   []SessionMessage `json:"messages"`
}

// GetSessionHistory retrieves message history for a session using /tools/invoke
func (c *Client) GetSessionHistory(ctx context.Context, sessionKey string, limit int) (*SessionHistoryResponse, error) {
	baseURL := c.gatewayURL
	if len(baseURL) > 5 && baseURL[:5] == "ws://" {
		baseURL = "http://" + baseURL[5:]
	}

	url := fmt.Sprintf("%s/tools/invoke", baseURL)

	args := map[string]interface{}{
		"sessionKey": sessionKey,
	}
	if limit > 0 {
		args["limit"] = limit
	}

	invokeReq := ToolInvokeRequest{
		Tool: "sessions_history",
		Args: args,
	}

	body, err := json.Marshal(invokeReq)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.gatewayToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.gatewayToken)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("history failed with status %d: %s", resp.StatusCode, string(body))
	}

	var invokeResp ToolInvokeResponse
	if err := json.NewDecoder(resp.Body).Decode(&invokeResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !invokeResp.OK {
		errMsg := "unknown error"
		if invokeResp.Error != nil {
			errMsg = invokeResp.Error.Message
		}
		return nil, fmt.Errorf("history failed: %s", errMsg)
	}

	var historyResp SessionHistoryResponse
	if err := json.Unmarshal(invokeResp.Result, &historyResp); err != nil {
		return nil, fmt.Errorf("failed to decode history result: %w", err)
	}

	return &historyResp, nil
}

// GetStatus checks the gateway connection status
func (c *Client) GetStatus(ctx context.Context) (bool, error) {
	baseURL := c.gatewayURL
	if len(baseURL) > 5 && baseURL[:5] == "ws://" {
		baseURL = "http://" + baseURL[5:]
	}

	url := fmt.Sprintf("%s/health", baseURL)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, err
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// ListAgents returns configured agents from OpenClaw config
type AgentInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Workspace string `json:"workspace"`
	Model     string `json:"model"`
}

func (c *Client) ListAgentsFromConfig() ([]AgentInfo, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(home, ".openclaw", "openclaw.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config struct {
		Agents struct {
			List []AgentInfo `json:"list"`
		} `json:"agents"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return config.Agents.List, nil
}
