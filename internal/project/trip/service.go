package trip

import (
	"errors"
	"time"

	"github.com/eduardnesterov/task-manager/internal/models"
)

// TripService — сервис для работы с командировками
type TripService struct {
	repo TripRepository
}

func NewTripService(repo TripRepository) *TripService {
	return &TripService{repo: repo}
}

func (s *TripService) CreateTrip(trip *models.Trip) error {
	if trip.ProjectID == 0 {
		return errors.New("project_id не указан")
	}
	if trip.PersonID == 0 {
		return errors.New("person_id не указан")
	}
	if trip.Place == "" {
		return errors.New("место командировки не указано")
	}
	return s.repo.Create(trip)
}

func (s *TripService) GetTrip(id uint) (*models.Trip, error) {
	entity, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return entity.(*models.Trip), nil
}

func (s *TripService) UpdateTrip(trip *models.Trip) error {
	if trip.ID == 0 {
		return errors.New("ID командировки не указан")
	}
	return s.repo.Update(trip)
}

func (s *TripService) DeleteTrip(id uint) error {
	return s.repo.Delete(id)
}

func (s *TripService) ListTrips() ([]*models.Trip, error) {
	entities, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	list := make([]*models.Trip, len(entities))
	for i, e := range entities {
		list[i] = e.(*models.Trip)
	}
	return list, nil
}

func (s *TripService) CountTrips() (int64, error) {
	return s.repo.Count()
}

func (s *TripService) ListTripsByProject(projectID uint) ([]*models.Trip, error) {
	entities, err := s.repo.ListByProject(projectID)
	if err != nil {
		return nil, err
	}
	list := make([]*models.Trip, len(entities))
	for i, e := range entities {
		list[i] = e.(*models.Trip)
	}
	return list, nil
}

func (s *TripService) ListTripsByPerson(personID uint) ([]models.Trip, error) {
	return s.repo.ListByPerson(personID)
}

func (s *TripService) ListTripsActiveInPeriod(start, end time.Time) ([]models.Trip, error) {
	return s.repo.ListActiveInPeriod(start, end)
}
