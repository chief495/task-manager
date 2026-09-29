package trip

import "github.com/gin-gonic/gin"

func (h *TripHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/trips", h.CreateTrip)
	router.GET("/trips/:id", h.GetTrip)
	router.GET("/trips", h.ListTrips)
	router.GET("/trips/by-project", h.ListTripsByProject)
	router.GET("/trips/by-person", h.ListTripsByPerson)
	router.GET("/trips/active", h.ListTripsActive)
	router.PUT("/trips", h.UpdateTrip)
	router.DELETE("/trips/:id", h.DeleteTrip)
}
