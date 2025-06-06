package commands

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/terzigolu/josepshbrain-go/internal/cli/interactive"
	"github.com/terzigolu/josepshbrain-go/pkg/models"
	"golang.org/x/term"
	"gorm.io/gorm"
)

// NewMemoryCmd creates the memory command with all subcommands
func NewMemoryCmd(db *gorm.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memory",
		Short: "Memory management commands",
		Long:  "Create, list, and manage memories",
	}

	// Add subcommands
	cmd.AddCommand(newMemoryAddCmd(db))
	cmd.AddCommand(newMemoryListCmd(db))
	cmd.AddCommand(newMemoryRecallCmd(db))
	cmd.AddCommand(newMemoryForgetCmd(db))
	cmd.AddCommand(newMemoryInfoCmd(db))
	cmd.AddCommand(newMemoryModifyCmd(db))

	return cmd
}

// NewRememberCmd creates the remember command (shortcut for memory add)
func NewRememberCmd(db *gorm.DB) *cobra.Command {
	return &cobra.Command{
		Use:   "remember [text]",
		Short: "Store a new memory",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			createMemory(db, strings.Join(args, " "))
		},
	}
}

// NewMemoriesCmd creates the memories command (shortcut for memory list)
func NewMemoriesCmd(db *gorm.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "memories",
		Short:   "List all memories",
		Aliases: []string{"recall"},
		Run: func(cmd *cobra.Command, args []string) {
			all, _ := cmd.Flags().GetBool("all")
			listMemories(db, "", all)
		},
	}
	
	cmd.Flags().BoolP("all", "a", false, "Show memories from all projects")
	return cmd
}

// memory add - create new memory
func newMemoryAddCmd(db *gorm.DB) *cobra.Command {
	return &cobra.Command{
		Use:     "add [text]",
		Short:   "Store a new memory",
		Aliases: []string{"create"},
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			createMemory(db, strings.Join(args, " "))
		},
	}
}

// memory list - list all memories
func newMemoryListCmd(db *gorm.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all memories",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			search, _ := cmd.Flags().GetString("search")
			all, _ := cmd.Flags().GetBool("all")
			listMemories(db, search, all)
		},
	}
	
	cmd.Flags().StringP("search", "s", "", "Search term to filter memories")
	cmd.Flags().BoolP("all", "a", false, "Show memories from all projects")
	return cmd
}

// memory recall - search memories
func newMemoryRecallCmd(db *gorm.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recall [search_term]",
		Short: "Search and recall memories",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			searchTerm := strings.Join(args, " ")
			all, _ := cmd.Flags().GetBool("all")
			listMemories(db, searchTerm, all)
		},
	}
	
	cmd.Flags().BoolP("all", "a", false, "Show memories from all projects")
	return cmd
}

// memory forget - delete memory
func newMemoryForgetCmd(db *gorm.DB) *cobra.Command {
	return &cobra.Command{
		Use:   "forget [memory_id]",
		Short: "Delete a memory",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			memoryID := args[0]
			forgetMemory(db, memoryID)
		},
	}
}

// memory info - show detailed memory information
func newMemoryInfoCmd(db *gorm.DB) *cobra.Command {
	return &cobra.Command{
		Use:   "info [id]",
		Short: "Show detailed memory information",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			memoryID := args[0]
			
			var memory models.MemoryItem
			if err := db.Preload("Project").Preload("Tags").Preload("TaskLinks").Where("id::text LIKE ?", memoryID+"%").First(&memory).Error; err != nil {
				log.Fatalf("Memory not found: %v", err)
			}

			fmt.Println("🧠 Memory Details:")
			fmt.Println("================================================================================")
			fmt.Printf("📝 ID:          %s\n", memory.ID.String())
			fmt.Printf("📄 Content:     %s\n", memory.Content)
			if memory.Project != nil {
				fmt.Printf("🏢 Project:     %s\n", memory.Project.Name)
			} else {
				fmt.Printf("🏢 Project:     N/A\n")
			}
			fmt.Printf("📅 Created:     %s\n", memory.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("🔄 Updated:     %s\n", memory.UpdatedAt.Format("2006-01-02 15:04:05"))
			
			if len(memory.Tags) > 0 {
				fmt.Printf("\n🏷️  Tags (%d):\n", len(memory.Tags))
				for i, tag := range memory.Tags {
					fmt.Printf("  %d. %s\n", i+1, tag.Name)
				}
			} else {
				fmt.Println("\n🏷️  Tags: None")
			}
			
			if len(memory.TaskLinks) > 0 {
				fmt.Printf("\n📋 Task Links (%d):\n", len(memory.TaskLinks))
				for i, taskLink := range memory.TaskLinks {
					fmt.Printf("  %d. Task ID: %s\n", i+1, taskLink.TaskID.String()[:8])
					if taskLink.RelationType != "" {
						fmt.Printf("     Type: %s\n", taskLink.RelationType)
					}
					if taskLink.Confidence > 0 {
						fmt.Printf("     Confidence: %.2f\n", taskLink.Confidence)
					}
				}
			} else {
				fmt.Println("\n📋 Task Links: None")
			}
			
			fmt.Println("================================================================================")
		},
	}
}

