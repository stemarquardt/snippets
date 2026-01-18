package todoist

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func (c *Client) GetCompletedTasks(opts TodoistAPIOpts) ([]CompletedTask, error) {
	endpoint := "/tasks/completed/by_completion_date"

	resp, err := c.doGetRequest(endpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("Failed to make completed tasks request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var completedResp struct {
		Items []CompletedTask `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&completedResp); err != nil {
		return nil, fmt.Errorf("failed to decode completed tasks response: %w", err)
	}

	return completedResp.Items, nil
}

func (c *Client) GetCompletedTasksInTimeWindow(since, until time.Time) ([]CompletedTask, error) {
	return c.GetCompletedTasks(TodoistAPIOpts{
		Since: since,
		Until: until,
	})
}

func (c *Client) GetCompletedTasksByProject(projectID string, since, until time.Time) ([]CompletedTask, error) {
	return c.GetCompletedTasks(TodoistAPIOpts{
		ProjectID: projectID,
		Since:     since,
		Until:     until,
	})
}

func (c *Client) GetCompletedTasksToday() ([]CompletedTask, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	return c.GetCompletedTasksInTimeWindow(startOfDay, endOfDay)
}

func (c *Client) GetCompletedTasksThisWeek() ([]CompletedTask, error) {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	startOfWeek := now.AddDate(0, 0, -(weekday - 1))
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())
	endOfWeek := startOfWeek.Add(7 * 24 * time.Hour)

	return c.GetCompletedTasksInTimeWindow(startOfWeek, endOfWeek)
}

// GetCompletedTasksForBusinessWeek returns completed tasks for a specific business week
func (c *Client) GetCompletedTasksForBusinessWeek(week BusinessWeek) ([]CompletedTask, error) {
	return c.GetCompletedTasksInTimeWindow(week.Start, week.End)
}

// GetCompletedTasksForCurrentBusinessWeek returns completed tasks for the current business week (Monday to today)
func (c *Client) GetCompletedTasksForCurrentBusinessWeek() ([]CompletedTask, error) {
	week := GetCurrentBusinessWeekToDate()
	return c.GetCompletedTasksForBusinessWeek(week)
}

func (c *Client) GetCompletedTasksForCurrentBusinessWeekByProject(p Project) ([]CompletedTask, error) {
	week := GetCurrentBusinessWeekToDate()
	return c.GetCompletedTasks(TodoistAPIOpts{
		Since:     week.Start,
		Until:     week.End,
		ProjectID: p.ID,
	})
}

// GetCompletedTasksForCurrentFullBusinessWeek returns completed tasks for the entire current business week (Monday to Sunday)
func (c *Client) GetCompletedTasksForCurrentFullBusinessWeek() ([]CompletedTask, error) {
	week := GetCurrentBusinessWeek()
	return c.GetCompletedTasksForBusinessWeek(week)
}

// GetCompletedTasksForPreviousBusinessWeeks returns completed tasks for N previous business weeks
// Returns a slice of slices, where each inner slice contains tasks for one week
// Weeks are in chronological order (oldest first)
func (c *Client) GetCompletedTasksForPreviousBusinessWeeks(n int) ([][]CompletedTask, error) {
	weeks := GetBusinessWeeksBack(n)
	result := make([][]CompletedTask, len(weeks))

	for i, week := range weeks {
		tasks, err := c.GetCompletedTasksForBusinessWeek(week)
		if err != nil {
			return nil, fmt.Errorf("failed to get tasks for week %s: %w", week.String(), err)
		}
		result[i] = tasks
	}

	return result, nil
}
