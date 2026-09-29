package report

import "time"

type CreateReportRequest struct {
	UserID    uint      `json:"user_id" binding:"required"`
	ProjectID uint      `json:"project_id"`
	Text      string    `json:"text" binding:"required,max=5000"`
	Overwork  int       `json:"overwork"`
	CreatedAt time.Time `json:"created_at"`
}

type UpdateReportRequest struct {
	UserID    *uint      `json:"user_id"`
	ProjectID *uint      `json:"project_id"` // Указатель на указатель для отличия nil и нуля
	Text      *string    `json:"text"`
	Overwork  *int       `json:"overwork"`
	CreatedAt *time.Time `json:"created_at"`
}
