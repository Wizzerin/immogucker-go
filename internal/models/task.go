package models

type TaskRequest struct {
	City     string `json:"city" binding:"required"`
	MinPrice int    `json:"min_price" binding:"required,gt=0"`
	MaxPrice int    `json:"max_price" binding:"required,gt=0"`
	Email    string `json:"email" binding:"required,email"`
}
