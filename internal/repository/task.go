package repository

import (
	"database/sql"
	"fmt"

	"github.com/Wizzerin/immogucker-go/internal/models"
)

// CreateTask inserts a new scraping task into the database and returns its UUID
func CreateTask(db *sql.DB, userID int, req models.TaskRequest) (string, error) {
	var taskID string

	query := `
			INSERT INTO tasks (user_id, city, max_price, min_price, min_size, max_size, min_rooms, max_rooms, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'pending')
			RETURNING id
	`
	err := db.QueryRow(query, userID, req.City, req.MaxPrice, req.MinPrice, req.MinSize, req.MaxSize, req.MinRooms, req.MaxRooms).Scan(&taskID)
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
func GetTaskByID(db *sql.DB, userID int, id string) (models.TaskRequest, error) {
	var req models.TaskRequest

	query := `SELECT city, max_price, min_price, min_size, max_size, min_rooms, max_rooms FROM tasks WHERE id = $1 AND user_id = $2`
	err := db.QueryRow(query, id, userID).Scan(&req.City, &req.MaxPrice, &req.MinPrice, &req.MinSize, &req.MaxSize, &req.MinRooms, &req.MaxRooms)
	if err != nil {
		return req, fmt.Errorf("failed to retrieve task %s: %w", id, err)
	}

	return req, nil
}

// GetTaskWithResults returns the task status and its associated apartments
func GetTaskWithResults(db *sql.DB, userID int, taskID string) (string, []models.Apartment, error) {
	var status string

	// 1. Retrieve the task status
	err := db.QueryRow(`SELECT status FROM tasks WHERE id = $1 AND user_id = $2`, taskID, userID).Scan(&status)
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

func GetTaskForWorker(db *sql.DB, taskID string) (models.WorkerTask, error) {
	var task models.WorkerTask

	query := `
	SELECT t.city, t.min_price, t.max_price, t.min_size, t.max_size, t.min_rooms, t.max_rooms, u.email
			FROM tasks t
			JOIN users u ON t.user_id = u.id
			WHERE t.id = $1
	`
	err := db.QueryRow(query, taskID).Scan(&task.City, &task.MinPrice, &task.MaxPrice, &task.MinSize, &task.MaxSize, &task.MinRooms, &task.MaxRooms, &task.Email)
	if err != nil {
		return task, fmt.Errorf("failed to retrieve task for worker %s: %w", taskID, err)
	}

	return task, nil
}
