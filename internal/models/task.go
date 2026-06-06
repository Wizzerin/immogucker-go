package models

type TaskRequest struct {
	City     string `json:"city" form:"city" binding:"required"`
	MinPrice int    `json:"min_price" form:"min_price" binding:"required,gt=0"`
	MaxPrice int    `json:"max_price" form:"max_price" binding:"required,gt=0"`
	Email    string `json:"email" form:"email" binding:"required,email"`
}
