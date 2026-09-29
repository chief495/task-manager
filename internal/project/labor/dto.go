package labor

type CreateCategoryRequest struct {
	Name        string `json:"name"        binding:"required"`
	Code        string `json:"code"        binding:"required"`
	Description string `json:"description"`
}

type UpdateCategoryRequest struct {
	Name        *string `json:"name"`
	Code        *string `json:"code"`
	Description *string `json:"description"`
}

type CreateNormRequest struct {
	CategoryID  uint    `json:"category_id"  binding:"required"`
	Code        string  `json:"code"         binding:"required"`
	Name        string  `json:"name"         binding:"required"`
	Unit        string  `json:"unit"`
	LaborHours  float64 `json:"labor_hours"  binding:"required"`
	Description string  `json:"description"`
	Source      string  `json:"source"`
}

type UpdateNormRequest struct {
	CategoryID  *uint    `json:"category_id"`
	Code        *string  `json:"code"`
	Name        *string  `json:"name"`
	Unit        *string  `json:"unit"`
	LaborHours  *float64 `json:"labor_hours"`
	Description *string  `json:"description"`
	Source      *string  `json:"source"`
}

type CreateCoefficientRequest struct {
	Name        string  `json:"name"        binding:"required"`
	Code        string  `json:"code"        binding:"required"`
	Value       float64 `json:"value"       binding:"required"`
	CategoryID  *uint   `json:"category_id"`
	Description string  `json:"description"`
}

type UpdateCoefficientRequest struct {
	Name        *string  `json:"name"`
	Code        *string  `json:"code"`
	Value       *float64 `json:"value"`
	CategoryID  *uint    `json:"category_id"`
	Description *string  `json:"description"`
}
