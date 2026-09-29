package product

import (
	"errors"

	"github.com/eduardnesterov/task-manager/internal/models"
)

// ProductService — сервис для работы с продуктами
type ProductService struct {
	repo ProductRepository
}

func NewProductService(repo ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) CreateProduct(product *models.Product) error {
	if product.Name == "" {
		return errors.New("название продукта не может быть пустым")
	}
	if product.SystemID == 0 {
		return errors.New("system_id не указан")
	}
	if product.StageID == 0 {
		return errors.New("stage_id не указан")
	}
	return s.repo.Create(product)
}

func (s *ProductService) GetProduct(id uint) (*models.Product, error) {
	entity, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return entity.(*models.Product), nil
}

func (s *ProductService) UpdateProduct(product *models.Product) error {
	if product.ID == 0 {
		return errors.New("ID продукта не указан")
	}
	return s.repo.Update(product)
}

func (s *ProductService) DeleteProduct(id uint) error {
	return s.repo.Delete(id)
}

func (s *ProductService) ListProducts() ([]*models.Product, error) {
	entities, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	list := make([]*models.Product, len(entities))
	for i, e := range entities {
		list[i] = e.(*models.Product)
	}
	return list, nil
}

func (s *ProductService) ListProductsBySystem(systemID uint) ([]models.Product, error) {
	return s.repo.ListBySystem(systemID)
}
