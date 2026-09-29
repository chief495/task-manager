package product

type CreateProductRequest struct {
	Name     string `json:"name" binding:"required"`
	SystemID uint   `json:"system_id" binding:"required"`
	StageID  uint   `json:"stage_id" binding:"required"`
}

type UpdateProductRequest struct {
	ID       uint    `json:"id" binding:"required"`
	Name     *string `json:"name"`
	SystemID *uint   `json:"system_id"`
	StageID  *uint   `json:"stage_id"`
}
