package work_type

import (
	"net/http"
	"strconv"

	"github.com/eduardnesterov/task-manager/internal/models"
	"github.com/gin-gonic/gin"
)

// WorkTypeHandler — обработчик HTTP-запросов для типов работ
type WorkTypeHandler struct {
	service *WorkTypeService
}

func NewWorkTypeHandler(service *WorkTypeService) *WorkTypeHandler {
	return &WorkTypeHandler{service: service}
}

// CreateWorkType — создание типа работ
func (h *WorkTypeHandler) CreateWorkType(c *gin.Context) {
	var input CreateWorkTypeRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	wt := &models.WorkType{Name: input.Name, Description: input.Description}
	if err := h.service.CreateWorkType(wt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, wt)
}

// GetWorkType — получение типа работ по ID
func (h *WorkTypeHandler) GetWorkType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	wt, err := h.service.GetWorkType(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "work type not found"})
		return
	}
	c.JSON(http.StatusOK, wt)
}

// ListWorkTypes — список всех типов работ
func (h *WorkTypeHandler) ListWorkTypes(c *gin.Context) {
	list, err := h.service.ListWorkTypes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// SearchWorkTypes — поиск по имени
func (h *WorkTypeHandler) SearchWorkTypes(c *gin.Context) {
	query := c.Query("query")
	list, err := h.service.SearchWorkTypes(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// UpdateWorkType — обновление типа работ
func (h *WorkTypeHandler) UpdateWorkType(c *gin.Context) {
	var input UpdateWorkTypeRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	wt, err := h.service.GetWorkType(input.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "work type not found"})
		return
	}
	if input.Name != nil {
		wt.Name = *input.Name
	}
	if input.Description != nil {
		wt.Description = *input.Description
	}
	if err := h.service.UpdateWorkType(wt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, wt)
}

// DeleteWorkType — удаление типа работ
func (h *WorkTypeHandler) DeleteWorkType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.DeleteWorkType(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
