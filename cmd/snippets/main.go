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
	"github.com/stemarquardt/snippets/internal/storage"
	"golang.org/x/term"
)

const (
	SummaryTypeWeekly    = "weekly"
	SummaryTypeBiweekly  = "biweekly"
	SummaryTypeQuarterly = "quarterly"
	// Special type of summary that outlines completed work, that also
	// summarizes the todo tasks for the rest of the week.
	SummaryTypeStandup = "standup"
)

var (
	dbPathFlag   string
	projectsFlag string
	todoClient   *todoist.Client
	claudeClient *claude.Client
	db           *storage.Store

	// Summarize Tasks Flags
	bizWeeksFlag      int
	summaryType       string
	forceRefresh      bool
	validSummaryTypes = []string{SummaryTypeStandup, SummaryTypeWeekly, SummaryTypeBiweekly, SummaryTypeQuarterly}
)

func main() {
	// Create a context that cancels on interrupt signals (Ctrl+C)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.Printf("!!ERROR: %v", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "snippets",
	Short: "A productivity analysis tool using Todoist and Claude AI",
	Long: `Snippets is a CLI tool that analyzes your Todoist tasks using Claude AI
to provide weekly summaries and productivity trend analysis.`,
	SilenceUsage:       true,
	SilenceErrors:      true,
	PersistentPreRunE:  initClients,
	PersistentPostRunE: cleanup,
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

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a project report with scheduled tasks and multi-period summaries",
	Long: `Generates a markdown report per project that includes:
  - Tasks scheduled for this week (due or deadline within Mon–Sun)
  - Weekly completed task summary
  - Biweekly (sprint) completed task summary
  - Quarterly completed task summary`,
	RunE: runReport,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dbPathFlag, "db-filepath", "snippets.db", "Path to SQLite database file")
	rootCmd.PersistentFlags().StringVar(&projectsFlag, "projects", "", "Comma-separated project names or IDs to include (e.g. \"Work,Personal\" or \"123,456\"). Omit to include all projects.")

	summarizeTasksCmd.Flags().IntVarP(&bizWeeksFlag, "weeks", "w", 1, "Number of weeks to look back for summarizing.")
	summarizeTasksCmd.Flags().StringVar(&summaryType, "summary-type", "weekly", fmt.Sprintf("Summary type to generate report for, options are:\n%s.", validSummaryTypes))
	summarizeTasksCmd.Flags().BoolVar(&forceRefresh, "force-refresh", false, "Bypass the cache and regenerate summaries from scratch.")
	reportCmd.Flags().BoolVar(&forceRefresh, "force-refresh", false, "Bypass the cache and regenerate summaries from scratch.")

	rootCmd.AddCommand(allTodoTasksCmd)
	rootCmd.AddCommand(getCompleTasksCmd)
	rootCmd.AddCommand(summarizeTasksCmd)
	rootCmd.AddCommand(reportCmd)
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
		if err := file.Close(); err != nil {
			return fmt.Errorf("failed to close database file: %w", err)
		}
		fmt.Printf("Created new database file: %s\n", absPath)
	} else {
		fmt.Printf("Using existing database: %s\n", absPath)
	}

	return nil
}

func promptForToken(ctx context.Context, prompt string) (string, error) {
	fmt.Print(prompt)

	// Channel to receive the result from ReadPassword
	type result struct {
		token []byte
		err   error
	}
	resultCh := make(chan result, 1)

	go func() {
		byteToken, err := term.ReadPassword(int(syscall.Stdin))
		resultCh <- result{byteToken, err}
	}()

	select {
	case <-ctx.Done():
		log.Println()
		return "", ctx.Err()
	case r := <-resultCh:
		if r.err != nil {
			return "", fmt.Errorf("failed to read token: %w", r.err)
		}
		log.Println()

		token := strings.TrimSpace(string(r.token))
		if token == "" {
			return "", fmt.Errorf("token cannot be empty")
		}
		return token, nil
	}
}

func getAPIToken(ctx context.Context, envVar, tokenName string) (string, error) {
	token := os.Getenv(envVar)
	if token != "" {
		return token, nil
	}

	fmt.Printf("\n%s API token not found in environment variable %s\n", tokenName, envVar)
	return promptForToken(ctx, fmt.Sprintf("Enter %s API token (input will be hidden): ", tokenName))
}

func initClients(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	if err := validateDatabase(dbPathFlag); err != nil {
		return fmt.Errorf("database validation failed: %w", err)
	}

	todoistToken, err := getAPIToken(ctx, "TODOIST_API_TOKEN", "Todoist")
	if err != nil {
		return fmt.Errorf("failed to get Todoist API token: %w", err)
	}

	claudeAPIKey, err := getAPIToken(ctx, "CLAUDE_API_KEY", "Claude")
	if err != nil {
		return fmt.Errorf("failed to get Claude API key: %w", err)
	}

	log.Println("\nValidating API credentials...")

	log.Print("Validating Todoist token... ")
	// Always load all projects first so we can resolve names.
	todoClient, err = todoist.NewClient(cmd.Context(), todoistToken, nil)
	if err != nil {
		return fmt.Errorf("error setting up Todoist client: %w", err)
	}
	if err := todoClient.ValidateToken(cmd.Context()); err != nil {
		return fmt.Errorf("error validating Todoist API token: %w", err)
	}

	// If --projects was given, resolve names/IDs and restrict the client.
	if projectsFlag != "" {
		refs := strings.Split(projectsFlag, ",")
		resolved, err := todoClient.ResolveProjectRefs(refs)
		if err != nil {
			return fmt.Errorf("--projects: %w", err)
		}
		todoClient.DelAllProjects()
		for _, p := range resolved {
			todoClient.AddProject(p)
		}
		log.Printf("Filtered to %d project(s): %s", len(resolved), projectNames(resolved))
	}

	log.Print("Validating Claude API key... ")
	claudeClient = claude.NewClient(claudeAPIKey)
	if err := claudeClient.ValidateAPIKey(); err != nil {
		return fmt.Errorf("invalid Claude API key: %w", err)
	}

	log.Printf("Validating database with path: %s... ", dbPathFlag)
	db, err = storage.New(dbPathFlag)
	if err != nil {
		return fmt.Errorf("unable to create database conn: %w", err)
	}

	return nil
}

func projectNames(projs []todoist.Project) string {
	names := make([]string, len(projs))
	for i, p := range projs {
		names[i] = p.Name
	}
	return strings.Join(names, ", ")
}

func cleanup(cmd *cobra.Command, args []string) error {
	if db != nil {
		return db.Close()
	}
	return nil
}
