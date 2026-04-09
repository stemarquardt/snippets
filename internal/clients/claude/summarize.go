package claude

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	todo "github.com/stemarquardt/snippets/internal/clients/todoist"
)

// SummarizeTasks generates an AI summary of completed tasks for the given period.
// periodLabel should be "weekly", "biweekly", or "quarterly".
func (c *Client) SummarizeTasks(tasks []todo.FullCtxTask, periodStart, periodEnd time.Time, periodLabel string) (*TaskSummary, error) {
	if len(tasks) == 0 {
		return &TaskSummary{
			PeriodStart:    periodStart,
			PeriodEnd:      periodEnd,
			PeriodLabel:    periodLabel,
			CompletedTasks: 0,
			Summary:        fmt.Sprintf("No tasks completed during this %s period.", periodLabel),
			KeyCategories:  []string{},
			Accomplishments: []string{},
		}, nil
	}

	systemPrompt := buildSystemPrompt(periodLabel)
	userPrompt := buildUserPrompt(tasks, periodStart, periodEnd, periodLabel)

	messages := []Message{
		{Role: "user", Content: userPrompt},
	}

	model := ModelHaiku
	maxTokens := 600
	if periodLabel == "quarterly" {
		model = ModelSonnet
		maxTokens = 1200
	} else if periodLabel == "biweekly" {
		maxTokens = 800
	}

	response, err := c.sendMessage(messages, systemPrompt, model, maxTokens)
	if err != nil {
		return nil, fmt.Errorf("failed to get task summary: %w", err)
	}

	var result struct {
		Summary            string   `json:"summary"`
		KeyCategories      []string `json:"key_categories"`
		Accomplishments    []string `json:"accomplishments"`
		ProductivityTrends string   `json:"productivity_trends"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("failed to parse summary response: %w", err)
	}

	return &TaskSummary{
		PeriodStart:        periodStart,
		PeriodEnd:          periodEnd,
		PeriodLabel:        periodLabel,
		CompletedTasks:     len(tasks),
		Summary:            result.Summary,
		KeyCategories:      result.KeyCategories,
		Accomplishments:    result.Accomplishments,
		ProductivityTrends: result.ProductivityTrends,
	}, nil
}

func buildSystemPrompt(periodLabel string) string {
	base := `You are an expert productivity analyst. Analyze completed tasks and provide a concise, high-quality summary.

Writing style — this is critical:
- Write in neutral, impersonal language. Describe work and outcomes directly.
- Do NOT use any subject: no "the team", "I", "we", "you", "they", or any pronoun or group noun.
- Use noun phrases and passive constructions instead. For example:
  - WRONG: "The team completed the API integration."
  - WRONG: "I focused on backend work."
  - RIGHT: "API integration completed."
  - RIGHT: "Focus on backend infrastructure and bug resolution."

Your response must be valid JSON with this exact structure:
{
  "summary": "overview of work accomplished",
  "key_categories": ["category1", "category2"],
  "accomplishments": ["specific accomplishment 1", "specific accomplishment 2"],
  "productivity_trends": "note about productivity patterns"
}`

	switch periodLabel {
	case "quarterly":
		return base + `

For a quarterly summary:
- "summary": 3-5 sentence executive-level overview of the quarter's work
- "key_categories": top 4-6 themes or work areas from the quarter
- "accomplishments": 4-6 most significant, concrete accomplishments
- "productivity_trends": patterns in how work evolved across the quarter, any notable shifts in focus`
	case "biweekly":
		return base + `

For a biweekly (sprint) summary:
- "summary": 2-3 sentence sprint-style overview of what was accomplished
- "key_categories": 3-5 key work themes from the two-week period
- "accomplishments": 3-5 notable deliverables or milestones completed
- "productivity_trends": brief note on pace and focus for this sprint`
	default: // weekly
		return base + `

For a weekly summary:
- "summary": 2-3 sentence overview of the week's work
- "key_categories": 2-4 main categories or themes
- "accomplishments": 2-3 notable completed items
- "productivity_trends": brief note about this week's patterns`
	}
}

func buildUserPrompt(tasks []todo.FullCtxTask, periodStart, periodEnd time.Time, periodLabel string) string {
	taskList := make([]string, 0, len(tasks))
	for _, task := range tasks {
		entry := fmt.Sprintf("- %s", task.Task.Content)
		if task.Task.Description != "" {
			entry += fmt.Sprintf(" [%s]", task.Task.Description)
		}
		if task.ParentTask.ID != "" && task.ParentTask.Content != "" {
			entry += fmt.Sprintf(" (under: %s", task.ParentTask.Content)
			if task.ParentTask.Description != "" {
				entry += fmt.Sprintf(" – %s", task.ParentTask.Description)
			}
			entry += ")"
		}
		if task.Task.CompletedAt != "" {
			if t, err := time.Parse(time.RFC3339, task.Task.CompletedAt); err == nil {
				entry += fmt.Sprintf(" [completed %s]", t.Format("Mon Jan 2"))
			}
		}
		taskList = append(taskList, entry)
	}

	periodRange := fmt.Sprintf("%s – %s", periodStart.Format("January 2"), periodEnd.Format("January 2, 2006"))

	return fmt.Sprintf(`Analyze these %d completed tasks from the %s period (%s):

%s

Provide a JSON summary tailored for a %s review.`,
		len(tasks),
		periodLabel,
		periodRange,
		strings.Join(taskList, "\n"),
		periodLabel,
	)
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
