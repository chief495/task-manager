package contact

import (
	"net/http"
	"strconv"

	"github.com/eduardnesterov/task-manager/internal/models"
	"github.com/gin-gonic/gin"
)

// OrganizationContactHandler — обработчик HTTP-запросов для контактов компаний
type OrganizationContactHandler struct {
	service *OrganizationContactService
}

func NewOrganizationContactHandler(service *OrganizationContactService) *OrganizationContactHandler {
	return &OrganizationContactHandler{service: service}
}

func (h *OrganizationContactHandler) CreateContact(c *gin.Context) {
	var input CreateContactRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	contact := &models.OrganizationContact{
		OrganizationID: input.OrganizationID,
		FullName:       input.FullName,
		Position:       input.Position,
		Phone:          input.Phone,
		Email:          input.Email,
		Notes:          input.Notes,
	}
	if err := h.service.CreateContact(contact); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, contact)
}

func (h *OrganizationContactHandler) GetContact(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	contact, err := h.service.GetContact(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "contact not found"})
		return
	}
	c.JSON(http.StatusOK, contact)
}

func (h *OrganizationContactHandler) ListContacts(c *gin.Context) {
	var organizationID uint
	if raw := c.Query("organization_id"); raw != "" {
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization_id"})
			return
		}
		organizationID = uint(id)
	}

	contacts, err := h.service.ListContacts(organizationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, contacts)
}

func (h *OrganizationContactHandler) UpdateContact(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var input UpdateContactRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	contact, err := h.service.GetContact(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "contact not found"})
		return
	}

	contact.OrganizationID = input.OrganizationID
	contact.FullName = input.FullName
	contact.Position = input.Position
	contact.Phone = input.Phone
	contact.Email = input.Email
	contact.Notes = input.Notes

	if err := h.service.UpdateContact(contact); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, contact)
}

func (h *OrganizationContactHandler) DeleteContact(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.DeleteContact(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
