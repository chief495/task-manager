package organization

type CreateOrganizationRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Address     string `json:"address"`
	Website     string `json:"website"`
	Notes       string `json:"notes"`
}

type UpdateOrganizationRequest struct {
	ID          uint    `json:"id" binding:"required"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Phone       *string `json:"phone"`
	Email       *string `json:"email"`
	Address     *string `json:"address"`
	Website     *string `json:"website"`
	Notes       *string `json:"notes"`
}
