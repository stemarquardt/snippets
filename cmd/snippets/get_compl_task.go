package main

import (
	"fmt"

	todo "github.com/stemarquardt/snippets/internal/clients/todoist"

	"github.com/spf13/cobra"
)

func runGetComplTasks(cmd *cobra.Command, args []string) error {
	var complTasks []todo.FullCtxTask
	for _, p := range todoClient.Projects {
		tasks, week, err := todoClient.GetComplTasksForCurrentBizWeekByProject(cmd.Context(), *p)
		if err != nil {
			fmt.Printf("error getting tasks for project %s: %v\n", p.Name, err)
			continue
		}
		fmt.Printf("[%s] Tasks for project %q:\n----------\n", week.String(), p.Name)
		for _, task := range tasks {
			fmt.Printf("Content: %s\nDescription: %s\n", task.Task.Content, task.Task.Description)
			if task.ParentTask.ID != "" {
				fmt.Printf("  Parent: %s\n  Parent description: %s\n", task.ParentTask.Content, task.ParentTask.Description)
			}
		}
		complTasks = append(complTasks, tasks...)
	}
	fmt.Printf("\nTotal completed tasks this week: %d\n", len(complTasks))
	return nil
}
