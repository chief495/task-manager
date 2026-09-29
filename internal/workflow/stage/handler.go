package stage

import (
	"net/http"
	"strconv"

	"github.com/eduardnesterov/task-manager/internal/models"
	"github.com/gin-gonic/gin"
)

// StageHandler — обработчик HTTP‑запросов для стадий проекта
type StageHandler struct {
	service *StageService
}

// NewStageHandler — конструктор обработчика
func NewStageHandler(service *StageService) *StageHandler {
	return &StageHandler{
		service: service,
	}
}

// CreateStage — создание новой стадии
// @Summary Создать стадию проекта
// @Description Создаёт новую стадию с указанными параметрами
// @Tags стадии
// @Accept json
// @Produce json
// @Param stage body struct{Name string `json:"name" binding:"required,max=100"` Description string `json:"description"` Position int `json:"position" binding:"required,min=0"`} true "Данные стадии"
// @Success 201 {object} models.Stage "Созданная стадия"
// @Failure 400 {object} gin.H "Ошибка валидации входных данных"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /stages [post]
func (h *StageHandler) CreateStage(c *gin.Context) {
	var input CreateStageRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stage := &models.Stage{
		Name:        input.Name,
		Description: input.Description,
		Position:    input.Position,
	}

	err := h.service.CreateStage(stage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, stage)
}

// GetStage — получение стадии по ID
// @Summary Получить стадию по ID
// @Description Возвращает стадию с указанным ID
// @Tags стадии
// @Produce json
// @Param id path int true "ID стадии"
// @Success 200 {object} models.Stage "Стадия"
// @Failure 400 {object} gin.H "Некорректный формат ID"
// @Failure 404 {object} gin.H "Стадия не найдена"
// @Router /stages/{id} [get]
func (h *StageHandler) GetStage(c *gin.Context) {
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

	stage, err := h.service.GetStage(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "stage not found"})
		return
	}

	c.JSON(http.StatusOK, stage)
}

// ListStages — получение списка всех стадий
// @Summary Получить список всех стадий
// @Description Возвращает все стадии проектов в системе
// @Tags стадии
// @Produce json
// @Success 200 {array} models.Stage "Список стадий"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /stages [get]
func (h *StageHandler) ListStages(c *gin.Context) {
	stages, err := h.service.ListStages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stages)
}

// UpdateStage — обновление стадии
// @Summary Обновить стадию
// @Description Обновляет стадию по ID (частичное обновление)
// @Tags стадии
// @Accept json
// @Produce json
// @Param stage body struct{ID uint `json:"id" binding:"required"` Name *string `json:"name"` Description *string `json:"description"` Position *int `json:"position"`} true "Данные для обновления"
// @Success 200 {object} models.Stage "Обновлённая стадия"
// @Failure 400 {object} gin.H "Ошибка валидации входных данных"
// @Failure 404 {object} gin.H "Стадия не найдена"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /stages [put]
func (h *StageHandler) UpdateStage(c *gin.Context) {
	var input UpdateStageRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stage, err := h.service.GetStage(input.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "stage not found"})
		return
	}

	// Обновляем только переданные поля
	if input.Name != nil {
		stage.Name = *input.Name
	}
	if input.Description != nil {
		stage.Description = *input.Description
	}
	if input.Position != nil {
		stage.Position = *input.Position
	}

	err = h.service.UpdateStage(stage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stage)
}

// DeleteStage — удаление стадии
// @Summary Удалить стадию
// @Description Удаляет стадию по ID
// @Tags стадии
// @Param id path int true "ID стадии"
// @Success 204 "Стадия удалена"
// @Failure 400 {object} gin.H "Некорректный формат ID"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /stages/{id} [delete]
func (h *StageHandler) DeleteStage(c *gin.Context) {
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

	err = h.service.DeleteStage(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
