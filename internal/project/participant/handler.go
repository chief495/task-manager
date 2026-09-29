package participant

import (
	"net/http"
	"strconv"

	"github.com/eduardnesterov/task-manager/internal/models"
	"github.com/gin-gonic/gin"
)

// ParticipantHandler — обработчик HTTP-запросов для участников проектов
type ParticipantHandler struct {
	service *ParticipantService
}

func NewParticipantHandler(service *ParticipantService) *ParticipantHandler {
	return &ParticipantHandler{service: service}
}

// CreateParticipant — добавление участника
func (h *ParticipantHandler) CreateParticipant(c *gin.Context) {
	var input CreateParticipantRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p := &models.Participant{
		ProjectID:      input.ProjectID,
		OrganizationID: input.OrganizationID,
		Role:           input.Role,
	}
	if err := h.service.CreateParticipant(p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

// GetParticipant — получение участника по ID
func (h *ParticipantHandler) GetParticipant(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	p, err := h.service.GetParticipant(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "participant not found"})
		return
	}
	c.JSON(http.StatusOK, p)
}

// ListParticipants — список всех участников
func (h *ParticipantHandler) ListParticipants(c *gin.Context) {
	list, err := h.service.ListParticipants()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// ListParticipantsByProject — участники по project_id
func (h *ParticipantHandler) ListParticipantsByProject(c *gin.Context) {
	projectIDStr := c.Query("project_id")
	pid, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil || pid == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}
	list, err := h.service.ListParticipantsByProject(uint(pid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// UpdateParticipant — обновление участника
func (h *ParticipantHandler) UpdateParticipant(c *gin.Context) {
	var input UpdateParticipantRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if pathID := c.Param("id"); pathID != "" {
		id, err := strconv.ParseUint(pathID, 10, 64)
		if err != nil || id == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		input.ID = uint(id)
	}
	if input.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}
	p, err := h.service.GetParticipant(input.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "participant not found"})
		return
	}
	if input.ProjectID != nil {
		p.ProjectID = *input.ProjectID
	}
	if input.OrganizationID != nil {
		p.OrganizationID = *input.OrganizationID
	}
	if input.Role != nil {
		p.Role = *input.Role
	}
	if err := h.service.UpdateParticipant(p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

// DeleteParticipant — удаление участника
func (h *ParticipantHandler) DeleteParticipant(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.DeleteParticipant(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
