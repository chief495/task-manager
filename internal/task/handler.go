package task

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/eduardnesterov/task-manager/internal/models"
	"github.com/gin-gonic/gin"
)

// TaskHandler — обработчик HTTP‑запросов для задач
type TaskHandler struct {
	service *TaskService
}

// NewTaskHandler — конструктор обработчика
func NewTaskHandler(service *TaskService) *TaskHandler {
	return &TaskHandler{
		service: service,
	}
}

// CreateTask — создание новой задачи
// @Summary Создать задачу
// @Description Создаёт новую задачу с указанными параметрами
// @Tags задачи
// @Accept json
// @Produce json
// @Param task body struct{Name string `json:"name" binding:"required,max=200"` Description string `json:"description"` ParentID *uint `json:"parent_id"` StatusID uint `json:"status_id" binding:"required"` ExecutorID uint `json:"executor_id" binding:"required"` ProjectID uint `json:"project_id" binding:"required"` StartTime *time.Time `json:"start_time"` EndTime *time.Time `json:"end_time"` Priority int `json:"priority" binding:"omitempty,min=0,max=2"`} true "Данные задачи"
// @Success 201 {object} models.Task "Созданная задача"
// @Failure 400 {object} gin.H "Ошибка валидации входных данных"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /tasks [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var input CreateTaskRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task := &models.Task{
		Name:                  input.Name,
		Description:           input.Description,
		Category:              input.Category,
		ParentID:              input.ParentID,
		StatusID:              input.StatusID,
		ExecutorID:            input.ExecutorID,
		ProjectID:             input.ProjectID,
		ProjectPhaseID:        input.ProjectPhaseID,
		StartTime:             input.StartTime,
		EndTime:               input.EndTime,
		Priority:              input.Priority,
		UrgencyCoefficient:    input.Urgency,
		ComplexityCoefficient: input.Complexity,
	}

	var assignerID uint
	if assignerIDValue, exists := c.Get("user_id"); exists {
		if parsedAssignerID, ok := assignerIDValue.(uint); ok {
			assignerID = parsedAssignerID
		}
	}

	err := h.service.CreateTask(task, assignerID)
	if err != nil {
		if err.Error() == "start_time and end_time are required" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// GetTask — получение задачи по ID
// @Summary Получить задачу по ID
// @Description Возвращает задачу с указанным ID
// @Tags задачи
// @Produce json
// @Param id path int true "ID задачи"
// @Success 200 {object} models.Task "Задача"
// @Failure 400 {object} gin.H "Некорректный формат ID"
// @Failure 404 {object} gin.H "Задача не найдена"
// @Router /tasks/{id} [get]
func (h *TaskHandler) GetTask(c *gin.Context) {
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

	task, err := h.service.GetTask(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// ListTasks — получение списка всех задач
// @Summary Получить список всех задач
// @Description Возвращает все задачи в системе
// @Tags задачи
// @Produce json
// @Success 200 {array} models.Task "Список задач"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /tasks [get]
func (h *TaskHandler) ListTasks(c *gin.Context) {

	id, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	tasks, err := h.service.ListTasks(id.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// UpdateTask — обновление задачи
// @Summary Обновить задачу
// @Description Обновляет задачу по ID (частичное обновление)
// @Tags задачи
// @Accept json
// @Produce json
// @Param task body struct{ID uint `json:"id" binding:"required"` Name *string `json:"name"` Description *string `json:"description"` ParentID *uint `json:"parent_id"` StatusID *uint `json:"status_id"` ExecutorID *uint `json:"executor_id"` ProjectID *uint `json:"project_id"` StartTime *time.Time `json:"start_time"` EndTime *time.Time `json:"end_time"` Priority *int `json:"priority" binding:"omitempty,min=0,max=2"`} true "Данные для обновления"
// @Success 200 {object} models.Task "Обновлённая задача"
// @Failure 400 {object} gin.H "Ошибка валидации входных данных"
// @Failure 404 {object} gin.H "Задача не найдена"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /tasks [put]
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	// ID из URL path
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}

	var input UpdateTaskRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ищем задачу по ID из URL, а не из body
	task, err := h.service.GetTask(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	// Обновляем только переданные поля
	if input.Name != nil {
		task.Name = *input.Name
	}
	if input.Description != nil {
		task.Description = *input.Description
	}
	if input.ParentID != nil {
		task.ParentID = input.ParentID
	}
	if input.StatusID != nil {
		task.StatusID = *input.StatusID
	}
	if input.ExecutorID != nil {
		task.ExecutorID = *input.ExecutorID
	}
	if input.ProjectID != nil {
		task.ProjectID = *input.ProjectID
	}
	if input.ProjectPhaseID != nil {
		if string(*input.ProjectPhaseID) == "null" {
			task.ProjectPhaseID = nil
		} else {
			var projectPhaseID uint
			if err := json.Unmarshal(*input.ProjectPhaseID, &projectPhaseID); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_phase_id"})
				return
			}
			task.ProjectPhaseID = &projectPhaseID
		}
	}
	if input.StartTime != nil {
		task.StartTime = input.StartTime
	}
	if input.EndTime != nil {
		task.EndTime = input.EndTime
	}
	if input.Priority != nil {
		task.Priority = *input.Priority
	}
	if input.Urgency != nil {
		task.UrgencyCoefficient = *input.Urgency
	}
	if input.Complexity != nil {
		task.ComplexityCoefficient = *input.Complexity
	}
	if input.Category != nil {
		task.Category = *input.Category
	}

	err = h.service.UpdateTask(task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updatedTask, err := h.service.GetTask(task.ID)
	if err != nil {
		c.JSON(http.StatusOK, task)
		return
	}

	c.JSON(http.StatusOK, updatedTask)
}

// DeleteTask — удаление задачи
// @Summary Удалить задачу
// @Description Удаляет задачу по ID
// @Tags задачи
// @Param id path int true "ID задачи"
// @Success 204 "Задача удалена"
// @Failure 400 {object} gin.H "Некорректный формат ID"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(c *gin.Context) {
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

	err = h.service.DeleteTask(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// UpdateTaskStatus — обновление статуса задачи
// @Summary Обновить статус задачи
// @Description Изменяет статус задачи. Проверяет разрешённые переходы.
// @Tags задачи
// @Accept json
// @Produce json
// @Param id path int true "ID задачи"
// @Param request body struct{StatusID uint `json:"status_id"`} true "Новый статус"
// @Success 200 {object} models.Task
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 403 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /tasks/{id}/status [put]
func (h *TaskHandler) UpdateTaskStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	var input UpdateTaskStatusRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.StatusID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status_id is required"})
		return
	}

	updatedTask, err := h.service.UpdateTaskStatus(uint(id), input.StatusID)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		} else if errors.Is(err, ErrInvalidStatusTransition) {
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid status transition"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, updatedTask)
}

// GetTasksByProject — получение задач по проекту
// @Summary Получить задачи проекта
// @Description Возвращает список задач, принадлежащих указанному проекту
// @Tags задачи
// @Produce json
// @Param project_id query int true "ID проекта"
// @Success 200 {array} models.Task "Список задач проекта"
// @Failure 400 {object} gin.H "Некорректный формат project_id"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /tasks/by-project [get]
func (h *TaskHandler) GetTasksByProject(c *gin.Context) {
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

	tasks, err := h.service.ListTasksByProject(uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func (h *TaskHandler) GetTasksByCategory(c *gin.Context) {
	category := c.Query("category")
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category is required"})
		return
	}

	tasks, err := h.service.ListTasksByCategory(category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func (h *TaskHandler) GetTasksByPhase(c *gin.Context) {
	phaseIDStr := c.Query("project_phase_id")
	if phaseIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_phase_id is required"})
		return
	}

	phaseID, err := strconv.ParseUint(phaseIDStr, 10, 64)
	if err != nil || phaseID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_phase_id"})
		return
	}

	tasks, err := h.service.ListTasksByPhase(uint(phaseID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}
