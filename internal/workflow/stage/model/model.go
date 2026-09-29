package model

import "gorm.io/gorm"

type Stage struct {
	gorm.Model
	ID          uint   `gorm:"primaryKey" json:"id,omitempty"`
	Name        string `gorm:"not null;size:100" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	Position    int    `gorm:"not null" json:"position"`
}
