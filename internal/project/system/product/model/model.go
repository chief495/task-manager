package model

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	ID       uint   `gorm:"primaryKey" json:"id,omitempty"`
	Name     string `gorm:"not null;size:200" json:"name"`
	SystemID uint   `gorm:"not null" json:"system_id"`
	StageID  uint   `gorm:"not null" json:"stage_id"`
}
