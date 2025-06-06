package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/terzigolu/josepshbrain-go/api/database"
	"github.com/terzigolu/josepshbrain-go/api/models"
)

// ListTasks retrieves all tasks from the database.
func ListTasks(c *gin.Context) {
	var tasks []models.Task
	if err := database.DB.Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tasks"})
		return
	}

	c.JSON(http.StatusOK, tasks)
} 