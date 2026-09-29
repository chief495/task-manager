package transition

type CreateStatusTransitionRequest struct {
	FromStatusID uint     `json:"from_status_id" binding:"required"`
	ToStatusID   uint     `json:"to_status_id" binding:"required"`
	Condition    string   `json:"condition" binding:"max=200"`
	AllowedRoles []string `json:"allowed_roles" binding:"required"`
}

type UpdateStatusTransitionRequest struct {
	ID           uint      `json:"id" binding:"required"`
	Condition    *string   `json:"condition"`
	AllowedRoles *[]string `json:"allowed_roles"`
}
