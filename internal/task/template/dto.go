package template

type CreateRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Priority    int     `json:"priority"`
	Urgency     float32 `json:"urgency"`
	Complexity  float32 `json:"complexity"`
	ProjectID   *uint   `json:"project_id"`
}

type UpdateRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Priority    int     `json:"priority"`
	Urgency     float32 `json:"urgency"`
	Complexity  float32 `json:"complexity"`
	ProjectID   *uint   `json:"project_id"`
}
