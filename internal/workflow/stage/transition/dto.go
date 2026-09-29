package transition

type CreateStageTransitionRequest struct {
	FromStageID  uint     `json:"from_stage_id" binding:"required"`
	ToStageID    uint     `json:"to_stage_id" binding:"required"`
	Condition    string   `json:"condition" binding:"max=200"`
	AllowedRoles []string `json:"allowed_roles" binding:"required"`
}

type UpdateStageTransitionRequest struct {
	ID           uint      `json:"id" binding:"required"`
	Condition    *string   `json:"condition"`
	AllowedRoles *[]string `json:"allowed_roles"`
}
