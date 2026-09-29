package auth

import "github.com/gin-gonic/gin"

func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/auth/login", h.Login)
	router.POST("/auth/register", h.Register)
	router.POST("/auth/reset-password", h.ResetPassword)
}
