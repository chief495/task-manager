package project

import (
	"errors"

	"github.com/eduardnesterov/task-manager/internal/models"
	"gorm.io/gorm"
)

// ProjectRepo — структура репозитория проектов
type ProjectRepo struct {
	db *gorm.DB
}

// NewProjectRepo — конструктор репозитория
func NewProjectRepo(db *gorm.DB) *ProjectRepo {
	return &ProjectRepo{db: db}
}

// === Реализация CommonRepository ===

func (r *ProjectRepo) Create(entity any) error {
	project, ok := entity.(*models.Project)
	if !ok {
		return errors.New("entity must be of type *models.Project")
	}
	return r.db.Create(project).Error
}

// GetDB — возвращает ссылку на *gorm.DB (необходимо для Preload и сложных запросов)
// Реализация интерфейса, если нужно
func (r *ProjectRepo) GetDB() *gorm.DB {
	return r.db
}
func (r *ProjectRepo) GetByID(id uint) (any, error) {
	var project models.Project
	err := r.db.
		First(&project, id).Error
	return &project, err
}

func (r *ProjectRepo) Update(entity any) error {
	project, ok := entity.(*models.Project)
	if !ok {
		return errors.New("entity must be of type *models.Project")
	}
	return r.db.Save(project).Error
}

func (r *ProjectRepo) Delete(id uint) error {
	return r.db.Delete(&models.Project{}, id).Error
}

func (r *ProjectRepo) List() ([]any, error) {
	var projects []models.Project
	err := r.db.
		Find(&projects).Error

	// Приводим к []any
	result := make([]any, len(projects))
	for i, p := range projects {
		result[i] = &p
	}
	return result, err
}

// === Реализация CountableRepository ===

func (r *ProjectRepo) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Project{}).Count(&count).Error
	return count, err
}

// === Реализация SearchableRepository ===

func (r *ProjectRepo) Search(query string) ([]any, error) {
	var projects []models.Project
	err := r.db.Where("name ILIKE ?", "%"+query+"%").
		Where("deleted_at IS NULL").
		Find(&projects).Error

	// Приводим к []any
	result := make([]any, len(projects))
	for i, p := range projects {
		result[i] = &p
	}
	return result, err
}

// === Реализация ByProjectRepository ===

func (r *ProjectRepo) ListByProject(projectID uint) ([]any, error) {
	// Для проектов этот метод не имеет смысла (проект не может быть "вложен" в другой проект)
	// Возвращаем пустой срез и nil-ошибку
	return []any{}, nil
}

// === Специфичные методы ProjectRepository ===

func (r *ProjectRepo) ListByResponsible(responsibleID uint) ([]models.Project, error) {
	var projects []models.Project
	err := r.db.Where("responsible_id = ?", responsibleID).
		Find(&projects).Error
	return projects, err
}

func (r *ProjectRepo) ListVisibleForUser(userID uint) ([]models.Project, error) {
	var projects []models.Project
	err := r.db.
		Where(
			"responsible_id = ? OR EXISTS (?) OR EXISTS (?) OR EXISTS (?)",
			userID,
			r.db.Model(&models.ProjectMember{}).
				Select("1").
				Where("project_members.project_id = projects.id").
				Where("project_members.user_id = ?", userID).
				Where("project_members.deleted_at IS NULL"),
			r.db.Model(&models.Task{}).
				Select("1").
				Where("tasks.project_id = projects.id").
				Where("tasks.executor_id = ?", userID).
				Where("tasks.deleted_at IS NULL"),
			r.db.Model(&models.ProjectAccess{}).
				Select("1").
				Where("project_accesses.project_id = projects.id").
				Where("project_accesses.user_id = ?", userID).
				Where("project_accesses.deleted_at IS NULL"),
		).
		Find(&projects).Error
	return projects, err
}

func (r *ProjectRepo) CreateMember(entity any) error {
	member, ok := entity.(*models.ProjectMember)
	if !ok {
		return errors.New("entity must be of type *models.ProjectMember")
	}
	return r.db.Create(member).Error
}

func (r *ProjectRepo) GetMemberByID(id uint) (*models.ProjectMember, error) {
	var member models.ProjectMember
	err := r.db.First(&member, id).Error
	return &member, err
}

func (r *ProjectRepo) UpdateMember(member *models.ProjectMember) error {
	return r.db.Save(member).Error
}

func (r *ProjectRepo) DeleteMember(id uint) error {
	return r.db.Delete(&models.ProjectMember{}, id).Error
}

func (r *ProjectRepo) ListMembersByProject(projectID uint) ([]models.ProjectMember, error) {
	var members []models.ProjectMember
	err := r.db.Where("project_id = ?", projectID).Find(&members).Error
	return members, err
}

func (r *ProjectRepo) ListByStage(stageID uint) ([]models.Project, error) {
	var projects []models.Project
	err := r.db.Where("stage_id = ?", stageID).
		Find(&projects).Error
	return projects, err
}
func (r *ProjectRepo) GetActiveProjects() ([]models.Project, error) {
	var projects []models.Project
	err := r.db.Where("deleted_at IS NULL").
		Find(&projects).Error
	return projects, err
}

func (r *ProjectRepo) GetProjectsByStage(stageID uint) ([]models.Project, error) {
	var projects []models.Project
	err := r.db.Where("stage_id = ?", stageID).
		Find(&projects).Error
	return projects, err
}
