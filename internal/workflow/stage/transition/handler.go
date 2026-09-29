package transition

import (
	"net/http"
	"strconv"

	"github.com/eduardnesterov/task-manager/internal/models"
	"github.com/gin-gonic/gin"
)

// StageTransitionHandler — обработчик HTTP‑запросов для переходов между стадиями
type StageTransitionHandler struct {
	service *StageTransitionService
}

// NewStageTransitionHandler — конструктор обработчика
func NewStageTransitionHandler(service *StageTransitionService) *StageTransitionHandler {
	return &StageTransitionHandler{
		service: service,
	}
}

// CreateStageTransition — создание нового разрешённого перехода
// @Summary Создать переход между стадиями
// @Description Создаёт новый разрешённый переход с указанными параметрами
// @Tags переходы-между-стадиями
// @Accept json
// @Produce json
// @Param transition body struct{FromStageID uint `json:"from_stage_id" binding:"required"` ToStageID uint `json:"to_stage_id" binding:"required"` Condition string `json:"condition" binding:"max=200"` AllowedRoles []string `json:"allowed_roles" binding:"required"`} true "Данные перехода"
// @Success 201 {object} models.StageTransition "Созданный переход"
// @Failure 400 {object} gin.H "Ошибка валидации входных данных"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /stage-transitions [post]
func (h *StageTransitionHandler) CreateStageTransition(c *gin.Context) {
	var input CreateStageTransitionRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transition := &models.StageTransition{
		FromStageID:  input.FromStageID,
		ToStageID:    input.ToStageID,
		Condition:    input.Condition,
		AllowedRoles: input.AllowedRoles,
	}

	err := h.service.CreateStageTransition(transition)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, transition)
}

// GetStageTransition — получение перехода по ID
// @Summary Получить переход по ID
// @Description Возвращает переход между стадиями с указанным ID
// @Tags переходы-между-стадиями
// @Produce json
// @Param id path int true "ID перехода"
// @Success 200 {object} models.StageTransition "Переход"
// @Failure 400 {object} gin.H "Некорректный формат ID"
// @Failure 404 {object} gin.H "Переход не найден"
// @Router /stage-transitions/{id} [get]
func (h *StageTransitionHandler) GetStageTransition(c *gin.Context) {
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

	transition, err := h.service.GetStageTransition(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transition not found"})
		return
	}

	c.JSON(http.StatusOK, transition)
}

// ListStageTransitions — получение списка всех переходов
// @Summary Получить список всех переходов
// @Description Возвращает все разрешённые переходы между стадиями в системе
// @Tags переходы-между-стадиями
// @Produce json
// @Success 200 {array} models.StageTransition "Список переходов"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /stage-transitions [get]
func (h *StageTransitionHandler) ListStageTransitions(c *gin.Context) {
	transitions, err := h.service.ListStageTransitions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transitions)
}

// ListTransitionsByFromStage — получение переходов, исходящих из определённой стадии
// @Summary Получить переходы из конкретной стадии
// @Description Возвращает список переходов, которые начинаются со стадии с указанным ID
// @Tags переходы-между-стадиями
// @Produce json
// @Param from_stage_id query int true "ID исходной стадии"
// @Success 200 {array} models.StageTransition "Список переходов из стадии"
// @Failure 400 {object} gin.H "Некорректный формат from_stage_id"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /stage-transitions/from [get]
func (h *StageTransitionHandler) ListTransitionsByFromStage(c *gin.Context) {
	stageIDStr := c.Query("from_stage_id")
	if stageIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from_stage_id is required"})
		return
	}

	stageID, err := strconv.ParseUint(stageIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from_stage_id format"})
		return
	}

	transitions, err := h.service.ListTransitionsByFromStage(uint(stageID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transitions)
}

// ListTransitionsByToStage — получение переходов, ведущих в определённую стадию
// @Summary Получить переходы в конкретную стадию
// @Description Возвращает список переходов, которые ведут в стадию с указанным ID
// @Tags переходы-между-стадиями
// @Produce json
// @Param to_stage_id query int true "ID целевой стадии"
// @Success 200 {array} models.StageTransition "Список переходов в стадию"
// @Failure 400 {object} gin.H "Некорректный формат to_stage_id"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /stage-transitions/to [get]
func (h *StageTransitionHandler) ListTransitionsByToStage(c *gin.Context) {
	stageIDStr := c.Query("to_stage_id")
	if stageIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "to_stage_id is required"})
		return
	}

	stageID, err := strconv.ParseUint(stageIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to_stage_id format"})
		return
	}

	transitions, err := h.service.ListTransitionsByToStage(uint(stageID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transitions)
}

// UpdateStageTransition — обновление перехода
// @Summary Обновить переход между стадиями
// @Description Обновляет существующий переход (частичное обновление)
// @Tags переходы-между-стадиями
// @Accept json
// @Produce json
// @Param transition body struct{ID uint `json:"id" binding:"required"` Condition *string `json:"condition"` AllowedRoles *[]string `json:"allowed_roles"`} true "Данные для обновления"
// @Success 200 {object} models.StageTransition "Обновлённый переход"
// @Failure 400 {object} gin.H "Ошибка валидации входных данных"
// @Failure 404 {object} gin.H "Переход не найден"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /stage-transitions [put]
func (h *StageTransitionHandler) UpdateStageTransition(c *gin.Context) {
	var input UpdateStageTransitionRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transition, err := h.service.GetStageTransition(input.ID)
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

	err = h.service.UpdateStageTransition(transition)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transition)
}

// DeleteStageTransition — удаление перехода
// @Summary Удалить переход между стадиями
// @Description Удаляет переход по указанному ID
// @Tags переходы-между-стадиями
// @Param id path int true "ID перехода"
// @Success 204 "Переход удалён"
// @Failure 400 {object} gin.H "Некорректный формат ID"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /stage-transitions/{id} [delete]
func (h *StageTransitionHandler) DeleteStageTransition(c *gin.Context) {
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

	err = h.service.DeleteStageTransition(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
