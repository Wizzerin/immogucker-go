package repository

import (
	"database/sql"
	"fmt"

	"github.com/Wizzerin/immogucker-go/internal/models"
)

// GetUserTasks retrieves all tasks belonging to a specific user for the UI Dashboard
func GetUserTasks(db *sql.DB, userID int) ([]models.Task, error) {
	var tasks []models.Task

	query := `SELECT id, user_id, city, max_price, min_price, status FROM tasks WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user tasks: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var t models.Task
		if err := rows.Scan(&t.ID, &t.UserID, &t.City, &t.MaxPrice, &t.MinPrice, &t.Status); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func CreateUser(db *sql.DB, email, passwordHash string) (int, error) {
	var userID int

	// The RETURNING id clause allows us to get the generated primary key immediately
	query := `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id
	`

	err := db.QueryRow(query, email, passwordHash).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	return userID, nil
}
func GetUserByEmail(db *sql.DB, email string) (models.User, error) {
	var u models.User
	query := `SELECT id, email, password_hash, role FROM users WHERE email = $1`
	err := db.QueryRow(query, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role)
	return u, err
}
