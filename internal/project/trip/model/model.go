package model

import (
	"time"

	"gorm.io/gorm"
)

type Trip struct {
	gorm.Model
	Place     string    `gorm:"not null;size:200"`
	ProjectID uint      `gorm:"not null"`
	PersonID  uint      `gorm:"not null"`
	BeginDate time.Time `gorm:"not null"`
	EndDate   time.Time `gorm:"not null"`
}
