package participant

import (
	"errors"

	"github.com/eduardnesterov/task-manager/internal/models"
)

// ParticipantService — сервис для работы с участниками проектов
type ParticipantService struct {
	repo ParticipantRepository
}

func NewParticipantService(repo ParticipantRepository) *ParticipantService {
	return &ParticipantService{repo: repo}
}

func (s *ParticipantService) CreateParticipant(p *models.Participant) error {
	if p.ProjectID == 0 {
		return errors.New("project_id не указан")
	}
	if p.OrganizationID == 0 {
		return errors.New("organization_id не указан")
	}
	if p.Role == "" {
		return errors.New("роль участника не указана")
	}
	return s.repo.Create(p)
}

func (s *ParticipantService) GetParticipant(id uint) (*models.Participant, error) {
	entity, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return entity.(*models.Participant), nil
}

func (s *ParticipantService) UpdateParticipant(p *models.Participant) error {
	if p.ID == 0 {
		return errors.New("ID участника не указан")
	}
	return s.repo.Update(p)
}

func (s *ParticipantService) DeleteParticipant(id uint) error {
	return s.repo.Delete(id)
}

func (s *ParticipantService) ListParticipants() ([]*models.Participant, error) {
	entities, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	list := make([]*models.Participant, len(entities))
	for i, e := range entities {
		list[i] = e.(*models.Participant)
	}
	return list, nil
}

func (s *ParticipantService) ListParticipantsByProject(projectID uint) ([]*models.Participant, error) {
	entities, err := s.repo.ListByProject(projectID)
	if err != nil {
		return nil, err
	}
	list := make([]*models.Participant, len(entities))
	for i, e := range entities {
		list[i] = e.(*models.Participant)
	}
	return list, nil
}
