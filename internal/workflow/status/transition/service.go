package transition

import (
	"errors"
	"fmt"

	"github.com/eduardnesterov/task-manager/internal/models"
	"gorm.io/gorm"
)

// StatusTransitionService — сервис для работы с переходами между статусами
type StatusTransitionService struct {
	repo StatusTransitionRepository
}

// NewStatusTransitionService — конструктор сервиса
func NewStatusTransitionService(repo StatusTransitionRepository) *StatusTransitionService {
	return &StatusTransitionService{
		repo: repo,
	}
}

// CreateStatusTransition — создание нового перехода
func (s *StatusTransitionService) CreateStatusTransition(transition *models.StatusTransition) error {
	err := s.repo.Create(transition)
	return err
}

// GetStatusTransition — получение перехода по ID
func (s *StatusTransitionService) GetStatusTransition(id uint) (*models.StatusTransition, error) {
	transition, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("transition not found")
		}
		return nil, err
	}
	return transition.(*models.StatusTransition), nil
}

// ListStatusTransitions — получение всех переходов
func (s *StatusTransitionService) ListStatusTransitions() ([]models.StatusTransition, error) {
	items, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	transitions := make([]models.StatusTransition, len(items))
	for i, item := range items {
		if transition, ok := item.(*models.StatusTransition); ok {
			transitions[i] = *transition
		} else {
			return nil, fmt.Errorf("item at index %d is not StatusTransition", i)
		}
	}
	return transitions, nil
}

// ListTransitionsByToStatus — получение переходов, ведущих в указанную стадию (to_Status_id)
func (s *StatusTransitionService) ListTransitionsByToStatus(toStatusID uint) ([]models.StatusTransition, error) {
	var transitions []models.StatusTransition
	transitions, err := s.repo.ListByToStatus(toStatusID)
	if err != nil {
		return nil, err
	}
	return transitions, nil
}

// UpdateStatusTransition — обновление перехода
func (s *StatusTransitionService) UpdateStatusTransition(transition *models.StatusTransition) error {
	// Обновляем только поля Name и Description
	err := s.repo.Update(transition)
	return err
}

// DeleteStatusTransition — удаление перехода по ID
func (s *StatusTransitionService) DeleteStatusTransition(id uint) error {
	err := s.repo.Delete(id)
	if err != nil {
		return err
	}
	return nil
}

// CreateTransition — создание нового перехода между статусами
// Возвращает ошибку, если:
// - fromID или toID равны 0
// - переход уже существует
func (s *StatusTransitionService) CreateTransition(
	fromID, toID uint,
	allowedRoles []string,
) error {
	if fromID == 0 || toID == 0 {
		return errors.New("ID статусов не могут быть нулевыми")
	}

	// Проверяем, существует ли уже такой переход
	exists, err := s.repo.ExistsTransition(fromID, toID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("переход между указанными статусами уже существует")
	}

	transition := &models.StatusTransition{
		FromStatusID: fromID,
		ToStatusID:   toID,
		AllowedRoles: allowedRoles,
	}

	return s.repo.Create(transition)
}

// GetTransition — получение перехода по парам статусов
func (s *StatusTransitionService) GetTransition(
	fromID, toID uint,
) (*models.StatusTransition, error) {
	if fromID == 0 || toID == 0 {
		return nil, errors.New("ID статусов не могут быть нулевыми")
	}

	return s.repo.GetByStatuses(fromID, toID)
}

// ListTransitionsByFromStatus — получение всех переходов, исходящих из указанного статуса
func (s *StatusTransitionService) ListTransitionsByFromStatus(
	fromID uint,
) ([]models.StatusTransition, error) {
	if fromID == 0 {
		return nil, errors.New("ID статуса не может быть нулевым")
	}

	transitions, err := s.repo.ListByFromStatus(fromID)
	if err != nil {
		return nil, err
	}

	return transitions, nil
}

// DeleteTransition — удаление перехода между статусами
func (s *StatusTransitionService) DeleteTransition(
	fromID, toID uint,
) error {
	if fromID == 0 || toID == 0 {
		return errors.New("ID статусов не могут быть нулевыми")
	}

	return s.repo.DeleteByStatuses(fromID, toID)
}

// CanTransition — проверка возможности перехода для конкретного пользователя
// Параметры:
// - task: текущая задача
// - toStatusID: целевой статус
// - userRole: роль пользователя, инициирующего переход
// Возвращает:
// - bool: можно ли выполнить переход
// - string: сообщение (причина разрешения/запрета)
// - error: ошибка БД
func (s *StatusTransitionService) CanTransition(
	task *models.Task,
	toStatusID uint,
	userRole string,
) (bool, string, error) {
	if task == nil {
		return false, "задача не указана", nil
	}
	if toStatusID == 0 {
		return false, "целевой статус не указан", nil
	}

	fromStatusID := task.StatusID

	// Если статус не меняется — разрешаем
	if fromStatusID == toStatusID {
		return true, "статус не изменяется", nil
	}

	// Проверяем существование перехода
	transition, err := s.repo.GetByStatuses(fromStatusID, toStatusID)
	if err != nil {
		return false, "переход между статусами не разрешён", err
	}

	// Проверяем права пользователя
	if len(transition.AllowedRoles) > 0 {
		for _, role := range transition.AllowedRoles {
			if role == userRole {
				return true, "переход разрешён", nil
			}
		}
		return false, "у вас нет прав для выполнения этого перехода", nil
	}

	// Если роли не ограничены — разрешаем
	return true, "переход разрешён (роли не ограничены)", nil
}

// UpdateTransition — обновление перехода (например, изменение списка разрешённых ролей)
func (s *StatusTransitionService) UpdateTransition(
	fromID, toID uint,
	allowedRoles []string,
) error {
	if fromID == 0 || toID == 0 {
		return errors.New("ID статусов не могут быть нулевыми")
	}

	transition, err := s.repo.GetByStatuses(fromID, toID)
	if err != nil {
		return err
	}

	transition.AllowedRoles = allowedRoles
	return s.repo.Update(transition)
}

// ExistsTransition — проверка существования перехода между статусами
func (s *StatusTransitionService) ExistsTransition(
	fromID, toID uint,
) (bool, error) {
	if fromID == 0 || toID == 0 {
		return false, errors.New("ID статусов не могут быть нулевыми")
	}

	return s.repo.ExistsTransition(fromID, toID)
}
