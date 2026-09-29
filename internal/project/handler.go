package project

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/eduardnesterov/task-manager/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ProjectHandler — обработчик HTTP‑запросов для проектов
type ProjectHandler struct {
	service *ProjectService
}

// NewProjectHandler — конструктор обработчика
func NewProjectHandler(service *ProjectService) *ProjectHandler {
	return &ProjectHandler{
		service: service,
	}
}

func (h *ProjectHandler) projectResponse(project *models.Project) (*ProjectResponse, error) {
	response := &ProjectResponse{Project: project}
	db := h.service.GetDB()
	if db == nil {
		return response, nil
	}

	if project.StageID != 0 {
		var stage models.Stage
		if err := db.First(&stage, project.StageID).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
		} else {
			response.Stage = &stage
		}
	}

	if project.ResponsibleID != 0 {
		var responsible models.User
		if err := db.First(&responsible, project.ResponsibleID).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
		} else {
			response.Responsible = &responsible
		}
	}

	if err := db.Where("project_id = ?", project.ID).Find(&response.Members).Error; err != nil {
		return nil, err
	}
	if err := db.Where("project_id = ?", project.ID).Find(&response.AccessUsers).Error; err != nil {
		return nil, err
	}

	return response, nil
}

func (h *ProjectHandler) projectResponses(projects []*models.Project) ([]*ProjectResponse, error) {
	response := make([]*ProjectResponse, 0, len(projects))
	for _, project := range projects {
		item, err := h.projectResponse(project)
		if err != nil {
			return nil, err
		}
		response = append(response, item)
	}
	return response, nil
}

// CreateProject — создание нового проекта
// @Summary Создать проект
// @Description Создаёт новый проект с указанными параметрами
// @Tags проекты
// @Accept json
// @Produce json
// @Param project body ProjectCreateRequest true "Данные проекта"
// @Success 201 {object} models.Project "Созданный проект"
// @Failure 400 {object} gin.H "Ошибка валидации входных данных"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /projects [post]
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var input ProjectCreateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	role, _ := c.Get("role")
	if role != "admin" && (input.IsHidden || len(input.AccessUserIDs) > 0) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admin can configure hidden projects"})
		return
	}

	var startDate, deadlineDate, conclusionDate, endDate time.Time
	var err error
	if input.StartDate != "" {
		startDate, err = time.Parse("2006-01-02", input.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date, expected YYYY-MM-DD"})
			return
		}
	}
	if input.DeadlineDate != "" {
		deadlineDate, err = time.Parse("2006-01-02", input.DeadlineDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid deadline_date, expected YYYY-MM-DD"})
			return
		}
	}
	if input.ConclusionDate != "" {
		conclusionDate, err = time.Parse("2006-01-02", input.ConclusionDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conclusion_date, expected YYYY-MM-DD"})
			return
		}
	}
	if input.EndDate != "" {
		endDate, err = time.Parse("2006-01-02", input.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date, expected YYYY-MM-DD"})
			return
		}
	}

	project := &models.Project{
		Name:          input.Name,
		Description:   input.Description,
		StageID:       input.Stage,
		ResponsibleID: input.Responsible,
		Address:       input.Address,
		StartDate:     startDate,
		DeadlineDate:  deadlineDate,
		Cost: func() float32 {
			if input.Cost != nil {
				return *input.Cost
			}
			return 0
		}(),
		Comments:       input.Comments,
		Assessment:     input.Assessment,
		ConclusionDate: conclusionDate,
		EndDate:        endDate,
		IsHidden:       input.IsHidden,
	}

	err = h.service.CreateProject(project)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err = h.service.ReplaceProjectAccess(project.ID, input.AccessUserIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response, err := h.projectResponse(project)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, response)
}

