package template

import "github.com/eduardnesterov/task-manager/internal/models"

type TaskTemplateRepository interface {
	Create(template *models.TaskTemplate) error
	GetByID(id uint) (*models.TaskTemplate, error)
	List() ([]*models.TaskTemplate, error)
	ListByProject(projectID uint) ([]*models.TaskTemplate, error)
	ListGlobal() ([]*models.TaskTemplate, error)
	ListForProject(projectID uint) ([]*models.TaskTemplate, error)
	Update(template *models.TaskTemplate) error
	Delete(id uint) error
}
