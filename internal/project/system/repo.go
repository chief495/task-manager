package system

import (
	"errors"

	"github.com/eduardnesterov/task-manager/internal/models"
	"gorm.io/gorm"
)

// SystemRepo — репозиторий для работы с системами
type SystemRepo struct {
	db *gorm.DB
}

func NewSystemRepo(db *gorm.DB) *SystemRepo {
	return &SystemRepo{db: db}
}

func (r *SystemRepo) GetDB() *gorm.DB {
	return r.db
}

func (r *SystemRepo) Create(entity any) error {
	system, ok := entity.(*models.System)
	if !ok {
		return errors.New("entity must be of type *models.System")
	}
	return r.db.Create(system).Error
}

func (r *SystemRepo) GetByID(id uint) (any, error) {
	var system models.System
	err := r.db.
		First(&system, id).Error
	return &system, err
}

func (r *SystemRepo) Update(entity any) error {
	system, ok := entity.(*models.System)
	if !ok {
		return errors.New("entity must be of type *models.System")
	}
	return r.db.Save(system).Error
}

func (r *SystemRepo) Delete(id uint) error {
	return r.db.Delete(&models.System{}, id).Error
}

func (r *SystemRepo) List() ([]any, error) {
	var systems []models.System
	err := r.db.
		Find(&systems).Error
	if err != nil {
		return nil, err
	}
	result := make([]any, len(systems))
	for i := range systems {
		result[i] = &systems[i]
	}
	return result, nil
}

func (r *SystemRepo) ListByProject(projectID uint) ([]any, error) {
	var systems []models.System
	err := r.db.Where("project_id = ?", projectID).
		Find(&systems).Error
	if err != nil {
		return nil, err
	}
	result := make([]any, len(systems))
	for i := range systems {
		result[i] = &systems[i]
	}
	return result, nil
}
