package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

// NewOverviewCommand creates the overview command.
func NewOverviewCommand() *cli.Command {
	return &cli.Command{
		Name:    "overview",
		Aliases: []string{"help-all"},
		Usage:   "Show all available features and commands",
		Action: func(c *cli.Context) error {
			fmt.Println(`
╔═══════════════════════════════════════════════════════════════════╗
║                    🧠 JosephsBrain CLI                            ║
║                    Feature Overview                               ║
╚═══════════════════════════════════════════════════════════════════╝

📋 TASK MANAGEMENT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  jbrain task list              List all tasks
  jbrain task create "title"    Create a new task
  jbrain task show <id>         Show task details
  jbrain task update <id>       Update task properties
  jbrain task start <id>        Start a task (IN_PROGRESS)
  jbrain task done <id>         Complete a task (COMPLETED)
  jbrain task delete <id>       Delete a task
  jbrain task duplicate <id>    Duplicate a task with notes
  jbrain task move <ids> -p X   Move tasks to another project
  jbrain task next              Show next tasks by priority
  jbrain task progress <id> N   Update task progress (0-100)
  jbrain task elaborate <id>    AI elaboration on task

📁 PROJECT MANAGEMENT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  jbrain project list           List all projects
  jbrain project create "name"  Create a new project
  jbrain project show <id>      Show project details
  jbrain project use <name>     Set active project
  jbrain project delete <id>    Delete a project

🧠 MEMORY (KNOWLEDGE BASE)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  jbrain remember "content"     Add a new memory
  jbrain memory list            List all memories
  jbrain memory search "term"   Search memories
  jbrain memory show <id>       Show memory details
  jbrain memory delete <id>     Delete a memory

🔗 LINKING (TASK ↔ MEMORY)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  jbrain link <task> <memory>   Link a task to a memory
  jbrain task-memories <id>     List memories for a task
  jbrain memory-tasks <id>      List tasks for a memory

📝 SUBTASKS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  jbrain subtask list <task>    List subtasks
  jbrain subtask add <task> "X" Add a subtask
  jbrain subtask done <t> <s>   Complete a subtask
  jbrain subtask delete <t> <s> Delete a subtask

📌 ANNOTATIONS (NOTES)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  jbrain annotate <id> "note"   Add a note to a task
  jbrain task-annotations <id>  List notes for a task

🎯 CONTEXTS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  jbrain context list           List all contexts
  jbrain context create "name"  Create a new context
  jbrain context use <name>     Set active context
  jbrain context delete <name>  Delete a context

📊 REPORTS & VIEWS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  jbrain kanban                 Kanban board view
  jbrain reports stats          Task statistics

⚙️  CONFIGURATION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  jbrain setup                  Configure authentication
  jbrain setup login            Login with credentials
  jbrain setup logout           Remove saved credentials
  jbrain setup status           Check auth status
  jbrain config                 View/edit configuration
  jbrain set-gemini-key         Set Gemini API key

🤖 MCP (Model Context Protocol)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  jbrain mcp start              Start MCP server
  jbrain mcp status             Check MCP server status

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
💡 TIP: Use 'jbrain <command> --help' for detailed command usage.
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`)
			return nil
		},
	}
}
