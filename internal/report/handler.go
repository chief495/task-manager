package report

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/eduardnesterov/task-manager/internal/models"
	"github.com/eduardnesterov/task-manager/internal/project"
	"github.com/gin-gonic/gin"
)

// ReportHandler — обработчик HTTP-запросов для отчётов
type ReportHandler struct {
	service        *ReportService
	projectService *project.ProjectService
}

// NewReportHandler — конструктор обработчика
func NewReportHandler(service *ReportService,
	projectService *project.ProjectService) *ReportHandler {
	return &ReportHandler{
		service:        service,
		projectService: projectService,
	}
}

// CreateReport — создание нового отчёта
// @Summary Создать отчёт
// @Description Создаёт новый отчёт с указанными параметрами
// @Tags отчёты
// @Accept json
// @Produce json
// @Param report body struct{UserID uint `json:"user_id" binding:"required"` ProjectID *uint `json:"project_id"` Text string `json:"text" binding:"required,max=5000"` CreatedAt *time.Time `json:"created_at"`} true "Данные отчёта"
// @Success 201 {object} models.Report "Созданный отчёт"
// @Failure 400 {object} gin.H "Ошибка валидации входных данных"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /reports [post]
func (h *ReportHandler) CreateReport(c *gin.Context) {
	var input CreateReportRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report := &models.Report{
		UserID:    input.UserID,
		ProjectID: input.ProjectID,
		Text:      input.Text,
		Overwork:  input.Overwork,
		CreatedAt: input.CreatedAt,
	}

	err := h.service.Create(report)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, report)
}

// GetReport — получение отчёта по ID
// @Summary Получить отчёт по ID
// @Description Возвращает отчёт с указанным ID
// @Tags отчёты
// @Produce json
// @Param id path int true "ID отчёта"
// @Success 200 {object} models.Report "Отчёт"
// @Failure 400 {object} gin.H "Некорректный формат ID"
// @Failure 404 {object} gin.H "Отчёт не найден"
// @Router /reports/{id} [get]
func (h *ReportHandler) GetReport(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
		return
	}

	report, err := h.service.GetByID(uint(id))
	if err != nil {
		if errors.Is(err, ErrReportNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, report)
}

// ListReports — получение списка всех отчётов
// @Summary Получить список всех отчётов
// @Description Возвращает все отчёты в системе
// @Tags отчёты
// @Produce json
// @Success 200 {array} models.Report "Список отчётов"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /reports [get]
func (h *ReportHandler) ListReports(c *gin.Context) {
	var (
		dateFrom        = c.Query("date_from")
		dateTo          = c.Query("date_to")
		projectID       = c.Query("project_id")
		currentUserID   = c.MustGet("user_id").(uint)
		currentUserRole = c.MustGet("role").(string)
	)
	// Админ видит всё, менеджер — только по своим проектам
	var allowedProjectIDs []uint
	if currentUserRole != "admin" {
		projects, err := h.projectService.ListProjectsByResponsible(currentUserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, p := range projects {
			allowedProjectIDs = append(allowedProjectIDs, p.ID)
		}
	}

	reports, err := h.service.FilterReports(dateFrom, dateTo, projectID, allowedProjectIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, reports)
}

// UpdateReport — обновление отчёта
// @Summary Обновить отчёт
// @Description Обновляет отчёт по ID (частичное обновление)
// @Tags отчёты
// @Accept json
// @Produce json
// @Param report body struct{ID uint `json:"id" binding:"required"` UserID *uint `json:"user_id"` ProjectID **uint `json:"project_id"` Text *string `json:"text"` CreatedAt *time.Time `json:"created_at"`} true "Данные для обновления"
// @Success 200 {object} models.Report "Обновлённый отчёт"
// @Failure 400 {object} gin.H "Ошибка валидации входных данных"
// @Failure 404 {object} gin.H "Отчёт не найден"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /reports [put]
func (h *ReportHandler) UpdateReport(c *gin.Context) {
	var input UpdateReportRequest
	idStr := c.Param("id")
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
		return
	}
	report, err := h.service.GetByID(uint(id))
	if err != nil {
		if errors.Is(err, ErrReportNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Обновляем только переданные поля
	if input.UserID != nil {
		report.UserID = *input.UserID
	}
	if input.ProjectID != nil {
		report.ProjectID = *input.ProjectID
	}
	if input.Text != nil {
		report.Text = *input.Text
	}
	if input.CreatedAt != nil {
		report.CreatedAt = *input.CreatedAt
	}
	if input.Overwork != nil {
		report.Overwork = *input.Overwork
	}
	err = h.service.Update(report)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// DeleteReport — удаление отчёта
// @Summary Удалить отчёт
// @Description Удаляет отчёт по ID
// @Tags отчёты
// @Param id path int true "ID отчёта"
// @Success 204 "Отчёт удалён"
// @Failure 400 {object} gin.H "Некорректный формат ID"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /reports/{id} [delete]
func (h *ReportHandler) DeleteReport(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
		return
	}

	err = h.service.Delete(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListReportsByUser — получение отчётов по пользователю
// @Summary Получить отчёты пользователя
// @Description Возвращает список отчётов, созданных указанным пользователем
// @Tags отчёты
// @Produce json
// @Param user_id query int true "ID пользователя"
// @Success 200 {array} models.Report "Список отчётов пользователя"
// @Failure 400 {object} gin.H "Некорректный формат user_id"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /reports/by-user [get]
func (h *ReportHandler) ListReportsByUser(c *gin.Context) {
	userIDStr := c.Query("user_id")
	dateStr := c.Query("date")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id format"})
		return
	}
	reports, err := h.service.ListByUser(uint(userID), dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, reports)
}

// ListReportsByProject — получение отчётов по проекту
// @Summary Получить отчёты проекта
// @Description Возвращает список отчётов, связанных с указанным проектом
// @Tags отчёты
// @Produce json
// @Param project_id query int true "ID проекта"
// @Success 200 {array} models.Report "Список отчётов проекта"
// @Failure 400 {object} gin.H "Некорректный формат project_id"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /reports/by-project [get]
func (h *ReportHandler) ListReportsByProject(c *gin.Context) {
	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
		return
	}

	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id format"})
		return
	}

	reports, err := h.service.ListByProject(uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, reports)
}
