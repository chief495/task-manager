package model

import "gorm.io/gorm"

type WorkType struct {
	gorm.Model
	ID          uint   `gorm:"primaryKey" json:"id,omitempty"`
	Name        string `gorm:"not null;size:200" json:"name"`
	Description string `gorm:"type:text" json:"description"`
}
