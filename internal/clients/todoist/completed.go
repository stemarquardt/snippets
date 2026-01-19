package todoist

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func (c *Client) GetComplTasks(opts TodoistAPIOpts) ([]Task, error) {
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
		Items []Task `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&completedResp); err != nil {
		return nil, fmt.Errorf("failed to decode completed tasks response: %w", err)
	}

	return completedResp.Items, nil
}

func (c *Client) GetComplTasksInTimeWindow(since, until time.Time) ([]Task, error) {
	return c.GetComplTasks(TodoistAPIOpts{
		Since: since,
		Until: until,
	})
}

func (c *Client) GetComplTasksToday() ([]Task, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	return c.GetComplTasksInTimeWindow(startOfDay, endOfDay)
}

func (c *Client) GetCompTasksThisWeek() ([]Task, error) {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	startOfWeek := now.AddDate(0, 0, -(weekday - 1))
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())
	endOfWeek := startOfWeek.Add(7 * 24 * time.Hour)

	return c.GetComplTasksInTimeWindow(startOfWeek, endOfWeek)
}

// GetTasksForBusinessWeek returns completed tasks for a specific business week
func (c *Client) GetComplTasksForBusinessWeek(week BusinessWeek) ([]Task, error) {
	return c.GetComplTasksInTimeWindow(week.Start, week.End)
}

// GetTasksForCurrentBusinessWeek returns completed tasks for the current business week (Monday to today)
func (c *Client) GetComplTasksForCurrentBusinessWeek() ([]Task, error) {
	week := GetCurrentBusinessWeekToDate()
	return c.GetComplTasksForBusinessWeek(week)
}

func (c *Client) GetComplTasksForCurrentBusinessWeekByProject(p Project) ([]Task, error) {
	week := GetCurrentBusinessWeekToDate()
	return c.GetComplTasks(TodoistAPIOpts{
		Since:     week.Start,
		Until:     week.End,
		ProjectID: p.ID,
	})
}

// GetTasksForCurrentFullBusinessWeek returns completed tasks for the entire current business week (Monday to Sunday)
func (c *Client) GetComplTasksForCurrentFullBusinessWeek() ([]Task, error) {
	week := GetCurrentBusinessWeek()
	return c.GetComplTasksForBusinessWeek(week)
}

// GetTasksForPreviousBusinessWeeks returns completed tasks for N previous business weeks
// Returns a slice of slices, where each inner slice contains tasks for one week
// Weeks are in chronological order (oldest first)
func (c *Client) GetComplTasksForPreviousBusinessWeeks(n int) ([][]Task, error) {
	weeks := GetBusinessWeeksBack(n)
	result := make([][]Task, len(weeks))

	for i, week := range weeks {
		tasks, err := c.GetComplTasksForBusinessWeek(week)
		if err != nil {
			return nil, fmt.Errorf("failed to get tasks for week %s: %w", week.String(), err)
		}
		result[i] = tasks
	}

	return result, nil
}
