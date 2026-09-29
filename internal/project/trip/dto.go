package trip

type CreateTripRequest struct {
	Place     string `json:"place" binding:"required"`
	ProjectID uint   `json:"project_id" binding:"required"`
	PersonID  uint   `json:"person_id" binding:"required"`
	BeginDate string `json:"begin_date" binding:"required"` // YYYY-MM-DD
	EndDate   string `json:"end_date" binding:"required"`   // YYYY-MM-DD
}

type UpdateTripRequest struct {
	ID        uint    `json:"id" binding:"required"`
	Place     *string `json:"place"`
	ProjectID *uint   `json:"project_id"`
	PersonID  *uint   `json:"person_id"`
	BeginDate *string `json:"begin_date"` // YYYY-MM-DD
	EndDate   *string `json:"end_date"`   // YYYY-MM-DD
}
