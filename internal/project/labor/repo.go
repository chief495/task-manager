package labor

import (
	"github.com/eduardnesterov/task-manager/internal/models"
	"gorm.io/gorm"
)

type LaborRepository struct {
	db *gorm.DB
}

type SavedCalculation struct {
	Calculation  models.LaborCalculation
	Items        []SavedCalculationEntry
	GlobalCoeffs []models.LaborCalculationGlobalCoeff
}

type SavedCalculationEntry struct {
	Entry                models.LaborCalculationEntry
	SelectedCoefficients []models.LaborCalculationEntryCoeff
}

func NewLaborRepository(db *gorm.DB) *LaborRepository {
	return &LaborRepository{db: db}
}

// Категории
func (r *LaborRepository) CreateCategory(cat *models.LaborCategory) error {
	return r.db.Create(cat).Error
}

func (r *LaborRepository) GetCategory(id uint) (*models.LaborCategory, error) {
	var cat models.LaborCategory
	err := r.db.First(&cat, id).Error
	return &cat, err
}

func (r *LaborRepository) ListCategories() ([]models.LaborCategory, error) {
	var cats []models.LaborCategory
	err := r.db.Order("code").Find(&cats).Error
	return cats, err
}

func (r *LaborRepository) UpdateCategory(cat *models.LaborCategory) error {
	return r.db.Save(cat).Error
}

func (r *LaborRepository) DeleteCategory(id uint) error {
	return r.db.Delete(&models.LaborCategory{}, id).Error
}

// Нормы

func (r *LaborRepository) CreateNorm(norm *models.LaborNorm) error {
	return r.db.Create(norm).Error
}

func (r *LaborRepository) GetNorm(id uint) (*models.LaborNorm, error) {
	var norm models.LaborNorm
	err := r.db.
		Joins("LEFT JOIN labor_categories ON labor_categories.id = labor_norms.category_id").
		Select("labor_norms.*, labor_categories.name AS category_name").
		First(&norm, id).Error
	return &norm, err
}

func (r *LaborRepository) ListNorms(categoryID uint) ([]models.LaborNorm, error) {
	var norms []models.LaborNorm

	q := r.db.
		Joins("LEFT JOIN labor_categories ON labor_categories.id = labor_norms.category_id").
		Select("labor_norms.*, labor_categories.name AS category_name").
		Order("labor_norms.code")

	if categoryID > 0 {
		q = q.Where("labor_norms.category_id = ?", categoryID)
	}

	err := q.Find(&norms).Error
	return norms, err
}

func (r *LaborRepository) UpdateNorm(norm *models.LaborNorm) error {
	return r.db.Save(norm).Error
}

func (r *LaborRepository) DeleteNorm(id uint) error {
	return r.db.Delete(&models.LaborNorm{}, id).Error
}

// ─── Коэффициенты ────────────────────────────────────────────────────────────

func (r *LaborRepository) CreateCoefficient(coeff *models.LaborCoefficient) error {
	return r.db.Create(coeff).Error
}

func (r *LaborRepository) GetCoefficient(id uint) (*models.LaborCoefficient, error) {
	var coeff models.LaborCoefficient
	err := r.db.First(&coeff, id).Error
	return &coeff, err
}

func (r *LaborRepository) ListCoefficients(categoryID uint) ([]models.LaborCoefficient, error) {
	var coeffs []models.LaborCoefficient

	q := r.db.Order("code")

	if categoryID > 0 {
		q = q.Where("category_id = ? OR category_id IS NULL", categoryID)
	}

	err := q.Find(&coeffs).Error
	return coeffs, err
}

func (r *LaborRepository) UpdateCoefficient(coeff *models.LaborCoefficient) error {
	return r.db.Save(coeff).Error
}

func (r *LaborRepository) DeleteCoefficient(id uint) error {
	return r.db.Delete(&models.LaborCoefficient{}, id).Error
}

// Сохранённые расчёты

