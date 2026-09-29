package template

import (
	"net/http"
	"strconv"

	"github.com/eduardnesterov/task-manager/internal/models"
	"github.com/gin-gonic/gin"
)

type TaskTemplateHandler struct {
	service *TaskTemplateService
}

func NewTaskTemplateHandler(s *TaskTemplateService) *TaskTemplateHandler {
	return &TaskTemplateHandler{service: s}
}

func (h *TaskTemplateHandler) Create(c *gin.Context) {
	var input CreateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user is not authenticated"})
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user context"})
		return
	}

	urgency := input.Urgency
	if urgency == 0.0 {
		urgency = 1.0
	}
	complexity := input.Complexity
	if complexity == 0.0 {
		complexity = 1.0
	}

	template := &models.TaskTemplate{
		Name:        input.Name,
		Description: input.Description,
		Category:    input.Category,
		Priority:    input.Priority,
		Urgency:     urgency,
		Complexity:  complexity,
		ProjectID:   input.ProjectID,
		CreatedByID: userID,
	}

	if err := h.service.Create(template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, template)
}

func (h *TaskTemplateHandler) List(c *gin.Context) {
	projectIDStr := c.Query("project_id")

	var templates []*models.TaskTemplate
	var err error

	if projectIDStr != "" {
		projectID, parseErr := strconv.ParseUint(projectIDStr, 10, 64)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_ID"})
			return
		}
		templates, err = h.service.ListForProject(uint(projectID))
	} else {
		templates, err = h.service.ListAll()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if templates == nil {
		templates = []*models.TaskTemplate{}
	}

	c.JSON(http.StatusOK, templates)
}

func (h *TaskTemplateHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	template, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
	}

	c.JSON(http.StatusOK, template)
}

func (h *TaskTemplateHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var input UpdateRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template := &models.TaskTemplate{
		Name:        input.Name,
		Description: input.Description,
		Category:    input.Category,
		Priority:    input.Priority,
		Urgency:     input.Urgency,
		Complexity:  input.Complexity,
		ProjectID:   input.ProjectID,
	}
	template.ID = uint(id)

	if err := h.service.Update(template); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, template)
}

func (h *TaskTemplateHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.service.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "task template deleted"})
}
