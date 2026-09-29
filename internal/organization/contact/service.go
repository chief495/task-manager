package contact

import (
	"errors"
	"regexp"
	"strings"

	"github.com/eduardnesterov/task-manager/internal/models"
)

var (
	emailPattern             = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]{2,}$`)
	phoneAllowedCharsPattern = regexp.MustCompile(`^\+?[0-9\s()\-]+$`)
)

// OrganizationContactService — сервис контактов компаний
type OrganizationContactService struct {
	repo OrganizationContactRepository
}

func NewOrganizationContactService(repo OrganizationContactRepository) *OrganizationContactService {
	return &OrganizationContactService{repo: repo}
}

func (s *OrganizationContactService) CreateContact(contact *models.OrganizationContact) error {
	if contact.OrganizationID == 0 {
		return errors.New("организация не указана")
	}
	if contact.FullName == "" {
		return errors.New("ФИО контакта не указано")
	}
	if err := validatePhoneAndEmail(contact.Phone, contact.Email); err != nil {
		return err
	}
	return s.repo.Create(contact)
}

func (s *OrganizationContactService) GetContact(id uint) (*models.OrganizationContact, error) {
	entity, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return entity.(*models.OrganizationContact), nil
}

func (s *OrganizationContactService) UpdateContact(contact *models.OrganizationContact) error {
	if contact.ID == 0 {
		return errors.New("ID контакта не указан")
	}
	if contact.OrganizationID == 0 {
		return errors.New("организация не указана")
	}
	if contact.FullName == "" {
		return errors.New("ФИО контакта не указано")
	}
	if err := validatePhoneAndEmail(contact.Phone, contact.Email); err != nil {
		return err
	}
	return s.repo.Update(contact)
}

func (s *OrganizationContactService) DeleteContact(id uint) error {
	return s.repo.Delete(id)
}

func (s *OrganizationContactService) ListContacts(organizationID uint) ([]*models.OrganizationContact, error) {
	if organizationID != 0 {
		contacts, err := s.repo.ListByOrganization(organizationID)
		if err != nil {
			return nil, err
		}
		result := make([]*models.OrganizationContact, len(contacts))
		for i := range contacts {
			result[i] = &contacts[i]
		}
		return result, nil
	}

	entities, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	result := make([]*models.OrganizationContact, len(entities))
	for i, entity := range entities {
		result[i] = entity.(*models.OrganizationContact)
	}
	return result, nil
}

func validatePhoneAndEmail(phone, email string) error {
	if !isValidPhone(phone) {
		return errors.New("некорректный номер телефона")
	}
	if !isValidEmail(email) {
		return errors.New("некорректный email")
	}
	return nil
}

func isValidEmail(value string) bool {
	value = strings.TrimSpace(value)
	return value == "" || emailPattern.MatchString(value)
}

func isValidPhone(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return true
	}
	if !phoneAllowedCharsPattern.MatchString(value) {
		return false
	}
	digitCount := 0
	for _, char := range value {
		if char >= '0' && char <= '9' {
			digitCount++
		}
	}
	return digitCount >= 7 && digitCount <= 15
}
