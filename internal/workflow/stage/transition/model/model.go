package model

import "gorm.io/gorm"

type StageTransition struct {
	gorm.Model
	ID           uint     `gorm:"primaryKey" json:"id,omitempty"`
	FromStageID  uint     `gorm:"not null" json:"from_stage_id"`
	ToStageID    uint     `gorm:"not null" json:"to_stage_id"`
	Condition    string   `gorm:"size:200" json:"condition"`
	AllowedRoles []string `gorm:"not null;type:text[]" json:"allowed_roles,omitempty"`
}
