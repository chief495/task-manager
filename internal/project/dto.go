package project

import "github.com/eduardnesterov/task-manager/internal/models"

type ProjectCreateRequest struct {
	Name           string   `json:"name" binding:"required"`
	Description    string   `json:"description"`
	Stage          uint     `json:"stage_id"`
	Responsible    uint     `json:"responsible_id" binding:"required"`
	Address        string   `json:"address"`
	StartDate      string   `json:"start_date"`
	DeadlineDate   string   `json:"deadline_date"`
	Cost           *float32 `json:"cost"`
	Comments       string   `json:"comments"`
	Assessment     string   `json:"assessment"`
	ConclusionDate string   `json:"conclusion_date"`
	EndDate        string   `json:"end_date"`
	IsHidden       bool     `json:"is_hidden"`
	AccessUserIDs  []uint   `json:"access_user_ids"`
}

type UpdateProjectRequest struct {
	Name           *string  `json:"name"`
	Description    *string  `json:"description"`
	StageID        *uint    `json:"stage_id"`
	ResponsibleID  *uint    `json:"responsible_id"`
	Address        *string  `json:"address"`
	StartDate      *string  `json:"start_date"`
	DeadlineDate   *string  `json:"deadline_date"`
	Cost           *float32 `json:"cost"`
	Comments       *string  `json:"comments"`
	Assessment     *string  `json:"assessment"`
	ConclusionDate *string  `json:"conclusion_date"`
	EndDate        *string  `json:"end_date"`
	IsHidden       *bool    `json:"is_hidden"`
	AccessUserIDs  []uint   `json:"access_user_ids"`
}

type ProjectResponse struct {
	*models.Project
	Stage       *models.Stage          `json:"stage,omitempty"`
	Responsible *models.User           `json:"responsible,omitempty"`
	Members     []models.ProjectMember `json:"members,omitempty"`
	AccessUsers []models.ProjectAccess `json:"access_users,omitempty"`
}

type AddProjectMemberRequest struct {
	ProjectID uint   `json:"project_id" binding:"required"`
	UserID    uint   `json:"user_id" binding:"required"`
	Role      string `json:"role"`
}

type UpdateProjectMemberRequest struct {
	ProjectID *uint   `json:"project_id"`
	UserID    *uint   `json:"user_id"`
	Role      *string `json:"role"`
}

type CreateStakeholderRequest struct {
	OrganizationID uint   `json:"organization_id" binding:"required"`
	FullName       string `json:"full_name" binding:"required"`
	Position       string `json:"position"`
	Phone          string `json:"phone"`
	Email          string `json:"email"`
	Notes          string `json:"notes"`
}

type UpdateStakeholderRequest struct {
	OrganizationID uint   `json:"organization_id" binding:"required"`
	FullName       string `json:"full_name" binding:"required"`
	Position       string `json:"position"`
	Phone          string `json:"phone"`
	Email          string `json:"email"`
	Notes          string `json:"notes"`
}

type CreateProjectMessageRequest struct {
	Content   string `json:"content" binding:"required"`
	ReplyToID *uint  `json:"reply_to_id"`
}

type UpdateProjectMessageRequest struct {
	Content string `json:"content" binding:"required"`
}
