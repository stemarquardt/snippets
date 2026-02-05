package todoist

import (
	"context"
	"encoding/json"
	"fmt"
)

const (
	endpoint string = "/tasks"
)

func (c *Client) GetTasksForProj(ctx context.Context, pId string) ([]FullCtxTask, error) {
	resp, err := c.doGetRequest(ctx, endpoint, TodoistAPIOpts{ProjectID: pId})
	if err != nil {
		return nil, fmt.Errorf("failed to make all tasks tasks request: %w", err)
	}
	defer resp.Body.Close()
	var allTasksResp struct {
		Results []Task `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&allTasksResp); err != nil {
		return nil, fmt.Errorf("failed to decode completed tasks response: %w", err)
	}

	allTasks, err := c.ConvertToFullCtx(ctx, allTasksResp.Results)
	if err != nil {
		return nil, nil
	}

	return allTasks, nil
}

func (c *Client) GetAllTasks(ctx context.Context) ([]FullCtxTask, error) {
	var tasks []FullCtxTask
	for _, p := range c.Projects {
		t, err := c.GetTasksForProj(ctx, p.ID)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t...)
	}
	return tasks, nil
}

func (c *Client) GetTaskNames(tasks []Task) []string {
	var names []string
	for _, t := range tasks {
		names = append(names, t.Content)
	}
	return names
}

// Gather the additional context for a slice of tasks.
func (c *Client) ConvertToFullCtx(ctx context.Context, tasks []Task) ([]FullCtxTask, error) {
	var allTasks []FullCtxTask
	for _, t := range tasks {
		fullT := FullCtxTask{Task: t}
		if t.ParentID != "" {
			parent, err := c.GetTask(ctx, t.ParentID)
			if err != nil {
				fmt.Println("non-fatal error getting parent task: ", err)
				continue
			}
			fullT.ParentTask = parent
		}
		allTasks = append(allTasks, fullT)
	}
	return allTasks, nil
}

func (c *Client) GetTask(ctx context.Context, id string) (Task, error) {
	resp, err := c.doGetRequest(ctx, fmt.Sprintf("%s/%s", endpoint, id), TodoistAPIOpts{TaskID: id})
	if err != nil {
		return Task{}, err
	}
	defer resp.Body.Close()
	var task Task
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return Task{}, fmt.Errorf("failed to decode task response: %w", err)
	}
	return task, nil
}
