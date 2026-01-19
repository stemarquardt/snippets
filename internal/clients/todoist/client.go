package todoist

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	BaseURL = "https://api.todoist.com/api/v1"
)

type Client struct {
	httpClient *http.Client
	token      string
	baseURL    string
	projects   map[string]Project
}

type APIError struct {
	StatusCode int
	Message    string
}

func NewClient(token string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		token:    token,
		baseURL:  BaseURL,
		projects: map[string]Project{},
	}
}

func (e APIError) Error() string {
	return fmt.Sprintf("Todoist API error (status %d): %s", e.StatusCode, e.Message)
}

func (c *Client) GetProductivityStats(opts TodoistAPIOpts) (ProductivityStats, error) {
	// Not sure what to do with this info, but could be fun!
	resp, err := c.doGetRequest("/tasks/completed/stats", TodoistAPIOpts{})
	if err != nil {
		return ProductivityStats{}, fmt.Errorf("problem getting productivity stats: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ProductivityStats{}, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}
	var stats ProductivityStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return ProductivityStats{}, fmt.Errorf("failed to decode completed tasks response: %w", err)
	}
	return stats, nil
}

func (c *Client) doGetRequest(endpoint string, opts TodoistAPIOpts) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.baseURL+endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)

	q := req.URL.Query()
	if opts.ProjectID != "" {
		q.Add("project_id", opts.ProjectID)
	}

	if !opts.Since.IsZero() {
		q.Add("since", opts.Since.Format("2006-01-02T15:04"))
	}

	if !opts.Until.IsZero() {
		q.Add("until", opts.Until.Format("2006-01-02T15:04"))
	}

	if opts.Limit > 0 {
		q.Add("limit", strconv.Itoa(opts.Limit))
	}

	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, APIError{
			StatusCode: resp.StatusCode,
			Message:    string(bodyBytes),
		}
	}

	return resp, nil
}

// func (c *Client) doRequest(method, endpoint string, opts TodoistAPIOpts) (*http.Response, error) {
// 	if method == "GET" {
// 		return c.doGetRequest(endpoint, opts)
// 	}

// 	req, err := http.NewRequest(method, c.baseURL+endpoint, reqBody)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create request: %w", err)
// 	}

// 	req.Header.Set("Authorization", "Bearer "+c.token)
// 	if body != nil {
// 		req.Header.Set("Content-Type", "application/json")
// 	}

// 	resp, err := c.httpClient.Do(req)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to execute request: %w", err)
// 	}

// 	if resp.StatusCode >= 400 {
// 		defer resp.Body.Close()
// 		bodyBytes, _ := io.ReadAll(resp.Body)
// 		return nil, APIError{
// 			StatusCode: resp.StatusCode,
// 			Message:    string(bodyBytes),
// 		}
// 	}

// 	return resp, nil
// }

func (c *Client) doFormRequest(endpoint string, formData url.Values) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+endpoint, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("Failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	fmt.Printf("Req: %+v", req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to execute request: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, APIError{
			StatusCode: resp.StatusCode,
			Message:    string(bodyBytes),
		}
	}

	return resp, nil
}

func (c *Client) ValidateToken() error {
	_, err := c.GetProductivityStats(TodoistAPIOpts{})
	return err
}
