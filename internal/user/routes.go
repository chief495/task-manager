package user

import "github.com/gin-gonic/gin"

func (h *UserHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/users", h.CreateUser)
	router.GET("/users/:id", h.GetUser)
	router.PUT("/users/:id", h.UpdateUser)
	router.DELETE("/users/:id", h.DeleteUser)
	router.GET("/users", h.ListUsers)
	router.GET("/users/:id/tasks-stats", h.GetUserTasksStats)
	router.GET("/users/:id/social-rating-history", h.GetUserSocialRatingHistory)
	router.POST("/users/authenticate", h.AuthenticateUser)
	router.POST("/users/change-password", h.ChangePassword)
	router.POST("/users/:id/activate", h.ActivateUser)
	router.POST("/users/:id/deactivate", h.DeactivateUser)
}
