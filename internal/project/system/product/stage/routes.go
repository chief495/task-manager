package stage

import "github.com/gin-gonic/gin"

func (h *ProductStageHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/product-stages", h.CreateProductStage)
	router.GET("/product-stages/:id", h.GetProductStage)
	router.GET("/product-stages", h.ListProductStages)
	router.GET("/product-stages/search", h.SearchProductStages)
	router.PUT("/product-stages", h.UpdateProductStage)
	router.DELETE("/product-stages/:id", h.DeleteProductStage)
}
