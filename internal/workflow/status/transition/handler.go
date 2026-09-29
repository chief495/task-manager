package transition

import (
	"net/http"
	"strconv"

	"github.com/eduardnesterov/task-manager/internal/models"
	"github.com/gin-gonic/gin"
)

// StatusTransitionHandler — обработчик HTTP‑запросов для переходов между статусами задач
type StatusTransitionHandler struct {
	service *StatusTransitionService
}

// NewStatusTransitionHandler — конструктор обработчика
func NewStatusTransitionHandler(service *StatusTransitionService) *StatusTransitionHandler {
	return &StatusTransitionHandler{
		service: service,
	}
}

// CreateStatusTransition — создание нового правила перехода между статусами
// @Summary Создать правило перехода между статусами
// @Description Создаёт новое правило перехода с указанными параметрами
// @Tags переходы-между-статусами
// @Accept json
// @Produce json
// @Param transition body struct{FromStatusID uint `json:"from_status_id" binding:"required"` ToStatusID uint `json:"to_status_id" binding:"required"` Condition string `json:"condition" binding:"max=200"` AllowedRoles []string `json:"allowed_roles" binding:"required"`} true "Данные правила перехода"
// @Success 201 {object} models.StatusTransition "Созданное правило перехода"
// @Failure 400 {object} gin.H "Ошибка валидации входных данных"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /status-transitions [post]
func (h *StatusTransitionHandler) CreateStatusTransition(c *gin.Context) {
	var input CreateStatusTransitionRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transition := &models.StatusTransition{
		FromStatusID: input.FromStatusID,
		ToStatusID:   input.ToStatusID,
		Condition:    input.Condition,
		AllowedRoles: input.AllowedRoles,
	}

	err := h.service.CreateStatusTransition(transition)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, transition)
}

// GetStatusTransition — получение правила перехода по ID
// @Summary Получить правило перехода по ID
// @Description Возвращает правило перехода между статусами с указанным ID
// @Tags переходы-между-статусами
// @Produce json
// @Param id path int true "ID правила перехода"
// @Success 200 {object} models.StatusTransition "Правило перехода"
// @Failure 400 {object} gin.H "Некорректный формат ID"
// @Failure 404 {object} gin.H "Правило перехода не найдено"
// @Router /status-transitions/{id} [get]
func (h *StatusTransitionHandler) GetStatusTransition(c *gin.Context) {
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

	transition, err := h.service.GetStatusTransition(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transition not found"})
		return
	}

	c.JSON(http.StatusOK, transition)
}

// ListStatusTransitions — получение всех правил переходов
// @Summary Получить список всех правил переходов
// @Description Возвращает все правила переходов между статусами в системе
// @Tags переходы-между-статусами
// @Produce json
// @Success 200 {array} models.StatusTransition "Список правил переходов"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /status-transitions [get]
func (h *StatusTransitionHandler) ListStatusTransitions(c *gin.Context) {
	transitions, err := h.service.ListStatusTransitions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transitions)
}

// ListTransitionsByFromStatus — получение переходов из указанного статуса
// @Summary Получить переходы из конкретного статуса
// @Description Возвращает список переходов, которые начинаются со статуса с указанным ID
// @Tags переходы-между-статусами
// @Produce json
// @Param from_status_id query int true "ID исходного статуса"
// @Success 200 {array} models.StatusTransition "Список переходов из статуса"
// @Failure 400 {object} gin.H "Некорректный формат from_status_id"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /status-transitions/from [get]
func (h *StatusTransitionHandler) ListTransitionsByFromStatus(c *gin.Context) {
	statusIDStr := c.Query("from_status_id")
	if statusIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from_status_id is required"})
		return
	}

	statusID, err := strconv.ParseUint(statusIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from_status_id format"})
		return
	}

	transitions, err := h.service.ListTransitionsByFromStatus(uint(statusID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transitions)
}

// ListTransitionsByToStatus — получение переходов в указанный статус
// @Summary Получить переходы в конкретный статус
// @Description Возвращает список переходов, которые ведут в статус с указанным ID
// @Tags переходы-между-статусами
// @Produce json
// @Param to_status_id query int true "ID целевого статуса"
// @Success 200 {array} models.StatusTransition "Список переходов в статус"
// @Failure 400 {object} gin.H "Некорректный формат to_status_id"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /status-transitions/to [get]
func (h *StatusTransitionHandler) ListTransitionsByToStatus(c *gin.Context) {
	statusIDStr := c.Query("to_status_id")
	if statusIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "to_status_id is required"})
		return
	}

	statusID, err := strconv.ParseUint(statusIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to_status_id format"})
		return
	}

	transitions, err := h.service.ListTransitionsByToStatus(uint(statusID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transitions)
}

// UpdateStatusTransition — обновление правила перехода
// @Summary Обновить правило перехода между статусами
// @Description Обновляет существующее правило перехода (частичное обновление)
// @Tags переходы-между-статусами
// @Accept json
// @Produce json
// @Param transition body struct{ID uint `json:"id" binding:"required"` Condition *string `json:"condition"` AllowedRoles *[]string `json:"allowed_roles"`} true "Данные для обновления"
// @Success 200 {object} models.StatusTransition "Обновлённое правило перехода"
// @Failure 400 {object} gin.H "Ошибка валидации входных данных"
// @Failure 404 {object} gin.H "Правило перехода не найдено"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /status-transitions [put]
func (h *StatusTransitionHandler) UpdateStatusTransition(c *gin.Context) {
	var input UpdateStatusTransitionRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transition, err := h.service.GetStatusTransition(input.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transition not found"})
		return
	}

	// Обновляем только переданные поля
	if input.Condition != nil {
		transition.Condition = *input.Condition
	}
	if input.AllowedRoles != nil {
		transition.AllowedRoles = *input.AllowedRoles
	}

	err = h.service.UpdateStatusTransition(transition)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transition)
}

// DeleteStatusTransition — удаление правила перехода
// @Summary Удалить правило перехода между статусами
// @Description Удаляет правило перехода по указанному ID
// @Tags переходы-между-статусами
// @Param id path int true "ID правила перехода"
// @Success 204 "Правило перехода удалено"
// @Failure 400 {object} gin.H "Некорректный формат ID"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /status-transitions/{id} [delete]
func (h *StatusTransitionHandler) DeleteStatusTransition(c *gin.Context) {
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

	err = h.service.DeleteStatusTransition(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
