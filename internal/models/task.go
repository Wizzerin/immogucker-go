package models

type TaskRequest struct {
	City     string `json:"city" form:"city" binding:"required"`
	MinPrice int    `json:"min_price" form:"min_price" binding:"required,gt=0"`
	MaxPrice int    `json:"max_price" form:"max_price" binding:"required,gt=0"`
	MinSize  int    `json:"min_size" form:"min_size"`
	MaxSize  int    `json:"max_size" form:"max_size"`
	MinRooms int    `json:"min_rooms" form:"min_rooms"`
	MaxRooms int    `json:"max_rooms" fomr:"max_rooms"`
}

type Task struct {
	ID       string `json:"id"`
	UserID   int    `json:"user_id"`
	City     string `json:"city"`
	MinPrice int    `json:"min_price"`
	MaxPrice int    `json:"max_price"`
	MinSize  int    `json:"min_size"`
	MaxSize  int    `json:"max_size"`
	MinRooms int    `json:"min_rooms"`
	MaxRooms int    `json:"max_rooms"`
	Status   string `json:"status"`
}

type WorkerTask struct {
	City     string
	MinPrice int
	MaxPrice int
	MinSize  int
	MaxSize  int
	MinRooms int
	MaxRooms int
	Email    string // Fetched via JOIN from the users table
}
