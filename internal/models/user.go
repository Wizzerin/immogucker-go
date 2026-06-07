package models

import (
	"database/sql"
)

type User struct {
	ID                int            `json:"id" binding:"required"`
	PasswordHash      string         `json:"password_hash" binding:"required"`
	Role              string         `json:"role"" binding:"required"`
	Email             string         `json:"email" form:"email" binding:"required,email"`
	IsEmailVerified   bool           `json:"is_email_verified"`
	VerificationToken sql.NullString `json:"verification_token"`
	Username          sql.NullString `json:"username"`
}
