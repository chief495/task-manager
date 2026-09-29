package systemtype

import "github.com/gin-gonic/gin"

func (h *SystemTypeHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/system-types", h.CreateSystemType)
	router.GET("/system-types/:id", h.GetSystemType)
	router.GET("/system-types", h.ListSystemTypes)
	router.GET("/system-types/search", h.SearchSystemTypes)
	router.PUT("/system-types", h.UpdateSystemType)
	router.DELETE("/system-types/:id", h.DeleteSystemType)
}
