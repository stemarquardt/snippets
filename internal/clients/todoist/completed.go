package todoist

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

func (c *Client) GetComplTasks(ctx context.Context, opts TodoistAPIOpts) ([]Task, error) {
	endpoint := "/tasks/completed/by_completion_date"

	resp, err := c.doGetRequest(ctx, endpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("Failed to make completed tasks request: %w", err)
	}
	defer resp.Body.Close()

	var completedResp struct {
		Items []Task `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&completedResp); err != nil {
		return nil, fmt.Errorf("failed to decode completed tasks response: %w", err)
	}

	return completedResp.Items, nil
}

func (c *Client) GetComplTasksInTimeWindow(ctx context.Context, since, until time.Time) ([]Task, error) {
	var tasks []Task
	for pId := range c.Projects {
		t, err := c.GetComplTasks(ctx, TodoistAPIOpts{
			Since:     since,
			Until:     until,
			ProjectID: pId,
		})
		if err != nil {
			return []Task{}, err
		}
		tasks = append(tasks, t...)
	}
	return tasks, nil
}

func (c *Client) GetComplTasksToday(ctx context.Context) ([]Task, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	return c.GetComplTasksInTimeWindow(ctx, startOfDay, endOfDay)
}

// Gather tasks for a calendar week, not the business week.
func (c *Client) GetCompTasksThisCalWeek(ctx context.Context) ([]Task, error) {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	startOfWeek := now.AddDate(0, 0, -(weekday - 1))
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())
	endOfWeek := startOfWeek.Add(7 * 24 * time.Hour)

	return c.GetComplTasksInTimeWindow(ctx, startOfWeek, endOfWeek)
}

// GetTasksForBusinessWeek returns completed tasks for a specific business week
func (c *Client) GetComplTasksForBizWeek(ctx context.Context, week BusinessWeek) ([]Task, error) {
	return c.GetComplTasksInTimeWindow(ctx, week.Start, week.End)
}

// GetTasksForCurrentBusinessWeek returns completed tasks for the current business week (Monday to today)
func (c *Client) GetComplTasksForCurrentBizWeek(ctx context.Context) ([]Task, error) {
	week := GetCurrentBusinessWeekToDate()
	return c.GetComplTasksForBizWeek(ctx, week)
}

func (c *Client) GetComplTasksForCurrentBizWeekByProject(ctx context.Context, p Project) ([]Task, BusinessWeek, error) {
	week := GetCurrentBusinessWeekToDate()
	t, err := c.GetComplTasks(ctx, TodoistAPIOpts{
		Since:     week.Start,
		Until:     week.End,
		ProjectID: p.ID,
	})
	return t, week, err
}

// GetTasksForCurrentFullBusinessWeek returns completed tasks for the entire current business week (Monday to Sunday)
func (c *Client) GetComplTasksForCurrentFullBizWeek(ctx context.Context) ([]Task, error) {
	week := GetCurrentBusinessWeek()
	return c.GetComplTasksForBizWeek(ctx, week)
}

// GetTasksForPreviousBusinessWeeks returns completed tasks for N previous business weeks
// Returns a slice of slices, where each inner slice contains tasks for one week
// Weeks are in chronological order (oldest first)
func (c *Client) GetComplTasksForPreviousBizWeeks(ctx context.Context, n int) (map[BusinessWeek][]Task, error) {
	weeks := GetBusinessWeeksBack(n)
	result := make(map[BusinessWeek][]Task, len(weeks))

	for _, week := range weeks {
		tasks, err := c.GetComplTasksForBizWeek(ctx, week)
		if err != nil {
			return nil, fmt.Errorf("failed to get tasks for week %s: %w", week.String(), err)
		}
		result[week] = tasks
	}

	return result, nil
}