// Helper functions

func createMemory(db *gorm.DB, text string) {
	// Get active project
	var activeProject models.Project
	result := db.Where("is_active = ? AND deleted_at IS NULL", true).First(&activeProject)
	if result.Error != nil {
		fmt.Println("❌ No active project found")
		fmt.Println("💡 Use 'jbraincli project init <n>' to create a project")
		return
	}

	// Create new memory in memory_items table (based on schema)
	memoryItem := models.MemoryItem{
		ProjectID: &activeProject.ID,
		Content:   text,
	}

	if err := db.Create(&memoryItem).Error; err != nil {
		log.Fatalf("Failed to create memory: %v", err)
	}

	fmt.Printf("🧠 Memory stored successfully!\n")
	fmt.Printf("📋 Memory ID: %s\n", memoryItem.ID.String()[:8])
	fmt.Printf("📝 Content: %s\n", truncateString(text, 100))
	fmt.Printf("📁 Project: %s\n", activeProject.Name)
}

func listMemories(db *gorm.DB, searchTerm string, showAll bool) {
	var activeProject models.Project
	var projectName string

	if !showAll {
		// Get active project
		result := db.Where("is_active = ? AND deleted_at IS NULL", true).First(&activeProject)
		if result.Error != nil {
			fmt.Println("❌ No active project found")
			fmt.Println("💡 Use --all flag to see all memories or set an active project")
			return
		}
		projectName = activeProject.Name
	}

	// Build query for memory_items table
	query := db.Order("created_at DESC")

	if showAll {
		query = query.Preload("Project")
	}

	if !showAll {
		query = query.Where("project_id = ?", activeProject.ID)
	}

	if searchTerm != "" {
		query = query.Where("content ILIKE ?", "%"+searchTerm+"%")
	}

	var memories []models.MemoryItem
	if err := query.Find(&memories).Error; err != nil {
		log.Fatalf("Failed to fetch memories: %v", err)
	}

	if len(memories) == 0 {
		if searchTerm != "" {
			fmt.Printf("🔍 No memories found matching '%s'\n", searchTerm)
		} else {
			if showAll {
				fmt.Println("🧠 No memories found. Create one with 'jbraincli remember <text>'")
			} else {
				fmt.Printf("🧠 No memories found for project '%s'. Create one with 'jbraincli remember <text>'\n", projectName)
			}
		}
		return
	}

	displayMemoryList(memories, projectName, showAll, searchTerm)
}

func forgetMemory(db *gorm.DB, memoryID string) {
	// Find memory by ID (partial match) in memory_items table
	var memory models.MemoryItem
	result := db.Where("id::text LIKE ? AND deleted_at IS NULL", memoryID+"%").First(&memory)
	if result.Error != nil {
		fmt.Printf("❌ Memory with ID '%s' not found\n", memoryID)
		return
	}

	// Soft delete the memory (GORM's default behavior with DeletedAt)
	if err := db.Delete(&memory).Error; err != nil {
		log.Fatalf("Failed to delete memory: %v", err)
	}

	fmt.Printf("🗑️  Memory deleted successfully!\n")
	fmt.Printf("📋 Memory ID: %s\n", memory.ID.String()[:8])
	fmt.Printf("📝 Content: %s\n", truncateString(memory.Content, 100))
}

// displayMemoryList shows memories in a beautiful, responsive format
func displayMemoryList(memories []models.MemoryItem, projectName string, showAll bool, searchTerm string) {
	var width int = 80 // default width
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width = w
	}

	// Header
	if searchTerm != "" {
		fmt.Printf("🔍 Search results for '%s' (%d found):\n", searchTerm, len(memories))
	} else {
		if showAll {
			fmt.Printf("🧠 All memories (%d total):\n", len(memories))
		} else {
			fmt.Printf("🧠 Memories for project '%s' (%d total):\n", projectName, len(memories))
		}
	}

	uniqueIDs := generateUniqueShortIDsForMemories(memories)

	if width < 100 {
		displayMemoryListCompact(memories, uniqueIDs, showAll)
	} else {
		displayMemoryListTable(memories, uniqueIDs, showAll, width)
	}
}

