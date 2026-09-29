package system

import "github.com/gin-gonic/gin"

func (h *SystemHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/systems", h.CreateSystem)
	router.GET("/systems/:id", h.GetSystem)
	router.GET("/systems", h.ListSystems)
	router.GET("/systems/by-project", h.ListSystemsByProject)
	router.PUT("/systems", h.UpdateSystem)
	router.DELETE("/systems/:id", h.DeleteSystem)
}
