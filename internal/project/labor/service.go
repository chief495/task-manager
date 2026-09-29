package labor

import (
	"fmt"
	"math"
	"time"

	"github.com/eduardnesterov/task-manager/internal/models"
	"github.com/eduardnesterov/task-manager/internal/project"
)

type LaborService struct {
	repo        *LaborRepository
	projectRepo project.ProjectRepository
}

func NewLaborService(repo *LaborRepository, projectRepo project.ProjectRepository) *LaborService {
	return &LaborService{repo: repo, projectRepo: projectRepo}
}

// Категории
func (s *LaborService) CreateCategory(cat *models.LaborCategory) error {
	return s.repo.CreateCategory(cat)
}

func (s *LaborService) GetCategory(id uint) (*models.LaborCategory, error) {
	return s.repo.GetCategory(id)
}

func (s *LaborService) ListCategories() ([]models.LaborCategory, error) {
	return s.repo.ListCategories()
}

func (s *LaborService) UpdateCategory(cat *models.LaborCategory) error {
	return s.repo.UpdateCategory(cat)
}

func (s *LaborService) DeleteCategory(id uint) error {
	return s.repo.DeleteCategory(id)
}

// Нормы
func (s *LaborService) CreateNorm(norm *models.LaborNorm) error {
	return s.repo.CreateNorm(norm)
}

func (s *LaborService) GetNorm(id uint) (*models.LaborNorm, error) {
	return s.repo.GetNorm(id)
}

func (s *LaborService) ListNorms(categoryID uint) ([]models.LaborNorm, error) {
	return s.repo.ListNorms(categoryID)
}

func (s *LaborService) UpdateNorm(norm *models.LaborNorm) error {
	return s.repo.UpdateNorm(norm)
}

func (s *LaborService) DeleteNorm(id uint) error {
	return s.repo.DeleteNorm(id)
}

// Коэффициенты
func (s *LaborService) CreateCoefficient(coeff *models.LaborCoefficient) error {
	return s.repo.CreateCoefficient(coeff)
}

func (s *LaborService) GetCoefficient(id uint) (*models.LaborCoefficient, error) {
	return s.repo.GetCoefficient(id)
}

func (s *LaborService) ListCoefficients(categoryID uint) ([]models.LaborCoefficient, error) {
	return s.repo.ListCoefficients(categoryID)
}

func (s *LaborService) UpdateCoefficient(coeff *models.LaborCoefficient) error {
	return s.repo.UpdateCoefficient(coeff)
}

func (s *LaborService) DeleteCoefficient(id uint) error {
	return s.repo.DeleteCoefficient(id)
}

// Расчёт + сохранение

