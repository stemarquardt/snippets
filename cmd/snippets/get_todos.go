package main

import (
	"encoding/json"
	"fmt"

	todo "github.com/stemarquardt/snippets/internal/clients/todoist"

	"github.com/spf13/cobra"
)

func runAllTodoTasks(cmd *cobra.Command, args []string) error {
	var tasks []todo.FullCtxTask
	for _, p := range todoClient.Projects {
		t, err := todoClient.GetTasksForProj(cmd.Context(), p.ID)
		if err != nil {
			return err
		}
		tasks = append(tasks, t...)
	}
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %w", err)
	}
	fmt.Println(string(data))
	return nil
}
