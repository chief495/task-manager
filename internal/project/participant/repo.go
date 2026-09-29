package participant

import (
	"errors"

	"github.com/eduardnesterov/task-manager/internal/models"
	"gorm.io/gorm"
)

// ParticipantRepo — репозиторий участников проектов
type ParticipantRepo struct {
	db *gorm.DB
}

func NewParticipantRepo(db *gorm.DB) *ParticipantRepo {
	return &ParticipantRepo{db: db}
}

func (r *ParticipantRepo) GetDB() *gorm.DB {
	return r.db
}

func (r *ParticipantRepo) Create(entity any) error {
	p, ok := entity.(*models.Participant)
	if !ok {
		return errors.New("entity must be of type *models.Participant")
	}
	return r.db.Create(p).Error
}

func (r *ParticipantRepo) GetByID(id uint) (any, error) {
	var p models.Participant
	err := r.db.First(&p, id).Error
	return &p, err
}

func (r *ParticipantRepo) Update(entity any) error {
	p, ok := entity.(*models.Participant)
	if !ok {
		return errors.New("entity must be of type *models.Participant")
	}
	return r.db.Save(p).Error
}

func (r *ParticipantRepo) Delete(id uint) error {
	return r.db.Delete(&models.Participant{}, id).Error
}

func (r *ParticipantRepo) List() ([]any, error) {
	var list []models.Participant
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

func (r *ParticipantRepo) ListByProject(projectID uint) ([]any, error) {
	var list []models.Participant
	err := r.db.Where("project_id = ?", projectID).
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	result := make([]any, len(list))
	for i := range list {
		result[i] = &list[i]
	}
	return result, nil
}
