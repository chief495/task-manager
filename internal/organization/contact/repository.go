package contact

import "github.com/eduardnesterov/task-manager/internal/models"

type OrganizationContactRepository interface {
	Create(entity any) error
	GetByID(id uint) (any, error)
	Update(entity any) error
	Delete(id uint) error
	List() ([]any, error)

	ListByOrganization(organizationID uint) ([]models.OrganizationContact, error)
}
