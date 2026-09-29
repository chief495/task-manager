package stage

import (
	"net/http"
	"strconv"

	"github.com/eduardnesterov/task-manager/internal/models"
	"github.com/gin-gonic/gin"
)

// ProductStageHandler — обработчик HTTP-запросов для стадий продукта
type ProductStageHandler struct {
	service *ProductStageService
}

func NewProductStageHandler(service *ProductStageService) *ProductStageHandler {
	return &ProductStageHandler{service: service}
}

// CreateProductStage — создание стадии продукта
func (h *ProductStageHandler) CreateProductStage(c *gin.Context) {
	var input CreateProductStageRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ps := &models.ProductStage{Name: input.Name, Description: input.Description}
	if err := h.service.CreateProductStage(ps); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, ps)
}

// GetProductStage — получение стадии продукта по ID
func (h *ProductStageHandler) GetProductStage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	ps, err := h.service.GetProductStage(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product stage not found"})
		return
	}
	c.JSON(http.StatusOK, ps)
}

// ListProductStages — список всех стадий продукта
func (h *ProductStageHandler) ListProductStages(c *gin.Context) {
	list, err := h.service.ListProductStages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// SearchProductStages — поиск по имени
func (h *ProductStageHandler) SearchProductStages(c *gin.Context) {
	query := c.Query("query")
	list, err := h.service.SearchProductStages(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// UpdateProductStage — обновление стадии продукта
func (h *ProductStageHandler) UpdateProductStage(c *gin.Context) {
	var input UpdateProductStageRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ps, err := h.service.GetProductStage(input.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product stage not found"})
		return
	}
	if input.Name != nil {
		ps.Name = *input.Name
	}
	if input.Description != nil {
		ps.Description = *input.Description
	}
	if err := h.service.UpdateProductStage(ps); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ps)
}

// DeleteProductStage — удаление стадии продукта
func (h *ProductStageHandler) DeleteProductStage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.DeleteProductStage(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
