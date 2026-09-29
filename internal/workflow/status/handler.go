package status

import (
	"net/http"
	"strconv"

	"github.com/eduardnesterov/task-manager/internal/models"
	"github.com/gin-gonic/gin"
)

// StatusHandler — обработчик HTTP‑запросов для статусов задач
type StatusHandler struct {
	service *StatusService
}

// NewStatusHandler — конструктор обработчика
func NewStatusHandler(service *StatusService) *StatusHandler {
	return &StatusHandler{
		service: service,
	}
}

// CreateStatus — создание нового статуса
// @Summary Создать статус
// @Description Создаёт новый статус с указанными параметрами
// @Tags статусы
// @Accept json
// @Produce json
// @Param status body struct{Name string `json:"name" binding:"required,max=50"` Description string `json:"description" binding:"max=200"` IsDefault bool `json:"is_default"`} true "Данные статуса"
// @Success 201 {object} models.Status "Созданный статус"
// @Failure 400 {object} gin.H "Ошибка валидации входных данных"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /statuses [post]
func (h *StatusHandler) CreateStatus(c *gin.Context) {
	var input CreateStatusRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := &models.Status{
		Name:        input.Name,
		Description: input.Description,
		IsDefault:   input.IsDefault,
		IsFinal:     input.IsFinal,
	}

	err := h.service.CreateStatus(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, status)
}

// GetStatus — получение статуса по ID
// @Summary Получить статус по ID
// @Description Возвращает статус с указанным ID
// @Tags статусы
// @Produce json
// @Param id path int true "ID статуса"
// @Success 200 {object} models.Status "Статус"
// @Failure 400 {object} gin.H "Некорректный формат ID"
// @Failure 404 {object} gin.H "Статус не найден"
// @Router /statuses/{id} [get]
func (h *StatusHandler) GetStatus(c *gin.Context) {
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

	status, err := h.service.GetStatus(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "status not found"})
		return
	}

	c.JSON(http.StatusOK, status)
}

// ListStatuses — получение списка всех статусов
// @Summary Получить список всех статусов
// @Description Возвращает все статусы в системе
// @Tags статусы
// @Produce json
// @Success 200 {array} models.Status "Список статусов"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /statuses [get]
func (h *StatusHandler) ListStatuses(c *gin.Context) {
	statuses, err := h.service.ListStatuses()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, statuses)
}

// UpdateStatus — обновление статуса
// @Summary Обновить статус
// @Description Обновляет статус по ID (частичное обновление)
// @Tags статусы
// @Accept json
// @Produce json
// @Param status body struct{ID uint `json:"id" binding:"required"` Name *string `json:"name"` Description *string `json:"description"` IsDefault *bool `json:"is_default"`} true "Данные для обновления"
// @Success 200 {object} models.Status "Обновлённый статус"
// @Failure 400 {object} gin.H "Ошибка валидации входных данных"
// @Failure 404 {object} gin.H "Статус не найден"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /statuses [put]
func (h *StatusHandler) UpdateStatus(c *gin.Context) {
	var input UpdateStatusRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status, err := h.service.GetStatus(input.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "status not found"})
		return
	}

	// Обновляем только переданные поля
	if input.Name != nil {
		status.Name = *input.Name
	}
	if input.Description != nil {
		status.Description = *input.Description
	}
	if input.IsDefault != nil {
		status.IsDefault = *input.IsDefault
	}

	err = h.service.UpdateStatus(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// DeleteStatus — удаление статуса
// @Summary Удалить статус
// @Description Удаляет статус по ID
// @Tags статусы
// @Param id path int true "ID статуса"
// @Success 204 "Статус удалён"
// @Failure 400 {object} gin.H "Некорректный формат ID"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /statuses/{id} [delete]
func (h *StatusHandler) DeleteStatus(c *gin.Context) {
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

	err = h.service.DeleteStatus(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetDefaultStatus — получение статуса по умолчанию
// @Summary Получить статус по умолчанию
// @Description Возвращает статус, отмеченный как стандартный/начальный для системы
// @Tags статусы
// @Produce json
// @Success 200 {object} models.Status "Статус по умолчанию"
// @Failure 404 {object} gin.H "Статус по умолчанию не найден"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /statuses/default [get]
func (h *StatusHandler) GetDefaultStatus(c *gin.Context) {
	status, err := h.service.GetDefaultStatus()
	if err != nil {
		if err.Error() == "default status not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, status)
}
