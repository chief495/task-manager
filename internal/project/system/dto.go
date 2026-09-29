package system

type CreateSystemRequest struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	ProjectID    uint   `json:"project_id" binding:"required"`
	SystemTypeID uint   `json:"system_type_id" binding:"required"`
}

type UpdateSystemRequest struct {
	ID           uint    `json:"id" binding:"required"`
	Name         *string `json:"name"`
	Description  *string `json:"description"`
	ProjectID    *uint   `json:"project_id"`
	SystemTypeID *uint   `json:"system_type_id"`
}
