package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/einarsundgren/sikta/internal/config"
)

const (
	apiVersion = "2023-06-01"
	maxRetries = 3
	baseRetryDelay = 1 * time.Second
)

// Client handles communication with the Claude API.
type Client struct {
	httpClient *http.Client
	apiKey     string
	apiURL     string
	logger     *slog.Logger
}

// NewClient creates a new Claude API client.
func NewClient(cfg *config.Config, logger *slog.Logger) *Client {
	apiURL := cfg.AnthropicAPIURL
	if apiURL == "" {
		apiURL = "https://api.anthropic.com"
	}
	apiURL = apiURL + "/v1/messages"

	return &Client{
		httpClient: &http.Client{
			Timeout: 300 * time.Second,
		},
		apiKey: cfg.AnthropicAPIKey,
		apiURL: apiURL,
		logger: logger,
	}
}

// Message represents a message in the conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Request represents an API request.
type Request struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
	System    string    `json:"system,omitempty"`
}

// Response represents an API response.
type Response struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason   string `json:"stop_reason"`
	StopSequence *int    `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// SendMessage sends a message to Claude and returns the response.
func (c *Client) SendMessage(ctx context.Context, req Request) (*Response, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY not configured")
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			c.logger.Debug("retrying request", "attempt", attempt+1, "error", lastErr)
			time.Sleep(baseRetryDelay * time.Duration(1<<attempt))
		}

		response, err := c.doRequest(ctx, reqBody)
		if err == nil {
			return response, nil
		}

		lastErr = err

		// Don't retry authentication errors
		if isAuthError(err) {
			return nil, err
		}
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// doRequest performs a single HTTP request.
func (c *Client) doRequest(ctx context.Context, reqBody []byte) (*Response, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", c.apiURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", apiVersion)

	// Log HTTP request details to stderr
	maskedKey := c.apiKey
	if len(maskedKey) > 10 {
		maskedKey = maskedKey[:10] + "..."
	}
	fmt.Fprintf(os.Stderr, "=== HTTP REQUEST ===\n")
	fmt.Fprintf(os.Stderr, "URL: %s\n", req.URL.String())
	fmt.Fprintf(os.Stderr, "Method: %s\n", req.Method)
	fmt.Fprintf(os.Stderr, "Headers:\n")
	fmt.Fprintf(os.Stderr, "  Content-Type: %s\n", req.Header.Get("Content-Type"))
	fmt.Fprintf(os.Stderr, "  x-api-key: %s\n", maskedKey)
	fmt.Fprintf(os.Stderr, "  anthropic-version: %s\n", req.Header.Get("anthropic-version"))
	fmt.Fprintf(os.Stderr, "Body length: %d bytes\n", len(reqBody))
	fmt.Fprintf(os.Stderr, "=== END HTTP REQUEST ===\n\n")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "HTTP request failed: %v\n", err)
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read response body: %v\n", err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Log HTTP response details
	fmt.Fprintf(os.Stderr, "=== HTTP RESPONSE ===\n")
	fmt.Fprintf(os.Stderr, "Status: %d %s\n", resp.StatusCode, resp.Status)
	fmt.Fprintf(os.Stderr, "Content-Type: %s\n", resp.Header.Get("Content-Type"))
	fmt.Fprintf(os.Stderr, "Body length: %d bytes\n", len(body))
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Error body: %s\n", string(body))
	}
	fmt.Fprintf(os.Stderr, "=== END HTTP RESPONSE ===\n\n")

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var apiResp Response
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	c.logger.Debug("API call successful",
		"input_tokens", apiResp.Usage.InputTokens,
		"output_tokens", apiResp.Usage.OutputTokens)

	return &apiResp, nil
}

// SendSystemPrompt sends a message with a system prompt.
func (c *Client) SendSystemPrompt(ctx context.Context, systemPrompt, userMessage string, model string) (*Response, error) {
	req := Request{
		Model:     model,
		MaxTokens: 8192, // Increased from 4096 to prevent truncation on long documents
		System:    systemPrompt,
		Messages: []Message{
			{
				Role:    "user",
				Content: userMessage,
			},
		},
	}

	return c.SendMessage(ctx, req)
}

// isAuthError checks if an error is an authentication error.
func isAuthError(err error) bool {
	return fmt.Sprintf("%v", err) == "API error (status 401)"
}
