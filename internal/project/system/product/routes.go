package product

import "github.com/gin-gonic/gin"

func (h *ProductHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/products", h.CreateProduct)
	router.GET("/products/:id", h.GetProduct)
	router.GET("/products", h.ListProducts)
	router.GET("/products/by-system", h.ListProductsBySystem)
	router.PUT("/products", h.UpdateProduct)
	router.DELETE("/products/:id", h.DeleteProduct)
}
