package labor

import (
	"net/http"
	"strconv"

	"github.com/eduardnesterov/task-manager/internal/models"
	"github.com/gin-gonic/gin"
)

type LaborHandler struct {
	service *LaborService
}

func NewLaborHandler(s *LaborService) *LaborHandler {
	return &LaborHandler{service: s}
}

// ─── Категории ───────────────────────────────────────────────────────────────

func (h *LaborHandler) CreateCategory(c *gin.Context) {
	var input CreateCategoryRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cat := &models.LaborCategory{
		Name:        input.Name,
		Code:        input.Code,
		Description: input.Description,
	}
	if err := h.service.CreateCategory(cat); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, cat)
}

func (h *LaborHandler) ListCategories(c *gin.Context) {
	list, err := h.service.ListCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if list == nil {
		list = []models.LaborCategory{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *LaborHandler) GetCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	cat, err := h.service.GetCategory(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}
	c.JSON(http.StatusOK, cat)
}

func (h *LaborHandler) UpdateCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	cat, err := h.service.GetCategory(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}
	var input UpdateCategoryRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.Name != nil {
		cat.Name = *input.Name
	}
	if input.Code != nil {
		cat.Code = *input.Code
	}
	if input.Description != nil {
		cat.Description = *input.Description
	}
	if err := h.service.UpdateCategory(cat); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cat)
}

func (h *LaborHandler) DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.DeleteCategory(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// ─── Нормы ───────────────────────────────────────────────────────────────────

func (h *LaborHandler) CreateNorm(c *gin.Context) {
	var input CreateNormRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	norm := &models.LaborNorm{
		CategoryID:  input.CategoryID,
		Code:        input.Code,
		Name:        input.Name,
		Unit:        input.Unit,
		LaborHours:  input.LaborHours,
		Description: input.Description,
		Source:      input.Source,
	}
	if norm.Unit == "" {
		norm.Unit = "чел.-ч"
	}
	if norm.Source == "" {
		norm.Source = "Пользовательская"
	}
	if err := h.service.CreateNorm(norm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, norm)
}

func (h *LaborHandler) ListNorms(c *gin.Context) {
	catID, _ := strconv.ParseUint(c.Query("category_id"), 10, 64)
	list, err := h.service.ListNorms(uint(catID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if list == nil {
		list = []models.LaborNorm{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *LaborHandler) GetNorm(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	norm, err := h.service.GetNorm(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "norm not found"})
		return
	}
	c.JSON(http.StatusOK, norm)
}

func (h *LaborHandler) UpdateNorm(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	norm, err := h.service.GetNorm(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "norm not found"})
		return
	}
	var input UpdateNormRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.CategoryID != nil {
		norm.CategoryID = *input.CategoryID
	}
	if input.Code != nil {
		norm.Code = *input.Code
	}
	if input.Name != nil {
		norm.Name = *input.Name
	}
	if input.Unit != nil {
		norm.Unit = *input.Unit
	}
	if input.LaborHours != nil {
		norm.LaborHours = *input.LaborHours
	}
	if input.Description != nil {
		norm.Description = *input.Description
	}
	if input.Source != nil {
		norm.Source = *input.Source
	}
	if err := h.service.UpdateNorm(norm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, norm)
}

func (h *LaborHandler) DeleteNorm(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.DeleteNorm(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// ─── Коэффициенты ────────────────────────────────────────────────────────────

func (h *LaborHandler) CreateCoefficient(c *gin.Context) {
	var input CreateCoefficientRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	coeff := &models.LaborCoefficient{
		Name:        input.Name,
		Code:        input.Code,
		Value:       input.Value,
		CategoryID:  input.CategoryID,
		Description: input.Description,
	}
	if err := h.service.CreateCoefficient(coeff); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, coeff)
}

func (h *LaborHandler) ListCoefficients(c *gin.Context) {
	catID, _ := strconv.ParseUint(c.Query("category_id"), 10, 64)
	list, err := h.service.ListCoefficients(uint(catID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if list == nil {
		list = []models.LaborCoefficient{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *LaborHandler) GetCoefficient(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	coeff, err := h.service.GetCoefficient(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "coefficient not found"})
		return
	}
	c.JSON(http.StatusOK, coeff)
}

func (h *LaborHandler) UpdateCoefficient(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	coeff, err := h.service.GetCoefficient(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "coefficient not found"})
		return
	}
	var input UpdateCoefficientRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.Name != nil {
		coeff.Name = *input.Name
	}
	if input.Code != nil {
		coeff.Code = *input.Code
	}
	if input.Value != nil {
		coeff.Value = *input.Value
	}
	if input.CategoryID != nil {
		coeff.CategoryID = input.CategoryID
	}
	if input.Description != nil {
		coeff.Description = *input.Description
	}
	if err := h.service.UpdateCoefficient(coeff); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, coeff)
}

func (h *LaborHandler) DeleteCoefficient(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.DeleteCoefficient(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// ─── Расчёт ──────────────────────────────────────────────────────────────────

func (h *LaborHandler) Calculate(c *gin.Context) {
	var req models.LaborCalculationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "список позиций пуст"})
		return
	}

	// Получаем ID пользователя из контекста
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	userID, ok := userIDRaw.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id in context"})
		return
	}

	result, err := h.service.Calculate(&req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// ─── Сохранённые расчёты ─────────────────────────────────────────────────────

// ListCalculations — GET /labor/calculations?project_id=N
func (h *LaborHandler) ListCalculations(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Query("project_id"), 10, 64)
	if err != nil || projectID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
		return
	}

	list, err := h.service.ListCalculationsByProject(uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if list == nil {
		list = []models.LaborCalculationSummary{}
	}
	c.JSON(http.StatusOK, list)
}

// GetCalculation — GET /labor/calculations/:id
func (h *LaborHandler) GetCalculation(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	result, err := h.service.GetCalculation(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "calculation not found"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// DeleteCalculation — DELETE /labor/calculations/:id
func (h *LaborHandler) DeleteCalculation(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.DeleteCalculation(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
