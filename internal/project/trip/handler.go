package trip

import (
	"net/http"
	"strconv"
	"time"

	"github.com/eduardnesterov/task-manager/internal/models"
	"github.com/gin-gonic/gin"
)

// TripHandler — обработчик HTTP-запросов для командировок
type TripHandler struct {
	service *TripService
}

func NewTripHandler(service *TripService) *TripHandler {
	return &TripHandler{service: service}
}

// CreateTrip — создание командировки
func (h *TripHandler) CreateTrip(c *gin.Context) {
	var input CreateTripRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	begin, err := time.Parse("2006-01-02", input.BeginDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid begin_date, expected YYYY-MM-DD"})
		return
	}
	end, err := time.Parse("2006-01-02", input.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date, expected YYYY-MM-DD"})
		return
	}
	trip := &models.Trip{
		Place:     input.Place,
		ProjectID: input.ProjectID,
		PersonID:  input.PersonID,
		BeginDate: begin,
		EndDate:   end,
	}
	if err := h.service.CreateTrip(trip); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, trip)
}

// GetTrip — получение командировки по ID
func (h *TripHandler) GetTrip(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	trip, err := h.service.GetTrip(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "trip not found"})
		return
	}
	c.JSON(http.StatusOK, trip)
}

// ListTrips — список всех командировок
func (h *TripHandler) ListTrips(c *gin.Context) {
	list, err := h.service.ListTrips()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// ListTripsByProject — командировки по project_id
func (h *TripHandler) ListTripsByProject(c *gin.Context) {
	projectIDStr := c.Query("project_id")
	pid, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil || pid == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}
	list, err := h.service.ListTripsByProject(uint(pid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// ListTripsByPerson — командировки по исполнителю
func (h *TripHandler) ListTripsByPerson(c *gin.Context) {
	personIDStr := c.Query("person_id")
	pid, err := strconv.ParseUint(personIDStr, 10, 64)
	if err != nil || pid == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid person_id"})
		return
	}
	list, err := h.service.ListTripsByPerson(uint(pid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// ListTripsActive — командировки, пересекающие период
func (h *TripHandler) ListTripsActive(c *gin.Context) {
	startStr := c.Query("start")
	endStr := c.Query("end")
	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start, expected YYYY-MM-DD"})
		return
	}
	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end, expected YYYY-MM-DD"})
		return
	}
	list, err := h.service.ListTripsActiveInPeriod(start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// UpdateTrip — обновление командировки
func (h *TripHandler) UpdateTrip(c *gin.Context) {
	var input UpdateTripRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	trip, err := h.service.GetTrip(input.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "trip not found"})
		return
	}
	if input.Place != nil {
		trip.Place = *input.Place
	}
	if input.ProjectID != nil {
		trip.ProjectID = *input.ProjectID
	}
	if input.PersonID != nil {
		trip.PersonID = *input.PersonID
	}
	if input.BeginDate != nil {
		if t, err := time.Parse("2006-01-02", *input.BeginDate); err == nil {
			trip.BeginDate = t
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid begin_date, expected YYYY-MM-DD"})
			return
		}
	}
	if input.EndDate != nil {
		if t, err := time.Parse("2006-01-02", *input.EndDate); err == nil {
			trip.EndDate = t
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date, expected YYYY-MM-DD"})
			return
		}
	}
	if err := h.service.UpdateTrip(trip); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, trip)
}

// DeleteTrip — удаление командировки
func (h *TripHandler) DeleteTrip(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.DeleteTrip(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
