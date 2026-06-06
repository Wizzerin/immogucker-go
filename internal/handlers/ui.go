package handlers

import (
	"log"
	"net/http"

	"github.com/Wizzerin/immogucker-go/internal/models"
	"github.com/Wizzerin/immogucker-go/internal/repository"
	"github.com/gin-gonic/gin"
)

// TaskView is a struct used specifically for rendering the HTML template
type TaskView struct {
	ID       string
	City     string
	MinPrice int
	MaxPrice int
	Status   string
}

// RenderDashboard serves the main HTML dashboard
func (api *API) RenderDashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"Tasks": []TaskView{}, // Here you could load the last 10 tasks from the DB if desired
	})
}

// HandleSubmitTask processes the HTMX form submission
func (api *API) HandleSubmitTask(c *gin.Context) {
	var req models.TaskRequest

	// ShouldBind handles application/x-www-form-urlencoded from the HTML form
	if err := c.ShouldBind(&req); err != nil {
		log.Printf("[UI Error] Form binding failed: %v", err) // Теперь вы увидите, какого поля не хватает
		c.String(http.StatusBadRequest, "Invalid form data")
		return
	}

	// Create task in DB
	taskID, err := repository.CreateTask(api.DB, req)
	if err != nil {
		log.Printf("[UI] Error creating task: %v", err)
		c.String(http.StatusInternalServerError, "Database error")
		return
	}

	// Send to worker pool
	api.TaskChan <- taskID

	// Return only the HTML snippet for the new table row
	view := TaskView{
		ID:       taskID,
		City:     req.City,
		MinPrice: req.MinPrice,
		MaxPrice: req.MaxPrice,
		Status:   "pending",
	}
	c.HTML(http.StatusOK, "task_row.html", view)
}

// HandleTaskStatus returns the updated HTML snippet for HTMX polling
func (api *API) HandleTaskStatus(c *gin.Context) {
	taskID := c.Param("id")

	req, err := repository.GetTaskByID(api.DB, taskID)
	if err != nil {
		c.String(http.StatusNotFound, "Task not found")
		return
	}

	status, _, _ := repository.GetTaskWithResults(api.DB, taskID)

	view := TaskView{
		ID:       taskID,
		City:     req.City,
		MinPrice: req.MinPrice,
		MaxPrice: req.MaxPrice,
		Status:   status,
	}

	c.HTML(http.StatusOK, "task_row.html", view)
}
