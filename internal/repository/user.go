package repository

import (
	"database/sql"

	"github.com/Wizzerin/immogucker-go/internal/models"
)

func CreateUser(db *sql.DB, email, hashedPassword string) error {
	query := `INSERT INTO users (email, password_hash) VALUES ($1, $2)`
	_, err := db.Exec(query, email, hashedPassword)
	return err
}

func GetUserByEmail(db *sql.DB, email string) (models.User, error) {
	var u models.User
	query := `SELECT id, email, password_hash, role FROM users WHERE email = $1`
	err := db.QueryRow(query, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role)
	return u, err
}
