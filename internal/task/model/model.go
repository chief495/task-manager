package model

import (
	"time"

	"gorm.io/gorm"
)

type Task struct {
	gorm.Model
	ID                    uint       `gorm:"primaryKey" json:"id,omitempty"`
	Name                  string     `gorm:"not null;size:200" json:"name,omitempty"`
	Description           string     `gorm:"type:text" json:"description,omitempty"`
	ParentID              *uint      `json:"parentID,omitempty"`
	StatusID              uint       `gorm:"not null" json:"status_id,omitempty"`
	Category              string     `gorm:"size:100;default:''" json:"category,omitempty"`
	Completed             bool       `json:"completed"`
	ProjectID             uint       `gorm:"not null" json:"project_id,omitempty" binding:"required"`
	ProjectPhaseID        *uint      `gorm:"index" json:"project_phase_id,omitempty"`
	StartTime             *time.Time `json:"start_time,omitempty"`
	EndTime               *time.Time `json:"end_time,omitempty"`
	FactEndTime           *time.Time `json:"fact_end_time"`
	FactTime              float32    `json:"fact_time"`
	PauseDate             *time.Time
	ComplexityCoefficient float32 `json:"complexity,omitempty"`
	UrgencyCoefficient    float32 `json:"urgency,omitempty"`
	Priority              int     `gorm:"default:0" json:"priority"`
	ExecutorID            uint    `gorm:"not null" json:"executor_id" binding:"required"`
	CreatedAt             time.Time
}
