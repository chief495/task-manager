package phase

import "github.com/gin-gonic/gin"

func (h *ProjectPhaseHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/project-phases", h.CreatePhase)
	router.GET("/project-phases/:id", h.GetPhase)
	router.GET("/project-phases", h.ListPhases)
	router.PUT("/project-phases/:id", h.UpdatePhase)
	router.DELETE("/project-phases/:id", h.DeletePhase)
}
