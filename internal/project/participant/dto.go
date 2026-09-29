package participant

type CreateParticipantRequest struct {
	ProjectID      uint   `json:"project_id" binding:"required"`
	OrganizationID uint   `json:"organization_id" binding:"required"`
	Role           string `json:"role" binding:"required"`
}

type UpdateParticipantRequest struct {
	ID             uint    `json:"id"`
	ProjectID      *uint   `json:"project_id"`
	OrganizationID *uint   `json:"organization_id"`
	Role           *string `json:"role"`
}
