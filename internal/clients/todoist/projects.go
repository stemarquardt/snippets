package todoist

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *Client) SetProjects(projs []Project) {
	for _, p := range projs {
		c.Projects[p.ID] = &p
	}
}

func (c *Client) DelAllProjects() {
	c.Projects = map[string]*Project{}
}

func (c *Client) DelProject(p Project) {
	c.Projects[p.ID] = &Project{}
}

func (c *Client) AddProject(p Project) {
	// Overwrites existing project in map, they shouldn't change but beware.
	c.Projects[p.ID] = &p
}

func (c *Client) GetProject(ctx context.Context, projectId string) (*Project, error) {
	p, ok := c.Projects[projectId]
	if ok {
		return p, nil
	}
	return c.getProjectFromAPI(ctx, projectId)
}

func (c *Client) GetProjects(ctx context.Context) ([]Project, error) {
	if c.Projects != nil {
		// This means the project flags were set, so we're filtering for these Projects only.
		p := make([]Project, 0, len(c.Projects))
		for k := range c.Projects {
			p = append(p, *c.Projects[k])
		}
		return p, nil
	}
	// Desired projects not set, or requesting every project for this user.
	return c.getAllProjectsFromAPI(ctx)
}

func (c *Client) getAllProjectsFromAPI(ctx context.Context) ([]Project, error) {
	resp, err := c.doGetRequest(ctx, "/projects", TodoistAPIOpts{})
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

func (c *Client) getProjectFromAPI(ctx context.Context, projectID string) (*Project, error) {
	resp, err := c.doGetRequest(ctx, fmt.Sprintf("/projects/%s", projectID), TodoistAPIOpts{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var p Project
	if err = json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return nil, err
	}

	return &p, nil
}
