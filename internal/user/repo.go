package user

import (
	"errors"
	"fmt"

	"github.com/eduardnesterov/task-manager/internal/models"
	"gorm.io/gorm"
)

// UserRepo — структура репозитория пользователей
type UserRepo struct {
	db *gorm.DB
}

func (r *UserRepo) GetDB() *gorm.DB {
	return r.db
}

// NewUserRepo — конструктор репозитория
func NewUserRepo(db *gorm.DB) *UserRepo {

	return &UserRepo{db: db}
}

// === Реализация CommonRepository ===

func (r *UserRepo) Create(entity any) error {
	user, ok := entity.(*models.User)
	fmt.Println(user)
	if !ok {
		return errors.New("entity must be of type *models.User")
	}
	return r.db.Create(user).Error
}

func (r *UserRepo) GetByID(id uint) (any, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	return &user, err
}

func (r *UserRepo) Update(entity any) error {
	user, ok := entity.(*models.User)
	if !ok {
		return errors.New("entity must be of type *models.User")
	}
	return r.db.Save(user).Error
}

func (r *UserRepo) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

func (r *UserRepo) List() ([]any, error) {
	var users []models.User
	err := r.db.Find(&users).Error

	// Приводим к []any
	result := make([]any, len(users))
	for i, u := range users {
		result[i] = &u
	}
	return result, err
}

// === Реализация CountableRepository ===

func (r *UserRepo) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Count(&count).Error
	return count, err
}

// === Реализация SearchableRepository ===

func (r *UserRepo) Search(query string) ([]any, error) {
	var users []models.User
	err := r.db.Where("name ILIKE ?", "%"+query+"%").
		Where("deleted_at IS NULL").
		Find(&users).Error

	// Приводим к []any
	result := make([]any, len(users))
	for i, u := range users {
		result[i] = &u
	}
	return result, err
}

// === Специфичные методы UserRepository ===

func (r *UserRepo) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *UserRepo) Exists(id uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).
		Where("id = ?", id).
		Count(&count).Error
	return count > 0, err
}

func (r *UserRepo) GetActiveUsers() ([]models.User, error) {
	var users []models.User
	err := r.db.Where("is_active = ? AND deleted_at IS NULL", true).Find(&users).Error
	return users, err
}

func (r *UserRepo) GetUsersByRole(role string) ([]models.User, error) {
	var users []models.User
	err := r.db.Where("role = ? AND deleted_at IS NULL", role).
		Find(&users).Error
	return users, err
}

// GetUserTaskStats — получает статистику по задачам пользователя через JOIN с таблицей статусов
func (r *UserRepo) GetUserTaskStats(userID uint) (*models.TaskStats, error) {
	var stats models.TaskStats
	stats.UserID = userID

	// Получаем общее количество задач
	var total int64
	if err := r.db.Model(&models.Task{}).Where("executor_id = ?", userID).Count(&total).Error; err != nil {
		return nil, err
	}
	stats.Total = int(total)

	// Получаем ID финальных статусов
	var finalStatusIDs []uint
	if err := r.db.Model(&models.Status{}).Where("is_final = ?", true).Pluck("id", &finalStatusIDs).Error; err != nil {
		return nil, err
	}

	// Если есть финальные статусы, считаем завершённые задачи
	if len(finalStatusIDs) > 0 {
		var done int64
		if err := r.db.Model(&models.Task{}).
			Where("executor_id = ? AND status_id IN ?", userID, finalStatusIDs).
			Count(&done).Error; err != nil {
			return nil, err
		}
		stats.Done = int(done)
	} // else stats.Done остаётся 0

	// В работе = всего минус завершённые
	stats.InProgress = stats.Total - stats.Done

	return &stats, nil
}
