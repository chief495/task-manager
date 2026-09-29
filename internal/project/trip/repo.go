package trip

import (
	"errors"
	"time"

	"github.com/eduardnesterov/task-manager/internal/models"
	"gorm.io/gorm"
)

// TripRepo — репозиторий командировок
type TripRepo struct {
	db *gorm.DB
}

func NewTripRepo(db *gorm.DB) *TripRepo {
	return &TripRepo{db: db}
}

func (r *TripRepo) GetDB() *gorm.DB {
	return r.db
}

func (r *TripRepo) Create(entity any) error {
	trip, ok := entity.(*models.Trip)
	if !ok {
		return errors.New("entity must be of type *models.Trip")
	}
	return r.db.Create(trip).Error
}

func (r *TripRepo) GetByID(id uint) (any, error) {
	var trip models.Trip
	err := r.db.First(&trip, id).Error
	return &trip, err
}

func (r *TripRepo) Update(entity any) error {
	trip, ok := entity.(*models.Trip)
	if !ok {
		return errors.New("entity must be of type *models.Trip")
	}
	return r.db.Save(trip).Error
}

func (r *TripRepo) Delete(id uint) error {
	return r.db.Delete(&models.Trip{}, id).Error
}

func (r *TripRepo) List() ([]any, error) {
	var trips []models.Trip
	err := r.db.Find(&trips).Error
	if err != nil {
		return nil, err
	}
	result := make([]any, len(trips))
	for i := range trips {
		result[i] = &trips[i]
	}
	return result, nil
}

func (r *TripRepo) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Trip{}).Count(&count).Error
	return count, err
}

func (r *TripRepo) ListByProject(projectID uint) ([]any, error) {
	var trips []models.Trip
	err := r.db.Where("project_id = ?", projectID).
		Find(&trips).Error
	if err != nil {
		return nil, err
	}
	result := make([]any, len(trips))
	for i := range trips {
		result[i] = &trips[i]
	}
	return result, nil
}

func (r *TripRepo) ListByPerson(personID uint) ([]models.Trip, error) {
	var trips []models.Trip
	err := r.db.Where("person_id = ?", personID).
		Find(&trips).Error
	return trips, err
}

func (r *TripRepo) ListActiveInPeriod(start, end time.Time) ([]models.Trip, error) {
	var trips []models.Trip
	// Пересечение периодов: trip.begin_date <= end AND trip.end_date >= start
	err := r.db.Where("begin_date <= ? AND end_date >= ?", end, start).
		Find(&trips).Error
	return trips, err
}
