package organization

import (
	"errors"

	"github.com/eduardnesterov/task-manager/internal/models"
	"gorm.io/gorm"
)

// OrganizationRepo — репозиторий организаций
type OrganizationRepo struct {
	db *gorm.DB
}

func NewOrganizationRepo(db *gorm.DB) *OrganizationRepo {
	return &OrganizationRepo{db: db}
}

func (r *OrganizationRepo) GetDB() *gorm.DB {
	return r.db
}

func (r *OrganizationRepo) Create(entity any) error {
	org, ok := entity.(*models.Organization)
	if !ok {
		return errors.New("entity must be of type *models.Organization")
	}
	return r.db.Create(org).Error
}

func (r *OrganizationRepo) GetByID(id uint) (any, error) {
	var org models.Organization
	err := r.db.First(&org, id).Error
	return &org, err
}

func (r *OrganizationRepo) Update(entity any) error {
	org, ok := entity.(*models.Organization)
	if !ok {
		return errors.New("entity must be of type *models.Organization")
	}
	return r.db.Save(org).Error
}

func (r *OrganizationRepo) Delete(id uint) error {
	return r.db.Delete(&models.Organization{}, id).Error
}

func (r *OrganizationRepo) List() ([]any, error) {
	var list []models.Organization
	err := r.db.Find(&list).Error
	if err != nil {
		return nil, err
	}
	result := make([]any, len(list))
	for i := range list {
		result[i] = &list[i]
	}
	return result, nil
}

func (r *OrganizationRepo) GetByName(name string) (any, error) {
	var org models.Organization
	err := r.db.Where("name = ?", name).First(&org).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}
