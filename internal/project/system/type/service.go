package systemtype

import (
	"errors"

	"github.com/eduardnesterov/task-manager/internal/models"
)

// SystemTypeService — сервис для работы с типами систем
type SystemTypeService struct {
	repo SystemTypeRepository
}

func NewSystemTypeService(repo SystemTypeRepository) *SystemTypeService {
	return &SystemTypeService{repo: repo}
}

func (s *SystemTypeService) CreateSystemType(st *models.SystemType) error {
	if st.Name == "" {
		return errors.New("название типа системы не может быть пустым")
	}
	return s.repo.Create(st)
}

func (s *SystemTypeService) GetSystemType(id uint) (*models.SystemType, error) {
	entity, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return entity.(*models.SystemType), nil
}

func (s *SystemTypeService) GetSystemTypeByName(name string) (*models.SystemType, error) {
	entity, err := s.repo.GetByName(name)
	if err != nil {
		return nil, err
	}
	if entity == nil {
		return nil, nil
	}
	return entity.(*models.SystemType), nil
}

func (s *SystemTypeService) UpdateSystemType(st *models.SystemType) error {
	if st.ID == 0 {
		return errors.New("ID типа системы не указан")
	}
	return s.repo.Update(st)
}

func (s *SystemTypeService) DeleteSystemType(id uint) error {
	return s.repo.Delete(id)
}

func (s *SystemTypeService) ListSystemTypes() ([]*models.SystemType, error) {
	entities, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	list := make([]*models.SystemType, len(entities))
	for i, e := range entities {
		list[i] = e.(*models.SystemType)
	}
	return list, nil
}

func (s *SystemTypeService) SearchSystemTypes(query string) ([]*models.SystemType, error) {
	entities, err := s.repo.Search(query)
	if err != nil {
		return nil, err
	}
	list := make([]*models.SystemType, len(entities))
	for i, e := range entities {
		list[i] = e.(*models.SystemType)
	}
	return list, nil
}