// GetTasksByProject — получение задач по проекту
// @Summary Получить задачи по проекту
// @Description Возвращает список задач по заданным ID проектов. Требуется аутентификация.
// @Tags проекты
// @Produce json
// @Param project_ids query string true "Список ID проектов через запятую"
// @Success 200 {array} models.Task "Список задач"
// @Failure 400 {object} gin.H "Некорректный формат ID"
// @Failure 401 {object} gin.H "Требуется аутентификация"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /projects/tasks [get]
func (h *ProjectHandler) GetTasksByProject(c *gin.Context) {
	projectIDsStr := c.Query("project_ids")
	if projectIDsStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "param 'project_ids' is required"})
		return
	}

	var projectIDs []uint
	userID := c.MustGet("user_id").(uint)
	role := c.MustGet("role").(string)
	for _, idStr := range strings.Split(projectIDsStr, ",") {
		id, err := strconv.ParseUint(strings.TrimSpace(idStr), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID: " + idStr})
			return
		}
		canAccess, err := h.service.UserCanAccessProject(uint(id), userID, role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if canAccess {
			projectIDs = append(projectIDs, uint(id))
		}
	}

	tasks, err := h.service.GetTasksByProjectIDs(projectIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// GetProject — получение проекта по ID
// @Summary Получить проект по ID
// @Description Возвращает проект с указанным ID
// @Tags проекты
// @Produce json
// @Param id path int true "ID проекта"
// @Success 200 {object} models.Project "Проект"
// @Failure 400 {object} gin.H "Некорректный формат ID"
// @Failure 404 {object} gin.H "Проект не найден"
// @Router /projects/{id} [get]
func (h *ProjectHandler) GetProject(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
		return
	}

	project, err := h.service.GetProject(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	if userIDValue, ok := c.Get("user_id"); ok {
		roleValue, _ := c.Get("role")
		userID, _ := userIDValue.(uint)
		role, _ := roleValue.(string)
		canAccess, err := h.service.UserCanAccessProject(project.ID, userID, role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !canAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
	}

	response, err := h.projectResponse(project)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// ListProjects — получение списка всех проектов
// @Summary Получить список всех проектов
// @Description Возвращает все проекты в системе
// @Tags проекты
// @Produce json
// @Success 200 {array} models.Project "Список проектов"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /projects [get]
func (h *ProjectHandler) ListProjects(c *gin.Context) {
	var (
		projects []*models.Project
		err      error
	)
	if userIDValue, ok := c.Get("user_id"); ok {
		roleValue, _ := c.Get("role")
		userID, _ := userIDValue.(uint)
		role, _ := roleValue.(string)
		projects, err = h.service.ListProjectsForUser(userID, role)
	} else {
		projects, err = h.service.ListProjects()
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response, err := h.projectResponses(projects)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateProject — обновление проекта
// @Summary Обновить проект
// @Description Обновляет проект по ID (частичное обновление)
// @Tags проекты
// @Accept json
// @Produce json
// @Param project body struct{ID uint `json:"id"` Name string `json:"name"` Description string `json:"description"` StageID uint `json:"stage_id"` ResponsibleID uint `json:"responsible_id"`} true "Данные для обновления"
// @Success 200 {object} models.Project "Обновлённый проект"
// @Failure 400 {object} gin.H "Ошибка валидации входных данных"
// @Failure 404 {object} gin.H "Проект не найден"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /projects [put]
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	// ID из URL path
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}

	var input UpdateProjectRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	role, _ := c.Get("role")
	if role != "admin" && (input.IsHidden != nil || input.AccessUserIDs != nil) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admin can configure hidden projects"})
		return
	}

	// Ищем проект по ID из URL
	project, err := h.service.GetProject(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	// Обновление полей (только переданных)
	if input.Name != nil {
		project.Name = *input.Name
	}
	if input.Description != nil {
		project.Description = *input.Description
	}
	if input.StageID != nil {
		project.StageID = *input.StageID
	}
	if input.ResponsibleID != nil {
		project.ResponsibleID = *input.ResponsibleID
	}
	if input.Address != nil {
		project.Address = *input.Address
	}
	if input.StartDate != nil {
		if *input.StartDate == "" {
			project.StartDate = time.Time{}
		} else {
			if t, err := time.Parse("2006-01-02", *input.StartDate); err == nil {
				project.StartDate = t
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date, expected YYYY-MM-DD"})
				return
			}
		}
	}
	if input.DeadlineDate != nil {
		if *input.DeadlineDate == "" {
			project.DeadlineDate = time.Time{}
		} else {
			if t, err := time.Parse("2006-01-02", *input.DeadlineDate); err == nil {
				project.DeadlineDate = t
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid deadline_date, expected YYYY-MM-DD"})
				return
			}
		}
	}
	if input.Cost != nil {
		project.Cost = *input.Cost
	}
	if input.Comments != nil {
		project.Comments = *input.Comments
	}
	if input.Assessment != nil {
		project.Assessment = *input.Assessment
	}
	if input.ConclusionDate != nil {
		if *input.ConclusionDate == "" {
			project.ConclusionDate = time.Time{}
		} else {
			if t, err := time.Parse("2006-01-02", *input.ConclusionDate); err == nil {
				project.ConclusionDate = t
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conclusion_date, expected YYYY-MM-DD"})
				return
			}
		}
	}
	if input.EndDate != nil {
		if *input.EndDate == "" {
			project.EndDate = time.Time{}
		} else {
			if t, err := time.Parse("2006-01-02", *input.EndDate); err == nil {
				project.EndDate = t
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date, expected YYYY-MM-DD"})
				return
			}
		}
	}
	if input.IsHidden != nil {
		project.IsHidden = *input.IsHidden
	}

	log.Println(project)
	err = h.service.UpdateProject(project)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if input.AccessUserIDs != nil {
		if err = h.service.ReplaceProjectAccess(project.ID, input.AccessUserIDs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		project, _ = h.service.GetProject(project.ID)
	}

	response, err := h.projectResponse(project)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// DeleteProject — удаление проекта
// @Summary Удалить проект
// @Description Удаляет проект по ID
// @Tags проекты
// @Param id path int true "ID проекта"
// @Success 204 "Проект удалён"
// @Failure 400 {object} gin.H "Некорректный формат ID"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /projects/{id} [delete]
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
		return
	}

	err = h.service.DeleteProject(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *ProjectHandler) CreateStakeholder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	var input CreateStakeholderRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stakeholder := &models.Stakeholder{
		ProjectID:      uint(id),
		OrganizationID: input.OrganizationID,
		FullName:       input.FullName,
		Position:       input.Position,
		Phone:          input.Phone,
		Email:          input.Email,
		Notes:          input.Notes,
	}
	if err := h.service.CreateStakeholder(stakeholder); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, stakeholder)
}

func (h *ProjectHandler) UpdateStakeholder(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || projectID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	stakeholderID, err := strconv.ParseUint(c.Param("stakeholder_id"), 10, 64)
	if err != nil || stakeholderID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid stakeholder ID"})
		return
	}

	var input UpdateStakeholderRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var stakeholder models.Stakeholder
	if err := h.service.GetDB().First(&stakeholder, uint(stakeholderID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "stakeholder not found"})
		return
	}
	if stakeholder.ProjectID != uint(projectID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "stakeholder not found"})
		return
	}

	stakeholder.OrganizationID = input.OrganizationID
	stakeholder.FullName = input.FullName
	stakeholder.Position = input.Position
	stakeholder.Phone = input.Phone
	stakeholder.Email = input.Email
	stakeholder.Notes = input.Notes

	if err := h.service.UpdateStakeholder(&stakeholder); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stakeholder)
}

