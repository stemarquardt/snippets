package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func runSummarizeTasks(cmd *cobra.Command, args []string) error {
	tasksMap, err := todoClient.GetComplTasksForPreviousBizWeeks(cmd.Context(), bizWeeksFlag)
	if err != nil {
		return err
	}
	for week, tasks := range tasksMap {
		summary, err := claudeClient.SummarizeTasks(tasks, week.Start)
		if err != nil {
			return err
		}
		fmt.Println("Claude summary:")
		fmt.Println("Week of ", summary.WeekOf.Format(time.DateOnly))
		fmt.Println("Number Tasks Completed: ", summary.CompletedTasks)
		fmt.Println("Key Categories ", summary.KeyCategories)
		fmt.Println("Summary: ", summary.Summary)
		fmt.Println("Productivity Trends: ", summary.ProductivityTrends)
	}

	return nil
}
