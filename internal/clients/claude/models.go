package claude

import "time"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type MessagesRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
	System    string    `json:"system,omitempty"`
}

type MessagesResponse struct {
	ID           string        `json:"id"`
	Type         string        `json:"type"`
	Role         string        `json:"role"`
	Content      []ContentBlock `json:"content"`
	Model        string        `json:"model"`
	StopReason   string        `json:"stop_reason"`
	StopSequence string        `json:"stop_sequence"`
	Usage        Usage         `json:"usage"`
}

type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type TaskSummary struct {
	WeekOf          time.Time `json:"week_of"`
	CompletedTasks  int       `json:"completed_tasks"`
	Summary         string    `json:"summary"`
	KeyCategories   []string  `json:"key_categories"`
	ProductivityTrends string `json:"productivity_trends"`
}

type TrendAnalysis struct {
	OverallSummary    string        `json:"overall_summary"`
	ProductivityTrend string        `json:"productivity_trend"`
	CategoryTrends    []CategoryTrend `json:"category_trends"`
	Recommendations   []string      `json:"recommendations"`
	WeeklyComparison  string        `json:"weekly_comparison"`
}

type CategoryTrend struct {
	Category    string `json:"category"`
	Trend       string `json:"trend"`
	Description string `json:"description"`
}