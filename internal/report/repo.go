package report

import (
	"errors"
	"strconv"
	"time"

	"github.com/eduardnesterov/task-manager/internal/models"
	"gorm.io/gorm"
)

type ReportRepo struct {
	db *gorm.DB
}

func (r *ReportRepo) GetDB() *gorm.DB {
	return r.db
}

func NewReportRepo(db *gorm.DB) *ReportRepo {
	return &ReportRepo{db: db}
}

func (r *ReportRepo) GetByID(id uint) (any, error) {
	var report models.Report
	err := r.db.First(&report, id).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *ReportRepo) Create(report any) error {
	return r.db.Create(report).Error
}

func (r *ReportRepo) Update(report any) error {
	return r.db.Save(report).Error
}

func (r *ReportRepo) Delete(id uint) error {
	return r.db.Delete(&models.Report{}, id).Error
}

func (r *ReportRepo) List() ([]any, error) {
	var reports []*models.Report
	err := r.db.Find(&reports).Error
	if err != nil {
		return nil, err
	}
	// Приводим к []any
	result := make([]any, len(reports))
	for i, s := range reports {
		result[i] = s
	}
	return result, err
}

func (r *ReportRepo) ListByUser(userID uint, dateStr string) ([]models.Report, error) {
	var reports []*models.Report
	query := r.db.
		Where("user_id = ?", userID)
	if dateStr != "" {
		from, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, errors.New("invalid date_from format, expected YYYY-MM-DD")
		}
		query = query.Where("created_at >= ?", from)
	}
	err := query.Find(&reports).Error
	if err != nil {
		return nil, err
	} // Приводим к []any
	result := make([]models.Report, len(reports))
	for i, s := range reports {
		result[i] = *s
	}
	return result, nil
}

func (r *ReportRepo) ListByProject(projectID uint) ([]any, error) {
	var reports []*models.Report
	err := r.db.
		Where("project_id = ?", projectID).
		Find(&reports).Error
	if err != nil {
		return nil, err
	}
	// Приводим к []any
	result := make([]any, len(reports))
	for i, s := range reports {
		result[i] = &s
	}
	return result, nil
}

// FilterReports — фильтрация отчётов по параметрам и списку разрешённых проектов
func (r *ReportRepo) FilterReports(dateFrom, dateTo, projectIDStr string, allowedProjectIDs []uint) ([]*models.Report, error) {
	var reports []*models.Report

	query := r.db.Model(&models.Report{})

	// Фильтр по дате начала
	if dateFrom != "" {
		from, err := time.Parse("2006-01-02", dateFrom)
		if err != nil {
			return nil, errors.New("invalid date_from format, expected YYYY-MM-DD")
		}
		query = query.Where("created_at >= ?", from)
	}

	// Фильтр по дате окончания
	if dateTo != "" {
		to, err := time.Parse("2006-01-02", dateTo)
		if err != nil {
			return nil, errors.New("invalid date_to format, expected YYYY-MM-DD")
		}
		// Включаем весь день
		to = to.AddDate(0, 0, 1).Add(-time.Nanosecond)
		query = query.Where("created_at <= ?", to)
	}

	// Фильтр по проекту
	if projectIDStr != "" {
		projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
		if err != nil {
			return nil, errors.New("invalid project_id")
		}
		query = query.Where("project_id = ?", uint(projectID))
	}

	// Ограничение по доступным проектам (для не-админов)
	if len(allowedProjectIDs) > 0 {
		if projectIDStr != "" {
			// Проверяем, есть ли запрашиваемый проект в списке доступных
			allowed := false
			projectID, _ := strconv.ParseUint(projectIDStr, 10, 64)
			for _, id := range allowedProjectIDs {
				if id == uint(projectID) {
					allowed = true
					break
				}
			}
			if !allowed {
				return []*models.Report{}, nil // пустой список, но не ошибка
			}
		}
		query = query.Where("project_id IN ?", allowedProjectIDs)
	}

	err := query.Order("created_at DESC").Find(&reports).Error
	if err != nil {
		return nil, err
	}

	return reports, nil
}