// SaveCalculation сохраняет расчёт со всеми позициями и коэффициентами
func (r *LaborRepository) SaveCalculation(calc *SavedCalculation) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&calc.Calculation).Error; err != nil {
			return err
		}
		for i := range calc.Items {
			calc.Items[i].Entry.CalculationID = calc.Calculation.ID
			if err := tx.Create(&calc.Items[i].Entry).Error; err != nil {
				return err
			}
			for j := range calc.Items[i].SelectedCoefficients {
				calc.Items[i].SelectedCoefficients[j].EntryID = calc.Items[i].Entry.ID
				if err := tx.Create(&calc.Items[i].SelectedCoefficients[j]).Error; err != nil {
					return err
				}
			}
		}
		for i := range calc.GlobalCoeffs {
			calc.GlobalCoeffs[i].CalculationID = calc.Calculation.ID
			if err := tx.Create(&calc.GlobalCoeffs[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetCalculation возвращает расчёт со всеми вложенными данными
func (r *LaborRepository) GetCalculation(id uint) (*SavedCalculation, error) {
	var calc models.LaborCalculation
	if err := r.db.First(&calc, id).Error; err != nil {
		return nil, err
	}

	var entries []models.LaborCalculationEntry
	if err := r.db.Where("calculation_id = ?", id).Find(&entries).Error; err != nil {
		return nil, err
	}
	items := make([]SavedCalculationEntry, len(entries))
	for i := range entries {
		items[i].Entry = entries[i]
		if err := r.db.Where("entry_id = ?", entries[i].ID).Find(&items[i].SelectedCoefficients).Error; err != nil {
			return nil, err
		}
	}

	var globalCoeffs []models.LaborCalculationGlobalCoeff
	if err := r.db.Where("calculation_id = ?", id).Find(&globalCoeffs).Error; err != nil {
		return nil, err
	}

	return &SavedCalculation{Calculation: calc, Items: items, GlobalCoeffs: globalCoeffs}, nil
}

// ListCalculationsByProject возвращает список расчётов для проекта
func (r *LaborRepository) ListCalculationsByProject(projectID uint) ([]models.LaborCalculation, error) {
	var calcs []models.LaborCalculation
	err := r.db.
		Where("project_id = ?", projectID).
		Order("created_at DESC").
		Find(&calcs).Error
	return calcs, err
}

// DeleteCalculation удаляет расчёт и все связанные записи
func (r *LaborRepository) DeleteCalculation(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Находим все entry этого расчёта
		var entryIDs []uint
		tx.Model(&models.LaborCalculationEntry{}).
			Where("calculation_id = ?", id).
			Pluck("id", &entryIDs)

		// Удаляем коэффициенты позиций
		if len(entryIDs) > 0 {
			if err := tx.Where("entry_id IN ?", entryIDs).
				Delete(&models.LaborCalculationEntryCoeff{}).Error; err != nil {
				return err
			}
		}

		// Удаляем позиции
		if err := tx.Where("calculation_id = ?", id).
			Delete(&models.LaborCalculationEntry{}).Error; err != nil {
			return err
		}

		// Удаляем глобальные коэффициенты
		if err := tx.Where("calculation_id = ?", id).
			Delete(&models.LaborCalculationGlobalCoeff{}).Error; err != nil {
			return err
		}

		// Удаляем сам расчёт
		return tx.Delete(&models.LaborCalculation{}, id).Error
	})
}

// UpdateCalculation обновляет расчёт: удаляет старые данные и сохраняет новые
func (r *LaborRepository) UpdateCalculation(calc *models.LaborCalculation) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Удаляем старые позиции и коэффициенты
		var entryIDs []uint
		tx.Model(&models.LaborCalculationEntry{}).
			Where("calculation_id = ?", calc.ID).
			Pluck("id", &entryIDs)

		if len(entryIDs) > 0 {
			if err := tx.Where("entry_id IN ?", entryIDs).
				Delete(&models.LaborCalculationEntryCoeff{}).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("calculation_id = ?", calc.ID).
			Delete(&models.LaborCalculationEntry{}).Error; err != nil {
			return err
		}

		if err := tx.Where("calculation_id = ?", calc.ID).
			Delete(&models.LaborCalculationGlobalCoeff{}).Error; err != nil {
			return err
		}

		// Обновляем основные поля
		if err := tx.Model(calc).Updates(map[string]interface{}{
			"name":        calc.Name,
			"grand_total": calc.GrandTotal,
		}).Error; err != nil {
			return err
		}

		return nil
	})
}