func (h *ProjectHandler) AddMember(c *gin.Context) {
	var input AddProjectMemberRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	member := &models.ProjectMember{
		ProjectID: input.ProjectID,
		UserID:    input.UserID,
		Role:      input.Role,
	}
	if err := h.service.AddMember(member); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, member)
}

func (h *ProjectHandler) ListMembers(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Query("project_id"), 10, 64)
	if err != nil || projectID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	members, err := h.service.ListMembersByProject(uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, members)
}

func (h *ProjectHandler) UpdateMember(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var input UpdateProjectMemberRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	member, err := h.service.GetMember(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project member not found"})
		return
	}
	if input.ProjectID != nil {
		member.ProjectID = *input.ProjectID
	}
	if input.UserID != nil {
		member.UserID = *input.UserID
	}
	if input.Role != nil {
		member.Role = *input.Role
	}
	if err := h.service.UpdateMember(member); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, member)
}

func (h *ProjectHandler) DeleteMember(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.service.DeleteMember(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *ProjectHandler) ListMessages(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || projectID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	role := c.MustGet("role").(string)
	canAccess, err := h.service.UserCanAccessProject(uint(projectID), userID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !canAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	messages, err := h.service.ListMessages(uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, messages)
}

func (h *ProjectHandler) CreateMessage(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || projectID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	role := c.MustGet("role").(string)
	canAccess, err := h.service.UserCanAccessProject(uint(projectID), userID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !canAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var input CreateProjectMessageRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message := &models.ProjectMessage{
		ProjectID: uint(projectID),
		UserID:    userID,
		ReplyToID: input.ReplyToID,
		Content:   input.Content,
	}
	if err := h.service.CreateMessage(message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	created, err := h.service.GetMessage(message.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *ProjectHandler) UpdateMessage(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || projectID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}
	messageID, err := strconv.ParseUint(c.Param("message_id"), 10, 64)
	if err != nil || messageID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message ID"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	role := c.MustGet("role").(string)
	canAccess, err := h.service.UserCanAccessProject(uint(projectID), userID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !canAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	message, err := h.service.GetMessage(uint(messageID))
	if err != nil || message.ProjectID != uint(projectID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}
	if role != "admin" && message.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only message author can edit message"})
		return
	}

	var input UpdateProjectMessageRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	message.Content = input.Content
	if err := h.service.UpdateMessage(message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.service.GetMessage(message.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *ProjectHandler) DeleteMessage(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || projectID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}
	messageID, err := strconv.ParseUint(c.Param("message_id"), 10, 64)
	if err != nil || messageID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message ID"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	role := c.MustGet("role").(string)
	canAccess, err := h.service.UserCanAccessProject(uint(projectID), userID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !canAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	message, err := h.service.GetMessage(uint(messageID))
	if err != nil || message.ProjectID != uint(projectID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}
	if role != "admin" && message.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only message author can delete message"})
		return
	}

	if err := h.service.DeleteMessage(message.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// GetProjectsByUser — получение проектов пользователя
// @Summary Получить проекты пользователя
// @Description Возвращает список проектов, назначенных указанному пользователю
// @Tags проекты
// @Produce json
// @Param user_id query int true "ID пользователя"
// @Success 200 {array} models.Project "Список проектов пользователя"
// @Failure 400 {object} gin.H "Некорректный формат user_id"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /projects/by-user [get]
func (h *ProjectHandler) GetProjectsByUser(c *gin.Context) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id format"})
		return
	}

	projects, err := h.service.ListProjectsByResponsible(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	projectPointers := make([]*models.Project, len(projects))
	for i := range projects {
		projectPointers[i] = &projects[i]
	}
	response, err := h.projectResponses(projectPointers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
