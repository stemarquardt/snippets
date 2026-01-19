# Snippets

A todo list summarizer that gives a running high level overview (snippets) of the work you've been doing!

Getting a daily, weekly, sprintly, quarterly update can be tricky - you have to look through all of your Jira tickets, PRs, notes, and todo list items. This is annoying. What if you just put more details in your todo list app and summarized it using GenAI? That would help keep a running high level overview of the work / goals / etc. you have been working on. It won't help with Jira tracking, but it should help you recognize the work you've been doing all week on Friday when you feel like you've forgotten every single thing you did during the week.

This is very much a WIP hobby project, but I like to keep a pulse on the themes of my work (especially when there's a lot of interrupt work going on). So I'm mostly writing this for me, but maybe it'll help someone else out too.

This is starting out using Todoist (because that's what I use) but I'm sure you could plug in some other service into `internal/clients/` and leverage it however you want.

Next things up:

- [ ] Extract more info from tasks, especially sub tasks
- [ ] Gen AI summary
- [ ] Store rollups in SQLite db
- [ ] Clean up the CLI tooling

## Project Structure

```
snippets/
├── README.md              
├── go.mod                
├── .gitignore            
├── Makefile              
├── cmd/
│   └── snippets/
│       └── main.go       # Main application entry point
├── internal/             # Private application code
│   ├── clients/          # Third-party API clients
│   │   ├── todoist/       # Todoist API integration
│   ├── models/           # Data models
```

## Getting Started

### Prerequisites

- Go 1.21 or later

### Building

Build the binary:
```bash
make build
```

The binary will be created in the `build/` directory.

### Running

Run the CLI tool:
```bash
# Using default database location
./build/snippets --db-filepath ./snippets.db

# Or with environment variables set
export TODOIST_API_TOKEN="your_todoist_token"
export CLAUDE_API_KEY="your_claude_key"
./build/snippets --db-filepath ./data/snippets.db

# Without environment variables (will prompt securely)
./build/snippets --db-filepath ./snippets.db
# You'll be prompted to enter API tokens with hidden input
```

### Development

Format code:
```bash
make fmt
```

Run static analysis:
```bash
make vet
```

Run tests:
```bash
make test
```

Clean build artifacts:
```bash
make clean
```

## Todoist Integration

The service includes a comprehensive Todoist API client for integrating with Todoist accounts.

### Getting Your Todoist API Token

1. Log in to [Todoist](https://todoist.com)
2. Go to **Settings** → **Integrations** → **Developer**
3. Scroll to the **API token** section
4. Copy your API token

Or visit directly: [https://todoist.com/prefs/integrations](https://todoist.com/prefs/integrations)

### Configuration

Set your Todoist API token as an environment variable:
```bash
export TODOIST_API_TOKEN="your_token_here"
```

The CLI will prompt you securely if the environment variable is not set.

### Usage

```go
import "snippets/internal/clients/todoist"

// Initialize client
client := todoist.NewClient("your_api_token")

// Validate token
err := client.ValidateToken()

// Get projects
projects, err := client.GetProjects()

// Get all tasks
tasks, err := client.GetAllTasks()

// Get tasks by project
projectTasks, err := client.GetTasksByProject("project_id")

// Get completed tasks in time window
since := time.Now().AddDate(0, 0, -7) // 7 days ago
until := time.Now()
completed, err := client.GetCompletedTasksInTimeWindow(since, until)

// Helper methods for common time ranges
todayCompleted, err := client.GetCompletedTasksToday()
weekCompleted, err := client.GetCompletedTasksThisWeek()
```

### Features

- **Authentication**: Bearer token authentication with validation
- **Projects**: Fetch all projects or get specific project by ID
- **Tasks**: Get all active tasks or filter by project
- **Completed Tasks**: Fetch completed tasks with flexible time window filtering
- **Business Week Support**: Monday-Sunday week tracking for weekly summaries
- **Error Handling**: Custom API error types with detailed error messages
- **Time Helpers**: Built-in methods for common time ranges (today, this week, business weeks)

### Business Week Concept

This tool uses a "business week" concept where each week runs from **Monday to Sunday**. When fetching weekly summaries:

- If it's **Thursday**, the current business week includes Monday-Thursday only
- If it's **Sunday**, the current business week includes the full Monday-Sunday
- This ensures summaries reflect actual completed work, not future days

**Example usage:**

```go
// Get tasks for current business week (Monday to today)
tasks, err := client.GetCompletedTasksForCurrentBusinessWeek()

// Get tasks for full business week (Monday to Sunday)
fullWeek, err := client.GetCompletedTasksForCurrentFullBusinessWeek()

// Get last 4 weeks of tasks for trend analysis
weeklyTasks, err := client.GetCompletedTasksForPreviousBusinessWeeks(4)

// Get business week boundaries
week := todoist.GetCurrentBusinessWeek()
fmt.Printf("Week: %s\n", week.String()) // "Dec 2 - Dec 8, 2024"
```

