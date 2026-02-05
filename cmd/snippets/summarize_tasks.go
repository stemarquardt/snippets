package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/stemarquardt/snippets/internal/clients/claude"
)

func runSummarizeTasks(cmd *cobra.Command, args []string) error {
	tasksMap, err := todoClient.GetComplTasksForPreviousBizWeeks(cmd.Context(), bizWeeksFlag)
	if err != nil {
		return err
	}
	for i := 0; i < len(tasksMap); i++ {
		tasks := tasksMap[i].Tasks
		week := tasksMap[i].WeekOf
		b, err := db.Read([]byte("weekly"), []byte(week.Start.Format("2006-01-02")))
		if err != nil {
			fmt.Printf("no entry found for week starting %s, generating summary - err: %s", week.Start, err)
		}
		var summary *claude.TaskSummary
		err = json.Unmarshal(b, &summary)
		if err != nil {
			fmt.Println("error unmarshalling db entry, generating summary - err :", err)
		}
		if summary == nil {
			fmt.Println("Generating summary...")
			summary, err = claudeClient.SummarizeTasks(tasks, week.Start)
			if err != nil {
				return err
			}
		} else {
			fmt.Println("Loaded summary from storage...")
		}
		fmt.Println("Claude summary:")
		fmt.Println("Week of", summary.WeekOf.Format(time.DateOnly))
		fmt.Println("Number Tasks Completed:", summary.CompletedTasks)
		fmt.Println("Key Categories", summary.KeyCategories)
		fmt.Println("Summary:", summary.Summary)
		fmt.Println("Productivity Trends:", summary.ProductivityTrends)

		data, err := json.Marshal(summary)
		if err != nil {
			fmt.Println("Error writing summary to db:", err)
			return err
		}

		db.Write([]byte("weekly"), []byte(summary.WeekOf.Format("2006-01-02")), data)
	}

	return nil
}
