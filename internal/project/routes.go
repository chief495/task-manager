package project

import "github.com/gin-gonic/gin"

func (h *ProjectHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/projects", h.CreateProject)
	router.GET("/projects/:id", h.GetProject)
	router.PUT("/projects/:id", h.UpdateProject)
	router.DELETE("/projects/:id", h.DeleteProject)
	router.POST("/projects/:id/stakeholders", h.CreateStakeholder)
	router.PUT("/projects/:id/stakeholders/:stakeholder_id", h.UpdateStakeholder)
	router.POST("/project-members", h.AddMember)
	router.GET("/project-members", h.ListMembers)
	router.PUT("/project-members/:id", h.UpdateMember)
	router.DELETE("/project-members/:id", h.DeleteMember)
	router.GET("/projects/:id/messages", h.ListMessages)
	router.POST("/projects/:id/messages", h.CreateMessage)
	router.PUT("/projects/:id/messages/:message_id", h.UpdateMessage)
	router.DELETE("/projects/:id/messages/:message_id", h.DeleteMessage)
	router.GET("/projects", h.ListProjects)
	router.GET("/projects/by-user", h.GetProjectsByUser)
	router.GET("/projects/tasks", h.GetTasksByProject)
}
