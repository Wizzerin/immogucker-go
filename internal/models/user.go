package models

type User struct {
	ID           int    `json:"id" binding:"required"`
	PasswordHash string `json:"password_hash" binding:"required"`
	Role         string `json:"role"" binding:"required"`
	Email        string `json:"email" form:"email" binding:"required,email"`
}
