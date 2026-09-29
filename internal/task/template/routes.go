package template

import "github.com/gin-gonic/gin"

func (h *TaskTemplateHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/task-templates", h.List)
	router.GET("/task-templates/:id", h.Get)
	router.PUT("/task-templates/:id", h.Update)
	router.DELETE("/task-templates/:id", h.Delete)
	router.POST("/task-templates", h.Create)
}
