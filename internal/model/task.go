package model

import "time"

type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
}

type UpdateTaskRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
	Priority    *string `json:"priority,omitempty"`
}

type TaskEvent struct {
	Type      string    `json:"type"`
	TaskID    string    `json:"task_id"`
	Timestamp time.Time `json:"timestamp"`
	Data      any       `json:"data,omitempty"`
}

// Valid statuses and priorities
var (
	ValidStatuses   = []string{"todo", "in_progress", "done"}
	ValidPriorities = []string{"low", "medium", "high"}
)

func (r *CreateTaskRequest) Validate() string {
	if r.Title == "" {
		return "title is required"
	}
	if r.Priority == "" {
		r.Priority = "medium"
	}
	if !contains(ValidPriorities, r.Priority) {
		return "priority must be low, medium, or high"
	}
	return ""
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
