package repository

import (
	"database/sql"
	"fmt"

	"github.com/Wizzerin/immogucker-go/internal/models"
)

// CreateTask inserts a new scraping task into the database and returns its UUID
func CreateTask(db *sql.DB, req models.TaskRequest) (string, error) {
	var taskID string

	query := `
			INSERT INTO tasks (city, max_price, min_price, email, status)
			VALUES ($1, $2, $3, $4, 'pending')
			RETURNING id
	`
	err := db.QueryRow(query, req.City, req.MaxPrice, req.MinPrice, req.Email).Scan(&taskID)
	if err != nil {
		return "", fmt.Errorf("failed to insert task into database: %w", err)
	}

	return taskID, nil
}

// UpdateTaskStatus updates the status of an existing task
func UpdateTaskStatus(db *sql.DB, taskID string, status string) error {
	query := `UPDATE tasks SET status = $1 WHERE id = $2`
	_, err := db.Exec(query, status, taskID)
	if err != nil {
		return fmt.Errorf("failed to update status for task %s: %w", taskID, err)
	}
	return nil
}

// GetTaskByID retrieves task parameters based on its UUID
func GetTaskByID(db *sql.DB, id string) (models.TaskRequest, error) {
	var req models.TaskRequest

	query := `SELECT city, max_price, min_price, email FROM tasks WHERE id = $1`
	err := db.QueryRow(query, id).Scan(&req.City, &req.MaxPrice, &req.MinPrice, &req.Email)
	if err != nil {
		return req, fmt.Errorf("failed to retrieve task %s: %w", id, err)
	}

	return req, nil
}

// GetTaskWithResults returns the task status and its associated apartments
func GetTaskWithResults(db *sql.DB, taskID string) (string, []models.Apartment, error) {
	var status string

	// 1. Retrieve the task status
	err := db.QueryRow(`SELECT status FROM tasks WHERE id = $1`, taskID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil, fmt.Errorf("task not found")
		}
		return "", nil, fmt.Errorf("failed to retrieve status: %w", err)
	}

	var apartments []models.Apartment

	// 2. If the task is completed, fetch the list of apartments
	if status == "completed" {
		rows, err := db.Query(`SELECT title, price, link FROM apartments WHERE task_id = $1`, taskID)
		if err != nil {
			return status, nil, fmt.Errorf("failed to retrieve apartments: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var apt models.Apartment
			if err := rows.Scan(&apt.Title, &apt.Price, &apt.Link); err != nil {
				return status, nil, fmt.Errorf("failed to scan row: %w", err)
			}
			apt.TaskID = taskID
			apartments = append(apartments, apt)
		}
	}

	return status, apartments, nil
}
