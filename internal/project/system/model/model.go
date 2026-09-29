package model

import "gorm.io/gorm"

type System struct {
	gorm.Model
	ID           uint   `gorm:"primaryKey" json:"id,omitempty"`
	Name         string `gorm:"not null;size:200" json:"name"`
	Description  string `gorm:"type:text" json:"description"`
	ProjectID    uint   `gorm:"not null" json:"project_id"`
	SystemTypeID uint   `gorm:"not null" json:"system_type_id"`
}
