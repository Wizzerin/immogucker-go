package repository

import (
	"database/sql"
	"fmt"

	"github.com/Wizzerin/immogucker-go/internal/models"
)

// SaveApartment inserts a batch of scraped apartments into the database
func SaveApartment(db *sql.DB, apartments []models.Apartment) error {
	if len(apartments) == 0 {
		return nil
	}

	// Begin a transaction for bulk insertion
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	query := `INSERT INTO apartments (task_id, title, price, link) VALUES ($1, $2, $3, $4)`

	for _, apt := range apartments {
		_, err := tx.Exec(query, apt.TaskID, apt.Title, apt.Price, apt.Link)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert apartment: %w", err)
		}
	}

	return tx.Commit()
}
