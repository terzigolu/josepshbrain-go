package commands

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/terzigolu/josepshbrain-go/internal/cli/interactive"
	"github.com/terzigolu/josepshbrain-go/pkg/models"
	"golang.org/x/term"
	"gorm.io/gorm"
)

// NewProjectCmd creates the project command with all subcommands
func NewProjectCmd(db *gorm.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Project management commands",
		Long:  "Create, list, and manage projects",
	}

	// Add subcommands
	cmd.AddCommand(newProjectInitCmd(db))
	cmd.AddCommand(newProjectUseCmd(db))
	cmd.AddCommand(newProjectListCmd(db))
	cmd.AddCommand(newProjectDeleteCmd(db))

	return cmd
}

// project init - create new project
func newProjectInitCmd(db *gorm.DB) *cobra.Command {
	return &cobra.Command{
		Use:     "init [name]",
		Short:   "Create a new project",
		Aliases: []string{"create"},
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			projectName := args[0]

			// Check if project already exists
			var existingProject models.Project
			result := db.Where("name = ? AND deleted_at IS NULL", projectName).First(&existingProject)
			if result.Error == nil {
				fmt.Printf("❌ Project '%s' already exists\n", projectName)
				return
			}

			// Deactivate all other projects first
			if err := db.Model(&models.Project{}).Where("is_active = ? AND deleted_at IS NULL", true).Update("is_active", false).Error; err != nil {
				log.Fatalf("Failed to deactivate existing projects: %v", err)
			}

			// Create new project
			project := models.Project{
				Name:        projectName,
				Description: stringPtr(fmt.Sprintf("Project: %s", projectName)),
				IsActive:    true,
			}

			if err := db.Create(&project).Error; err != nil {
				log.Fatalf("Failed to create project: %v", err)
			}

			fmt.Printf("✨ Created and activated project: %s\n", projectName)
			fmt.Printf("📋 Project ID: %s\n", project.ID.String())
		},
	}
}

// project use - set active project or show current
func newProjectUseCmd(db *gorm.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use [name]",
		Short: "Set active project or show current active project",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			isInteractive, _ := cmd.Flags().GetBool("interactive")

			if isInteractive {
				// Interactive mode
				projects, err := getAllProjects(db)
				if err != nil {
					log.Fatalf("Failed to fetch projects for interactive selection: %v", err)
				}
				if len(projects) == 0 {
					fmt.Println("No projects to select. Use 'project init' to create one.")
					return
				}
				selectedProject, err := interactive.SelectProject(projects, "Select project to activate:")
				if err != nil {
					// User probably cancelled (Ctrl+C)
					fmt.Println("Project selection cancelled.")
					return
				}
				activateProject(db, selectedProject.Name)
				return
			}

			if len(args) == 0 {
				// Show current active project
				showActiveProject(db)
				return
			}

			// Activate by name
			activateProject(db, args[0])
		},
	}

	cmd.Flags().BoolP("interactive", "i", false, "Select project interactively")
	return cmd
}

func getAllProjects(db *gorm.DB) ([]models.Project, error) {
	var projects []models.Project
	err := db.Order("name asc").Find(&projects).Error
	return projects, err
}

func showActiveProject(db *gorm.DB) {
	var activeProject models.Project
	result := db.Where("is_active = ?", true).First(&activeProject)
	if result.Error != nil {
		fmt.Println("❌ No active project found.")
		fmt.Println("💡 Use 'jbraincli project use <name>' or 'jbraincli project use -i'")
		return
	}
	fmt.Printf("🎯 Active project: %s\n", activeProject.Name)
}

func activateProject(db *gorm.DB, projectName string) {
	var project models.Project
	result := db.Where("name = ?", projectName).First(&project)
	if result.Error != nil {
		fmt.Printf("❌ Project '%s' not found.\n", projectName)
		return
	}

	// Deactivate all projects
	db.Model(&models.Project{}).Where("is_active = ?", true).Update("is_active", false)

	// Activate the selected one
	project.IsActive = true
	if err := db.Save(&project).Error; err != nil {
		log.Fatalf("Failed to activate project '%s': %v", projectName, err)
	}

	fmt.Printf("✅ Activated project: %s\n", projectName)
}

// project list - list all projects
func newProjectListCmd(db *gorm.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all projects",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			var projects []models.Project
			if err := db.Order("created_at desc").Find(&projects).Error; err != nil {
				log.Fatalf("Failed to fetch projects: %v", err)
			}

			if len(projects) == 0 {
				fmt.Println("📋 No projects found. Create one with 'jbraincli project init <name>'")
				return
			}

			displayProjectList(projects)
		},
	}
	return cmd
}

func displayProjectList(projects []models.Project) {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 80 // default
	}

	fmt.Println("🏢 Project List:")

	// Column widths
	nameWidth := 25
	descWidth := 40
	if width > 120 {
		nameWidth = 30
		descWidth = width - nameWidth - 25 // Adjust for other columns
	}

	// Header
	fmt.Printf("┌─%s─┬─%s─┬─%s─┐\n", strings.Repeat("─", nameWidth), strings.Repeat("─", descWidth), "──────────")
	fmt.Printf("│ %-*s │ %-*s │ %-8s │\n", nameWidth, "PROJECT NAME", descWidth, "DESCRIPTION", "ACTIVE")
	fmt.Printf("├─%s─┼─%s─┼─%s─┤\n", strings.Repeat("─", nameWidth), strings.Repeat("─", descWidth), "──────────")

	for _, p := range projects {
		activeIcon := "  "
		if p.IsActive {
			activeIcon = "🎯"
		}

		desc := ""
		if p.Description != nil {
			desc = *p.Description
		}

		fmt.Printf("│ %-*s │ %-*s │    %-6s │\n",
			nameWidth, truncateString(p.Name, nameWidth),
			descWidth, truncateString(desc, descWidth),
			activeIcon)
	}

	fmt.Printf("└─%s─┴─%s─┴─%s─┘\n", strings.Repeat("─", nameWidth), strings.Repeat("─", descWidth), "──────────")
	fmt.Printf("\n💡 To switch projects, use 'jbraincli project use <name>'\n")
}

func newProjectDeleteCmd(db *gorm.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete [name]",
		Short:   "Delete a project and all its associated tasks",
		Aliases: []string{"rm", "del"},
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			projectName := args[0]

			// Find the project
			var project models.Project
			if err := db.Where("name = ?", projectName).First(&project).Error; err != nil {
				fmt.Printf("❌ Project '%s' not found.\n", projectName)
				return
			}

			// Count associated tasks
			var taskCount int64
			db.Model(&models.Task{}).Where("project_id = ?", project.ID).Count(&taskCount)

			// Confirmation prompt
			warningMessage := fmt.Sprintf("⚠️ You are about to delete the project '%s'.", projectName)
			details := fmt.Sprintf("This will permanently delete the project and its %d associated task(s). This action cannot be undone.", taskCount)

			confirmed, err := interactive.ConfirmAction(warningMessage, details)
			if err != nil || !confirmed {
				fmt.Println("🚫 Delete operation cancelled.")
				return
			}

			// Perform deletion
			if err := db.Delete(&project).Error; err != nil {
				log.Fatalf("❌ Failed to delete project '%s': %v", projectName, err)
			}

			fmt.Printf("✅ Successfully deleted project '%s' and its %d tasks.\n", projectName, taskCount)

			// If the deleted project was active, ensure no project is active
			if project.IsActive {
				fmt.Println("💡 The active project was deleted. Use 'jbraincli project use' to select a new one.")
			}
		},
	}
	return cmd
}

 