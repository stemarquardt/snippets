package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/stemarquardt/snippets/internal/clients/claude"
	"github.com/stemarquardt/snippets/internal/clients/todoist"
	"golang.org/x/term"
)

var (
	dbPathFlag   string
	projectsFlag string
	bizWeeksFlag int
	todoClient   *todoist.Client
	claudeClient *claude.Client
)

func main() {
	// Create a context that cancels on interrupt signals (Ctrl+C)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "snippets",
	Short: "A productivity analysis tool using Todoist and Claude AI",
	Long: `Snippets is a CLI tool that analyzes your Todoist tasks using Claude AI
to provide weekly summaries and productivity trend analysis.`,
	SilenceUsage:      true,
	PersistentPreRunE: initClients,
}

var allTodoTasksCmd = &cobra.Command{
	Use:   "get-todos",
	Short: "Fetch all todo tasks from specified projects",
	Long:  `Retrieves all active (not completed) tasks from the specified Todoist projects.`,
	RunE:  runAllTodoTasks,
}

var getCompleTasksCmd = &cobra.Command{
	Use:   "get-completed-tasks",
	Short: "Fetch all completed tasks from specified projects in time window",
	Long:  `Retrieves all completed tasks from the specified Todoist projects for the specified time window.`,
	RunE:  runGetComplTasks,
}

var summarizeTasksCmd = &cobra.Command{
	Use:   "summarize-tasks",
	Short: "Summarize tasks from the given time window",
	Long:  `Retrieves all completed tasks from the specified Todoist projects and summarizes for the specified time window.`,
	RunE:  runSummarizeTasks,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dbPathFlag, "db-filepath", "snippets.db", "Path to SQLite database file")
	rootCmd.PersistentFlags().StringVar(&projectsFlag, "projects", "", "Project IDs to include in task gathering.")

	summarizeTasksCmd.Flags().IntVarP(&bizWeeksFlag, "weeks", "w", 1, "Number of weeks to look back for summarizing.")

	rootCmd.AddCommand(allTodoTasksCmd)
	rootCmd.AddCommand(getCompleTasksCmd)
	rootCmd.AddCommand(summarizeTasksCmd)
}

func validateDatabase(dbPath string) error {
	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		return fmt.Errorf("invalid database path: %w", err)
	}

	dir := filepath.Dir(absPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		file, err := os.Create(absPath)
		if err != nil {
			return fmt.Errorf("failed to create database file: %w", err)
		}
		file.Close()
		fmt.Printf("Created new database file: %s\n", absPath)
	} else {
		fmt.Printf("Using existing database: %s\n", absPath)
	}

	return nil
}

func promptForToken(prompt string) (string, error) {
	fmt.Print(prompt)
	byteToken, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("failed to read token: %w", err)
	}
	fmt.Println()

	token := strings.TrimSpace(string(byteToken))
	if token == "" {
		return "", fmt.Errorf("token cannot be empty")
	}

	return token, nil
}

func getAPIToken(envVar, tokenName string) (string, error) {
	token := os.Getenv(envVar)
	if token != "" {
		return token, nil
	}

	fmt.Printf("\n%s API token not found in environment variable %s\n", tokenName, envVar)
	return promptForToken(fmt.Sprintf("Enter %s API token (input will be hidden): ", tokenName))
}

func initClients(cmd *cobra.Command, args []string) error {
	if err := validateDatabase(dbPathFlag); err != nil {
		return fmt.Errorf("database validation failed: %w", err)
	}

	todoistToken, err := getAPIToken("TODOIST_API_TOKEN", "Todoist")
	if err != nil {
		return fmt.Errorf("failed to get Todoist API token: %w", err)
	}

	claudeAPIKey, err := getAPIToken("CLAUDE_API_KEY", "Claude")
	if err != nil {
		return fmt.Errorf("failed to get Claude API key: %w", err)
	}

	fmt.Println("\nValidating API credentials...")

	fmt.Print("Validating Todoist token... ")
	projs := strings.Split(projectsFlag, ",")
	todoClient, err = todoist.NewClient(cmd.Context(), todoistToken, projs)
	if err != nil {
		return fmt.Errorf("error setting up Todoist client: %w", err)
	}
	if err := todoClient.ValidateToken(cmd.Context()); err != nil {
		fmt.Println("✗")
		return fmt.Errorf("error validating Todoist API token: %w", err)
	}
	fmt.Println("✓")

	fmt.Print("Validating Claude API key... ")
	claudeClient = claude.NewClient(claudeAPIKey)
	if err := claudeClient.ValidateAPIKey(); err != nil {
		fmt.Println("✗")
		return fmt.Errorf("invalid Claude API key: %w", err)
	}
	fmt.Println("✓")

	fmt.Printf("\nDatabase path: %s\n", dbPathFlag)
	return nil
}
