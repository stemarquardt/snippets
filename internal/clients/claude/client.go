package claude

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	BaseURL     = "https://api.anthropic.com/v1"
	APIVersion  = "2023-06-01"
	// Using Claude 3 Haiku - most cost-effective model for summarization tasks
	// Input: $0.25 per million tokens, Output: $1.25 per million tokens
	ModelHaiku  = "claude-3-haiku-20240307"
	// Claude 3 Sonnet - higher quality but more expensive
	ModelSonnet = "claude-3-sonnet-20240229"
)

type Client struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
}

type APIError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("Claude API error (%s): %s", e.Type, e.Message)
}

func NewClient(apiKey string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		apiKey:  apiKey,
		baseURL: BaseURL,
	}
}

func (c *Client) ValidateAPIKey() error {
	messages := []Message{
		{Role: "user", Content: "Hello"},
	}

	_, err := c.sendMessage(messages, "", ModelHaiku, 10)
	return err
}

func (c *Client) doRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, c.baseURL+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", APIVersion)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		var apiErr APIError
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
		}
		return nil, apiErr
	}

	return resp, nil
}