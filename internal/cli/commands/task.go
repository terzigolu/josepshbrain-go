package commands

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/terzigolu/josepshbrain-go/internal/cli/interactive"
	"github.com/terzigolu/josepshbrain-go/pkg/models"
	"golang.org/x/term"
	"gorm.io/gorm"
)

// NewTaskCmd creates the task command with all subcommands
func NewTaskCmd(db *gorm.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Task management commands",
		Long:  "Create, list, update, and manage tasks",
	}

	// Add subcommands with database
	cmd.AddCommand(newTaskCreateCmd(db))
	cmd.AddCommand(newTaskListCmd(db))
	cmd.AddCommand(newTaskStartCmd(db))
	cmd.AddCommand(newTaskDoneCmd(db))
	cmd.AddCommand(newTaskInfoCmd(db))
	cmd.AddCommand(newTaskProgressCmd(db))
	cmd.AddCommand(newTaskDeleteCmd(db))
	cmd.AddCommand(newTaskModifyCmd(db))

	return cmd
}

// task create
func newTaskCreateCmd(db *gorm.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create [description]",
		Short:   "Create a new task",
		Aliases: []string{"add"},
		Run: func(cmd *cobra.Command, args []string) {
			isInteractive, _ := cmd.Flags().GetBool("interactive")
			priority, _ := cmd.Flags().GetString("priority")
			description := ""

			if isInteractive {
				// Interactive mode
				task, err := interactive.CreateTaskInteractive()
				if err != nil {
					log.Fatalf("Interactive task creation failed: %v", err)
				}
				description = task.Description
				// Priority from interactive mode overrides flag
				if task.Priority != "" {
					priority = task.Priority
				}
			} else {
				// Traditional CLI mode
				if len(args) < 1 {
					fmt.Println("❌ Description is required when not in interactive mode.")
					fmt.Println("💡 Use 'jbraincli task create \"My new task\"' or 'jbraincli task create -i'")
					return
				}
				description = strings.Join(args, " ")
			}
			
			// Get active project - require one to exist
			var project models.Project
			result := db.Where("is_active = ? AND deleted_at IS NULL", true).First(&project)
			if result.Error != nil {
				fmt.Println("❌ No active project found")
				fmt.Println("💡 Use 'jbraincli init <name>' to create a project first")
				return
			}

			// Create new task
			task := models.Task{
				ProjectID:   project.ID,
				Description: description,
				Status:      string(models.TaskStatusTODO),
				Priority:    strings.ToUpper(priority),
				Progress:    0,
			}

			if err := db.Create(&task).Error; err != nil {
				log.Fatalf("Failed to create task: %v", err)
			}

			fmt.Printf("🔄 Created task: %s\n", description)
			fmt.Printf("✅ Task ID: %s\n", task.ID.String())
		},
	}
	
	// Add interactive flag and priority flag
	cmd.Flags().BoolP("interactive", "i", false, "Use interactive mode for task creation")
	cmd.Flags().StringP("priority", "p", "M", "Set task priority (L, M, H)")
	
	// If not in interactive mode, description is required
	cmd.Args = func(cmd *cobra.Command, args []string) error {
		isInteractive, _ := cmd.Flags().GetBool("interactive")
		if !isInteractive && len(args) < 1 {
			return fmt.Errorf("requires a description when not in interactive mode")
		}
		return nil
	}
	
	return cmd
}

// task list
func newTaskListCmd(db *gorm.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List tasks for the active project",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			allProjects, _ := cmd.Flags().GetBool("all")
			status, _ := cmd.Flags().GetString("status")
			
			// Get active project unless --all flag is used
			var project models.Project
			if !allProjects {
				result := db.Where("is_active = ? AND deleted_at IS NULL", true).First(&project)
				if result.Error != nil {
					fmt.Println("❌ No active project found")
					fmt.Println("💡 Use 'jbraincli use <project>' to set an active project")
					fmt.Println("💡 Or use --all flag to see tasks from all projects")
					return
				}
			}

			// Build query for tasks
			query := db.Preload("Project")
			if !allProjects {
				query = query.Where("project_id = ?", project.ID)
			}
			if status != "" {
				query = query.Where("status = ?", strings.ToUpper(status))
			}

			var tasks []models.Task
			if err := query.Find(&tasks).Error; err != nil {
				log.Fatalf("Failed to fetch tasks: %v", err)
			}

			if len(tasks) == 0 {
				if !allProjects {
					fmt.Printf("📋 No tasks found in project '%s'\n", project.Name)
				} else {
					fmt.Println("📋 No tasks found in any project")
				}
				fmt.Println("💡 Create one with 'jbraincli task create <description>'")
				return
			}

			// Display beautiful task list
			displayTaskList(tasks, project.Name, allProjects, status)
		},
	}
	
	cmd.Flags().BoolP("all", "a", false, "Show tasks from all projects")
	cmd.Flags().StringP("status", "s", "", "Filter by status (TODO, IN_PROGRESS, IN_REVIEW, COMPLETED)")
	
	return cmd
}

