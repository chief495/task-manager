package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID             uint    `gorm:"primaryKey" json:"id,omitempty"`
	Name           string  `gorm:"not null;size:100" json:"name,omitempty"`
	Email          string  `gorm:"unique;not null;size:100" json:"email,omitempty"`
	Password       string  `gorm:"not null" json:"-"`
	Role           string  `gorm:"not null;size:20" json:"role,omitempty"`
	IsActive       bool    `gorm:"not null" json:"is_active,omitempty"`
	SocialRating   float32 `gorm:"not null" json:"social_rating"`
	MiddlePlanTime float32 `gorm:"not null" json:"middle_plan_time"`
}

type TaskStats struct {
	UserID     uint `json:"user_id"`
	Total      int  `json:"total"`
	InProgress int  `json:"in_progress"`
	Done       int  `json:"done"`
}

type SocialRatingLog struct {
	gorm.Model
	UserID       uint      `gorm:"not null" json:"user_id"`
	Rating       float32   `gorm:"not null" json:"rating"`
	ChangeReason string    `gorm:"type:text" json:"change_reason"`
	CreatedAt    time.Time `gorm:"index" json:"date"`
}

func (u User) CheckPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) == nil
}
