package stage

import "github.com/gin-gonic/gin"

func (h *StageHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/stages", h.CreateStage)
	router.GET("/stages/:id", h.GetStage)
	router.GET("/stages", h.ListStages)
	router.PUT("/stages/:id", h.UpdateStage)
	router.DELETE("/stages/:id", h.DeleteStage)
}
