package stage

import (
	"github.com/eduardnesterov/task-manager/internal/models"
)

type StageService struct {
	repo StageRepository
}

func NewStageService(repo StageRepository) *StageService {
	return &StageService{repo: repo}
}

// CreateStage — создание стадии
func (s *StageService) CreateStage(stage *models.Stage) error {
	return s.repo.Create(stage)
}

// GetStage — получение стадии по ID
func (s *StageService) GetStage(id uint) (*models.Stage, error) {
	entity, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return entity.(*models.Stage), nil
}

// UpdateStage — обновление стадии
func (s *StageService) UpdateStage(stage *models.Stage) error {
	return s.repo.Update(stage)
}

// DeleteStage — удаление стадии
func (s *StageService) DeleteStage(id uint) error {
	return s.repo.Delete(id)
}

// ListStages — список всех стадий
func (s *StageService) ListStages() ([]*models.Stage, error) {
	entities, err := s.repo.List()
	if err != nil {
		return nil, err
	}

	stages := make([]*models.Stage, len(entities))
	for i, e := range entities {
		stages[i] = e.(*models.Stage)
	}
	return stages, nil
}

// SearchStages — поиск стадий по названию
func (s *StageService) SearchStages(query string) ([]*models.Stage, error) {
	entities, err := s.repo.Search(query)
	if err != nil {
		return nil, err
	}

	stages := make([]*models.Stage, len(entities))
	for i, e := range entities {
		stages[i] = e.(*models.Stage)
	}
	return stages, nil
}
