package main

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/spf13/cobra"
	"github.com/stemarquardt/snippets/internal/clients/claude"
	todo "github.com/stemarquardt/snippets/internal/clients/todoist"
)

func runSummarizeTasks(cmd *cobra.Command, args []string) error {
	if !slices.Contains(validSummaryTypes, summaryType) {
		return fmt.Errorf("provided summary type (%s) is not valid — options: %v", summaryType, validSummaryTypes)
	}

	switch summaryType {
	case SummaryTypeWeekly:
		return runWeeklySummaries(cmd)
	case SummaryTypeBiweekly:
		return runBiweeklySummaries(cmd)
	case SummaryTypeQuarterly:
		return runQuarterlySummary(cmd)
	default:
		return fmt.Errorf("summary type %q not yet implemented", summaryType)
	}
}

func runWeeklySummaries(cmd *cobra.Command) error {
	weeks := todo.GetBusinessWeeksBack(bizWeeksFlag)
	for _, w := range weeks {
		summary, err := loadOrGenSummary(cmd, SummaryTypeWeekly, w.Start.Format("2006-01-02"), func() (*claude.TaskSummary, error) {
			tasks, err := todoClient.GetComplTasksForBizWeek(cmd.Context(), w)
			if err != nil {
				return nil, fmt.Errorf("error getting tasks for week %s: %w", w.Start, err)
			}
			return claudeClient.SummarizeTasks(tasks, w.Start, w.End, SummaryTypeWeekly)
		})
		if err != nil {
			return err
		}
		fmt.Println(FormatSummary(summary))
	}
	return nil
}

func runBiweeklySummaries(cmd *cobra.Command) error {
	windows := todo.GetBiweeklyWindowsBack(bizWeeksFlag)
	for _, w := range windows {
		summary, err := loadOrGenSummary(cmd, SummaryTypeBiweekly, w.Start.Format("2006-01-02"), func() (*claude.TaskSummary, error) {
			tasks, err := todoClient.GetComplTasksInTimeWindow(cmd.Context(), w.Start, w.End)
			if err != nil {
				return nil, fmt.Errorf("error getting tasks for biweekly window %s: %w", w.Start, err)
			}
			return claudeClient.SummarizeTasks(tasks, w.Start, w.End, SummaryTypeBiweekly)
		})
		if err != nil {
			return err
		}
		fmt.Println(FormatSummary(summary))
	}
	return nil
}

func runQuarterlySummary(cmd *cobra.Command) error {
	w := todo.GetQuarterlyWindow()
	key := todo.QuarterLabel(w.Start)

	summary, err := loadOrGenSummary(cmd, SummaryTypeQuarterly, key, func() (*claude.TaskSummary, error) {
		tasks, err := todoClient.GetComplTasksInTimeWindow(cmd.Context(), w.Start, w.End)
		if err != nil {
			return nil, fmt.Errorf("error getting tasks for quarter %s: %w", key, err)
		}
		return claudeClient.SummarizeTasks(tasks, w.Start, w.End, SummaryTypeQuarterly)
	})
	if err != nil {
		return err
	}
	fmt.Println(FormatSummary(summary))
	return nil
}

// loadOrGenSummary reads a summary from the DB if available; otherwise calls gen to produce it
// and writes the result back to the DB. Pass --force-refresh to skip the cache read.
func loadOrGenSummary(cmd *cobra.Command, bucket, key string, gen func() (*claude.TaskSummary, error)) (*claude.TaskSummary, error) {
	if !forceRefresh {
		b, err := db.Read([]byte(bucket), []byte(key))
		if err == nil {
			var s claude.TaskSummary
			if jsonErr := json.Unmarshal(b, &s); jsonErr == nil {
				fmt.Printf("Loaded %s summary for %q from cache.\n", bucket, key)
				return &s, nil
			}
		}
	}

	fmt.Printf("Generating %s summary for %q...\n", bucket, key)
	s, err := gen()
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(s)
	if err != nil {
		fmt.Println("Warning: could not cache summary:", err)
		return s, nil
	}
	if err := db.Write([]byte(bucket), []byte(key), data); err != nil {
		fmt.Println("Warning: could not write summary to db:", err)
	}
	return s, nil
}
