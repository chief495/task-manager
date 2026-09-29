package stage

import (
	"errors"

	"github.com/eduardnesterov/task-manager/internal/models"
)

// ProductStageService — сервис для работы со стадиями продукта
type ProductStageService struct {
	repo ProductStageRepository
}

func NewProductStageService(repo ProductStageRepository) *ProductStageService {
	return &ProductStageService{repo: repo}
}

func (s *ProductStageService) CreateProductStage(ps *models.ProductStage) error {
	if ps.Name == "" {
		return errors.New("название стадии продукта не может быть пустым")
	}
	return s.repo.Create(ps)
}

func (s *ProductStageService) GetProductStage(id uint) (*models.ProductStage, error) {
	entity, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return entity.(*models.ProductStage), nil
}

func (s *ProductStageService) GetProductStageByName(name string) (*models.ProductStage, error) {
	entity, err := s.repo.GetByName(name)
	if err != nil {
		return nil, err
	}
	if entity == nil {
		return nil, nil
	}
	return entity.(*models.ProductStage), nil
}

func (s *ProductStageService) UpdateProductStage(ps *models.ProductStage) error {
	if ps.ID == 0 {
		return errors.New("ID стадии продукта не указан")
	}
	return s.repo.Update(ps)
}

func (s *ProductStageService) DeleteProductStage(id uint) error {
	return s.repo.Delete(id)
}

func (s *ProductStageService) ListProductStages() ([]*models.ProductStage, error) {
	entities, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	list := make([]*models.ProductStage, len(entities))
	for i, e := range entities {
		list[i] = e.(*models.ProductStage)
	}
	return list, nil
}

func (s *ProductStageService) SearchProductStages(query string) ([]*models.ProductStage, error) {
	entities, err := s.repo.Search(query)
	if err != nil {
		return nil, err
	}
	list := make([]*models.ProductStage, len(entities))
	for i, e := range entities {
		list[i] = e.(*models.ProductStage)
	}
	return list, nil
}
