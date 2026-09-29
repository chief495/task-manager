package model

import (
	"time"

	"gorm.io/gorm"
)

type ProjectPhase struct {
	gorm.Model
	ID            uint       `gorm:"primaryKey" json:"id,omitempty"`
	ProjectID     uint       `gorm:"not null;index" json:"project_id"`
	Name          string     `gorm:"not null;size:200" json:"name"`
	Description   string     `gorm:"type:text" json:"description"`
	StartDate     *time.Time `json:"start_date,omitempty"`
	EndDate       *time.Time `json:"end_date,omitempty"`
	Position      int        `gorm:"not null;default:0" json:"position"`
	ResponsibleID uint       `gorm:"not null;index" json:"responsible_id"`
}
