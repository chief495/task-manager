package status

type CreateStatusRequest struct {
	Name        string `json:"name" binding:"required,max=50"`
	Description string `json:"description" binding:"max=200"`
	IsDefault   bool   `json:"is_default"`
	IsFinal     bool   `json:"is_final"`
}

type UpdateStatusRequest struct {
	ID          uint    `json:"id" binding:"required"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	IsDefault   *bool   `json:"is_default"`
}
