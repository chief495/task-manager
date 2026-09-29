package phase

import "time"

type CreatePhaseRequest struct {
	ProjectID     uint       `json:"project_id" binding:"required"`
	Name          string     `json:"name" binding:"required"`
	Description   string     `json:"description"`
	StartDate     *time.Time `json:"start_date"`
	EndDate       *time.Time `json:"end_date"`
	Position      int        `json:"position"`
	ResponsibleID uint       `json:"responsible_id" binding:"required"`
}

type UpdatePhaseRequest struct {
	ProjectID     *uint   `json:"project_id"`
	Name          *string `json:"name"`
	Description   *string `json:"description"`
	StartDate     *string `json:"start_date"`
	EndDate       *string `json:"end_date"`
	Position      *int    `json:"position"`
	ResponsibleID *uint   `json:"responsible_id"`
}
