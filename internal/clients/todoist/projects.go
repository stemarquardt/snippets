package todoist

import (
	"encoding/json"
	"fmt"
)

func (c *Client) SetProjects(projs []Project) {
	for _, p := range projs {
		c.projects[p.ID] = p
	}
}

func (c *Client) DelProjects() {
	c.projects = map[string]Project{}
}

func (c *Client) DelProject(p Project) {
	c.projects[p.ID] = Project{}
}

func (c *Client) AddProject(p Project) {
	// Overwrites existing project in map, they shouldn't change but beware.
	c.projects[p.ID] = p
}

func (c *Client) GetProject(projectId string) (*Project, error) {
	p, ok := c.projects[projectId]
	if ok {
		return &p, nil
	}
	return c.getProjectFromAPI(projectId)
}

func (c *Client) GetProjects() ([]Project, error) {
	if c.projects != nil {
		// This means the project flags were set, so we're filtering for these projects only.
		p := make([]Project, 0, len(c.projects))
		for k := range c.projects {
			p = append(p, c.projects[k])
		}
		return p, nil
	}
	// Desired projects not set, or requesting every project for this user.
	return c.getAllProjectsFromAPI()
}

func (c *Client) getAllProjectsFromAPI() ([]Project, error) {
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

func (c *Client) getProjectFromAPI(projectID string) (*Project, error) {
	resp, err := c.doGetRequest(fmt.Sprintf("/projects/%s", projectID), TodoistAPIOpts{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("Project (%s) not found", projectID)
	}

	var p Project
	if err = json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return nil, err
	}

	return &p, nil
}
