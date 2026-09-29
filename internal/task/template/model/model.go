package model

import "gorm.io/gorm"

type TaskTemplate struct {
	gorm.Model
	ID          uint    `gorm:"primaryKey" json:"id"`
	Name        string  `gorm:"not null;size:200" json:"name"`
	Description string  `gorm:"type:text" json:"description"`
	Category    string  `gorm:"size:100" json:"category"`
	Priority    int     `gorm:"default:0" json:"priority"`
	Urgency     float32 `gorm:"default:1.0" json:"urgency"`
	Complexity  float32 `gorm:"default:1.0" json:"complexity"`
	ProjectID   *uint   `json:"project_id,omitempty"`
	CreatedByID uint    `gorm:"not null" json:"created_by_id"`
}