// task start
func newTaskStartCmd(db *gorm.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start [id]",
		Short: "Start working on a task",
		Args:  cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {
			isInteractive, _ := cmd.Flags().GetBool("interactive")
			
			var task models.Task
			
			if isInteractive {
				// Interactive mode - select from TODO tasks
				var todoTasks []models.Task
				if err := db.Where("status = ?", "TODO").Find(&todoTasks).Error; err != nil {
					log.Fatalf("Failed to fetch TODO tasks: %v", err)
				}
				
				if len(todoTasks) == 0 {
					fmt.Println("📋 No TODO tasks available to start")
					return
				}
				
				selectedTask, err := interactive.SelectTask(todoTasks, "Select task to start:")
				if err != nil {
					log.Fatalf("Task selection failed: %v", err)
				}
				task = *selectedTask
			} else {
				// Traditional CLI mode
				if len(args) == 0 {
					fmt.Println("❌ Task ID required (or use --interactive)")
					return
				}
				taskID := args[0]
				
				if err := db.Where("id::text LIKE ?", taskID+"%").First(&task).Error; err != nil {
					log.Fatalf("Task not found: %v", err)
				}
			}

			task.Status = string(models.TaskStatusInProgress)
			if err := db.Save(&task).Error; err != nil {
				log.Fatalf("Failed to update task: %v", err)
			}

			fmt.Printf("▶️ Started task: %s\n", task.Description)
			fmt.Println("✅ Task status updated to IN_PROGRESS!")
		},
	}
	
	// Add interactive flag
	cmd.Flags().BoolP("interactive", "i", false, "Use interactive mode for task selection")
	
	return cmd
}

// task done
func newTaskDoneCmd(db *gorm.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "done [id]",
		Short: "Mark task as completed",
		Args:  cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {
			isInteractive, _ := cmd.Flags().GetBool("interactive")
			
			var task models.Task
			
			if isInteractive {
				// Interactive mode - select from active tasks
				var activeTasks []models.Task
				if err := db.Where("status IN (?)", []string{"IN_PROGRESS", "IN_REVIEW"}).Find(&activeTasks).Error; err != nil {
					log.Fatalf("Failed to fetch active tasks: %v", err)
				}
				
				if len(activeTasks) == 0 {
					fmt.Println("📋 No active tasks to complete")
					return
				}
				
				selectedTask, err := interactive.SelectTask(activeTasks, "Select task to complete:")
				if err != nil {
					log.Fatalf("Task selection failed: %v", err)
				}
				task = *selectedTask
			} else {
				// Traditional CLI mode
				if len(args) == 0 {
					fmt.Println("❌ Task ID required (or use --interactive)")
					return
				}
				taskID := args[0]
				
				if err := db.Where("id::text LIKE ?", taskID+"%").First(&task).Error; err != nil {
					log.Fatalf("Task not found: %v", err)
				}
			}

			task.Status = string(models.TaskStatusCompleted)
			task.Progress = 100
			if err := db.Save(&task).Error; err != nil {
				log.Fatalf("Failed to update task: %v", err)
			}

			fmt.Printf("✅ Task completed: %s\n", task.Description)
			fmt.Println("🎉 Great job!")
		},
	}
	
	// Add interactive flag
	cmd.Flags().BoolP("interactive", "i", false, "Use interactive mode for task selection")
	
	return cmd
}

