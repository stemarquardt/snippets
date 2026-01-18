package todoist

import (
	"encoding/json"
	"fmt"
)

// NOTES:
// `parent_id` denotes a subtask, it'll show up like `"parent_id": "6fpCwRh45C7pXgHm"`, probably useful for gathering context about tasks

func (c *Client) GetTasksForProj(p string) ([]Task, error) {
	endpoint := "/tasks"
	resp, err := c.doGetRequest(endpoint, TodoistAPIOpts{ProjectID: p})
	if err != nil {
		return nil, fmt.Errorf("Failed to make all tasks tasks request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Non-200 status code (%d), resp: %s", resp.StatusCode, resp.Body)
	}
	var allTasksResp struct {
		Results []Task `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&allTasksResp); err != nil {
		return nil, fmt.Errorf("failed to decode completed tasks response: %w", err)
	}

	return allTasksResp.Results, nil
}

func (c *Client) GetAllTasks() ([]Task, error) {
	return c.GetTasksForProj("")
}

func (c *Client) GetTasksByProject(projectID string) ([]Task, error) {
	return c.GetTasksForProj(projectID)
}

func (c *Client) GetTaskNames(tasks []Task) []string {
	var names []string
	for _, t := range tasks {
		names = append(names, t.Content)
	}
	return names
}
