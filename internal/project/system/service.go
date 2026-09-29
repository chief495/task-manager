package system

import (
	"errors"

	"github.com/eduardnesterov/task-manager/internal/models"
)

// SystemService — сервис для работы с системами
type SystemService struct {
	repo SystemRepository
}

func NewSystemService(repo SystemRepository) *SystemService {
	return &SystemService{repo: repo}
}

func (s *SystemService) CreateSystem(system *models.System) error {
	if system.Name == "" {
		return errors.New("название системы не может быть пустым")
	}
	if system.ProjectID == 0 {
		return errors.New("project_id не указан")
	}
	return s.repo.Create(system)
}

func (s *SystemService) GetSystem(id uint) (*models.System, error) {
	entity, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return entity.(*models.System), nil
}

func (s *SystemService) UpdateSystem(system *models.System) error {
	if system.ID == 0 {
		return errors.New("ID системы не указан")
	}
	return s.repo.Update(system)
}

func (s *SystemService) DeleteSystem(id uint) error {
	return s.repo.Delete(id)
}

func (s *SystemService) ListSystems() ([]*models.System, error) {
	entities, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	list := make([]*models.System, len(entities))
	for i, e := range entities {
		list[i] = e.(*models.System)
	}
	return list, nil
}

func (s *SystemService) ListSystemsByProject(projectID uint) ([]*models.System, error) {
	entities, err := s.repo.ListByProject(projectID)
	if err != nil {
		return nil, err
	}
	list := make([]*models.System, len(entities))
	for i, e := range entities {
		list[i] = e.(*models.System)
	}
	return list, nil
}
