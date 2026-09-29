package work_type

import "github.com/gin-gonic/gin"

func (h *WorkTypeHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/work-types", h.CreateWorkType)
	router.GET("/work-types/:id", h.GetWorkType)
	router.GET("/work-types", h.ListWorkTypes)
	router.GET("/work-types/search", h.SearchWorkTypes)
	router.PUT("/work-types", h.UpdateWorkType)
	router.DELETE("/work-types/:id", h.DeleteWorkType)
}
