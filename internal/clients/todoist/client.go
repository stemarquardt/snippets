package todoist

import (
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
	projects   []Project
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
		token:   token,
		baseURL: BaseURL,
	}
}

func (e APIError) Error() string {
	return fmt.Sprintf("Todoist API error (status %d): %s", e.StatusCode, e.Message)
}

func (c *Client) SetProjects(projs []Project) {
	c.projects = projs
}

func (c *Client) AddProject(p Project) {
	c.projects = append(c.projects, p)
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
	_, err := c.GetProjects()
	return err
}
