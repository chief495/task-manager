package stage

type CreateStageRequest struct {
	Name        string `json:"name" binding:"required,max=100"`
	Description string `json:"description"`
	Position    int    `json:"position" binding:"required,min=0"`
}

type UpdateStageRequest struct {
	ID          uint    `json:"id" binding:"required"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Position    *int    `json:"position"`
}
