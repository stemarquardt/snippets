package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stemarquardt/snippets/internal/clients/claude"
	todo "github.com/stemarquardt/snippets/internal/clients/todoist"
)

func runReport(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Fetch scheduled tasks once, then bucket by project.
	allScheduled, err := todoClient.GetTasksScheduledThisWeek(ctx)
	if err != nil {
		return fmt.Errorf("error fetching scheduled tasks: %w", err)
	}
	scheduledByProject := make(map[string][]todo.FullCtxTask)
	for _, t := range allScheduled {
		scheduledByProject[t.Task.ProjectID] = append(scheduledByProject[t.Task.ProjectID], t)
	}

	currentWeek := todo.GetCurrentBusinessWeek()

	// Biweekly and quarterly windows are the same across all projects; compute once.
	biweekWindows := todo.GetBiweeklyWindowsBack(1)
	biweekWindow := biweekWindows[0]
	quarterWindow := todo.GetQuarterlyWindow()
	quarterKey := todo.QuarterLabel(quarterWindow.Start)

	for _, project := range todoClient.SortedProjects() {
		pID := project.ID

		// 1. Weekly summary — current business week to date.
		weekWindow := todo.GetCurrentBusinessWeekToDate()
		weeklySummary, err := loadOrGenSummary(
			cmd, SummaryTypeWeekly, weekWindow.Start.Format("2006-01-02"),
			func() (*claude.TaskSummary, error) {
				tasks, err := todoClient.GetComplTasksInTimeWindow(ctx, weekWindow.Start, weekWindow.End)
				if err != nil {
					return nil, err
				}
				return claudeClient.SummarizeTasks(filterByProject(tasks, pID), weekWindow.Start, weekWindow.End, SummaryTypeWeekly)
			},
		)
		if err != nil {
			fmt.Printf("warning: could not generate weekly summary for %q: %v\n", project.Name, err)
		}

		// 2. Biweekly summary — last two business weeks.
		biweeklySummary, err := loadOrGenSummary(
			cmd, SummaryTypeBiweekly, biweekWindow.Start.Format("2006-01-02"),
			func() (*claude.TaskSummary, error) {
				tasks, err := todoClient.GetComplTasksInTimeWindow(ctx, biweekWindow.Start, biweekWindow.End)
				if err != nil {
					return nil, err
				}
				return claudeClient.SummarizeTasks(filterByProject(tasks, pID), biweekWindow.Start, biweekWindow.End, SummaryTypeBiweekly)
			},
		)
		if err != nil {
			fmt.Printf("warning: could not generate biweekly summary for %q: %v\n", project.Name, err)
		}

		// 3. Quarterly summary — current calendar quarter.
		quarterlySummary, err := loadOrGenSummary(
			cmd, SummaryTypeQuarterly, quarterKey,
			func() (*claude.TaskSummary, error) {
				tasks, err := todoClient.GetComplTasksInTimeWindow(ctx, quarterWindow.Start, quarterWindow.End)
				if err != nil {
					return nil, err
				}
				return claudeClient.SummarizeTasks(filterByProject(tasks, pID), quarterWindow.Start, quarterWindow.End, SummaryTypeQuarterly)
			},
		)
		if err != nil {
			fmt.Printf("warning: could not generate quarterly summary for %q: %v\n", project.Name, err)
		}

		fmt.Println(FormatReport(
			project.Name,
			scheduledByProject[pID],
			&currentWeek,
			weeklySummary,
			biweeklySummary,
			quarterlySummary,
		))
	}
	return nil
}

func filterByProject(tasks []todo.FullCtxTask, projectID string) []todo.FullCtxTask {
	filtered := make([]todo.FullCtxTask, 0, len(tasks))
	for _, t := range tasks {
		if t.Task.ProjectID == projectID {
			filtered = append(filtered, t)
		}
	}
	return filtered
}