// task info
func newTaskInfoCmd(db *gorm.DB) *cobra.Command {
	return &cobra.Command{
		Use:   "info [id]",
		Short: "Show detailed task information",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			taskID := args[0]
			
			var task models.Task
			if err := db.Preload("Project").Preload("Annotations").Where("id::text LIKE ?", taskID+"%").First(&task).Error; err != nil {
				log.Fatalf("Task not found: %v", err)
			}

			fmt.Println("🔍 Task Details:")
			fmt.Println("================================================================================")
			fmt.Printf("📝 ID:          %s\n", task.ID.String())
			fmt.Printf("📋 Description: %s\n", task.Description)
			fmt.Printf("📊 Status:      %s\n", task.Status)
			fmt.Printf("⚡ Priority:    %s\n", task.Priority)
			fmt.Printf("📈 Progress:    %d%%\n", task.Progress)
			fmt.Printf("🏢 Project:     %s\n", task.Project.Name)
			fmt.Printf("📅 Created:     %s\n", task.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("🔄 Updated:     %s\n", task.UpdatedAt.Format("2006-01-02 15:04:05"))
			
			if len(task.Annotations) > 0 {
				fmt.Printf("\n📝 Annotations (%d):\n", len(task.Annotations))
				for i, annotation := range task.Annotations {
					fmt.Printf("  %d. %s\n", i+1, annotation.Content)
					fmt.Printf("     📅 %s\n", annotation.CreatedAt.Format("2006-01-02 15:04:05"))
				}
			} else {
				fmt.Println("\n📝 Annotations: None")
			}
			
			fmt.Println("================================================================================")
		},
	}
}

// task progress
func newTaskProgressCmd(db *gorm.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "progress [id] [percentage]",
		Short: "Update task progress (0-100)",
		Args:  cobra.RangeArgs(0, 2),
		Run: func(cmd *cobra.Command, args []string) {
			var task *models.Task
			var progress int
			var err error

			isInteractive, _ := cmd.Flags().GetBool("interactive")

			if isInteractive {
				// Interactive Mode
				var tasksToUpdate []models.Task
				db.Where("status = ?", "IN_PROGRESS").Find(&tasksToUpdate)
				if len(tasksToUpdate) == 0 {
					fmt.Println("No 'IN_PROGRESS' tasks to update.")
					return
				}
				task, err = interactive.SelectTask(tasksToUpdate, "Select task to update progress:")
				if err != nil {
					fmt.Println("Task selection cancelled.")
					return
				}

				prompt := &survey.Input{Message: "Enter progress percentage (0-100):"}
				var progressStr string
				survey.AskOne(prompt, &progressStr, survey.WithValidator(survey.Required))
				progress, err = strconv.Atoi(progressStr)
				if err != nil || progress < 0 || progress > 100 {
					fmt.Println("Invalid percentage. Please enter a number between 0 and 100.")
					return
				}
			} else {
				// Command-line Mode
				if len(args) < 2 {
					fmt.Println("Task ID and percentage are required in non-interactive mode.")
					return
				}
				task, err = getTaskByIDPrefix(db, args[0])
				if err != nil {
					log.Fatalf(err.Error())
				}
				progress, err = strconv.Atoi(args[1])
				if err != nil || progress < 0 || progress > 100 {
					log.Fatalf("Invalid percentage. Must be a number between 0 and 100.")
				}
			}

			// Update task progress
			task.Progress = progress
			if progress == 100 {
				task.Status = "COMPLETED"
			} else if progress > 0 && task.Status == "TODO" {
				task.Status = "IN_PROGRESS"
			}

			if err := db.Save(task).Error; err != nil {
				log.Fatalf("Failed to update task progress: %v", err)
			}

			fmt.Printf("✅ Updated progress for task: %s\n", truncateString(task.Description, 50))
			fmt.Printf("   New Progress: %d%%, Status: %s\n", task.Progress, task.Status)
		},
	}
	cmd.Flags().BoolP("interactive", "i", false, "Update progress interactively")
	return cmd
}

func getTaskByIDPrefix(db *gorm.DB, idPrefix string) (*models.Task, error) {
	var task models.Task
	if err := db.Where("id::text LIKE ?", idPrefix+"%").First(&task).Error; err != nil {
		return nil, fmt.Errorf("task with ID prefix '%s' not found", idPrefix)
	}
	return &task, nil
}

// displayTaskList shows tasks in a beautiful, responsive format
func displayTaskList(tasks []models.Task, projectName string, allProjects bool, statusFilter string) {
	// Import terminal width detection
	var width int = 80 // default width
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width = w
	}

	// Header with project info
	if allProjects {
		if statusFilter != "" {
			fmt.Printf("📋 %s Tasks from All Projects (%d)\n", strings.ToUpper(statusFilter), len(tasks))
		} else {
			fmt.Printf("📋 All Tasks from All Projects (%d)\n", len(tasks))
		}
	} else {
		if statusFilter != "" {
			fmt.Printf("📋 %s Tasks - %s (%d)\n", strings.ToUpper(statusFilter), projectName, len(tasks))
		} else {
			fmt.Printf("📋 Tasks - %s (%d)\n", projectName, len(tasks))
		}
	}

	// Generate unique short IDs (reuse from kanban)
	uniqueIDs := generateUniqueShortIDsForTasks(tasks)

	// Responsive design
	if width < 100 {
		// Compact view for narrow terminals
		displayTaskListCompact(tasks, uniqueIDs, allProjects)
	} else {
		// Full table view for wide terminals
		displayTaskListTable(tasks, uniqueIDs, allProjects, width)
	}
}

