package claude

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	todo "github.com/stemarquardt/snippets/internal/clients/todoist"
)

func (c *Client) AnalyzeTrends(currentWeekTasks []todo.Task, historicalSummaries []TaskSummary) (*TrendAnalysis, error) {
	if len(historicalSummaries) == 0 {
		return c.createBasicAnalysis(currentWeekTasks)
	}

	systemPrompt := `You are an expert productivity analyst specializing in trend analysis. Analyze current week tasks alongside historical weekly summaries to identify patterns, trends, and provide actionable insights.

Your response must be valid JSON with this exact structure:
{
  "overall_summary": "High-level overview of recent productivity patterns",
  "productivity_trend": "increasing/decreasing/stable with brief explanation",
  "category_trends": [
    {
      "category": "category name",
      "trend": "increasing/decreasing/stable/new/disappeared",
      "description": "brief explanation of the trend"
    }
  ],
  "recommendations": ["actionable recommendation 1", "actionable recommendation 2"],
  "weekly_comparison": "How this week compares to recent weeks"
}

Focus on:
- Overall productivity patterns and changes
- Category/theme trends over time
- Workload distribution changes
- Areas of growing or declining focus
- Actionable insights for improvement`

	currentWeekSummary := fmt.Sprintf("CURRENT WEEK (%s): %d tasks completed",
		time.Now().Format("Jan 2"), len(currentWeekTasks))

	if len(currentWeekTasks) > 0 {
		taskList := make([]string, len(currentWeekTasks))
		for i, task := range currentWeekTasks {
			taskList[i] = fmt.Sprintf("- %s", task.Content)
		}
		currentWeekSummary += "\nTasks:\n" + strings.Join(taskList, "\n")
	}

	historicalData := make([]string, len(historicalSummaries))
	for i, summary := range historicalSummaries {
		categories := "none"
		if len(summary.KeyCategories) > 0 {
			categories = strings.Join(summary.KeyCategories, ", ")
		}
		historicalData[i] = fmt.Sprintf("Week of %s: %d tasks, Categories: %s\nSummary: %s",
			summary.WeekOf.Format("Jan 2"),
			summary.CompletedTasks,
			categories,
			summary.Summary)
	}

	userPrompt := fmt.Sprintf(`%s

HISTORICAL WEEKS:
%s

Analyze trends across these weeks and provide insights for productivity optimization.`,
		currentWeekSummary,
		strings.Join(historicalData, "\n\n"))

	messages := []Message{
		{Role: "user", Content: userPrompt},
	}

	response, err := c.sendMessage(messages, systemPrompt, ModelHaiku, 800)
	if err != nil {
		return nil, fmt.Errorf("failed to get trend analysis: %w", err)
	}

	var result TrendAnalysis
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("failed to parse trend analysis response: %w", err)
	}

	return &result, nil
}

func (c *Client) createBasicAnalysis(currentWeekTasks []todo.Task) (*TrendAnalysis, error) {
	if len(currentWeekTasks) == 0 {
		return &TrendAnalysis{
			OverallSummary:    "No tasks completed this week.",
			ProductivityTrend: "stable",
			CategoryTrends:    []CategoryTrend{},
			Recommendations:   []string{"Consider setting up weekly task goals", "Review task planning process"},
			WeeklyComparison:  "No historical data available for comparison.",
		}, nil
	}

	systemPrompt := `Analyze this week's completed tasks and provide initial insights.

Your response must be valid JSON with this structure:
{
  "overall_summary": "Brief overview of this week's accomplishments",
  "productivity_trend": "stable",
  "category_trends": [
    {
      "category": "category name",
      "trend": "new",
      "description": "brief description"
    }
  ],
  "recommendations": ["recommendation 1", "recommendation 2"],
  "weekly_comparison": "Baseline week - no historical comparison available"
}`

	taskList := make([]string, len(currentWeekTasks))
	for i, task := range currentWeekTasks {
		taskList[i] = fmt.Sprintf("- %s", task.Content)
	}

	userPrompt := fmt.Sprintf(`This week's %d completed tasks:
%s

Provide initial analysis for tracking future trends.`, len(currentWeekTasks), strings.Join(taskList, "\n"))

	messages := []Message{
		{Role: "user", Content: userPrompt},
	}

	response, err := c.sendMessage(messages, systemPrompt, ModelHaiku, 500)
	if err != nil {
		return nil, fmt.Errorf("failed to get basic analysis: %w", err)
	}

	var result TrendAnalysis
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("failed to parse basic analysis response: %w", err)
	}

	return &result, nil
}
