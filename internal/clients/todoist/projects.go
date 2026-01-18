package todoist

import (
	"encoding/json"
	"fmt"
)

func (c *Client) GetProjects() ([]Project, error) {
	resp, err := c.doGetRequest("/projects", TodoistAPIOpts{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var d struct {
		Results []Project `json:"results"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return nil, err
	}
	return d.Results, nil
}

func (c *Client) GetProject(projectID string) (*Project, error) {
	projects, err := c.GetProjects()
	if err != nil {
		return nil, err
	}

	for _, project := range projects {
		if project.ID == projectID {
			return &project, nil
		}
	}

	return nil, fmt.Errorf("project with ID %s not found", projectID)
}
