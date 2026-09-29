package user

import (
	"github.com/eduardnesterov/task-manager/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(entity any) error
	GetByID(id uint) (any, error)
	Update(entity any) error
	Delete(id uint) error
	List() ([]any, error)
	GetDB() *gorm.DB
	Search(query string) ([]any, error)
	Count() (int64, error)

	GetByEmail(email string) (*models.User, error)
	Exists(id uint) (bool, error)
	GetActiveUsers() ([]models.User, error)
	GetUsersByRole(role string) ([]models.User, error)
	GetUserTaskStats(userID uint) (*models.TaskStats, error)
}
