package contact

import (
	"errors"

	"github.com/eduardnesterov/task-manager/internal/models"
	"gorm.io/gorm"
)

// OrganizationContactRepo — репозиторий контактов компаний
type OrganizationContactRepo struct {
	db *gorm.DB
}

func NewOrganizationContactRepo(db *gorm.DB) *OrganizationContactRepo {
	return &OrganizationContactRepo{db: db}
}

func (r *OrganizationContactRepo) GetDB() *gorm.DB {
	return r.db
}

func (r *OrganizationContactRepo) Create(entity any) error {
	contact, ok := entity.(*models.OrganizationContact)
	if !ok {
		return errors.New("entity must be of type *models.OrganizationContact")
	}
	return r.db.Create(contact).Error
}

func (r *OrganizationContactRepo) GetByID(id uint) (any, error) {
	var contact models.OrganizationContact
	err := r.db.First(&contact, id).Error
	return &contact, err
}

func (r *OrganizationContactRepo) Update(entity any) error {
	contact, ok := entity.(*models.OrganizationContact)
	if !ok {
		return errors.New("entity must be of type *models.OrganizationContact")
	}
	return r.db.Save(contact).Error
}

func (r *OrganizationContactRepo) Delete(id uint) error {
	return r.db.Delete(&models.OrganizationContact{}, id).Error
}

func (r *OrganizationContactRepo) List() ([]any, error) {
	var contacts []models.OrganizationContact
	err := r.db.Order("organization_id ASC, full_name ASC").Find(&contacts).Error
	if err != nil {
		return nil, err
	}
	result := make([]any, len(contacts))
	for i := range contacts {
		result[i] = &contacts[i]
	}
	return result, nil
}

func (r *OrganizationContactRepo) ListByOrganization(organizationID uint) ([]models.OrganizationContact, error) {
	var contacts []models.OrganizationContact
	err := r.db.Where("organization_id = ?", organizationID).Order("full_name ASC").Find(&contacts).Error
	return contacts, err
}
