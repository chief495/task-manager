package status

import "github.com/gin-gonic/gin"

func (h *StatusHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/statuses", h.CreateStatus)
	router.GET("/statuses/:id", h.GetStatus)
	router.GET("/statuses", h.ListStatuses)
	router.PUT("/statuses/:id", h.UpdateStatus)
	router.DELETE("/statuses/:id", h.DeleteStatus)
}
