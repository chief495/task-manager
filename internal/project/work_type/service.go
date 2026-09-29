package work_type

import (
	"errors"

	"github.com/eduardnesterov/task-manager/internal/models"
)

// WorkTypeService — сервис для работы с типами работ
type WorkTypeService struct {
	repo WorkTypeRepository
}

func NewWorkTypeService(repo WorkTypeRepository) *WorkTypeService {
	return &WorkTypeService{repo: repo}
}

func (s *WorkTypeService) CreateWorkType(wt *models.WorkType) error {
	if wt.Name == "" {
		return errors.New("название типа работ не может быть пустым")
	}
	return s.repo.Create(wt)
}

func (s *WorkTypeService) GetWorkType(id uint) (*models.WorkType, error) {
	entity, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return entity.(*models.WorkType), nil
}

func (s *WorkTypeService) GetWorkTypeByName(name string) (*models.WorkType, error) {
	entity, err := s.repo.GetByName(name)
	if err != nil {
		return nil, err
	}
	if entity == nil {
		return nil, nil
	}
	return entity.(*models.WorkType), nil
}

func (s *WorkTypeService) UpdateWorkType(wt *models.WorkType) error {
	if wt.ID == 0 {
		return errors.New("ID типа работ не указан")
	}
	return s.repo.Update(wt)
}

func (s *WorkTypeService) DeleteWorkType(id uint) error {
	return s.repo.Delete(id)
}

func (s *WorkTypeService) ListWorkTypes() ([]*models.WorkType, error) {
	entities, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	list := make([]*models.WorkType, len(entities))
	for i, e := range entities {
		list[i] = e.(*models.WorkType)
	}
	return list, nil
}

func (s *WorkTypeService) SearchWorkTypes(query string) ([]*models.WorkType, error) {
	entities, err := s.repo.Search(query)
	if err != nil {
		return nil, err
	}
	list := make([]*models.WorkType, len(entities))
	for i, e := range entities {
		list[i] = e.(*models.WorkType)
	}
	return list, nil
}