// displayTaskListCompact shows tasks in compact format
func displayTaskListCompact(tasks []models.Task, uniqueIDs map[string]string, allProjects bool) {
	fmt.Println()
	for i, task := range tasks {
		// Priority and status icons
		priorityIcon := getPriorityIconForTask(task.Priority)
		statusIcon := getStatusIconForTask(task.Status)
		
		// Progress indicator
		progressBar := getProgressBar(task.Progress, 8)
		
		fmt.Printf("%s %s %s %s\n", 
			priorityIcon, 
			statusIcon, 
			uniqueIDs[task.ID.String()], 
			task.Description)
		
		if allProjects && task.Project != nil {
			fmt.Printf("   🏢 %s", task.Project.Name)
		}
		
		if task.Progress > 0 {
			fmt.Printf("   %s %d%%", progressBar, task.Progress)
		}
		
		fmt.Println()
		
		// Add separator between tasks (except last)
		if i < len(tasks)-1 {
			fmt.Println("   ─────────────────────────────────────")
		}
	}
}

// displayTaskListTable shows tasks in full table format  
func displayTaskListTable(tasks []models.Task, uniqueIDs map[string]string, allProjects bool, termWidth int) {
	// Calculate dynamic column widths
	idWidth := 12
	priorityWidth := 4
	statusWidth := 12
	progressWidth := 12
	projectWidth := 0
	if allProjects {
		projectWidth = 20
	}
	
	// Remaining width for description
	usedWidth := idWidth + priorityWidth + statusWidth + progressWidth + projectWidth + 8 // borders and spaces
	descWidth := termWidth - usedWidth
	if descWidth < 30 {
		descWidth = 30
	}

	// Table header
	fmt.Println()
	if allProjects {
		fmt.Printf("┌─%-*s─┬─%-*s─┬─%-*s─┬─%-*s─┬─%-*s─┬─%-*s─┐\n", 
			idWidth, strings.Repeat("─", idWidth),
			priorityWidth, strings.Repeat("─", priorityWidth),
			statusWidth, strings.Repeat("─", statusWidth),
			progressWidth, strings.Repeat("─", progressWidth),
			projectWidth, strings.Repeat("─", projectWidth),
			descWidth, strings.Repeat("─", descWidth))
		
		fmt.Printf("│ %-*s │ %-*s │ %-*s │ %-*s │ %-*s │ %-*s │\n",
			idWidth, "ID",
			priorityWidth, "PRI",
			statusWidth, "STATUS",
			progressWidth, "PROGRESS",
			projectWidth, "PROJECT",
			descWidth, "DESCRIPTION")
	} else {
		fmt.Printf("┌─%-*s─┬─%-*s─┬─%-*s─┬─%-*s─┬─%-*s─┐\n", 
			idWidth, strings.Repeat("─", idWidth),
			priorityWidth, strings.Repeat("─", priorityWidth),
			statusWidth, strings.Repeat("─", statusWidth),
			progressWidth, strings.Repeat("─", progressWidth),
			descWidth, strings.Repeat("─", descWidth))
		
		fmt.Printf("│ %-*s │ %-*s │ %-*s │ %-*s │ %-*s │\n",
			idWidth, "ID",
			priorityWidth, "PRI", 
			statusWidth, "STATUS",
			progressWidth, "PROGRESS",
			descWidth, "DESCRIPTION")
	}

	// Separator
	if allProjects {
		fmt.Printf("├─%-*s─┼─%-*s─┼─%-*s─┼─%-*s─┼─%-*s─┼─%-*s─┤\n",
			idWidth, strings.Repeat("─", idWidth),
			priorityWidth, strings.Repeat("─", priorityWidth),
			statusWidth, strings.Repeat("─", statusWidth),
			progressWidth, strings.Repeat("─", progressWidth),
			projectWidth, strings.Repeat("─", projectWidth),
			descWidth, strings.Repeat("─", descWidth))
	} else {
		fmt.Printf("├─%-*s─┼─%-*s─┼─%-*s─┼─%-*s─┼─%-*s─┤\n",
			idWidth, strings.Repeat("─", idWidth),
			priorityWidth, strings.Repeat("─", priorityWidth),
			statusWidth, strings.Repeat("─", statusWidth),
			progressWidth, strings.Repeat("─", progressWidth),
			descWidth, strings.Repeat("─", descWidth))
	}

	// Task rows
	for _, task := range tasks {
		priorityIcon := getPriorityIconForTask(task.Priority)
		statusIcon := getStatusIconForTask(task.Status)
		progressBar := getProgressBar(task.Progress, 10)
		
		shortID := uniqueIDs[task.ID.String()]
		description := truncateString(task.Description, descWidth)
		
		if allProjects {
			projectName := ""
			if task.Project != nil {
				projectName = truncateString(task.Project.Name, projectWidth)
			}
			
			fmt.Printf("│ %-*s │ %-*s │ %-*s │ %-*s │ %-*s │ %-*s │\n",
				idWidth, shortID,
				priorityWidth, priorityIcon,
				statusWidth, statusIcon,
				progressWidth, progressBar,
				projectWidth, projectName,
				descWidth, description)
		} else {
			fmt.Printf("│ %-*s │ %-*s │ %-*s │ %-*s │ %-*s │\n",
				idWidth, shortID,
				priorityWidth, priorityIcon,
				statusWidth, statusIcon,
				progressWidth, progressBar,
				descWidth, description)
		}
	}

	// Table footer
	if allProjects {
		fmt.Printf("└─%-*s─┴─%-*s─┴─%-*s─┴─%-*s─┴─%-*s─┴─%-*s─┘\n",
			idWidth, strings.Repeat("─", idWidth),
			priorityWidth, strings.Repeat("─", priorityWidth),
			statusWidth, strings.Repeat("─", statusWidth),
			progressWidth, strings.Repeat("─", progressWidth),
			projectWidth, strings.Repeat("─", projectWidth),
			descWidth, strings.Repeat("─", descWidth))
	} else {
		fmt.Printf("└─%-*s─┴─%-*s─┴─%-*s─┴─%-*s─┴─%-*s─┘\n",
			idWidth, strings.Repeat("─", idWidth),
			priorityWidth, strings.Repeat("─", priorityWidth),
			statusWidth, strings.Repeat("─", statusWidth),
			progressWidth, strings.Repeat("─", progressWidth),
			descWidth, strings.Repeat("─", descWidth))
	}
}

