package model

import "time"

type LaborCategory struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null;size:200" json:"name"`
	Code        string    `gorm:"not null;unique;size:50" json:"code"`
	Description string    `gorm:"type:text;default:''" json:"description"`
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}

type LaborNorm struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	CategoryID   uint      `gorm:"not null;index" json:"category_id"`
	Code         string    `gorm:"not null;size:50;index" json:"code"`
	Name         string    `gorm:"not null;size:300" json:"name"`
	Unit         string    `gorm:"not null;size:20;default:'чел.-ч'" json:"unit"`
	LaborHours   float64   `gorm:"not null" json:"labor_hours"`
	Description  string    `gorm:"type:text;default:''" json:"description"`
	Source       string    `gorm:"size:100;default:'ГЭСН'" json:"source"`
	CategoryName string    `gorm:"-" json:"category_name,omitempty"`
	CreatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}

type LaborCoefficient struct {
	ID          uint    `gorm:"primaryKey" json:"id"`
	Name        string  `gorm:"not null;size:200" json:"name"`
	Code        string  `gorm:"not null;size:50" json:"code"`
	Value       float64 `gorm:"not null;default:1.0" json:"value"`
	CategoryID  *uint   `gorm:"index" json:"category_id"`
	Description string  `gorm:"type:text;default:''" json:"description"`
}

type LaborCalculation struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ProjectID   uint      `gorm:"not null;index" json:"project_id"`
	CreatedByID uint      `gorm:"not null" json:"created_by_id"`
	Name        string    `gorm:"not null;size:300;default:'Расчёт трудозатрат'" json:"name"`
	GrandTotal  float64   `gorm:"not null;default:0" json:"grand_total"`
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type LaborCalculationEntry struct {
	ID            uint    `gorm:"primaryKey" json:"id"`
	CalculationID uint    `gorm:"not null;index" json:"calculation_id"`
	NormID        uint    `gorm:"not null" json:"norm_id"`
	Quantity      float64 `gorm:"not null;default:1" json:"quantity"`
	CustomCoeff   float64 `gorm:"not null;default:1" json:"custom_coeff"`
}

type LaborCalculationEntryCoeff struct {
	ID            uint    `gorm:"primaryKey" json:"id"`
	EntryID       uint    `gorm:"not null;index" json:"entry_id"`
	CoefficientID uint    `gorm:"not null" json:"coefficient_id"`
	Value         float64 `gorm:"not null" json:"value"`
}

type LaborCalculationGlobalCoeff struct {
	ID            uint    `gorm:"primaryKey" json:"id"`
	CalculationID uint    `gorm:"not null;index" json:"calculation_id"`
	CategoryID    uint    `gorm:"not null" json:"category_id"`
	Value         float64 `gorm:"not null;default:1" json:"value"`
}

type LaborCalculationRequest struct {
	ProjectID    uint                     `json:"project_id" binding:"required"`
	Name         string                   `json:"name"`
	Items        []LaborCalculationItem   `json:"items"`
	GlobalCoeffs []LaborGlobalCoefficient `json:"global_coefficients"`
}

type LaborCalculationItem struct {
	NormID       uint                   `json:"norm_id"`
	Quantity     float64                `json:"quantity"`
	CustomCoeff  float64                `json:"custom_coeff"`
	Coefficients []LaborItemCoefficient `json:"coefficients"`
}

type LaborItemCoefficient struct {
	CoefficientID uint    `json:"coefficient_id"`
	Value         float64 `json:"value"`
}

type LaborGlobalCoefficient struct {
	CategoryID uint    `json:"category_id"`
	Value      float64 `json:"value"`
	Name       string  `json:"name"`
}

type LaborCalculationResult struct {
	ID             uint                         `json:"id"`
	ProjectID      uint                         `json:"project_id"`
	ProjectName    string                       `json:"project_name"`
	Name           string                       `json:"name"`
	Items          []LaborCalculationResultItem `json:"items"`
	CategoryTotals []LaborCategoryTotal         `json:"category_totals"`
	GrandTotal     float64                      `json:"grand_total"`
	CalculatedAt   string                       `json:"calculated_at"`
}

type LaborCalculationResultItem struct {
	NormCode        string  `json:"norm_code"`
	NormName        string  `json:"norm_name"`
	CategoryName    string  `json:"category_name"`
	Unit            string  `json:"unit"`
	Quantity        float64 `json:"quantity"`
	BaseLaborHours  float64 `json:"base_labor_hours"`
	CoefficientMult float64 `json:"coefficient_mult"`
	TotalLaborHours float64 `json:"total_labor_hours"`
	TotalManDays    float64 `json:"total_man_days"`
}

type LaborCategoryTotal struct {
	CategoryName    string  `json:"category_name"`
	TotalLaborHours float64 `json:"total_labor_hours"`
	TotalManDays    float64 `json:"total_man_days"`
	GlobalCoeff     float64 `json:"global_coeff"`
	AdjustedHours   float64 `json:"adjusted_hours"`
	AdjustedDays    float64 `json:"adjusted_days"`
}

type LaborCalculationSummary struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	GrandTotal  float64 `json:"grand_total"`
	ItemsCount  int     `json:"items_count"`
	CreatedByID uint    `json:"created_by_id"`
	CreatedBy   string  `json:"created_by"`
	CreatedAt   string  `json:"created_at"`
}
