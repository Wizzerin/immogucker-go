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

// RenderDashboard serves the main HTML dashboard with user's specific tasks
func (api *API) RenderDashboard(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Load existing tasks for the authorized user
	domainTasks, err := repository.GetUserTasks(api.DB, userID)
	if err != nil {
		log.Printf("[UI Error] Failed to load user tasks: %v", err)
		domainTasks = []models.Task{} // Fallback to empty list on error
	}

	// Map domain models to view models
	var tasksView []TaskView
	for _, t := range domainTasks {
		tasksView = append(tasksView, TaskView{
			ID:       t.ID,
			City:     t.City,
			MinPrice: t.MinPrice,
			MaxPrice: t.MaxPrice,
			Status:   t.Status,
		})
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"Tasks": tasksView,
	})
}

// HandleSubmitTask processes the HTMX form submission
func (api *API) HandleSubmitTask(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.String(http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.TaskRequest
	if err := c.ShouldBind(&req); err != nil {
		log.Printf("[UI Error] Form binding failed: %v", err)
		c.String(http.StatusBadRequest, "Invalid form data")
		return
	}

	taskID, err := repository.CreateTask(api.DB, userID, req)
	if err != nil {
		log.Printf("[UI] Error creating task: %v", err)
		c.String(http.StatusInternalServerError, "Database error")
		return
	}

	api.TaskChan <- taskID

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
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.String(http.StatusUnauthorized, "Unauthorized")
		return
	}

	taskID := c.Param("id")

	req, err := repository.GetTaskByID(api.DB, userID, taskID)
	if err != nil {
		c.String(http.StatusNotFound, "Task not found")
		return
	}

	status, _, _ := repository.GetTaskWithResults(api.DB, userID, taskID)

	view := TaskView{
		ID:       taskID,
		City:     req.City,
		MinPrice: req.MinPrice,
		MaxPrice: req.MaxPrice,
		Status:   status,
	}

	c.HTML(http.StatusOK, "task_row.html", view)
}
