package repository

import (
	"database/sql"
	"fmt"
	"time"
)

func CreateSession(db *sql.DB, sessionID string, userID int) error {
	expiresAt := time.Now().Add(24 * time.Hour)
	query := `INSERT INTO sessions (id, user_id, expires_at) VALUES($1, $2, $3)`
	_, err := db.Exec(query, sessionID, userID, expiresAt)
	return err
}

func GetUserIDBySession(db *sql.DB, sessionID string) (int, error) {
	var userID int
	var expiresAt time.Time

	query := `SELECT user_id, expires_at FROM sessions WHERE id = $1 AND expires_at > NOW()`
	err := db.QueryRow(query, sessionID).Scan(&userID, &expiresAt)

	if err != nil {
		return 0, fmt.Errorf("session not found in db: %w", err)
	}

	if time.Now().After(expiresAt) {
		return 0, fmt.Errorf("session expired")
	}

	return userID, err
}

func DeleteSession(db *sql.DB, sessionID string) error {
	_, err := db.Exec(`DELETE FROM sessions WHERE id = $1`, sessionID)
	return err
}
