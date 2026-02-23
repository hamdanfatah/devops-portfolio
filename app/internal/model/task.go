package model

import (
	"time"
)

// Task represents a task in the system
type Task struct {
	ID          string    `json:"id" db:"id"`
	Title       string    `json:"title" db:"title" binding:"required,min=1,max=255"`
	Description string    `json:"description" db:"description"`
	Status      string    `json:"status" db:"status"`
	Priority    string    `json:"priority" db:"priority"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// TaskCreateRequest represents a request to create a task
type TaskCreateRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=255"`
	Description string `json:"description"`
	Priority    string `json:"priority" binding:"omitempty,oneof=low medium high critical"`
}

// TaskUpdateRequest represents a request to update a task
type TaskUpdateRequest struct {
	Title       *string `json:"title" binding:"omitempty,min=1,max=255"`
	Description *string `json:"description"`
	Status      *string `json:"status" binding:"omitempty,oneof=pending in_progress completed cancelled"`
	Priority    *string `json:"priority" binding:"omitempty,oneof=low medium high critical"`
}

// TaskResponse wraps a single task response
type TaskResponse struct {
	Data Task `json:"data"`
}

// TaskListResponse wraps a list of tasks
type TaskListResponse struct {
	Data       []Task `json:"data"`
	Total      int    `json:"total"`
	Page       int    `json:"page"`
	PerPage    int    `json:"per_page"`
}

// ActivityLog represents an activity log entry stored in MongoDB
type ActivityLog struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	TaskID    string    `json:"task_id" bson:"task_id"`
	Action    string    `json:"action" bson:"action"`
	Details   string    `json:"details" bson:"details"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status   string            `json:"status"`
	Version  string            `json:"version"`
	Services map[string]string `json:"services"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}
