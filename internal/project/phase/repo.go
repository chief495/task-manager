package phase

import (
	"errors"

	"github.com/eduardnesterov/task-manager/internal/models"
	"gorm.io/gorm"
)

// ProjectPhaseRepo — репозиторий этапов проекта
type ProjectPhaseRepo struct {
	db *gorm.DB
}

func NewProjectPhaseRepo(db *gorm.DB) *ProjectPhaseRepo {
	return &ProjectPhaseRepo{db: db}
}

func (r *ProjectPhaseRepo) GetDB() *gorm.DB {
	return r.db
}

func (r *ProjectPhaseRepo) Create(entity any) error {
	phase, ok := entity.(*models.ProjectPhase)
	if !ok {
		return errors.New("entity must be of type *models.ProjectPhase")
	}
	return r.db.Create(phase).Error
}

func (r *ProjectPhaseRepo) GetByID(id uint) (any, error) {
	var phase models.ProjectPhase
	err := r.db.
		First(&phase, id).Error
	return &phase, err
}

func (r *ProjectPhaseRepo) Update(entity any) error {
	phase, ok := entity.(*models.ProjectPhase)
	if !ok {
		return errors.New("entity must be of type *models.ProjectPhase")
	}
	return r.db.Save(phase).Error
}

func (r *ProjectPhaseRepo) Delete(id uint) error {
	return r.db.Delete(&models.ProjectPhase{}, id).Error
}

func (r *ProjectPhaseRepo) List() ([]any, error) {
	var phases []models.ProjectPhase
	err := r.db.
		Order("project_id ASC, position ASC, id ASC").
		Find(&phases).Error
	if err != nil {
		return nil, err
	}
	result := make([]any, len(phases))
	for i := range phases {
		result[i] = &phases[i]
	}
	return result, nil
}

func (r *ProjectPhaseRepo) ListByProject(projectID uint) ([]any, error) {
	var phases []models.ProjectPhase
	err := r.db.Where("project_id = ?", projectID).
		Order("position ASC, id ASC").
		Find(&phases).Error
	if err != nil {
		return nil, err
	}
	result := make([]any, len(phases))
	for i := range phases {
		result[i] = &phases[i]
	}
	return result, nil
}
