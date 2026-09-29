package task

import "github.com/gin-gonic/gin"

func (h *TaskHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/tasks", h.CreateTask)
	router.GET("/tasks/:id", h.GetTask)
	router.PUT("/tasks/:id", h.UpdateTask)
	router.DELETE("/tasks/:id", h.DeleteTask)
	router.GET("/tasks", h.ListTasks)
	router.GET("/tasks/by-project", h.GetTasksByProject)
	router.GET("/tasks/by-phase", h.GetTasksByPhase)
	router.GET("/tasks/by-category", h.GetTasksByCategory)
	router.PUT("/tasks/:id/status", h.UpdateTaskStatus)
}
