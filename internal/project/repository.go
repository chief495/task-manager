package project

import (
	"github.com/eduardnesterov/task-manager/internal/models"
	"gorm.io/gorm"
)

type ProjectRepository interface {
	Create(entity any) error
	GetByID(id uint) (any, error)
	Update(entity any) error
	Delete(id uint) error
	List() ([]any, error)
	GetDB() *gorm.DB
	Search(query string) ([]any, error)
	Count() (int64, error)
	ListByProject(projectID uint) ([]any, error)

	ListByStage(stageID uint) ([]models.Project, error)
	ListByResponsible(responsibleID uint) ([]models.Project, error)
	ListVisibleForUser(userID uint) ([]models.Project, error)
	GetActiveProjects() ([]models.Project, error)
	GetProjectsByStage(stageID uint) ([]models.Project, error)

	CreateMember(entity any) error
	GetMemberByID(id uint) (*models.ProjectMember, error)
	UpdateMember(member *models.ProjectMember) error
	DeleteMember(id uint) error
	ListMembersByProject(projectID uint) ([]models.ProjectMember, error)
}
