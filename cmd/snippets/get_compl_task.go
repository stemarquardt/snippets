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
			fmt.Printf("Error getting tasks for project %s: %+v", p.Name, err)
			continue
		}
		complTasks = append(complTasks, tasks...)
		fmt.Printf("[%s] Tasks for project \"%s\":\n----------\n", week.String(), p.Name)
		for _, task := range complTasks {
			fmt.Printf("Content: %s\nDescription: %s\n", task.Task.Content, task.Task.Description)
			if task.ParentTask.ID != "" {
				fmt.Printf("!!Parent Info!!\nContent: %s\nDescription: %s\n", task.ParentTask.Content, task.ParentTask.Description)
			}
		}
	}
	return nil
}
