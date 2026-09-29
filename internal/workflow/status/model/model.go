package model

import "gorm.io/gorm"

type Status struct {
	gorm.Model
	ID          uint   `gorm:"primaryKey" json:"id,omitempty"`
	Name        string `gorm:"not null;size:50" json:"name,omitempty"`
	Description string `gorm:"type:text" json:"description,omitempty"`
	IsDefault   bool   `gorm:"not null;default:false" json:"is_default"`
	IsFinal     bool   `gorm:"not null;default:false" json:"is_final,omitempty"`
}
