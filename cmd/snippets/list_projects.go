package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stemarquardt/snippets/internal/clients/todoist"
)

var listProjectsCmd = &cobra.Command{
	Use:   "list-projects",
	Short: "List all Todoist projects with their names and IDs",
	Long: `Fetches all projects from your Todoist account and prints their names and IDs.
Use the IDs (or names) with the --projects flag to filter other commands to specific projects.`,
	// Override the root PersistentPreRunE so only the Todoist token is required —
	// no Claude key or database needed just to list projects.
	PersistentPreRunE: initTodoistOnly,
	RunE:              runListProjects,
}

func init() {
	rootCmd.AddCommand(listProjectsCmd)
}

func initTodoistOnly(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	token, err := getAPIToken(ctx, "TODOIST_API_TOKEN", "Todoist")
	if err != nil {
		return fmt.Errorf("failed to get Todoist API token: %w", err)
	}
	todoClient, err = todoist.NewClient(ctx, token, nil)
	if err != nil {
		return fmt.Errorf("error setting up Todoist client: %w", err)
	}
	return todoClient.ValidateToken(ctx)
}

func runListProjects(cmd *cobra.Command, args []string) error {
	projects := todoClient.SortedProjects()
	if len(projects) == 0 {
		fmt.Println("No projects found.")
		return nil
	}

	maxNameLen := 0
	for _, p := range projects {
		if len(p.Name) > maxNameLen {
			maxNameLen = len(p.Name)
		}
	}

	fmt.Printf("%-*s  %s\n", maxNameLen, "NAME", "ID")
	fmt.Printf("%s  %s\n", strings.Repeat("-", maxNameLen), strings.Repeat("-", 12))
	for _, p := range projects {
		fmt.Printf("%-*s  %s\n", maxNameLen, p.Name, p.ID)
	}
	return nil
}
