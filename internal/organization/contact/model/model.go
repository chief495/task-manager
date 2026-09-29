package model

import "gorm.io/gorm"

type OrganizationContact struct {
	gorm.Model
	ID             uint   `gorm:"primaryKey" json:"id,omitempty"`
	OrganizationID uint   `gorm:"not null;index" json:"organization_id"`
	FullName       string `gorm:"not null;size:200" json:"full_name"`
	Position       string `gorm:"size:200" json:"position"`
	Phone          string `gorm:"size:50" json:"phone"`
	Email          string `gorm:"size:100" json:"email"`
	Notes          string `gorm:"type:text" json:"notes"`
}
