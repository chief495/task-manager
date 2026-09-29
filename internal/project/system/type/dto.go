package systemtype

type CreateSystemTypeRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateSystemTypeRequest struct {
	ID          uint    `json:"id" binding:"required"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
}
