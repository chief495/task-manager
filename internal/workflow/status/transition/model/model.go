package model

import (
	"github.com/eduardnesterov/task-manager/internal/types"
	"gorm.io/gorm"
)

type StatusTransition struct {
	gorm.Model
	ID           uint            `gorm:"primaryKey" json:"id,omitempty"`
	FromStatusID uint            `gorm:"not null" json:"from_status_id,omitempty"`
	ToStatusID   uint            `gorm:"not null" json:"to_status_id,omitempty"`
	Condition    string          `gorm:"size:200" json:"condition,omitempty"`
	AllowedRoles types.TextArray `gorm:"not null;type:text[]" json:"allowed_roles,omitempty"`
}
