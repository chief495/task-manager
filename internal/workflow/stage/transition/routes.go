package transition

import "github.com/gin-gonic/gin"

func (h *StageTransitionHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/stage-transitions", h.CreateStageTransition)
	router.GET("/stage-transitions/:id", h.GetStageTransition)
	router.GET("/stage-transitions", h.ListStageTransitions)
	router.GET("/stage-transitions/from", h.ListTransitionsByFromStage)
	router.GET("/stage-transitions/to", h.ListTransitionsByToStage)
	router.PUT("/stage-transitions/:id", h.UpdateStageTransition)
	router.DELETE("/stage-transitions/:id", h.DeleteStageTransition)
}
