package transition

import (
	"errors"

	"github.com/eduardnesterov/task-manager/internal/models"
	"gorm.io/gorm"
)

// StageTransitionRepo — структура репозитория переходов между стадиями
type StageTransitionRepo struct {
	db *gorm.DB
}

func (r *StageTransitionRepo) GetDB() *gorm.DB {
	return r.db
}

// NewStageTransitionRepo — конструктор репозитория
func NewStageTransitionRepo(db *gorm.DB) *StageTransitionRepo {
	return &StageTransitionRepo{db: db}
}

// === Реализация CommonRepository ===

func (r *StageTransitionRepo) Create(entity any) error {
	transition, ok := entity.(*models.StageTransition)
	if !ok {
		return errors.New("entity must be of type *models.StageTransition")
	}
	return r.db.Create(transition).Error
}

func (r *StageTransitionRepo) GetByID(id uint) (any, error) {
	var transition models.StageTransition
	err := r.db.
		First(&transition, id).Error
	return &transition, err
}

func (r *StageTransitionRepo) Update(entity any) error {
	transition, ok := entity.(*models.StageTransition)
	if !ok {
		return errors.New("entity must be of type *models.StageTransition")
	}
	return r.db.Save(transition).Error
}

func (r *StageTransitionRepo) Delete(id uint) error {
	return r.db.Delete(&models.StageTransition{}, id).Error
}

func (r *StageTransitionRepo) List() ([]any, error) {
	var transitions []models.StageTransition
	err := r.db.
		Find(&transitions).Error

	// Приводим к []any
	result := make([]any, len(transitions))
	for i, t := range transitions {
		result[i] = &t
	}
	return result, err
}

// === Специфичные методы StageTransitionRepository ===

// GetByStages — получение перехода по парам стадий (fromID → toID)
func (r *StageTransitionRepo) GetByStages(fromID, toID uint) (*models.StageTransition, error) {
	var transition models.StageTransition
	err := r.db.Where("from_stage_id = ? AND to_stage_id = ?", fromID, toID).
		First(&transition).Error
	return &transition, err
}

// ListByFromStage — получение всех переходов, исходящих из указанной стадии
func (r *StageTransitionRepo) ListByFromStage(fromID uint) ([]models.StageTransition, error) {
	var transitions []models.StageTransition
	err := r.db.Where("from_stage_id = ?", fromID).
		Find(&transitions).Error
	return transitions, err
}

// ListByToStage — получение всех переходов, входящих в указанную стадию
func (r *StageTransitionRepo) ListByToStage(toID uint) ([]models.StageTransition, error) {
	var transitions []models.StageTransition
	err := r.db.Where("to_stage_id = ?", toID).
		Find(&transitions).Error
	return transitions, err
}

// DeleteByStages — удаление перехода по паре стадий
func (r *StageTransitionRepo) DeleteByStages(fromID, toID uint) error {
	return r.db.Where("from_stage_id = ? AND to_stage_id = ?", fromID, toID).
		Delete(&models.StageTransition{}).Error
}

// ExistsTransition — проверка существования перехода между стадиями
func (r *StageTransitionRepo) ExistsTransition(fromID, toID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.StageTransition{}).
		Where("from_stage_id = ? AND to_stage_id = ?", fromID, toID).
		Count(&count).Error
	return count > 0, err
}
