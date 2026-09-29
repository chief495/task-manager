package status

import (
	"errors"

	"github.com/eduardnesterov/task-manager/internal/models"
	"gorm.io/gorm"
)

// StatusRepo — структура репозитория статусов
type StatusRepo struct {
	db *gorm.DB
}

func (r *StatusRepo) GetDB() *gorm.DB {
	return r.db
}

// NewStatusRepo — конструктор репозитория
func NewStatusRepo(db *gorm.DB) *StatusRepo {
	return &StatusRepo{db: db}
}

// === Реализация CommonRepository ===

func (r *StatusRepo) Create(entity any) error {
	status, ok := entity.(*models.Status)
	if !ok {
		return errors.New("entity must be of type *models.Status")
	}
	return r.db.Create(status).Error
}

func (r *StatusRepo) GetByID(id uint) (any, error) {
	var status models.Status
	err := r.db.First(&status, id).Error
	return &status, err
}

func (r *StatusRepo) Update(entity any) error {
	status, ok := entity.(*models.Status)
	if !ok {
		return errors.New("entity must be of type *models.Status")
	}
	return r.db.Save(status).Error
}

func (r *StatusRepo) Delete(id uint) error {
	return r.db.Delete(&models.Status{}, id).Error
}

func (r *StatusRepo) List() ([]any, error) {
	var statuses []models.Status
	err := r.db.Find(&statuses).Error

	// Приводим к []any
	result := make([]any, len(statuses))
	for i, s := range statuses {
		result[i] = &s
	}
	return result, err
}

// === Реализация CountableRepository ===

func (r *StatusRepo) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Status{}).Count(&count).Error
	return count, err
}

// === Реализация SearchableRepository ===

func (r *StatusRepo) Search(query string) ([]any, error) {
	var statuses []models.Status
	err := r.db.Where("name ILIKE ?", "%"+query+"%").
		Find(&statuses).Error

	// Приводим к []any
	result := make([]any, len(statuses))
	for i, s := range statuses {
		result[i] = &s
	}
	return result, err
}

// === Реализация ByNameRepository ===

func (r *StatusRepo) GetByName(name string) (any, error) {
	var status models.Status
	err := r.db.Where("name = ?", name).First(&status).Error
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// === Специфичные методы StatusRepository ===

func (r *StatusRepo) IsFinal(statusID uint) (bool, error) {
	var isFinal bool
	err := r.db.Model(&models.Status{}).
		Where("id = ? AND is_final = TRUE", statusID).
		Select("TRUE").
		Take(&isFinal).Error

	if err == gorm.ErrRecordNotFound {
		return false, nil // статус найден, но не финальный
	} else if err != nil {
		return false, err // другая ошибка БД
	}

	return isFinal, nil
}

func (r *StatusRepo) GetDefaultStatus() (*models.Status, error) {
	var status models.Status

	result := r.db.Where("is_default = ?", true).First(&status)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("default status not found")
		}
		return nil, result.Error
	}
	return &status, nil
}
