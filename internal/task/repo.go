package task

import (
	"errors"
	"time"

	"github.com/eduardnesterov/task-manager/internal/models"
	"gorm.io/gorm"
)

// TaskRepo — структура репозитория задач
type TaskRepo struct {
	db *gorm.DB
}

func (r *TaskRepo) GetDB() *gorm.DB {
	return r.db
}

// NewTaskRepo — конструктор репозитория
func NewTaskRepo(db *gorm.DB) *TaskRepo {
	return &TaskRepo{db: db}
}

// === Реализация CommonRepository ===

func (r *TaskRepo) Create(entity any) error {
	task, ok := entity.(*models.Task)
	if !ok {
		return errors.New("entity must be of type *models.Task")
	}
	return r.db.Create(task).Error
}

func (r *TaskRepo) GetByID(id uint) (any, error) {
	var task models.Task
	err := r.db.
		First(&task, id).Error
	return &task, err
}

func (r *TaskRepo) Update(entity any) error {
	task, ok := entity.(*models.Task)
	if !ok {
		return errors.New("entity must be of type *models.Task")
	}
	return r.db.Model(&models.Task{}).
		Where("id = ?", task.ID).
		Updates(map[string]any{"name": task.Name,
			"description":            task.Description,
			"category":               task.Category,
			"parent_id":              task.ParentID,
			"status_id":              task.StatusID,
			"executor_id":            task.ExecutorID,
			"project_id":             task.ProjectID,
			"project_phase_id":       task.ProjectPhaseID,
			"start_time":             task.StartTime,
			"end_time":               task.EndTime,
			"priority":               task.Priority,
			"urgency_coefficient":    task.UrgencyCoefficient,
			"complexity_coefficient": task.ComplexityCoefficient,
			"completed":              task.Completed,
			"fact_end_time":          task.FactEndTime,
			"fact_time":              task.FactTime,
			"pause_date":             task.PauseDate,
			"updated_at":             time.Now(),
		}).Error
}

func (r *TaskRepo) Delete(id uint) error {
	return r.db.Delete(&models.Task{}, id).Error
}

func (r *TaskRepo) List() ([]any, error) {
	var tasks []models.Task
	err := r.db.
		Find(&tasks).Error

	// Приводим к []any
	result := make([]any, len(tasks))
	for i, t := range tasks {
		result[i] = &t
	}
	return result, err
}

// === Реализация CountableRepository ===

func (r *TaskRepo) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Task{}).Count(&count).Error
	return count, err
}

// === Реализация ByProjectRepository ===

func (r *TaskRepo) ListByProject(projectID uint) ([]any, error) {
	var tasks []models.Task
	err := r.db.Where("project_id = ?", projectID).
		Find(&tasks).Error

	// Приводим к []any
	result := make([]any, len(tasks))
	for i, t := range tasks {
		result[i] = &t
	}
	return result, err
}

// === Специфичные методы TaskRepository ===

func (r *TaskRepo) ListByExecutor(executorID uint) ([]models.Task, error) {
	var tasks []models.Task
	err := r.db.Where("executor_id = ?", executorID).
		Find(&tasks).Error
	return tasks, err
}

func (r *TaskRepo) ListActiveByProject(projectID uint) ([]models.Task, error) {
	var tasks []models.Task
	err := r.db.Where("project_id = ? AND deleted_at IS NULL", projectID).
		Find(&tasks).Error
	return tasks, err
}

func (r *TaskRepo) CountByStatus(statusID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Task{}).
		Where("status_id = ? AND deleted_at IS NULL", statusID).
		Count(&count).Error
	return count, err
}

func (r *TaskRepo) ListByCategory(category string) ([]any, error) {
	var tasks []models.Task
	err := r.db.Where("category = ?", category).
		Find(&tasks).Error

	result := make([]any, len(tasks))
	for i, t := range tasks {
		result[i] = &t
	}
	return result, err
}

func (r *TaskRepo) ListByPhase(phaseID uint) ([]any, error) {
	var tasks []models.Task
	err := r.db.Where("project_phase_id = ?", phaseID).
		Find(&tasks).Error

	result := make([]any, len(tasks))
	for i, t := range tasks {
		result[i] = &t
	}
	return result, err
}
