package repository

import (
	"database/sql"
	"fmt"

	"github.com/Wizzerin/immogucker-go/internal/models"
)

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
