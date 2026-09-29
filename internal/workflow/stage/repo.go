package stage

import (
	"errors"

	"github.com/eduardnesterov/task-manager/internal/models"
	"gorm.io/gorm"
)

// StageRepo — структура репозитория стадий
type StageRepo struct {
	db *gorm.DB
}

func (r *StageRepo) GetDB() *gorm.DB {
	return r.db
}

// NewStageRepo — конструктор репозитория
func NewStageRepo(db *gorm.DB) *StageRepo {
	return &StageRepo{db: db}
}

// === Реализация CommonRepository ===

func (r *StageRepo) Create(entity any) error {
	stage, ok := entity.(*models.Stage)
	if !ok {
		return errors.New("entity must be of type *models.Stage")
	}
	return r.db.Create(stage).Error
}

func (r *StageRepo) GetByID(id uint) (any, error) {
	var stage models.Stage
	err := r.db.First(&stage, id).Error
	return &stage, err
}

func (r *StageRepo) Update(entity any) error {
	stage, ok := entity.(*models.Stage)
	if !ok {
		return errors.New("entity must be of type *models.Stage")
	}
	return r.db.Save(stage).Error
}

func (r *StageRepo) Delete(id uint) error {
	return r.db.Delete(&models.Stage{}, id).Error
}

func (r *StageRepo) List() ([]any, error) {
	var stages []models.Stage
	err := r.db.Find(&stages).Error

	// Приводим к []any
	result := make([]any, len(stages))
	for i, s := range stages {
		result[i] = &s
	}
	return result, err
}

// === Реализация CountableRepository ===

func (r *StageRepo) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Stage{}).Count(&count).Error
	return count, err
}

// === Реализация SearchableRepository ===

func (r *StageRepo) Search(query string) ([]any, error) {
	var stages []models.Stage
	err := r.db.Where("name ILIKE ?", "%"+query+"%").
		Find(&stages).Error

	// Приводим к []any
	result := make([]any, len(stages))
	for i, s := range stages {
		result[i] = &s
	}
	return result, err
}

// === Реализация ByNameRepository ===

func (r *StageRepo) GetByName(name string) (any, error) {
	var stage models.Stage
	err := r.db.Where("name = ?", name).First(&stage).Error
	if err != nil {
		return nil, err
	}
	return &stage, nil
}

// === Специфичные методы StageRepository (если потребуются в будущем) ===
// В текущей реализации все методы покрываются базовыми интерфейсами.
// Здесь можно добавить специфичные методы при расширении функционала.
