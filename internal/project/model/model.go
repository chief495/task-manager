package model

import (
	"time"

	usermodel "github.com/eduardnesterov/task-manager/internal/user/model"
	"gorm.io/gorm"
)

type Project struct {
	gorm.Model
	ID             uint      `gorm:"primaryKey" json:"id,omitempty"`
	Name           string    `gorm:"not null;size:200" json:"name"`
	Description    string    `gorm:"type:text" json:"description"`
	StartDate      time.Time `gorm:"not null" json:"start_date"`
	DeadlineDate   time.Time `gorm:"not null" json:"deadline_date"`
	ResponsibleID  uint      `gorm:"not null" json:"responsible_id"`
	StageID        uint      `gorm:"not null" json:"stage_id"`
	Assessment     string    `gorm:"type:text" json:"assessment"`
	Cost           float32   `gorm:"type:float" json:"cost"`
	ConclusionDate time.Time `gorm:"index" json:"date"`
	EndDate        time.Time `gorm:"index" json:"end_date"`
	Comments       string    `gorm:"type:text" json:"comments"`
	Address        string    `gorm:"size:500" json:"address"`
	Latitude       float64   `gorm:"type:double precision" json:"latitude"`
	Longitude      float64   `gorm:"type:double precision" json:"longitude"`
	IsHidden       bool      `gorm:"not null;default:false;index" json:"is_hidden"`
}

type ProjectMember struct {
	gorm.Model
	ID        uint   `gorm:"primaryKey" json:"id,omitempty"`
	ProjectID uint   `gorm:"not null;uniqueIndex:idx_project_member_user" json:"project_id"`
	UserID    uint   `gorm:"not null;uniqueIndex:idx_project_member_user" json:"user_id"`
	Role      string `gorm:"not null;size:100" json:"role"`
}

type ProjectAccess struct {
	gorm.Model
	ID        uint `gorm:"primaryKey" json:"id,omitempty"`
	ProjectID uint `gorm:"not null;uniqueIndex:idx_project_access_user" json:"project_id"`
	UserID    uint `gorm:"not null;uniqueIndex:idx_project_access_user" json:"user_id"`
}

type Stakeholder struct {
	gorm.Model
	ID             uint   `gorm:"primaryKey" json:"id,omitempty"`
	ProjectID      uint   `gorm:"not null;index" json:"project_id"`
	OrganizationID uint   `gorm:"not null;index" json:"organization_id"`
	FullName       string `gorm:"not null;size:200" json:"full_name"`
	Position       string `gorm:"size:200" json:"position"`
	Phone          string `gorm:"size:50" json:"phone"`
	Email          string `gorm:"size:100" json:"email"`
	Notes          string `gorm:"type:text" json:"notes"`
}

type ProjectMessage struct {
	gorm.Model
	ID        uint            `gorm:"primaryKey" json:"id,omitempty"`
	ProjectID uint            `gorm:"not null;index" json:"project_id"`
	UserID    uint            `gorm:"not null;index" json:"user_id"`
	ReplyToID *uint           `gorm:"index" json:"reply_to_id,omitempty"`
	Content   string          `gorm:"not null;type:text" json:"content"`
	User      *usermodel.User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ReplyTo   *ProjectMessage `gorm:"foreignKey:ReplyToID" json:"reply_to,omitempty"`
}
