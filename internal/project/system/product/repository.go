package product

import "github.com/eduardnesterov/task-manager/internal/models"

type ProductRepository interface {
	Create(entity any) error
	GetByID(id uint) (any, error)
	Update(entity any) error
	Delete(id uint) error
	List() ([]any, error)

	ListBySystem(systemID uint) ([]models.Product, error)
}
