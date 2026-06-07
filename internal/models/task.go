package models

type TaskRequest struct {
	City     string `json:"city" form:"city" binding:"required"`
	MinPrice int    `json:"min_price" form:"min_price" binding:"required,gt=0"`
	MaxPrice int    `json:"max_price" form:"max_price" binding:"required,gt=0"`
}

type Task struct {
	ID       string `json:"id"`
	UserID   int    `json:"user_id"`
	City     string `json:"city"`
	MinPrice int    `json:"min_price"`
	MaxPrice int    `json:"max_price"`
	Status   string `json:"status"`
}

type WorkerTask struct {
	City     string
	MinPrice int
	MaxPrice int
	Email    string // Fetched via JOIN from the users table
}
