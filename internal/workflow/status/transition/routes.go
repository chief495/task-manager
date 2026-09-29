package transition

import "github.com/gin-gonic/gin"

func (h *StatusTransitionHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/status-transitions", h.CreateStatusTransition)
	router.GET("/status-transitions/:id", h.GetStatusTransition)
	router.GET("/status-transitions", h.ListStatusTransitions)
	router.GET("/status-transitions/from", h.ListTransitionsByFromStatus)
	router.GET("/status-transitions/to", h.ListTransitionsByToStatus)
	router.PUT("/status-transitions/:id", h.UpdateStatusTransition)
	router.DELETE("/status-transitions/:id", h.DeleteStatusTransition)
}
