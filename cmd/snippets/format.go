package main

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/stemarquardt/snippets/internal/clients/claude"
	todo "github.com/stemarquardt/snippets/internal/clients/todoist"
)

func FormatSummary(s *claude.TaskSummary) string {
	periodRange := fmt.Sprintf("%s – %s",
		s.PeriodStart.Format(time.DateOnly),
		s.PeriodEnd.Format(time.DateOnly),
	)

	lines := []string{
		"",
		fmt.Sprintf("=== %s Summary (%s) ===", titleCase(s.PeriodLabel), periodRange),
		fmt.Sprintf("Tasks Completed: %d", s.CompletedTasks),
		fmt.Sprintf("Categories:      %s", strings.Join(s.KeyCategories, ", ")),
		"",
		"Summary:",
		"  " + s.Summary,
	}

	if len(s.Accomplishments) > 0 {
		lines = append(lines, "", "Key Accomplishments:")
		for _, a := range s.Accomplishments {
			lines = append(lines, "  • "+a)
		}
	}

	if s.ProductivityTrends != "" {
		lines = append(lines, "", "Trends: "+s.ProductivityTrends)
	}

	return strings.Join(lines, "\n")
}

// FormatReport renders a markdown report for a single project.
func FormatReport(
	projectName string,
	scheduled []todo.FullCtxTask,
	week *todo.BusinessWeek,
	weeklySummary *claude.TaskSummary,
	biweeklySummary *claude.TaskSummary,
	quarterlySummary *claude.TaskSummary,
) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# Project Report: %s\n", projectName))
	b.WriteString(fmt.Sprintf("_Generated %s_\n\n", time.Now().Format("Monday, January 2, 2006")))
	b.WriteString("---\n\n")

	// Scheduled tasks this week
	b.WriteString(fmt.Sprintf("## Scheduled This Week (%s)\n\n", week.String()))
	if len(scheduled) == 0 {
		b.WriteString("_No tasks scheduled for this week._\n\n")
	} else {
		for _, t := range scheduled {
			dueInfo := ""
			if t.Task.Due != nil && t.Task.Due.Date != "" {
				dueInfo = fmt.Sprintf(" _(due %s)_", t.Task.Due.Date)
			} else if dateStr, ok := t.Task.Deadline["date"].(string); ok && dateStr != "" {
				dueInfo = fmt.Sprintf(" _(deadline %s)_", dateStr)
			}
			b.WriteString(fmt.Sprintf("- **%s**%s\n", t.Task.Content, dueInfo))
			if t.Task.Description != "" {
				b.WriteString(fmt.Sprintf("  %s\n", t.Task.Description))
			}
			if t.ParentTask.ID != "" && t.ParentTask.Content != "" {
				b.WriteString(fmt.Sprintf("  _Parent: %s_\n", t.ParentTask.Content))
			}
		}
		b.WriteString("\n")
	}

	// Weekly summary
	b.WriteString("## Weekly Summary\n\n")
	writeSummarySection(&b, weeklySummary)

	// Biweekly summary
	b.WriteString("## Biweekly Summary\n\n")
	writeSummarySection(&b, biweeklySummary)

	// Quarterly summary
	b.WriteString("## Quarterly Summary\n\n")
	writeSummarySection(&b, quarterlySummary)

	return b.String()
}

// titleCase uppercases the first rune of s.
func titleCase(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func writeSummarySection(b *strings.Builder, s *claude.TaskSummary) {
	if s == nil {
		b.WriteString("_No summary available._\n\n")
		return
	}

	periodRange := fmt.Sprintf("%s – %s",
		s.PeriodStart.Format("Jan 2"),
		s.PeriodEnd.Format("Jan 2, 2006"),
	)
	b.WriteString(fmt.Sprintf("**Period:** %s | **Tasks completed:** %d\n\n", periodRange, s.CompletedTasks))

	if s.Summary != "" {
		b.WriteString(s.Summary + "\n\n")
	}

	if len(s.Accomplishments) > 0 {
		b.WriteString("**Key accomplishments:**\n")
		for _, a := range s.Accomplishments {
			b.WriteString(fmt.Sprintf("- %s\n", a))
		}
		b.WriteString("\n")
	}

	if len(s.KeyCategories) > 0 {
		b.WriteString(fmt.Sprintf("**Categories:** %s\n\n", strings.Join(s.KeyCategories, ", ")))
	}

	if s.ProductivityTrends != "" {
		b.WriteString(fmt.Sprintf("**Trends:** %s\n\n", s.ProductivityTrends))
	}
}
