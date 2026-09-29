package product

import (
	"errors"

	"github.com/eduardnesterov/task-manager/internal/models"
	"gorm.io/gorm"
)

// ProductRepo — репозиторий продуктов
type ProductRepo struct {
	db *gorm.DB
}

func NewProductRepo(db *gorm.DB) *ProductRepo {
	return &ProductRepo{db: db}
}

func (r *ProductRepo) GetDB() *gorm.DB {
	return r.db
}

func (r *ProductRepo) Create(entity any) error {
	product, ok := entity.(*models.Product)
	if !ok {
		return errors.New("entity must be of type *models.Product")
	}
	return r.db.Create(product).Error
}

func (r *ProductRepo) GetByID(id uint) (any, error) {
	var product models.Product
	err := r.db.First(&product, id).Error
	return &product, err
}

func (r *ProductRepo) Update(entity any) error {
	product, ok := entity.(*models.Product)
	if !ok {
		return errors.New("entity must be of type *models.Product")
	}
	return r.db.Save(product).Error
}

func (r *ProductRepo) Delete(id uint) error {
	return r.db.Delete(&models.Product{}, id).Error
}

func (r *ProductRepo) List() ([]any, error) {
	var products []models.Product
	err := r.db.Find(&products).Error
	if err != nil {
		return nil, err
	}
	result := make([]any, len(products))
	for i := range products {
		result[i] = &products[i]
	}
	return result, nil
}

func (r *ProductRepo) ListBySystem(systemID uint) ([]models.Product, error) {
	var products []models.Product
	err := r.db.Where("system_id = ?", systemID).
		Find(&products).Error
	return products, err
}
