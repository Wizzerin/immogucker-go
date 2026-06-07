package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/Wizzerin/immogucker-go/internal/models"
	"github.com/Wizzerin/immogucker-go/internal/repository"
	"github.com/gin-gonic/gin"
)

type API struct {
	DB       *sql.DB
	TaskChan chan string // Channel to pass task UUIDs to workers
}

func getUserIDFromContext(c *gin.Context, db *sql.DB) (int, error) {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		return 0, fmt.Errorf("cookie missing or expired: %w", err)
	}

	userID, err := repository.GetUserIDBySession(db, sessionID)
	if err != nil {
		return 0, fmt.Errorf("invalid session: %w", err)
	}

	return userID, nil
}

// CreateTask godoc
// @Summary      Create a parsing task
// @Description  Adds a new asynchronous scraping task to the queue based on city and max price.
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Param        request body models.TaskRequest true "Task parameters (e.g., city, max_price, email)"
// @Success      202 {object} map[string]interface{} "Task successfully added to queue"
// @Failure      400 {object} map[string]string "Invalid input data"
// @Router       /tasks [post]
func (api *API) CreateTask(c *gin.Context) {
	var req models.TaskRequest

	userID, err := getUserIDFromContext(c, api.DB)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taskID, err := repository.CreateTask(api.DB, userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Send task ID to the channel for background parsing (non-blocking)
	// A worker goroutine will pick up this value
	api.TaskChan <- taskID

	// Return a fast response (202 Accepted)
	c.JSON(http.StatusAccepted, gin.H{
		"message": "Task successfully added to queue",
		"task_id": taskID,
		"status":  "pending",
	})
}

// GetTaskStatus godoc
// @Summary      Get task status
// @Description  Retrieves the current status of a parsing task and the list of apartments if completed.
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Param        id path string true "Task UUID"
// @Success      200 {object} map[string]interface{} "Task status and results"
// @Failure      404 {object} map[string]string "Task not found"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /tasks/{id} [get]
func (api *API) GetTaskStatus(c *gin.Context) {
	userID, err := getUserIDFromContext(c, api.DB)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	// Extract UUID from the URL path
	taskID := c.Param("id")

	// Query the DB for status and results
	status, apartments, err := repository.GetTaskWithResults(api.DB, userID, taskID)
	if err != nil {
		// Compatible check for both English and Russian DB errors until repository is translated
		if err.Error() == "task not found" || err.Error() == "задача не найдена" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task with this ID not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Build the successful response
	response := gin.H{
		"task_id": taskID,
		"status":  status,
	}

	// Include the array of apartments in the JSON only if the task is completed
	if status == "completed" {
		response["results"] = apartments
		response["count"] = len(apartments)
	}

	c.JSON(http.StatusOK, response)
}

// HealthCheck godoc
// @Summary      Show the status of server
// @Description  Get the health status of the API
// @Tags         System
// @Produce      json
// @Success      200 {object} map[string]string "Server is up and running"
// @Router       /health [get]
func (api *API) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "up",
	})
}