// displayMemoryListCompact shows memories in a compact format for narrow terminals
func displayMemoryListCompact(memories []models.MemoryItem, uniqueIDs map[string]string, showAll bool) {
	fmt.Println()
	for i, memory := range memories {
		fmt.Printf("🆔 %s  📅 %s\n",
			uniqueIDs[memory.ID.String()],
			memory.CreatedAt.Format("2006-01-02 15:04"))

		if showAll && memory.Project != nil {
			fmt.Printf("   📁 %s\n", memory.Project.Name)
		}

		fmt.Printf("   📝 %s\n", memory.Content)

		if i < len(memories)-1 {
			fmt.Println("   " + strings.Repeat("─", 40))
		}
	}
}

// displayMemoryListTable shows memories in a full table format for wide terminals
func displayMemoryListTable(memories []models.MemoryItem, uniqueIDs map[string]string, showAll bool, termWidth int) {
	idWidth := 10
	createdWidth := 18
	projectWidth := 0
	if showAll {
		projectWidth = 20
	}

	usedWidth := idWidth + createdWidth + projectWidth + 8 // borders and spaces
	contentWidth := termWidth - usedWidth
	if contentWidth < 30 {
		contentWidth = 30
	}

	// Table Header
	fmt.Println()
	if showAll {
		fmt.Printf("┌─%-*s─┬─%-*s─┬─%-*s─┬─%-*s─┐\n",
			idWidth, strings.Repeat("─", idWidth),
			createdWidth, strings.Repeat("─", createdWidth),
			projectWidth, strings.Repeat("─", projectWidth),
			contentWidth, strings.Repeat("─", contentWidth))

		fmt.Printf("│ %-*s │ %-*s │ %-*s │ %-*s │\n",
			idWidth, "ID",
			createdWidth, "CREATED",
			projectWidth, "PROJECT",
			contentWidth, "CONTENT")

		fmt.Printf("├─%-*s─┼─%-*s─┼─%-*s─┼─%-*s─┤\n",
			idWidth, strings.Repeat("─", idWidth),
			createdWidth, strings.Repeat("─", createdWidth),
			projectWidth, strings.Repeat("─", projectWidth),
			contentWidth, strings.Repeat("─", contentWidth))
	} else {
		fmt.Printf("┌─%-*s─┬─%-*s─┬─%-*s─┐\n",
			idWidth, strings.Repeat("─", idWidth),
			createdWidth, strings.Repeat("─", createdWidth),
			contentWidth, strings.Repeat("─", contentWidth))

		fmt.Printf("│ %-*s │ %-*s │ %-*s │\n",
			idWidth, "ID",
			createdWidth, "CREATED",
			contentWidth, "CONTENT")

		fmt.Printf("├─%-*s─┼─%-*s─┼─%-*s─┤\n",
			idWidth, strings.Repeat("─", idWidth),
			createdWidth, strings.Repeat("─", createdWidth),
			contentWidth, strings.Repeat("─", contentWidth))
	}

	// Table Rows
	for _, memory := range memories {
		wrappedContent := wrapString(memory.Content, contentWidth)
		lines := strings.Split(wrappedContent, "\n")

		projectName := ""
		if showAll && memory.Project != nil {
			projectName = truncateString(memory.Project.Name, projectWidth)
		}

		for i, line := range lines {
			if i == 0 {
				if showAll {
					fmt.Printf("│ %-*s │ %-*s │ %-*s │ %-*s │\n",
						idWidth, uniqueIDs[memory.ID.String()],
						createdWidth, memory.CreatedAt.Format("2006-01-02 15:04"),
						projectWidth, projectName,
						contentWidth, line)
				} else {
					fmt.Printf("│ %-*s │ %-*s │ %-*s │\n",
						idWidth, uniqueIDs[memory.ID.String()],
						createdWidth, memory.CreatedAt.Format("2006-01-02 15:04"),
						contentWidth, line)
				}
			} else {
				if showAll {
					fmt.Printf("│ %-*s │ %-*s │ %-*s │ %-*s │\n",
						idWidth, "", createdWidth, "", projectWidth, "", contentWidth, line)
				} else {
					fmt.Printf("│ %-*s │ %-*s │ %-*s │\n",
						idWidth, "", createdWidth, "", contentWidth, line)
				}
			}
		}
	}

	// Table Footer
	if showAll {
		fmt.Printf("└─%-*s─┴─%-*s─┴─%-*s─┴─%-*s─┘\n",
			idWidth, strings.Repeat("─", idWidth),
			createdWidth, strings.Repeat("─", createdWidth),
			projectWidth, strings.Repeat("─", projectWidth),
			contentWidth, strings.Repeat("─", contentWidth))
	} else {
		fmt.Printf("└─%-*s─┴─%-*s─┴─%-*s─┘\n",
			idWidth, strings.Repeat("─", idWidth),
			createdWidth, strings.Repeat("─", createdWidth),
			contentWidth, strings.Repeat("─", contentWidth))
	}
}

