package todoist

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
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

// GetTasksScheduledThisWeek returns active tasks where the due date or deadline
// falls within the current business week (Monday–Sunday).
func (c *Client) GetTasksScheduledThisWeek(ctx context.Context) ([]FullCtxTask, error) {
	week := GetCurrentBusinessWeek()
	all, err := c.GetAllTasks(ctx)
	if err != nil {
		return nil, err
	}
	var scheduled []FullCtxTask
	for _, t := range all {
		if isScheduledInWindow(t.Task, week) {
			scheduled = append(scheduled, t)
		}
	}
	return scheduled, nil
}

// isScheduledInWindow reports whether a task's due date or deadline falls within window.
// Todoist date-only fields represent the user's local calendar date, so time.Local is correct.
func isScheduledInWindow(t Task, window BusinessWeek) bool {
	inWindow := func(dateStr string) bool {
		d, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
		if err != nil {
			return false
		}
		return window.IsInBusinessWeek(d)
	}

	if t.Due != nil && inWindow(t.Due.Date) {
		return true
	}
	if dateStr, ok := t.Deadline["date"].(string); ok {
		return inWindow(dateStr)
	}
	return false
}

func (c *Client) GetTaskNames(tasks []Task) []string {
	var names []string
	for _, t := range tasks {
		names = append(names, t.Content)
	}
	return names
}

// ConvertToFullCtx enriches a slice of tasks with parent task context.
// Parent tasks are fetched at most once per unique parent ID.
// If a parent fetch fails the task is still included, just without parent context.
func (c *Client) ConvertToFullCtx(ctx context.Context, tasks []Task) ([]FullCtxTask, error) {
	parentCache := make(map[string]Task)
	allTasks := make([]FullCtxTask, 0, len(tasks))
	for _, t := range tasks {
		fullT := FullCtxTask{Task: t}
		if t.ParentID != "" {
			if cached, ok := parentCache[t.ParentID]; ok {
				fullT.ParentTask = cached
			} else {
				parent, err := c.GetTask(ctx, t.ParentID)
				if err != nil {
					log.Printf("warning: could not fetch parent %s for task %s: %v", t.ParentID, t.ID, err)
				} else {
					parentCache[t.ParentID] = parent
					fullT.ParentTask = parent
				}
			}
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
