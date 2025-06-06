package main

import (
	"github.com/gin-gonic/gin"
	"github.com/terzigolu/josepshbrain-go/api/database"
	"github.com/terzigolu/josepshbrain-go/api/handlers"
)

func main() {
	// Initialize database connection
	database.Connect()

	r := gin.Default()

	// Ping endpoint for health check
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// API v1 routes
	v1 := r.Group("/v1")
	{
		v1.GET("/tasks", handlers.ListTasks)
	}

	r.Run() // listen and serve on 0.0.0.0:8080
} 