package organization

import (
	"net/http"
	"strconv"

	"github.com/eduardnesterov/task-manager/internal/models"
	"github.com/gin-gonic/gin"
)

// OrganizationHandler — обработчик HTTP-запросов для организаций
type OrganizationHandler struct {
	service *OrganizationService
}

func NewOrganizationHandler(service *OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{service: service}
}

// CreateOrganization — создание организации
func (h *OrganizationHandler) CreateOrganization(c *gin.Context) {
	var input CreateOrganizationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	org := &models.Organization{
		Name:        input.Name,
		Description: input.Description,
		Phone:       input.Phone,
		Email:       input.Email,
		Address:     input.Address,
		Website:     input.Website,
		Notes:       input.Notes,
	}
	if err := h.service.CreateOrganization(org); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, org)
}

// GetOrganization — получение организации по ID
func (h *OrganizationHandler) GetOrganization(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	org, err := h.service.GetOrganization(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		return
	}
	c.JSON(http.StatusOK, org)
}

// ListOrganizations — список всех организаций
func (h *OrganizationHandler) ListOrganizations(c *gin.Context) {
	list, err := h.service.ListOrganizations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// UpdateOrganization — обновление организации
func (h *OrganizationHandler) UpdateOrganization(c *gin.Context) {
	var input UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	org, err := h.service.GetOrganization(input.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		return
	}
	if input.Name != nil {
		org.Name = *input.Name
	}
	if input.Description != nil {
		org.Description = *input.Description
	}
	if input.Phone != nil {
		org.Phone = *input.Phone
	}
	if input.Email != nil {
		org.Email = *input.Email
	}
	if input.Address != nil {
		org.Address = *input.Address
	}
	if input.Website != nil {
		org.Website = *input.Website
	}
	if input.Notes != nil {
		org.Notes = *input.Notes
	}
	if err := h.service.UpdateOrganization(org); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, org)
}

// DeleteOrganization — удаление организации
func (h *OrganizationHandler) DeleteOrganization(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.DeleteOrganization(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
