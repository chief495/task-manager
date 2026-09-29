package model

import (
	"time"

	"gorm.io/gorm"
)

type Report struct {
	gorm.Model
	ID        uint      `gorm:"primaryKey" json:"id,omitempty"`
	Text      string    `gorm:"type:text" json:"text,omitempty"`
	Overwork  int       `json:"overwork,omitempty"`
	ProjectID uint      `gorm:"not null" json:"project_id,omitempty"`
	UserID    uint      `gorm:"not null" json:"user_id,omitempty"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at,omitempty"`
}
