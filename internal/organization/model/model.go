package model

import "gorm.io/gorm"

type Organization struct {
	gorm.Model
	ID          uint   `gorm:"primaryKey" json:"id,omitempty"`
	Name        string `gorm:"not null;size:200" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	Phone       string `gorm:"size:50" json:"phone"`
	Email       string `gorm:"size:100" json:"email"`
	Address     string `gorm:"size:500" json:"address"`
	Website     string `gorm:"size:200" json:"website"`
	Notes       string `gorm:"type:text" json:"notes"`
}
