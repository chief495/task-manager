package report

import (
	"errors"

	"github.com/eduardnesterov/task-manager/internal/models"
)

// Ошибки
var (
	ErrEmptyReportContent = errors.New("содержание отчёта не может быть пустым")
	ErrInvalidUserID      = errors.New("некорректный ID пользователя")
	ErrReportIDRequired   = errors.New("требуется ID отчёта для обновления")
	ErrReportNotFound     = errors.New("отчёт не найден")
	ErrPermissionDenied   = errors.New("доступ запрещён")
)

type ReportService struct {
	repo ReportRepository
}

func NewReportService(repo ReportRepository) *ReportService {
	return &ReportService{repo: repo}
}

func (s *ReportService) GetByID(id uint) (*models.Report, error) {
	entity, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return entity.(*models.Report), nil
}

func (s *ReportService) Create(report *models.Report) error {
	if report.Text == "" {
		return ErrEmptyReportContent
	}
	if report.UserID == 0 {
		return ErrInvalidUserID
	}
	return s.repo.Create(report)
}

func (s *ReportService) Update(report *models.Report) error {
	if report.ID == 0 {
		return ErrReportIDRequired
	}
	return s.repo.Update(report)
}

func (s *ReportService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *ReportService) List() ([]*models.Report, error) {
	entities, err := s.repo.List()
	if err != nil {
		return nil, err
	}

	reports := make([]*models.Report, len(entities))
	for i, e := range entities {
		reports[i] = e.(*models.Report)
	}
	return reports, nil
}

func (s *ReportService) ListByUser(userID uint, dateStr string) ([]*models.Report, error) {
	entities, err := s.repo.ListByUser(userID, dateStr)
	if err != nil {
		return nil, err
	}
	reports := make([]*models.Report, len(entities))
	for i, e := range entities {
		reports[i] = &e
	}
	return reports, nil
}

func (s *ReportService) ListByProject(projectID uint) ([]*models.Report, error) {
	entities, err := s.repo.ListByProject(projectID)
	if err != nil {
		return nil, err
	}

	reports := make([]*models.Report, len(entities))
	for i, e := range entities {
		reports[i] = e.(*models.Report)
	}
	return reports, nil
}

// FilterReports — фильтрация отчётов по дате, проекту и доступу
func (s *ReportService) FilterReports(dateFrom, dateTo, projectIDStr string, allowedProjectIDs []uint) ([]*models.Report, error) {
	return s.repo.FilterReports(dateFrom, dateTo, projectIDStr, allowedProjectIDs)
}
