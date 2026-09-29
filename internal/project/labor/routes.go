package labor

import "github.com/gin-gonic/gin"

func (h *LaborHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/labor/categories", h.CreateCategory)
	router.GET("/labor/categories", h.ListCategories)
	router.GET("/labor/categories/:id", h.GetCategory)
	router.PUT("/labor/categories/:id", h.UpdateCategory)
	router.DELETE("/labor/categories/:id", h.DeleteCategory)

	router.POST("/labor/norms", h.CreateNorm)
	router.GET("/labor/norms", h.ListNorms)
	router.GET("/labor/norms/:id", h.GetNorm)
	router.PUT("/labor/norms/:id", h.UpdateNorm)
	router.DELETE("/labor/norms/:id", h.DeleteNorm)

	router.POST("/labor/coefficients", h.CreateCoefficient)
	router.GET("/labor/coefficients", h.ListCoefficients)
	router.GET("/labor/coefficients/:id", h.GetCoefficient)
	router.PUT("/labor/coefficients/:id", h.UpdateCoefficient)
	router.DELETE("/labor/coefficients/:id", h.DeleteCoefficient)

	router.POST("/labor/calculate", h.Calculate)
	router.GET("/labor/calculations", h.ListCalculations)
	router.GET("/labor/calculations/:id", h.GetCalculation)
	router.DELETE("/labor/calculations/:id", h.DeleteCalculation)
}
