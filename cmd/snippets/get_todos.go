package main

import (
	"encoding/json"
	"fmt"

	todo "github.com/stemarquardt/snippets/internal/clients/todoist"

	"github.com/spf13/cobra"
)

func runAllTodoTasks(cmd *cobra.Command, args []string) error {
	var tasks []todo.Task
	var err error
	for _, p := range todoClient.Projects {
		tasks, err = todoClient.GetTasksForProj(cmd.Context(), p.ID)
		if err != nil {
			return err
		}
	}
	fmt.Println("All tasks:")
	data, _ := json.MarshalIndent(tasks, "", "  ")
	fmt.Println(string(data))
	return nil
}
