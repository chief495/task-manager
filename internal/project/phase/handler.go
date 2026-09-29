package phase

import (
	"net/http"
	"strconv"
	"time"

	"github.com/eduardnesterov/task-manager/internal/models"
	"github.com/gin-gonic/gin"
)

// ProjectPhaseHandler — обработчик HTTP-запросов для этапов проекта
type ProjectPhaseHandler struct {
	service *ProjectPhaseService
}

func NewProjectPhaseHandler(service *ProjectPhaseService) *ProjectPhaseHandler {
	return &ProjectPhaseHandler{service: service}
}

func (h *ProjectPhaseHandler) CreatePhase(c *gin.Context) {
	var input CreatePhaseRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	phase := &models.ProjectPhase{
		ProjectID:     input.ProjectID,
		Name:          input.Name,
		Description:   input.Description,
		StartDate:     input.StartDate,
		EndDate:       input.EndDate,
		Position:      input.Position,
		ResponsibleID: input.ResponsibleID,
	}
	if err := h.service.CreatePhase(phase); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, phase)
}

func (h *ProjectPhaseHandler) GetPhase(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	phase, err := h.service.GetPhase(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project phase not found"})
		return
	}
	c.JSON(http.StatusOK, phase)
}

func (h *ProjectPhaseHandler) ListPhases(c *gin.Context) {
	projectIDStr := c.Query("project_id")
	if projectIDStr != "" {
		projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
		if err != nil || projectID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
			return
		}
		list, err := h.service.ListPhasesByProject(uint(projectID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, list)
		return
	}

	list, err := h.service.ListPhases()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *ProjectPhaseHandler) UpdatePhase(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input UpdatePhaseRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	phase, err := h.service.GetPhase(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project phase not found"})
		return
	}
	if input.ProjectID != nil {
		phase.ProjectID = *input.ProjectID
	}
	if input.Name != nil {
		phase.Name = *input.Name
	}
	if input.Description != nil {
		phase.Description = *input.Description
	}
	if input.StartDate != nil {
		startDate, err := parseOptionalPhaseDate(*input.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date, expected YYYY-MM-DD or RFC3339"})
			return
		}
		phase.StartDate = startDate
	}
	if input.EndDate != nil {
		endDate, err := parseOptionalPhaseDate(*input.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date, expected YYYY-MM-DD or RFC3339"})
			return
		}
		phase.EndDate = endDate
	}
	if input.Position != nil {
		phase.Position = *input.Position
	}
	if input.ResponsibleID != nil {
		phase.ResponsibleID = *input.ResponsibleID
	}

	if err := h.service.UpdatePhase(phase); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, phase)
}

func parseOptionalPhaseDate(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return &t, nil
	}
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (h *ProjectPhaseHandler) DeletePhase(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.DeletePhase(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
