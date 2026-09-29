package task

import (
	"encoding/json"
	"time"
)

type CreateTaskRequest struct {
	Name           string     `json:"name" binding:"required,max=200"`
	Description    string     `json:"description"`
	Category       string     `json:"category"`
	ParentID       *uint      `json:"parent_id"`
	StatusID       uint       `json:"status_id" binding:"required"`
	ExecutorID     uint       `json:"executor_id" binding:"required"`
	ProjectID      uint       `json:"project_id" binding:"required"`
	ProjectPhaseID *uint      `json:"project_phase_id"`
	StartTime      *time.Time `json:"start_time"`
	EndTime        *time.Time `json:"end_time"`
	Priority       int        `json:"priority" binding:"omitempty,min=0,max=2"`
	Urgency        float32    `json:"urgency" binding:"omitempty,min=0"`
	Complexity     float32    `json:"complexity" binding:"omitempty,min=0"`
}

type UpdateTaskRequest struct {
	Name           *string          `json:"name"`
	Description    *string          `json:"description"`
	Category       *string          `json:"category"`
	ParentID       *uint            `json:"parent_id"`
	StatusID       *uint            `json:"status_id"`
	ExecutorID     *uint            `json:"executor_id"`
	ProjectID      *uint            `json:"project_id"`
	ProjectPhaseID *json.RawMessage `json:"project_phase_id"`
	StartTime      *time.Time       `json:"start_time"`
	EndTime        *time.Time       `json:"end_time"`
	Priority       *int             `json:"priority" binding:"omitempty,min=0,max=2"`
	Urgency        *float32         `json:"urgency" binding:"omitempty,min=0"`
	Complexity     *float32         `json:"complexity" binding:"omitempty,min=0"`
}

type UpdateTaskStatusRequest struct {
	StatusID uint `json:"status_id"`
}
