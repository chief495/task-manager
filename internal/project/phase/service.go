package phase

import (
	"errors"

	"github.com/eduardnesterov/task-manager/internal/models"
	"github.com/eduardnesterov/task-manager/internal/project"
	"github.com/eduardnesterov/task-manager/internal/user"
)

// ProjectPhaseService — сервис для работы с этапами проекта
type ProjectPhaseService struct {
	phaseRepo   ProjectPhaseRepository
	projectRepo project.ProjectRepository
	userRepo    user.UserRepository
}

func NewProjectPhaseService(
	phaseRepo ProjectPhaseRepository,
	projectRepo project.ProjectRepository,
	userRepo user.UserRepository,
) *ProjectPhaseService {
	return &ProjectPhaseService{
		phaseRepo:   phaseRepo,
		projectRepo: projectRepo,
		userRepo:    userRepo,
	}
}

func (s *ProjectPhaseService) CreatePhase(phase *models.ProjectPhase) error {
	if phase.ProjectID == 0 {
		return errors.New("project_id не указан")
	}
	if phase.ResponsibleID == 0 {
		return errors.New("responsible_id не указан")
	}
	if phase.Name == "" {
		return errors.New("название этапа не может быть пустым")
	}
	if _, err := s.projectRepo.GetByID(phase.ProjectID); err != nil {
		return err
	}
	if _, err := s.userRepo.GetByID(phase.ResponsibleID); err != nil {
		return err
	}

	return s.phaseRepo.Create(phase)
}

func (s *ProjectPhaseService) GetPhase(id uint) (*models.ProjectPhase, error) {
	entity, err := s.phaseRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return entity.(*models.ProjectPhase), nil
}

func (s *ProjectPhaseService) UpdatePhase(phase *models.ProjectPhase) error {
	if phase.ID == 0 {
		return errors.New("ID этапа не указан")
	}
	if phase.ProjectID == 0 {
		return errors.New("project_id не указан")
	}
	if phase.ResponsibleID == 0 {
		return errors.New("responsible_id не указан")
	}
	if phase.Name == "" {
		return errors.New("название этапа не может быть пустым")
	}
	if _, err := s.userRepo.GetByID(phase.ResponsibleID); err != nil {
		return err
	}
	return s.phaseRepo.Update(phase)
}

func (s *ProjectPhaseService) DeletePhase(id uint) error {
	return s.phaseRepo.Delete(id)
}

func (s *ProjectPhaseService) ListPhases() ([]*models.ProjectPhase, error) {
	entities, err := s.phaseRepo.List()
	if err != nil {
		return nil, err
	}
	list := make([]*models.ProjectPhase, len(entities))
	for i, e := range entities {
		list[i] = e.(*models.ProjectPhase)
	}
	return list, nil
}

func (s *ProjectPhaseService) ListPhasesByProject(projectID uint) ([]*models.ProjectPhase, error) {
	entities, err := s.phaseRepo.ListByProject(projectID)
	if err != nil {
		return nil, err
	}
	list := make([]*models.ProjectPhase, len(entities))
	for i, e := range entities {
		list[i] = e.(*models.ProjectPhase)
	}
	return list, nil
}
