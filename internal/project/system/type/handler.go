package systemtype

import (
	"net/http"
	"strconv"

	"github.com/eduardnesterov/task-manager/internal/models"
	"github.com/gin-gonic/gin"
)

// SystemTypeHandler — обработчик HTTP-запросов для типов систем
type SystemTypeHandler struct {
	service *SystemTypeService
}

func NewSystemTypeHandler(service *SystemTypeService) *SystemTypeHandler {
	return &SystemTypeHandler{service: service}
}

// CreateSystemType — создание типа системы
func (h *SystemTypeHandler) CreateSystemType(c *gin.Context) {
	var input CreateSystemTypeRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	st := &models.SystemType{Name: input.Name, Description: input.Description}
	if err := h.service.CreateSystemType(st); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, st)
}

// GetSystemType — получение типа системы по ID
func (h *SystemTypeHandler) GetSystemType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	st, err := h.service.GetSystemType(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "system type not found"})
		return
	}
	c.JSON(http.StatusOK, st)
}

// ListSystemTypes — список всех типов систем
func (h *SystemTypeHandler) ListSystemTypes(c *gin.Context) {
	list, err := h.service.ListSystemTypes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// SearchSystemTypes — поиск по имени
func (h *SystemTypeHandler) SearchSystemTypes(c *gin.Context) {
	query := c.Query("query")
	list, err := h.service.SearchSystemTypes(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// UpdateSystemType — обновление типа системы
func (h *SystemTypeHandler) UpdateSystemType(c *gin.Context) {
	var input UpdateSystemTypeRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	st, err := h.service.GetSystemType(input.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "system type not found"})
		return
	}
	if input.Name != nil {
		st.Name = *input.Name
	}
	if input.Description != nil {
		st.Description = *input.Description
	}
	if err := h.service.UpdateSystemType(st); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, st)
}

// DeleteSystemType — удаление типа системы
func (h *SystemTypeHandler) DeleteSystemType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.DeleteSystemType(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