// Calculate выполняет расчёт и сохраняет его в базу
func (s *LaborService) Calculate(req *models.LaborCalculationRequest, userID uint) (*models.LaborCalculationResult, error) {
	// Проверяем что проект существует
	projectEntities, err := s.projectRepo.List()
	if err != nil {
		return nil, fmt.Errorf("failed to load projects: %w", err)
	}

	var projectName string
	for _, pe := range projectEntities {
		p, ok := pe.(*models.Project)
		if ok && p.ID == req.ProjectID {
			projectName = p.Name
			break
		}
	}
	if projectName == "" {
		return nil, fmt.Errorf("project with id %d not found", req.ProjectID)
	}

	// Загружаем нормы
	norms, err := s.repo.ListNorms(0)
	if err != nil {
		return nil, fmt.Errorf("failed to load norms: %w", err)
	}
	normMap := make(map[uint]models.LaborNorm, len(norms))
	for _, n := range norms {
		normMap[n.ID] = n
	}

	// Глобальные коэффициенты
	globalCoeffMap := make(map[uint]float64, len(req.GlobalCoeffs))
	for _, gc := range req.GlobalCoeffs {
		if gc.Value > 0 {
			globalCoeffMap[gc.CategoryID] = gc.Value
		}
	}

	calcName := req.Name
	if calcName == "" {
		calcName = fmt.Sprintf("Расчёт от %s", time.Now().Format("02.01.2006 15:04"))
	}

	result := &models.LaborCalculationResult{
		ProjectID:    req.ProjectID,
		ProjectName:  projectName,
		Name:         calcName,
		CalculatedAt: time.Now().Format("02.01.2006 15:04"),
	}

	// Модель для сохранения
	savedCalc := &SavedCalculation{
		Calculation: models.LaborCalculation{
			ProjectID:   req.ProjectID,
			CreatedByID: userID,
			Name:        calcName,
		},
	}

	// Глобальные коэффициенты для сохранения
	for _, gc := range req.GlobalCoeffs {
		if gc.Value > 0 && gc.Value != 1.0 {
			savedCalc.GlobalCoeffs = append(savedCalc.GlobalCoeffs, models.LaborCalculationGlobalCoeff{
				CategoryID: gc.CategoryID,
				Value:      gc.Value,
			})
		}
	}

	type catAccum struct {
		categoryID uint
		totalHours float64
		totalDays  float64
	}
	accumMap := make(map[string]*catAccum)

	for _, item := range req.Items {
		norm, ok := normMap[item.NormID]
		if !ok {
			continue
		}

		// Коэффициенты позиции (без пользовательского)
		coeffMult := 1.0
		var selectedCoeffs []models.LaborCalculationEntryCoeff

		for _, c := range item.Coefficients {
			if c.Value > 0 {
				coeffMult *= c.Value
				if c.CoefficientID > 0 {
					selectedCoeffs = append(selectedCoeffs, models.LaborCalculationEntryCoeff{
						CoefficientID: c.CoefficientID,
						Value:         c.Value,
					})
				}
			}
		}

		// Пользовательский коэффициент
		customCoeff := item.CustomCoeff
		if customCoeff <= 0 {
			customCoeff = 1.0
		}
		coeffMult *= customCoeff

		baseLH := norm.LaborHours * item.Quantity
		totalLH := round2(baseLH * coeffMult)

		catName := norm.CategoryName
		if catName == "" {
			catName = fmt.Sprintf("Категория #%d", norm.CategoryID)
		}

		result.Items = append(result.Items, models.LaborCalculationResultItem{
			NormCode:        norm.Code,
			NormName:        norm.Name,
			CategoryName:    catName,
			Unit:            norm.Unit,
			Quantity:        item.Quantity,
			BaseLaborHours:  round2(baseLH),
			CoefficientMult: round3(coeffMult),
			TotalLaborHours: totalLH,
			TotalManDays:    round2(totalLH / 8),
		})

		// Позиция для сохранения
		savedCalc.Items = append(savedCalc.Items, SavedCalculationEntry{
			Entry: models.LaborCalculationEntry{
				NormID:      item.NormID,
				Quantity:    item.Quantity,
				CustomCoeff: customCoeff,
			},
			SelectedCoefficients: selectedCoeffs,
		})

		if _, ok := accumMap[catName]; !ok {
			accumMap[catName] = &catAccum{categoryID: norm.CategoryID}
		}
		accumMap[catName].totalHours += totalLH
		accumMap[catName].totalDays += round2(totalLH / 8)
	}

	// Итоги по категориям
	grandTotal := 0.0
	for catName, acc := range accumMap {
		acc.totalHours = round2(acc.totalHours)
		acc.totalDays = round2(acc.totalDays)

		gc := globalCoeffMap[acc.categoryID]
		if gc == 0 {
			gc = 1.0
		}

		adjusted := round2(acc.totalHours * gc)

		result.CategoryTotals = append(result.CategoryTotals, models.LaborCategoryTotal{
			CategoryName:    catName,
			TotalLaborHours: acc.totalHours,
			TotalManDays:    acc.totalDays,
			GlobalCoeff:     gc,
			AdjustedHours:   adjusted,
			AdjustedDays:    round2(adjusted / 8),
		})
		grandTotal += adjusted
	}

	result.GrandTotal = round2(grandTotal)
	savedCalc.Calculation.GrandTotal = result.GrandTotal

	// Сохраняем в базу
	if err := s.repo.SaveCalculation(savedCalc); err != nil {
		return nil, fmt.Errorf("failed to save calculation: %w", err)
	}

	result.ID = savedCalc.Calculation.ID
	return result, nil
}

// GetCalculation возвращает сохранённый расчёт и пересчитывает результат
func (s *LaborService) GetCalculation(id uint) (*models.LaborCalculationResult, error) {
	calc, err := s.repo.GetCalculation(id)
	if err != nil {
		return nil, err
	}

	// Восстанавливаем запрос из сохранённых данных и пересчитываем
	req := &models.LaborCalculationRequest{
		ProjectID: calc.Calculation.ProjectID,
		Name:      calc.Calculation.Name,
	}

	for _, entry := range calc.Items {
		item := models.LaborCalculationItem{
			NormID:      entry.Entry.NormID,
			Quantity:    entry.Entry.Quantity,
			CustomCoeff: entry.Entry.CustomCoeff,
		}
		for _, sc := range entry.SelectedCoefficients {
			item.Coefficients = append(item.Coefficients, models.LaborItemCoefficient{
				CoefficientID: sc.CoefficientID,
				Value:         sc.Value,
			})
		}
		req.Items = append(req.Items, item)
	}

	for _, gc := range calc.GlobalCoeffs {
		req.GlobalCoeffs = append(req.GlobalCoeffs, models.LaborGlobalCoefficient{
			CategoryID: gc.CategoryID,
			Value:      gc.Value,
		})
	}

	result, err := s.Calculate(req, calc.Calculation.CreatedByID)
	if err != nil {
		return nil, err
	}

	// Подставляем ID оригинального расчёта, а не нового
	result.ID = calc.Calculation.ID
	return result, nil
}

// ListCalculationsByProject возвращает список расчётов для проекта
func (s *LaborService) ListCalculationsByProject(projectID uint) ([]models.LaborCalculationSummary, error) {
	calcs, err := s.repo.ListCalculationsByProject(projectID)
	if err != nil {
		return nil, err
	}

	summaries := make([]models.LaborCalculationSummary, 0, len(calcs))
	for _, c := range calcs {
		summaries = append(summaries, models.LaborCalculationSummary{
			ID:          c.ID,
			Name:        c.Name,
			GrandTotal:  c.GrandTotal,
			ItemsCount:  0,
			CreatedByID: c.CreatedByID,
			CreatedBy:   "",
			CreatedAt:   c.CreatedAt.Format("02.01.2006 15:04"),
		})
	}

	return summaries, nil
}

// DeleteCalculation удаляет расчёт
func (s *LaborService) DeleteCalculation(id uint) error {
	return s.repo.DeleteCalculation(id)
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func round3(v float64) float64 {
	return math.Round(v*1000) / 1000
}
