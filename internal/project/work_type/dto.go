package work_type

type CreateWorkTypeRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateWorkTypeRequest struct {
	ID          uint    `json:"id" binding:"required"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
}
