package organization

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

// OrganizationService — сервис для работы с организациями
type OrganizationService struct {
	repo OrganizationRepository
}

func NewOrganizationService(repo OrganizationRepository) *OrganizationService {
	return &OrganizationService{repo: repo}
}

func (s *OrganizationService) CreateOrganization(org *models.Organization) error {
	if org.Name == "" {
		return errors.New("название организации не может быть пустым")
	}
	if err := validatePhoneAndEmail(org.Phone, org.Email); err != nil {
		return err
	}
	return s.repo.Create(org)
}

func (s *OrganizationService) GetOrganization(id uint) (*models.Organization, error) {
	entity, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return entity.(*models.Organization), nil
}

func (s *OrganizationService) GetOrganizationByName(name string) (*models.Organization, error) {
	entity, err := s.repo.GetByName(name)
	if err != nil {
		return nil, err
	}
	if entity == nil {
		return nil, nil
	}
	return entity.(*models.Organization), nil
}

func (s *OrganizationService) UpdateOrganization(org *models.Organization) error {
	if org.ID == 0 {
		return errors.New("ID организации не указан")
	}
	if err := validatePhoneAndEmail(org.Phone, org.Email); err != nil {
		return err
	}
	return s.repo.Update(org)
}

func (s *OrganizationService) DeleteOrganization(id uint) error {
	return s.repo.Delete(id)
}

func (s *OrganizationService) ListOrganizations() ([]*models.Organization, error) {
	entities, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	list := make([]*models.Organization, len(entities))
	for i, e := range entities {
		list[i] = e.(*models.Organization)
	}
	return list, nil
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
