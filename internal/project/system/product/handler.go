package product

import (
	"net/http"
	"strconv"

	"github.com/eduardnesterov/task-manager/internal/models"
	"github.com/gin-gonic/gin"
)

// ProductHandler — обработчик HTTP-запросов для продуктов
type ProductHandler struct {
	service *ProductService
}

// NewProductHandler — конструктор обработчика
func NewProductHandler(service *ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

// CreateProduct — создание продукта
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var input CreateProductRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product := &models.Product{
		Name:     input.Name,
		SystemID: input.SystemID,
		StageID:  input.StageID,
	}
	if err := h.service.CreateProduct(product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, product)
}

// GetProduct — получение продукта по ID
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	product, err := h.service.GetProduct(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}
	c.JSON(http.StatusOK, product)
}

// ListProducts — список всех продуктов
func (h *ProductHandler) ListProducts(c *gin.Context) {
	list, err := h.service.ListProducts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// ListProductsBySystem — список продуктов по system_id
func (h *ProductHandler) ListProductsBySystem(c *gin.Context) {
	systemIDStr := c.Query("system_id")
	sid, err := strconv.ParseUint(systemIDStr, 10, 64)
	if err != nil || sid == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid system_id"})
		return
	}
	list, err := h.service.ListProductsBySystem(uint(sid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// UpdateProduct — обновление продукта
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	var input UpdateProductRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.service.GetProduct(input.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}
	if input.Name != nil {
		product.Name = *input.Name
	}
	if input.SystemID != nil {
		product.SystemID = *input.SystemID
	}
	if input.StageID != nil {
		product.StageID = *input.StageID
	}
	if err := h.service.UpdateProduct(product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, product)
}

// DeleteProduct — удаление продукта
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.DeleteProduct(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
