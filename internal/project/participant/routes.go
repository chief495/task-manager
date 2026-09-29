package participant

import "github.com/gin-gonic/gin"

func (h *ParticipantHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/participants", h.CreateParticipant)
	router.GET("/participants/:id", h.GetParticipant)
	router.GET("/participants", h.ListParticipants)
	router.GET("/participants/by-project", h.ListParticipantsByProject)
	router.PUT("/participants", h.UpdateParticipant)
	router.PUT("/participants/:id", h.UpdateParticipant)
	router.DELETE("/participants/:id", h.DeleteParticipant)
}
