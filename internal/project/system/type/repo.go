package systemtype

import (
	"errors"

	"github.com/eduardnesterov/task-manager/internal/models"
	"gorm.io/gorm"
)

// SystemTypeRepo — репозиторий типов систем
type SystemTypeRepo struct {
	db *gorm.DB
}

func NewSystemTypeRepo(db *gorm.DB) *SystemTypeRepo {
	return &SystemTypeRepo{db: db}
}

func (r *SystemTypeRepo) GetDB() *gorm.DB {
	return r.db
}

func (r *SystemTypeRepo) Create(entity any) error {
	st, ok := entity.(*models.SystemType)
	if !ok {
		return errors.New("entity must be of type *models.SystemType")
	}
	return r.db.Create(st).Error
}

func (r *SystemTypeRepo) GetByID(id uint) (any, error) {
	var st models.SystemType
	err := r.db.First(&st, id).Error
	return &st, err
}

func (r *SystemTypeRepo) Update(entity any) error {
	st, ok := entity.(*models.SystemType)
	if !ok {
		return errors.New("entity must be of type *models.SystemType")
	}
	return r.db.Save(st).Error
}

func (r *SystemTypeRepo) Delete(id uint) error {
	return r.db.Delete(&models.SystemType{}, id).Error
}

func (r *SystemTypeRepo) List() ([]any, error) {
	var list []models.SystemType
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

func (r *SystemTypeRepo) GetByName(name string) (any, error) {
	var st models.SystemType
	err := r.db.Where("name = ?", name).First(&st).Error
	if err != nil {
		return nil, err
	}
	return &st, nil
}

func (r *SystemTypeRepo) Search(query string) ([]any, error) {
	var list []models.SystemType
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
