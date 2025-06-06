package commands

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/terzigolu/josepshbrain-go/pkg/models"
	"github.com/terzigolu/josepshbrain-go/pkg/repository"
	"gorm.io/gorm"
)

// NewAnnotationCmd creates the annotation command
func NewAnnotationCmd(db *gorm.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "annotate",
		Short:   "Add annotation to a task",
		Example: `jbraincli annotate a2e35246 "This is important to remember"`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return createAnnotation(db, args)
		},
	}

	return cmd
}

// NewTaskAnnotationsCmd creates the task-annotations command
func NewTaskAnnotationsCmd(db *gorm.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "task-annotations",
		Short:   "List annotations for a task",
		Example: `jbraincli task-annotations a2e35246`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return listTaskAnnotations(db, args)
		},
	}

	return cmd
}

func createAnnotation(db *gorm.DB, args []string) error {
	repo := repository.NewRepository(db)
	
	taskIDStr := args[0]
	content := args[1]

	// Find the task by UUID prefix
	var task *models.Task
	var err error
	
	// Try to parse as full UUID first
	if taskUUID, parseErr := uuid.Parse(taskIDStr); parseErr == nil {
		task, err = repo.Task.GetByID(taskUUID)
	} else {
		// Search by prefix
		var tasks []models.Task
		if err := db.Where("id::text LIKE ?", taskIDStr+"%").Find(&tasks).Error; err != nil {
			return fmt.Errorf("failed to search tasks: %v", err)
		}
		if len(tasks) == 0 {
			return fmt.Errorf("task not found with prefix: %s", taskIDStr)
		}
		if len(tasks) > 1 {
			return fmt.Errorf("multiple tasks found with prefix %s, please be more specific", taskIDStr)
		}
		task = &tasks[0]
	}
	
	if err != nil {
		return fmt.Errorf("task not found: %v", err)
	}

	// Create annotation
	annotation := &models.Annotation{
		ID:        uuid.New(),
		TaskID:    task.ID,
		Content:   content,
		CreatedAt: time.Now(),
	}

	if err := repo.Annotation.Create(annotation); err != nil {
		return fmt.Errorf("failed to create annotation: %v", err)
	}

	fmt.Printf("✅ Annotation added to task: %s\n", truncateString(task.Description, 50))
	fmt.Printf("   📝 %s\n", content)
	fmt.Printf("   🆔 Task ID: %s\n", task.ID.String()[:8]+"...")

	return nil
}

func listTaskAnnotations(db *gorm.DB, args []string) error {
	repo := repository.NewRepository(db)
	
	taskIDStr := args[0]

	// Find the task by UUID prefix
	var task *models.Task
	var err error
	
	// Try to parse as full UUID first
	if taskUUID, parseErr := uuid.Parse(taskIDStr); parseErr == nil {
		task, err = repo.Task.GetByID(taskUUID)
	} else {
		// Search by prefix
		var tasks []models.Task
		if err := db.Where("id::text LIKE ?", taskIDStr+"%").Find(&tasks).Error; err != nil {
			return fmt.Errorf("failed to search tasks: %v", err)
		}
		if len(tasks) == 0 {
			return fmt.Errorf("task not found with prefix: %s", taskIDStr)
		}
		if len(tasks) > 1 {
			return fmt.Errorf("multiple tasks found with prefix %s, please be more specific", taskIDStr)
		}
		task = &tasks[0]
	}
	
	if err != nil {
		return fmt.Errorf("task not found: %v", err)
	}

	// Get annotations for this task
	annotations, err := repo.Annotation.GetByTaskID(task.ID)
	if err != nil {
		return fmt.Errorf("failed to get annotations: %v", err)
	}

	if len(annotations) == 0 {
		fmt.Printf("📝 No annotations found for task: %s\n", truncateString(task.Description, 50))
		return nil
	}

	fmt.Printf("📝 Annotations for task: %s\n", truncateString(task.Description, 50))
	fmt.Printf("   🆔 Task ID: %s\n\n", task.ID.String()[:8]+"...")
	
	for i, annotation := range annotations {
		fmt.Printf("%d. %s\n", i+1, annotation.Content)
		fmt.Printf("   🕒 %s\n\n", annotation.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	return nil
} 