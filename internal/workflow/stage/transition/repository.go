package transition

import "github.com/eduardnesterov/task-manager/internal/models"

type StageTransitionRepository interface {
	Create(entity any) error
	GetByID(id uint) (any, error)
	Update(entity any) error
	Delete(id uint) error
	List() ([]any, error)

	GetByStages(fromID, toID uint) (*models.StageTransition, error)
	ListByFromStage(fromID uint) ([]models.StageTransition, error)
	ListByToStage(toID uint) ([]models.StageTransition, error)
	DeleteByStages(fromID, toID uint) error
	ExistsTransition(fromID, toID uint) (bool, error)
}
