package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

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

func CreateUser(db *sql.DB, username, email, passwordHash, verificationToken string) (int, error) {
	var userID int

	// The RETURNING id clause allows us to get the generated primary key immediately
	query := `
		INSERT INTO users (username, email, password_hash, verification_token)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	err := db.QueryRow(query, username, email, passwordHash, verificationToken).Scan(&userID)
	if err != nil {
		if strings.Contains(err.Error(), "users_username_key") {
			return 0, errors.New("username_taken")
		}
		if strings.Contains(err.Error(), "users_email_key") {
			return 0, errors.New("email_taken")
		}
		if strings.Contains(err.Error(), "users_email_key") {
			return 0, errors.New("email_taken")
		}
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	return userID, nil
}
func GetUserByEmail(db *sql.DB, email string) (models.User, error) {
	var u models.User
	query := `SELECT id, email, password_hash, role, is_email_verified FROM users WHERE email = $1`
	err := db.QueryRow(query, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.IsEmailVerified)
	return u, err
}

func GetUserByID(db *sql.DB, id int) (models.User, error) {
	var u models.User
	query := `SELECT id, username, email, password_hash, role, is_email_verified FROM users WHERE id = $1`
	err := db.QueryRow(query, id).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Role, &u.IsEmailVerified)
	return u, err
}

func VerifyEmail(db *sql.DB, token string) error {
	query := `UPDATE users SET is_email_verified = true, verification_token = NULL WHERE verification_token = $1 AND is_email_verified = FALSE`
	res, err := db.Exec(query, token)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("invalid_or_expire_token")
	}
	return nil
}
