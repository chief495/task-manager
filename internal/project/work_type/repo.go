package work_type

import (
	"errors"

	"github.com/eduardnesterov/task-manager/internal/models"
	"gorm.io/gorm"
)

// WorkTypeRepo — репозиторий типов работ
type WorkTypeRepo struct {
	db *gorm.DB
}

func NewWorkTypeRepo(db *gorm.DB) *WorkTypeRepo {
	return &WorkTypeRepo{db: db}
}

func (r *WorkTypeRepo) GetDB() *gorm.DB {
	return r.db
}

func (r *WorkTypeRepo) Create(entity any) error {
	wt, ok := entity.(*models.WorkType)
	if !ok {
		return errors.New("entity must be of type *models.WorkType")
	}
	return r.db.Create(wt).Error
}

func (r *WorkTypeRepo) GetByID(id uint) (any, error) {
	var wt models.WorkType
	err := r.db.First(&wt, id).Error
	return &wt, err
}

func (r *WorkTypeRepo) Update(entity any) error {
	wt, ok := entity.(*models.WorkType)
	if !ok {
		return errors.New("entity must be of type *models.WorkType")
	}
	return r.db.Save(wt).Error
}

func (r *WorkTypeRepo) Delete(id uint) error {
	return r.db.Delete(&models.WorkType{}, id).Error
}

func (r *WorkTypeRepo) List() ([]any, error) {
	var list []models.WorkType
	err := r.db.Find(&list).Error
	if err != nil {
		return nil, err
	}
	result := make([]any, len(list))
	for i := range list {
		result[i] = &list[i]
	}
	return result, nil
}

func (r *WorkTypeRepo) GetByName(name string) (any, error) {
	var wt models.WorkType
	err := r.db.Where("name = ?", name).First(&wt).Error
	if err != nil {
		return nil, err
	}
	return &wt, nil
}

func (r *WorkTypeRepo) Search(query string) ([]any, error) {
	var list []models.WorkType
	err := r.db.Where("name ILIKE ?", "%"+query+"%").
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	result := make([]any, len(list))
	for i := range list {
		result[i] = &list[i]
	}
	return result, nil
}