// Helper functions for display

func generateUniqueShortIDsForMemories(memories []models.MemoryItem) map[string]string {
	uniqueIDs := make(map[string]string)
	usedShortIDs := make(map[string][]string)

	// First pass: try 8-character IDs
	for _, memory := range memories {
		fullID := memory.ID.String()
		shortID := fullID[:8]
		usedShortIDs[shortID] = append(usedShortIDs[shortID], fullID)
	}

	// Second pass: resolve collisions
	for shortID, fullIDs := range usedShortIDs {
		if len(fullIDs) == 1 {
			uniqueIDs[fullIDs[0]] = shortID
		} else {
			for _, fullID := range fullIDs {
				// Start with a longer length to resolve collision
				for length := 9; length < 36; length++ {
					candidate := fullID[:length]
					isUnique := true
					for _, otherID := range fullIDs {
						if otherID != fullID && strings.HasPrefix(otherID, candidate) {
							isUnique = false
							break
						}
					}
					if isUnique {
						uniqueIDs[fullID] = candidate
						break
					}
				}
			}
		}
	}
	return uniqueIDs
}

func wrapString(text string, lineWidth int) string {
	if len(text) <= lineWidth {
		return text
	}
	var result strings.Builder
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}
	currentLine := words[0]
	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) > lineWidth {
			result.WriteString(currentLine + "\n")
			currentLine = word
		} else {
			currentLine += " " + word
		}
	}
	result.WriteString(currentLine)
	return result.String()
}

// memory modify - modify memory content
func newMemoryModifyCmd(db *gorm.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "modify [id]",
		Short:   "Modify a memory's content",
		Aliases: []string{"update", "edit"},
		Args:    cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {
			isInteractive, _ := cmd.Flags().GetBool("interactive")
			newContent, _ := cmd.Flags().GetString("content")

			var memory *models.MemoryItem
			var err error

			if isInteractive {
				var allMemories []models.MemoryItem
				db.Order("updated_at desc").Find(&allMemories)
				memory, err = interactive.SelectMemory(allMemories, "Select memory to modify:")
				if err != nil {
					fmt.Println("Memory selection cancelled.")
					return
				}
				// Now, let's get the modifications interactively
				if newContent == "" {
					prompt := &survey.Multiline{
						Message: "New content (leave blank to keep current):",
						Default: memory.Content,
					}
					survey.AskOne(prompt, &newContent)
				}

			} else {
				if len(args) < 1 {
					fmt.Println("Memory ID is required for non-interactive modification.")
					return
				}
				memory, err = getMemoryByIDPrefix(db, args[0])
				if err != nil {
					log.Fatalf(err.Error())
				}
			}

			// Apply modifications
			modified := false
			if newContent != "" && newContent != memory.Content {
				memory.Content = newContent
				modified = true
			}

			if !modified {
				fmt.Println("No changes specified. Memory not modified.")
				return
			}

			if err := db.Save(memory).Error; err != nil {
				log.Fatalf("Failed to modify memory: %v", err)
			}
			fmt.Printf("✅ Successfully modified memory: %s\n", truncateString(memory.Content, 50))
		},
	}

	cmd.Flags().BoolP("interactive", "i", false, "Modify a memory interactively")
	cmd.Flags().StringP("content", "c", "", "New memory content")

	return cmd
}

// Helper function to get memory by ID prefix
func getMemoryByIDPrefix(db *gorm.DB, memoryID string) (*models.MemoryItem, error) {
	var memory models.MemoryItem
	result := db.Where("id::text LIKE ?", memoryID+"%").First(&memory)
	if result.Error != nil {
		return nil, fmt.Errorf("memory with ID '%s' not found", memoryID)
	}
	return &memory, nil
}