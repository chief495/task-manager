package transition

import "github.com/eduardnesterov/task-manager/internal/models"

type StatusTransitionRepository interface {
	Create(entity any) error
	GetByID(id uint) (any, error)
	Update(entity any) error
	Delete(id uint) error
	List() ([]any, error)

	GetByStatuses(fromID, toID uint) (*models.StatusTransition, error)
	ListByFromStatus(fromID uint) ([]models.StatusTransition, error)
	ListByToStatus(toID uint) ([]models.StatusTransition, error)
	DeleteByStatuses(fromID, toID uint) error
	CanTransition(fromID, toID uint, task *models.Task) (bool, string, error)
	ExistsTransition(fromID, toID uint) (bool, error)
}
