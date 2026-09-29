package stage

import (
	"errors"

	"github.com/eduardnesterov/task-manager/internal/models"
	"gorm.io/gorm"
)

// ProductStageRepo — репозиторий стадий продукта
type ProductStageRepo struct {
	db *gorm.DB
}

func NewProductStageRepo(db *gorm.DB) *ProductStageRepo {
	return &ProductStageRepo{db: db}
}

func (r *ProductStageRepo) GetDB() *gorm.DB {
	return r.db
}

func (r *ProductStageRepo) Create(entity any) error {
	ps, ok := entity.(*models.ProductStage)
	if !ok {
		return errors.New("entity must be of type *models.ProductStage")
	}
	return r.db.Create(ps).Error
}

func (r *ProductStageRepo) GetByID(id uint) (any, error) {
	var ps models.ProductStage
	err := r.db.First(&ps, id).Error
	return &ps, err
}

func (r *ProductStageRepo) Update(entity any) error {
	ps, ok := entity.(*models.ProductStage)
	if !ok {
		return errors.New("entity must be of type *models.ProductStage")
	}
	return r.db.Save(ps).Error
}

func (r *ProductStageRepo) Delete(id uint) error {
	return r.db.Delete(&models.ProductStage{}, id).Error
}

func (r *ProductStageRepo) List() ([]any, error) {
	var list []models.ProductStage
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

func (r *ProductStageRepo) GetByName(name string) (any, error) {
	var ps models.ProductStage
	err := r.db.Where("name = ?", name).First(&ps).Error
	if err != nil {
		return nil, err
	}
	return &ps, nil
}

func (r *ProductStageRepo) Search(query string) ([]any, error) {
	var list []models.ProductStage
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
