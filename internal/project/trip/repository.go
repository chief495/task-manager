package trip

import (
	"time"

	"github.com/eduardnesterov/task-manager/internal/models"
)

type TripRepository interface {
	Create(entity any) error
	GetByID(id uint) (any, error)
	Update(entity any) error
	Delete(id uint) error
	List() ([]any, error)
	Count() (int64, error)
	ListByProject(projectID uint) ([]any, error)

	ListByPerson(personID uint) ([]models.Trip, error)
	ListActiveInPeriod(start, end time.Time) ([]models.Trip, error)
}
