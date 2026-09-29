package report

import "github.com/eduardnesterov/task-manager/internal/models"

type ReportRepository interface {
	Create(entity any) error
	GetByID(id uint) (any, error)
	Update(entity any) error
	Delete(id uint) error
	List() ([]any, error)
	ListByProject(projectID uint) ([]any, error)

	ListByUser(userID uint, dateStr string) ([]models.Report, error)
	FilterReports(dateFrom, dateTo, projectIDStr string, allowedProjectIDs []uint) ([]*models.Report, error)
}
