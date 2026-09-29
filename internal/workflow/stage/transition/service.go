package transition

import (
	"errors"
	"fmt"
	"slices"

	"github.com/eduardnesterov/task-manager/internal/models"
	"gorm.io/gorm"
)

// StageTransitionService — сервис для работы с переходами между стадиями
type StageTransitionService struct {
	repo StageTransitionRepository
}

// NewStageTransitionService — конструктор сервиса
func NewStageTransitionService(repo StageTransitionRepository) *StageTransitionService {
	return &StageTransitionService{
		repo: repo,
	}
}

// CreateStageTransition — создание нового перехода
func (s *StageTransitionService) CreateStageTransition(transition *models.StageTransition) error {
	err := s.repo.Create(transition)
	return err
}

// GetStageTransition — получение перехода по ID
func (s *StageTransitionService) GetStageTransition(id uint) (*models.StageTransition, error) {
	transition, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("transition not found")
		}
		return nil, err
	}
	return transition.(*models.StageTransition), nil
}

// ListStageTransitions — получение всех переходов
func (s *StageTransitionService) ListStageTransitions() ([]models.StageTransition, error) {
	items, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	transitions := make([]models.StageTransition, len(items))
	for i, item := range items {
		if transition, ok := item.(models.StageTransition); ok {
			transitions[i] = transition
		} else {
			return nil, fmt.Errorf("item at index %d is not StageTransition", i)
		}
	}
	return transitions, nil
}

// ListTransitionsByToStage — получение переходов, ведущих в указанную стадию (to_stage_id)
func (s *StageTransitionService) ListTransitionsByToStage(toStageID uint) ([]models.StageTransition, error) {
	var transitions []models.StageTransition
	transitions, err := s.repo.ListByToStage(toStageID)
	if err != nil {
		return nil, err
	}
	return transitions, nil
}

// UpdateStageTransition — обновление перехода
func (s *StageTransitionService) UpdateStageTransition(transition *models.StageTransition) error {
	// Обновляем только поля Name и Description
	err := s.repo.Update(transition)
	return err
}

// DeleteStageTransition — удаление перехода по ID
func (s *StageTransitionService) DeleteStageTransition(id uint) error {
	err := s.repo.Delete(id)
	if err != nil {
		return err
	}
	return nil
}

// CreateTransition — создание нового перехода между стадиями
// Возвращает ошибку, если:
// - fromID или toID равны 0
// - переход уже существует
func (s *StageTransitionService) CreateTransition(
	fromID, toID uint,
	allowedRoles []string,
	condition string,
) error {
	if fromID == 0 || toID == 0 {
		return errors.New("ID стадий не могут быть нулевыми")
	}

	// Проверяем, существует ли уже такой переход
	exists, err := s.repo.ExistsTransition(fromID, toID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("переход между указанными стадиями уже существует")
	}

	transition := &models.StageTransition{
		FromStageID:  fromID,
		ToStageID:    toID,
		AllowedRoles: allowedRoles,
		Condition:    condition,
	}

	return s.repo.Create(transition)
}

// GetTransition — получение перехода по парам стадий
func (s *StageTransitionService) GetTransition(
	fromID, toID uint,
) (*models.StageTransition, error) {
	if fromID == 0 || toID == 0 {
		return nil, errors.New("ID стадий не могут быть нулевыми")
	}

	return s.repo.GetByStages(fromID, toID)
}

// ListTransitionsByFromStage — получение всех переходов, исходящих из указанной стадии
func (s *StageTransitionService) ListTransitionsByFromStage(
	fromID uint,
) ([]models.StageTransition, error) {
	if fromID == 0 {
		return nil, errors.New("ID стадии не может быть нулевым")
	}

	transitions, err := s.repo.ListByFromStage(fromID)
	if err != nil {
		return nil, err
	}

	return transitions, nil
}

// DeleteTransition — удаление перехода между стадиями
func (s *StageTransitionService) DeleteTransition(
	fromID, toID uint,
) error {
	if fromID == 0 || toID == 0 {
		return errors.New("ID стадий не могут быть нулевыми")
	}

	return s.repo.DeleteByStages(fromID, toID)
}

// CanTransition — проверка возможности перехода для конкретного пользователя и проекта
// Параметры:
// - project: текущий проект
// - toStageID: целевая стадия
// - userRole: роль пользователя, инициирующего переход
// Возвращает:
// - bool: можно ли выполнить переход
// - string: сообщение (причина разрешения/запрета)
// - error: ошибка БД
func (s *StageTransitionService) CanTransition(
	project *models.Project,
	toStageID uint,
	userRole string,
) (bool, string, error) {
	if project == nil {
		return false, "проект не указан", nil
	}
	if toStageID == 0 {
		return false, "целевая стадия не указана", nil
	}

	fromStageID := project.StageID

	// Если стадия не меняется — разрешаем
	if fromStageID == toStageID {
		return true, "стадия не изменяется", nil
	}

	// Проверяем существование перехода
	transition, err := s.repo.GetByStages(fromStageID, toStageID)
	if err != nil {
		return false, "переход между стадиями не разрешён", err
	}

	// Проверяем права пользователя
	if len(transition.AllowedRoles) > 0 {
		if slices.Contains(transition.AllowedRoles, userRole) {
			return true, "переход разрешён", nil
		}
		return false, "у вас нет прав для выполнения этого перехода", nil
	}

	// Проверяем условие перехода (если задано)
	if transition.Condition != "" {
		// Здесь можно реализовать логику проверки условия
		// Например: условие может содержать выражение на DSL,
		// или ключ к бизнес‑правилу, которое нужно интерпретировать.
		// Для примера: если условие — это имя поля, проверяем его заполненность

	}

	// Если роли и условие не заданы — разрешаем
	return true, "переход разрешён (правила не ограничены)", nil
}

// UpdateTransition — обновление перехода (изменение ролей и условия)
func (s *StageTransitionService) UpdateTransition(
	fromID, toID uint,
	allowedRoles []string,
	condition string,
) error {
	if fromID == 0 || toID == 0 {
		return errors.New("ID стадий не могут быть нулевыми")
	}

	transition, err := s.repo.GetByStages(fromID, toID)
	if err != nil {
		return err
	}

	transition.AllowedRoles = allowedRoles
	transition.Condition = condition
	return s.repo.Update(transition)
}

// ExistsTransition — проверка существования перехода между стадиями
func (s *StageTransitionService) ExistsTransition(
	fromID, toID uint,
) (bool, error) {
	if fromID == 0 || toID == 0 {
		return false, errors.New("ID стадий не могут быть нулевыми")
	}

	return s.repo.ExistsTransition(fromID, toID)
}
