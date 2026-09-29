package organization

import "github.com/gin-gonic/gin"

func (h *OrganizationHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/organizations", h.CreateOrganization)
	router.GET("/organizations/:id", h.GetOrganization)
	router.GET("/organizations", h.ListOrganizations)
	router.PUT("/organizations", h.UpdateOrganization)
	router.DELETE("/organizations/:id", h.DeleteOrganization)
}
