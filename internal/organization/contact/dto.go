package contact

type CreateContactRequest struct {
	OrganizationID uint   `json:"organization_id" binding:"required"`
	FullName       string `json:"full_name" binding:"required"`
	Position       string `json:"position"`
	Phone          string `json:"phone"`
	Email          string `json:"email"`
	Notes          string `json:"notes"`
}

type UpdateContactRequest struct {
	OrganizationID uint   `json:"organization_id" binding:"required"`
	FullName       string `json:"full_name" binding:"required"`
	Position       string `json:"position"`
	Phone          string `json:"phone"`
	Email          string `json:"email"`
	Notes          string `json:"notes"`
}
