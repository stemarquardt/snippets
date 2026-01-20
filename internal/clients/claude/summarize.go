package claude

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	todo "github.com/stemarquardt/snippets/internal/clients/todoist"
)

func (c *Client) SummarizeTasks(tasks []todo.Task, weekOf time.Time) (*TaskSummary, error) {
	if len(tasks) == 0 {
		return &TaskSummary{
			WeekOf:         weekOf,
			CompletedTasks: 0,
			Summary:        "No tasks completed this week.",
			KeyCategories:  []string{},
		}, nil
	}

	systemPrompt := `You are an expert productivity analyst. Analyze completed tasks and provide a concise summary.

Your response must be valid JSON with this exact structure:
{
  "summary": "2-3 sentence overview of work accomplished",
  "key_categories": ["category1", "category2", "category3"],
  "productivity_trends": "Brief note about productivity patterns"
}

Focus on:
- Main themes and categories of work
- Notable accomplishments
- Work patterns or focus areas`

	taskList := make([]string, len(tasks))
	for i, task := range tasks {
		complAt, err := time.Parse(time.RFC3339, task.CompletedAt)
		if err != nil {
			return nil, err
		}
		completedDate := complAt.Format("Mon Jan 2")
		taskList[i] = fmt.Sprintf("- %s (completed %s)", task.Content, completedDate)
	}

	userPrompt := fmt.Sprintf(`Analyze these %d completed tasks from the week of %s:

%s

Provide a JSON summary focusing on key themes, accomplishments, and productivity patterns.`,
		len(tasks),
		weekOf.Format("January 2, 2006"),
		strings.Join(taskList, "\n"))

	messages := []Message{
		{Role: "user", Content: userPrompt},
	}

	response, err := c.sendMessage(messages, systemPrompt, ModelHaiku, 500)
	if err != nil {
		return nil, fmt.Errorf("failed to get task summary: %w", err)
	}

	var result struct {
		Summary            string   `json:"summary"`
		KeyCategories      []string `json:"key_categories"`
		ProductivityTrends string   `json:"productivity_trends"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("failed to parse summary response: %w", err)
	}

	return &TaskSummary{
		WeekOf:             weekOf,
		CompletedTasks:     len(tasks),
		Summary:            result.Summary,
		KeyCategories:      result.KeyCategories,
		ProductivityTrends: result.ProductivityTrends,
	}, nil
}

func (c *Client) sendMessage(messages []Message, system, model string, maxTokens int) (string, error) {
	req := MessagesRequest{
		Model:     model,
		MaxTokens: maxTokens,
		Messages:  messages,
		System:    system,
	}

	resp, err := c.doRequest(http.MethodPost, "/messages", req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var msgResp MessagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&msgResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(msgResp.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return msgResp.Content[0].Text, nil
}
