package contact

import "github.com/gin-gonic/gin"

func (h *OrganizationContactHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/organization-contacts", h.CreateContact)
	router.GET("/organization-contacts/:id", h.GetContact)
	router.GET("/organization-contacts", h.ListContacts)
	router.PUT("/organization-contacts/:id", h.UpdateContact)
	router.DELETE("/organization-contacts/:id", h.DeleteContact)
}
