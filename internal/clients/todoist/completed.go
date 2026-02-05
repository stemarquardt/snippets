package todoist

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

func (c *Client) GetComplTasks(ctx context.Context, opts TodoistAPIOpts) ([]FullCtxTask, error) {
	endpoint := "/tasks/completed/by_completion_date"

	resp, err := c.doGetRequest(ctx, endpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to make completed tasks request: %w", err)
	}
	defer resp.Body.Close()

	var completedResp struct {
		Items []Task `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&completedResp); err != nil {
		return nil, fmt.Errorf("failed to decode completed tasks response: %w", err)
	}
	tasks, err := c.ConvertToFullCtx(ctx, completedResp.Items)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (c *Client) GetComplTasksInTimeWindow(ctx context.Context, since, until time.Time) ([]FullCtxTask, error) {
	var tasks []FullCtxTask
	for pId := range c.Projects {
		t, err := c.GetComplTasks(ctx, TodoistAPIOpts{
			Since:     since,
			Until:     until,
			ProjectID: pId,
		})
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t...)
	}
	return tasks, nil
}

func (c *Client) GetComplTasksToday(ctx context.Context) ([]FullCtxTask, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	return c.GetComplTasksInTimeWindow(ctx, startOfDay, endOfDay)
}

// Gather tasks for a calendar week, not the business week.
func (c *Client) GetCompTasksThisCalWeek(ctx context.Context) ([]FullCtxTask, error) {
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
func (c *Client) GetComplTasksForBizWeek(ctx context.Context, week BusinessWeek) ([]FullCtxTask, error) {
	return c.GetComplTasksInTimeWindow(ctx, week.Start, week.End)
}

// GetTasksForCurrentBusinessWeek returns completed tasks for the current business week (Monday to today)
func (c *Client) GetComplTasksForCurrentBizWeek(ctx context.Context) ([]FullCtxTask, error) {
	week := GetCurrentBusinessWeekToDate()
	return c.GetComplTasksForBizWeek(ctx, week)
}

func (c *Client) GetComplTasksForCurrentBizWeekByProject(ctx context.Context, p Project) ([]FullCtxTask, BusinessWeek, error) {
	week := GetCurrentBusinessWeekToDate()
	t, err := c.GetComplTasks(ctx, TodoistAPIOpts{
		Since:     week.Start,
		Until:     week.End,
		ProjectID: p.ID,
	})
	return t, week, err
}

// GetTasksForCurrentFullBusinessWeek returns completed tasks for the entire current business week (Monday to Sunday)
func (c *Client) GetComplTasksForCurrentFullBizWeek(ctx context.Context) ([]FullCtxTask, error) {
	week := GetCurrentBusinessWeek()
	return c.GetComplTasksForBizWeek(ctx, week)
}

// GetTasksForPreviousBusinessWeeks returns completed tasks for N previous business weeks
// Returns a slice of slices, where each inner slice contains tasks for one week
// Weeks are in chronological order (oldest first)
func (c *Client) GetComplTasksForPreviousBizWeeks(ctx context.Context, n int) (map[int]BizWeekTasks, error) {
	weeks := GetBusinessWeeksBack(n)
	result := make(map[int]BizWeekTasks, len(weeks))

	for i, week := range weeks {
		tasks, err := c.GetComplTasksForBizWeek(ctx, week)
		if err != nil {
			return nil, fmt.Errorf("failed to get tasks for week %s: %w", week.String(), err)
		}
		result[i] = BizWeekTasks{Tasks: tasks, WeekOf: week}
	}

	return result, nil
}