// Helper functions for task list display
func generateUniqueShortIDsForTasks(tasks []models.Task) map[string]string {
	uniqueIDs := make(map[string]string)
	usedShortIDs := make(map[string][]string)
	
	// First pass: try 8-character IDs
	for _, task := range tasks {
		fullID := task.ID.String()
		shortID := fullID[:8]
		usedShortIDs[shortID] = append(usedShortIDs[shortID], fullID)
	}
	
	// Second pass: resolve collisions
	for shortID, fullIDs := range usedShortIDs {
		if len(fullIDs) == 1 {
			uniqueIDs[fullIDs[0]] = shortID
		} else {
			for _, fullID := range fullIDs {
				uniqueLen := 8
				for uniqueLen < len(fullID) {
					candidate := fullID[:uniqueLen]
					isUnique := true
					for _, otherID := range fullIDs {
						if otherID != fullID && len(otherID) > uniqueLen && otherID[:uniqueLen] == candidate {
							isUnique = false
							break
						}
					}
					if isUnique {
						break
					}
					uniqueLen++
				}
				uniqueIDs[fullID] = fullID[:uniqueLen]
			}
		}
	}
	
	return uniqueIDs
}

func getPriorityIconForTask(priority string) string {
	icons := map[string]string{
		"H": "🔴",
		"M": "🟡",
		"L": "🟢",
	}
	if icon, exists := icons[priority]; exists {
		return icon
	}
	return "⚪"
}

func getStatusIconForTask(status string) string {
	icons := map[string]string{
		"TODO":        "📋",
		"IN_PROGRESS": "🚀", 
		"IN_REVIEW":   "👀",
		"COMPLETED":   "✅",
	}
	if icon, exists := icons[status]; exists {
		return icon
	}
	return "❓"
}

func getProgressBar(progress int, width int) string {
	if progress == 0 {
		return strings.Repeat("░", width)
	}
	if progress == 100 {
		return "✅ 100%"
	}
	
	filled := (progress * width) / 100
	bar := strings.Repeat("▓", filled) + strings.Repeat("░", width-filled)
	return fmt.Sprintf("%s %d%%", bar, progress)
}

