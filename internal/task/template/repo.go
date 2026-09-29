package template

import (
	"errors"

	"github.com/eduardnesterov/task-manager/internal/models"
	"gorm.io/gorm"
)

type TaskTemplateRepo struct {
	db *gorm.DB
}

func NewTaskTemplateRepo(db *gorm.DB) *TaskTemplateRepo {
	return &TaskTemplateRepo{db: db}
}

func (r *TaskTemplateRepo) Create(template *models.TaskTemplate) error {
	return r.db.Create(template).Error
}

func (r *TaskTemplateRepo) GetByID(id uint) (*models.TaskTemplate, error) {
	var template models.TaskTemplate
	err := r.db.
		First(&template, id).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *TaskTemplateRepo) List() ([]*models.TaskTemplate, error) {
	var templates []*models.TaskTemplate
	err := r.db.
		Find(&templates).Error
	if err != nil {
		return nil, err
	}
	return templates, nil
}

// Шаблоны конкретного проекта
func (r *TaskTemplateRepo) ListByProject(projectID uint) ([]*models.TaskTemplate, error) {
	var templates []*models.TaskTemplate
	err := r.db.
		Where("project_id = ?", projectID).
		Find(&templates).Error
	if err != nil {
		return nil, err
	}
	return templates, nil
}

// Общие шаблоны
func (r *TaskTemplateRepo) ListGlobal() ([]*models.TaskTemplate, error) {
	var templates []*models.TaskTemplate
	err := r.db.
		Where("project_id IS NULL").
		Find(&templates).Error
	if err != nil {
		return nil, err
	}
	return templates, nil
}

// Общие шаблоны + конкретного проекта
func (r *TaskTemplateRepo) ListForProject(projectID uint) ([]*models.TaskTemplate, error) {
	var templates []*models.TaskTemplate
	err := r.db.
		Where("project_id IS NULL OR project_id = ?", projectID).
		Find(&templates).Error
	if err != nil {
		return nil, err
	}
	return templates, nil
}

func (r *TaskTemplateRepo) Update(template *models.TaskTemplate) error {
	if template.ID == 0 {
		return errors.New("cannot update empty template")
	}
	return r.db.Model(&models.TaskTemplate{}).
		Where("id = ?", template.ID).
		Updates(map[string]any{
			"name":        template.Name,
			"description": template.Description,
			"category":    template.Category,
			"priority":    template.Priority,
			"urgency":     template.Urgency,
			"complexity":  template.Complexity,
			"project_id":  template.ProjectID,
			"updated_at":  gorm.Expr("NOW()"),
		}).Error
}

func (r *TaskTemplateRepo) Delete(id uint) error {
	return r.db.Delete(&models.TaskTemplate{}, id).Error
}
