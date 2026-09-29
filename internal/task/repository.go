package task

import (
	"github.com/eduardnesterov/task-manager/internal/models"
	"gorm.io/gorm"
)

type TaskRepository interface {
	Create(entity any) error
	GetByID(id uint) (any, error)
	Update(entity any) error
	Delete(id uint) error
	List() ([]any, error)
	GetDB() *gorm.DB
	Count() (int64, error)
	ListByProject(projectID uint) ([]any, error)

	ListByExecutor(executorID uint) ([]models.Task, error)
	ListActiveByProject(projectID uint) ([]models.Task, error)
	CountByStatus(statusID uint) (int64, error)
	ListByCategory(category string) ([]any, error)
	ListByPhase(phaseID uint) ([]any, error)
}
