package todoist

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
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

// GetProjects returns the projects currently loaded in the client.
func (c *Client) GetProjects() []Project {
	p := make([]Project, 0, len(c.Projects))
	for _, proj := range c.Projects {
		p = append(p, *proj)
	}
	return p
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

// ResolveProjectRefs resolves a slice of project names or IDs to Project values.
// Each ref is matched in order: exact ID → exact name (case-insensitive) → substring name.
// Returns an error if a ref matches no projects or is ambiguous.
// c.Projects must already be populated (e.g. via NewClient with no ID filter).
func (c *Client) ResolveProjectRefs(refs []string) ([]Project, error) {
	var resolved []Project
	for _, ref := range refs {
		ref = strings.TrimSpace(ref)
		if ref == "" {
			continue
		}
		p, err := c.resolveRef(ref)
		if err != nil {
			return nil, err
		}
		resolved = append(resolved, p)
	}
	return resolved, nil
}

func (c *Client) resolveRef(ref string) (Project, error) {
	// Exact ID match.
	if p, ok := c.Projects[ref]; ok {
		return *p, nil
	}

	// Exact name match (case-insensitive).
	refLower := strings.ToLower(ref)
	for _, p := range c.Projects {
		if strings.ToLower(p.Name) == refLower {
			return *p, nil
		}
	}

	// Fuzzy: substring match on name.
	var matches []Project
	for _, p := range c.Projects {
		if strings.Contains(strings.ToLower(p.Name), refLower) {
			matches = append(matches, *p)
		}
	}
	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		return Project{}, fmt.Errorf("no project found matching %q", ref)
	default:
		sort.Slice(matches, func(i, j int) bool { return matches[i].Name < matches[j].Name })
		labels := make([]string, len(matches))
		for i, m := range matches {
			labels[i] = fmt.Sprintf("%q (ID: %s)", m.Name, m.ID)
		}
		return Project{}, fmt.Errorf("ambiguous project ref %q — matches: %s", ref, strings.Join(labels, ", "))
	}
}

// SortedProjects returns all loaded projects sorted alphabetically by name.
func (c *Client) SortedProjects() []Project {
	projs := make([]Project, 0, len(c.Projects))
	for _, p := range c.Projects {
		projs = append(projs, *p)
	}
	sort.Slice(projs, func(i, j int) bool { return projs[i].Name < projs[j].Name })
	return projs
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