func newTaskDeleteCmd(db *gorm.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete [id]",
		Short:   "Delete a task",
		Aliases: []string{"rm", "del"},
		Args:    cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {
			var task *models.Task
			var err error

			isInteractive, _ := cmd.Flags().GetBool("interactive")

			if isInteractive {
				var allTasks []models.Task
				db.Find(&allTasks)
				if len(allTasks) == 0 {
					fmt.Println("No tasks to delete.")
					return
				}
				task, err = interactive.SelectTask(allTasks, "Select task to DELETE:")
				if err != nil {
					fmt.Println("Task selection cancelled.")
					return
				}
			} else {
				if len(args) < 1 {
					fmt.Println("Task ID is required in non-interactive mode.")
					return
				}
				task, err = getTaskByIDPrefix(db, args[0])
				if err != nil {
					log.Fatalf(err.Error())
				}
			}

			// Confirmation
			warningMessage := fmt.Sprintf("⚠️ You are about to permanently delete task '%s'.", truncateString(task.Description, 40))
			confirmed, err := interactive.ConfirmAction(warningMessage, "This action cannot be undone.")
			if err != nil || !confirmed {
				fmt.Println("🚫 Delete operation cancelled.")
				return
			}

			// Deletion
			if err := db.Delete(task).Error; err != nil {
				log.Fatalf("Failed to delete task: %v", err)
			}

			fmt.Printf("✅ Successfully deleted task: %s\n", task.Description)
		},
	}
	cmd.Flags().BoolP("interactive", "i", false, "Delete a task interactively")
	return cmd
}

func newTaskModifyCmd(db *gorm.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "modify [id]",
		Short:   "Modify a task's attributes",
		Aliases: []string{"update", "edit"},
		Args:    cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {
			isInteractive, _ := cmd.Flags().GetBool("interactive")
			newDesc, _ := cmd.Flags().GetString("description")
			newPriority, _ := cmd.Flags().GetString("priority")
			newStatus, _ := cmd.Flags().GetString("status")

			var task *models.Task
			var err error

			if isInteractive {
				var allTasks []models.Task
				db.Order("updated_at desc").Find(&allTasks)
				task, err = interactive.SelectTask(allTasks, "Select task to modify:")
				if err != nil {
					fmt.Println("Task selection cancelled.")
					return
				}
				// Now, let's get the modifications interactively
				if newDesc == "" {
					prompt := &survey.Input{Message: "New description (leave blank to keep current):", Default: task.Description}
					survey.AskOne(prompt, &newDesc)
				}
				if newPriority == "" {
					prompt := &survey.Select{
						Message: "New priority (leave blank to keep current):",
						Options: []string{"", "H", "M", "L"},
						Default: task.Priority,
					}
					survey.AskOne(prompt, &newPriority)
				}
				if newStatus == "" {
					prompt := &survey.Select{
						Message: "New status (leave blank to keep current):",
						Options: []string{"", "TODO", "IN_PROGRESS", "IN_REVIEW", "COMPLETED"},
						Default: task.Status,
					}
					survey.AskOne(prompt, &newStatus)
				}

			} else {
				if len(args) < 1 {
					fmt.Println("Task ID is required for non-interactive modification.")
					return
				}
				task, err = getTaskByIDPrefix(db, args[0])
				if err != nil {
					log.Fatalf(err.Error())
				}
			}

			// Apply modifications
			modified := false
			if newDesc != "" && newDesc != task.Description {
				task.Description = newDesc
				modified = true
			}
			if newPriority != "" && newPriority != task.Priority {
				task.Priority = newPriority
				modified = true
			}
			if newStatus != "" && newStatus != task.Status {
				task.Status = newStatus
				modified = true
			}

			if !modified {
				fmt.Println("No changes specified. Task not modified.")
				return
			}

			if err := db.Save(task).Error; err != nil {
				log.Fatalf("Failed to modify task: %v", err)
			}
			fmt.Printf("✅ Successfully modified task: %s\n", truncateString(task.Description, 50))
		},
	}

	cmd.Flags().BoolP("interactive", "i", false, "Modify a task interactively")
	cmd.Flags().StringP("description", "d", "", "New task description")
	cmd.Flags().StringP("priority", "p", "", "New priority (H, M, L)")
	cmd.Flags().StringP("status", "s", "", "New status (TODO, IN_PROGRESS, etc.)")

	return cmd
} 