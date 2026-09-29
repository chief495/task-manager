package template

import (
	"errors"

	"github.com/eduardnesterov/task-manager/internal/models"
)

type TaskTemplateService struct {
	repo TaskTemplateRepository
}

func NewTaskTemplateService(repo TaskTemplateRepository) *TaskTemplateService {
	return &TaskTemplateService{repo: repo}
}

func (s *TaskTemplateService) Create(template *models.TaskTemplate) error {
	if template.Name == "" {
		return errors.New("name is required")
	}
	return s.repo.Create(template)
}

func (s *TaskTemplateService) GetByID(id uint) (*models.TaskTemplate, error) {
	return s.repo.GetByID(id)
}

func (s *TaskTemplateService) ListAll() ([]*models.TaskTemplate, error) {
	return s.repo.List()
}

func (s *TaskTemplateService) ListForProject(projectID uint) ([]*models.TaskTemplate, error) {
	if projectID == 0 {
		return s.repo.ListGlobal()
	}
	return s.repo.ListForProject(projectID)
}

func (s *TaskTemplateService) Update(template *models.TaskTemplate) error {
	if template.Name == "" {
		return errors.New("name is required")
	}
	if template.ID == 0 {
		return errors.New("id is required")
	}
	return s.repo.Update(template)
}

func (s *TaskTemplateService) Delete(id uint) error {
	return s.repo.Delete(id)
}
