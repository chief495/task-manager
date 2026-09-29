package status

import (
	"errors"

	"github.com/eduardnesterov/task-manager/internal/models"
)

type StatusService struct {
	repo StatusRepository
}

func NewStatusService(repo StatusRepository) *StatusService {
	return &StatusService{repo: repo}
}

// CreateStatus — создание статуса
func (s *StatusService) CreateStatus(status *models.Status) error {
	if status.Name == "" {
		return errors.New("название статуса не может быть пустым")
	}
	return s.repo.Create(status)
}

// GetStatus — получение статуса по ID
func (s *StatusService) GetStatus(id uint) (*models.Status, error) {
	entity, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return entity.(*models.Status), nil
}

// GetStatus — получение статуса по ID
func (s *StatusService) GetDefaultStatus() (*models.Status, error) {
	entity, err := s.repo.GetDefaultStatus()
	if err != nil {
		return nil, err
	}
	return entity, nil
}

// UpdateStatus — обновление статуса
func (s *StatusService) UpdateStatus(status *models.Status) error {
	if status.ID == 0 {
		return errors.New("ID статуса не указан")
	}
	return s.repo.Update(status)
}

// DeleteStatus — удаление статуса
func (s *StatusService) DeleteStatus(id uint) error {
	return s.repo.Delete(id)
}

// ListStatuses — список всех статусов
func (s *StatusService) ListStatuses() ([]*models.Status, error) {
	entities, err := s.repo.List()
	if err != nil {
		return nil, err
	}

	statuses := make([]*models.Status, len(entities))
	for i, e := range entities {
		statuses[i] = e.(*models.Status)
	}
	return statuses, nil
}

// IsFinalStatus — проверка, является ли статус финальным
func (s *StatusService) IsFinalStatus(id uint) (bool, error) {
	return s.repo.IsFinal(id)
}
