package transition

import (
	"errors"

	"github.com/eduardnesterov/task-manager/internal/models"
	"gorm.io/gorm"
)

// StatusTransitionRepo — структура репозитория переходов между статусами
type StatusTransitionRepo struct {
	db *gorm.DB
}

func (r *StatusTransitionRepo) GetDB() *gorm.DB {
	return r.db
}

// NewStatusTransitionRepo — конструктор репозитория
func NewStatusTransitionRepo(db *gorm.DB) *StatusTransitionRepo {
	return &StatusTransitionRepo{db: db}
}

// === Реализация CommonRepository ===

func (r *StatusTransitionRepo) Create(entity any) error {
	transition, ok := entity.(*models.StatusTransition)
	if !ok {
		return errors.New("entity must be of type *models.StatusTransition")
	}
	return r.db.Create(transition).Error
}

func (r *StatusTransitionRepo) GetByID(id uint) (any, error) {
	var transition models.StatusTransition
	err := r.db.
		First(&transition, id).Error
	return &transition, err
}

func (r *StatusTransitionRepo) Update(entity any) error {
	transition, ok := entity.(*models.StatusTransition)
	if !ok {
		return errors.New("entity must be of type *models.StatusTransition")
	}
	return r.db.Save(transition).Error
}

func (r *StatusTransitionRepo) Delete(id uint) error {
	return r.db.Delete(&models.StatusTransition{}, id).Error
}

func (r *StatusTransitionRepo) List() ([]any, error) {
	var transitions []models.StatusTransition
	err := r.db.
		Find(&transitions).Error

	// Приводим к []any
	result := make([]any, len(transitions))
	for i, t := range transitions {
		result[i] = &t
	}
	return result, err
}

// === Специфичные методы StatusTransitionRepository ===

// GetByStatuses — получение перехода по парам статусов (fromID → toID)
func (r *StatusTransitionRepo) GetByStatuses(fromID, toID uint) (*models.StatusTransition, error) {
	var transition models.StatusTransition
	err := r.db.Where("from_status_id = ? AND to_status_id = ?", fromID, toID).
		First(&transition).Error
	return &transition, err
}

// ListByFromStatus — получение всех переходов, исходящих из указанного статуса
func (r *StatusTransitionRepo) ListByFromStatus(fromID uint) ([]models.StatusTransition, error) {
	var transitions []models.StatusTransition
	err := r.db.Where("from_status_id = ?", fromID).
		Find(&transitions).Error
	return transitions, err
}

// ListByToStatus — получение всех переходов, исходящих из указанного статуса
func (r *StatusTransitionRepo) ListByToStatus(toID uint) ([]models.StatusTransition, error) {
	var transitions []models.StatusTransition
	err := r.db.Where("to_status_id = ?", toID).
		Find(&transitions).Error
	return transitions, err
}

// DeleteByStatuses — удаление перехода по паре статусов
func (r *StatusTransitionRepo) DeleteByStatuses(fromID, toID uint) error {
	return r.db.Where("from_status_id = ? AND to_status_id = ?", fromID, toID).
		Delete(&models.StatusTransition{}).Error
}

// CanTransition — проверка возможности перехода для конкретной задачи
// Возвращает:
// - bool: можно ли выполнить переход
// - string: сообщение (например, причина запрета)
// - error: ошибка БД
func (r *StatusTransitionRepo) CanTransition(
	fromID, toID uint,
	task *models.Task,
) (bool, string, error) {

	// Проверяем существование перехода
	exists, err := r.ExistsTransition(fromID, toID)
	if err != nil {
		return false, "", err
	}
	if !exists {
		return false, "переход между указанными статусами не разрешён", nil
	}

	return true, "переход разрешён", nil
}

// ExistsTransition — проверка существования перехода между статусами
func (r *StatusTransitionRepo) ExistsTransition(fromID, toID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.StatusTransition{}).
		Where("from_status_id = ? AND to_status_id = ?", fromID, toID).
		Count(&count).Error
	return count > 0, err
}
