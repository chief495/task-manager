package system

import (
	"net/http"
	"strconv"

	"github.com/eduardnesterov/task-manager/internal/models"
	"github.com/gin-gonic/gin"
)

// SystemHandler — обработчик HTTP-запросов для систем
type SystemHandler struct {
	service *SystemService
}

// NewSystemHandler — конструктор обработчика
func NewSystemHandler(service *SystemService) *SystemHandler {
	return &SystemHandler{service: service}
}

// CreateSystem — создание системы
func (h *SystemHandler) CreateSystem(c *gin.Context) {
	var input CreateSystemRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	system := &models.System{
		Name:         input.Name,
		Description:  input.Description,
		ProjectID:    input.ProjectID,
		SystemTypeID: input.SystemTypeID,
	}
	if err := h.service.CreateSystem(system); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, system)
}

// GetSystem — получение системы по ID
func (h *SystemHandler) GetSystem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	system, err := h.service.GetSystem(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "system not found"})
		return
	}
	c.JSON(http.StatusOK, system)
}

// ListSystems — список всех систем
func (h *SystemHandler) ListSystems(c *gin.Context) {
	list, err := h.service.ListSystems()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// ListSystemsByProject — список систем по project_id
func (h *SystemHandler) ListSystemsByProject(c *gin.Context) {
	projectIDStr := c.Query("project_id")
	pid, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil || pid == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}
	list, err := h.service.ListSystemsByProject(uint(pid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// UpdateSystem — обновление системы
func (h *SystemHandler) UpdateSystem(c *gin.Context) {
	var input UpdateSystemRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	system, err := h.service.GetSystem(input.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "system not found"})
		return
	}
	if input.Name != nil {
		system.Name = *input.Name
	}
	if input.Description != nil {
		system.Description = *input.Description
	}
	if input.ProjectID != nil {
		system.ProjectID = *input.ProjectID
	}
	if input.SystemTypeID != nil {
		system.SystemTypeID = *input.SystemTypeID
	}
	if err := h.service.UpdateSystem(system); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, system)
}

// DeleteSystem — удаление системы
func (h *SystemHandler) DeleteSystem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.DeleteSystem(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
