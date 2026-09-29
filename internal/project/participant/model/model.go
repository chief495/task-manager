package model

import "gorm.io/gorm"

type Participant struct {
	gorm.Model
	ID             uint   `gorm:"primaryKey" json:"id,omitempty"`
	ProjectID      uint   `gorm:"not null" json:"project_id"`
	OrganizationID uint   `gorm:"not null" json:"organization_id"`
	Role           string `gorm:"not null;size:20" json:"role"`
}
