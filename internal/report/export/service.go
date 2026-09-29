package export

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/eduardnesterov/task-manager/internal/models"
	"github.com/eduardnesterov/task-manager/internal/project"
	"github.com/eduardnesterov/task-manager/internal/report"
	"github.com/eduardnesterov/task-manager/internal/task"
	"github.com/eduardnesterov/task-manager/internal/user"
	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
)

type ExportService struct {
	taskRepo    task.TaskRepository
	projectRepo project.ProjectRepository
	userRepo    user.UserRepository
	reportRepo  report.ReportRepository
}

func NewExportService(
	taskRepo task.TaskRepository,
	projectRepo project.ProjectRepository,
	userRepo user.UserRepository,
	reportRepo report.ReportRepository,
) *ExportService {
	return &ExportService{
		taskRepo:    taskRepo,
		projectRepo: projectRepo,
		userRepo:    userRepo,
		reportRepo:  reportRepo,
	}
}

// Стили
func createHeaderStyle(f *excelize.File) (int, error) {
	return f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "FFFFFF",
			Size:  11,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"1A73E8"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
		},
	})
}

func createRowStyle(f *excelize.File, even bool) (int, error) {
	color := "F8F9FA"
	if even {
		color = "FFFFFF"
	}
	return f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{color},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Vertical: "center",
			WrapText: true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "DEE2E6", Style: 1},
			{Type: "right", Color: "DEE2E6", Style: 1},
			{Type: "top", Color: "DEE2E6", Style: 1},
			{Type: "bottom", Color: "DEE2E6", Style: 1},
		},
	})
}

// Вспомогательные
func colName(index int) string {
	name := ""
	for index >= 0 {
		name = string(rune('A'+index%26)) + name
		index = index/26 - 1
	}
	return name
}

func cellRef(col, row int) string {
	return fmt.Sprintf("%s%d", colName(col), row)
}

func formatExportTime(t *time.Time) string {
	if t == nil {
		return "—"
	}
	return t.Format("02.01.2006 15:04")
}

func formatExportDate(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	return t.Format("02.01.2006")
}

func boolToString(v bool, yes, no string) string {
	if v {
		return yes
	}
	return no
}

func priorityText(p int) string {
	switch p {
	case 2:
		return "Высокий"
	case 1:
		return "Средний"
	default:
		return "Низкий"
	}
}

func writeSheet(
	f *excelize.File,
	sheet string,
	headers []string,
	widths []float64,
	rows [][]interface{},
) error {
	f.SetSheetName("Sheet1", sheet)

	headerStyle, err := createHeaderStyle(f)
	if err != nil {
		return err
	}

	// Записываем заголовки
	for i, h := range headers {
		cell := cellRef(i, 1)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	// Ширина колонок
	for i, w := range widths {
		col := colName(i)
		f.SetColWidth(sheet, col, col, w)
	}

	// Высота заголовка
	f.SetRowHeight(sheet, 1, 30)

	// Записываем данные
	for i, row := range rows {
		rowNum := i + 2

		rStyle, err := createRowStyle(f, i%2 == 0)
		if err != nil {
			return err
		}

		for j, val := range row {
			cell := cellRef(j, rowNum)
			f.SetCellValue(sheet, cell, val)
			f.SetCellStyle(sheet, cell, cell, rStyle)
		}

		f.SetRowHeight(sheet, rowNum, 22)
	}

	// Закрепляем шапку
	f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	})

	return nil
}

// Экспорт задач
func (s *ExportService) ExportTasks() (*excelize.File, error) {
	entities, err := s.taskRepo.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	headers := []string{
		"№", "Название", "Категория", "Проект", "Исполнитель",
		"Статус", "Приоритет", "Начало", "Срок",
		"Факт. завершение", "Факт. время (ч)", "Выполнена",
	}

	widths := []float64{
		5, 35, 18, 22, 22,
		15, 12, 18, 18,
		20, 16, 12,
	}

	rows := make([][]interface{}, 0, len(entities))

	for i, entity := range entities {
		task, ok := entity.(*models.Task)
		if !ok {
			continue
		}

		rows = append(rows, []interface{}{
			i + 1,
			task.Name,
			task.Category,
			task.ProjectID,
			task.ExecutorID,
			task.StatusID,
			priorityText(task.Priority),
			formatExportTime(task.StartTime),
			formatExportTime(task.EndTime),
			formatExportTime(task.FactEndTime),
			fmt.Sprintf("%.1f", task.FactTime),
			boolToString(task.Completed, "Да", "Нет"),
		})
	}

	f := excelize.NewFile()
	if err := writeSheet(f, "Задачи", headers, widths, rows); err != nil {
		return nil, err
	}

	return f, nil
}

// Экспорт проектов
func (s *ExportService) ExportProjects() (*excelize.File, error) {
	entities, err := s.projectRepo.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	headers := []string{
		"№", "Название", "Описание", "Ответственный", "Стадия",
		"Дата начала", "Дедлайн", "Дата завершения",
		"Стоимость", "Адрес",
	}

	widths := []float64{
		5, 30, 35, 22, 18,
		16, 16, 18,
		14, 35,
	}

	rows := make([][]interface{}, 0, len(entities))

	for i, entity := range entities {
		project, ok := entity.(*models.Project)
		if !ok {
			continue
		}

		rows = append(rows, []interface{}{
			i + 1,
			project.Name,
			project.Description,
			project.ResponsibleID,
			project.StageID,
			formatExportDate(project.StartDate),
			formatExportDate(project.DeadlineDate),
			formatExportDate(project.EndDate),
			fmt.Sprintf("%.2f", project.Cost),
			project.Address,
		})
	}

	f := excelize.NewFile()
	if err := writeSheet(f, "Проекты", headers, widths, rows); err != nil {
		return nil, err
	}

	return f, nil
}

// Экспорт отчётов
func (s *ExportService) ExportReports() (*excelize.File, error) {
	entities, err := s.reportRepo.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list reports: %w", err)
	}

	headers := []string{
		"№", "Дата", "Пользователь", "Проект",
		"Текст отчёта", "Переработки (ч)",
	}

	widths := []float64{
		5, 18, 22, 22,
		50, 16,
	}

	rows := make([][]interface{}, 0, len(entities))

	for i, entity := range entities {
		report, ok := entity.(*models.Report)
		if !ok {
			continue
		}

		rows = append(rows, []interface{}{
			i + 1,
			report.CreatedAt.Format("02.01.2006 15:04"),
			report.UserID,
			report.ProjectID,
			report.Text,
			report.Overwork,
		})
	}

	f := excelize.NewFile()
	if err := writeSheet(f, "Отчёты", headers, widths, rows); err != nil {
		return nil, err
	}

	return f, nil
}

func (s *ExportService) ExportTimesheet(dateFrom, dateTo string) (*excelize.File, error) {
	users, err := s.userRepo.GetActiveUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	reports, err := s.reportRepo.FilterReports(dateFrom, dateTo, "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list reports: %w", err)
	}

	from, err := time.Parse("2006-01-02", dateFrom)
	if err != nil {
		return nil, fmt.Errorf("invalid date_from format, expected YYYY-MM-DD")
	}
	to, err := time.Parse("2006-01-02", dateTo)
	if err != nil {
		return nil, fmt.Errorf("invalid date_to format, expected YYYY-MM-DD")
	}
	if to.Before(from) {
		return nil, fmt.Errorf("date_to must be greater than or equal to date_from")
	}

	dates := make([]time.Time, 0)
	for current := from; !current.After(to); current = current.AddDate(0, 0, 1) {
		dates = append(dates, current)
	}

	overworkByUserDate := make(map[uint]map[string]int)
	for _, report := range reports {
		if report.Overwork <= 0 {
			continue
		}
		dateKey := report.CreatedAt.Format("2006-01-02")
		if _, ok := overworkByUserDate[report.UserID]; !ok {
			overworkByUserDate[report.UserID] = make(map[string]int)
		}
		overworkByUserDate[report.UserID][dateKey] += report.Overwork
	}

	headers := []string{"Сотрудник"}
	widths := []float64{28}
	for _, date := range dates {
		headers = append(headers, date.Format("02.01.2006"))
		widths = append(widths, 12)
	}
	headers = append(headers, "Итого")
	widths = append(widths, 12)

	rows := make([][]interface{}, 0, len(users))
	for _, user := range users {
		row := []interface{}{user.Name}
		total := 0
		for _, date := range dates {
			hours := overworkByUserDate[user.ID][date.Format("2006-01-02")]
			total += hours
			if hours > 0 {
				row = append(row, hours)
			} else {
				row = append(row, "")
			}
		}
		row = append(row, total)
		rows = append(rows, row)
	}

	f := excelize.NewFile()
	if err := writeSheet(f, "Табель", headers, widths, rows); err != nil {
		return nil, err
	}

	return f, nil
}

// ExportLaborExcel генерирует Excel-файл с результатами расчёта трудозатрат.
func (s *ExportService) ExportLaborExcel(result *models.LaborCalculationResult) (*bytes.Buffer, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Трудозатраты"
	index, err := f.NewSheet(sheet)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheet: %w", err)
	}
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	// ── Стили ────────────────────────────────────────────────────────────────

	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"2E5090"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    fullBorder("000000"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create header style: %w", err)
	}

	cellStyle, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
		Border:    fullBorder("000000"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create cell style: %w", err)
	}

	numberStyle, err := f.NewStyle(&excelize.Style{
		NumFmt:    2, // #,##0.00
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border:    fullBorder("000000"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create number style: %w", err)
	}

	totalStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11},
		NumFmt:    2,
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"E8F0FE"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border:    fullBorder("000000"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create total style: %w", err)
	}

	titleStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create title style: %w", err)
	}

	// ── Ширины колонок ───────────────────────────────────────────────────────

	colWidths := map[string]float64{
		"A": 5, "B": 20, "C": 45, "D": 18,
		"E": 12, "F": 10, "G": 16, "H": 12,
		"I": 16, "J": 14,
	}
	for col, width := range colWidths {
		f.SetColWidth(sheet, col, col, width)
	}

	// ── Заголовок документа ──────────────────────────────────────────────────

	f.MergeCell(sheet, "A1", "J1")
	f.SetCellValue(sheet, "A1", fmt.Sprintf("РАСЧЁТ ТРУДОЗАТРАТ: %s", result.ProjectName))
	f.SetCellStyle(sheet, "A1", "J1", titleStyle)

	f.MergeCell(sheet, "A2", "J2")
	f.SetCellValue(sheet, "A2", fmt.Sprintf("Дата расчёта: %s", result.CalculatedAt))

	// ── Шапка таблицы позиций ────────────────────────────────────────────────

	const headerRow = 4
	headers := []string{
		"№", "Код нормы", "Наименование работ", "Категория",
		"Ед. изм.", "Кол-во", "База (чел.-ч)", "Коэфф.",
		"Итого (чел.-ч)", "Итого (чел.-дн.)",
	}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, headerRow)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	// ── Строки позиций ───────────────────────────────────────────────────────

	for i, item := range result.Items {
		row := headerRow + 1 + i
		cells := []interface{}{
			i + 1,
			item.NormCode,
			item.NormName,
			item.CategoryName,
			item.Unit,
			item.Quantity,
			item.BaseLaborHours,
			item.CoefficientMult,
			item.TotalLaborHours,
			item.TotalManDays,
		}
		for j, val := range cells {
			cell, _ := excelize.CoordinatesToCellName(j+1, row)
			f.SetCellValue(sheet, cell, val)
			if j >= 5 { // числовые колонки: Кол-во, База, Коэфф., Итого...
				f.SetCellStyle(sheet, cell, cell, numberStyle)
			} else {
				f.SetCellStyle(sheet, cell, cell, cellStyle)
			}
		}
	}

	// ── Итого по категориям ──────────────────────────────────────────────────

	catTitleRow := headerRow + len(result.Items) + 2

	f.MergeCell(sheet, fmt.Sprintf("A%d", catTitleRow), fmt.Sprintf("J%d", catTitleRow))
	f.SetCellValue(sheet, fmt.Sprintf("A%d", catTitleRow), "ИТОГО ПО КАТЕГОРИЯМ")
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", catTitleRow), fmt.Sprintf("J%d", catTitleRow), headerStyle)

	catHeaderRow := catTitleRow + 1
	catHeaders := []string{
		"", "Категория", "", "", "", "",
		"Чел.-ч", "Глоб.коэфф.", "С коэфф.(чел.-ч)", "С коэфф.(чел.-дн.)",
	}
	for i, h := range catHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, catHeaderRow)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	for i, ct := range result.CategoryTotals {
		row := catHeaderRow + 1 + i

		// Объединяем B:F для названия категории
		f.MergeCell(sheet,
			fmt.Sprintf("B%d", row),
			fmt.Sprintf("F%d", row),
		)

		setCellStyled := func(col int, val interface{}, style int) {
			cell, _ := excelize.CoordinatesToCellName(col, row)
			f.SetCellValue(sheet, cell, val)
			f.SetCellStyle(sheet, cell, cell, style)
		}

		setCellStyled(1, i+1, cellStyle)
		setCellStyled(2, ct.CategoryName, cellStyle)
		setCellStyled(7, ct.TotalLaborHours, numberStyle)
		setCellStyled(8, ct.GlobalCoeff, numberStyle)
		setCellStyled(9, ct.AdjustedHours, numberStyle)
		setCellStyled(10, ct.AdjustedDays, numberStyle)
	}

	// ── Гранд-тотал ──────────────────────────────────────────────────────────

	grandRow := catHeaderRow + len(result.CategoryTotals) + 2

	f.MergeCell(sheet,
		fmt.Sprintf("A%d", grandRow),
		fmt.Sprintf("H%d", grandRow),
	)
	f.SetCellValue(sheet, fmt.Sprintf("A%d", grandRow), "ИТОГО ТРУДОЗАТРАТЫ:")
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", grandRow), fmt.Sprintf("H%d", grandRow), totalStyle)

	iCell, _ := excelize.CoordinatesToCellName(9, grandRow)
	f.SetCellValue(sheet, iCell, result.GrandTotal)
	f.SetCellStyle(sheet, iCell, iCell, totalStyle)

	jCell, _ := excelize.CoordinatesToCellName(10, grandRow)
	f.SetCellValue(sheet, jCell, fmt.Sprintf("%.2f", result.GrandTotal/8))
	f.SetCellStyle(sheet, jCell, jCell, totalStyle)

	// ── Запись в буфер ───────────────────────────────────────────────────────

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write excel buffer: %w", err)
	}
	return &buf, nil
}

// ExportLaborPDF генерирует PDF-файл с результатами расчёта трудозатрат.
func (s *ExportService) ExportLaborPDF(result *models.LaborCalculationResult) (*bytes.Buffer, error) {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 15)

	// ── Шрифты ───────────────────────────────────────────────────────────────

	fontFamily := "DejaVu"
	regularPath := findLaborFont()
	boldPath := findLaborFontBold()

	if regularPath != "" {
		pdf.AddUTF8Font("DejaVu", "", regularPath)
		if boldPath != "" {
			pdf.AddUTF8Font("DejaVu", "B", boldPath)
		}
	} else {
		// Fallback: встроенный Helvetica (без кириллицы, но не упадёт)
		fontFamily = "Helvetica"
	}

	pdf.AddPage()
	pdf.SetFont(fontFamily, "", 10)

	// ── Заголовок ────────────────────────────────────────────────────────────

	pdf.SetFont(fontFamily, "B", 14)
	pdf.CellFormat(0, 10,
		fmt.Sprintf("РАСЧЁТ ТРУДОЗАТРАТ: %s", result.ProjectName),
		"", 1, "C", false, 0, "")

	pdf.SetFont(fontFamily, "", 9)
	pdf.CellFormat(0, 6,
		fmt.Sprintf("Дата расчёта: %s", result.CalculatedAt),
		"", 1, "C", false, 0, "")
	pdf.Ln(5)

	// ── Таблица позиций ──────────────────────────────────────────────────────

	pdf.SetFont(fontFamily, "B", 8)
	pdf.SetFillColor(46, 80, 144)
	pdf.SetTextColor(255, 255, 255)

	colWidthsPDF := []float64{8, 30, 70, 30, 15, 15, 25, 18, 30, 28}
	pdfHeaders := []string{
		"№", "Код", "Наименование", "Категория",
		"Ед.", "Кол.", "База(ч-ч)", "Коэфф", "Итого(ч-ч)", "Итого(ч-д)",
	}

	for i, h := range pdfHeaders {
		pdf.CellFormat(colWidthsPDF[i], 8, h, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	// Строки позиций
	pdf.SetFont(fontFamily, "", 7)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFillColor(245, 245, 245)

	for i, item := range result.Items {
		fill := i%2 == 0
		cells := []string{
			fmt.Sprintf("%d", i+1),
			item.NormCode,
			item.NormName,
			item.CategoryName,
			item.Unit,
			fmt.Sprintf("%.1f", item.Quantity),
			fmt.Sprintf("%.2f", item.BaseLaborHours),
			fmt.Sprintf("%.3f", item.CoefficientMult),
			fmt.Sprintf("%.2f", item.TotalLaborHours),
			fmt.Sprintf("%.2f", item.TotalManDays),
		}
		for j, cell := range cells {
			align := "L"
			if j == 0 || j >= 5 {
				align = "C"
			}
			pdf.CellFormat(colWidthsPDF[j], 6, cell, "1", 0, align, fill, 0, "")
		}
		pdf.Ln(-1)
	}

	// ── Итого по категориям ──────────────────────────────────────────────────

	pdf.Ln(5)
	pdf.SetFont(fontFamily, "B", 10)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(0, 8, "ИТОГО ПО КАТЕГОРИЯМ", "", 1, "L", false, 0, "")

	pdf.SetFont(fontFamily, "B", 8)
	pdf.SetFillColor(46, 80, 144)
	pdf.SetTextColor(255, 255, 255)

	catColW := []float64{60, 35, 30, 40, 40}
	catPDFHeaders := []string{
		"Категория", "Чел.-часы", "Глоб.коэфф.",
		"С коэфф.(ч-ч)", "С коэфф.(ч-д)",
	}
	for i, h := range catPDFHeaders {
		pdf.CellFormat(catColW[i], 8, h, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont(fontFamily, "", 9)
	pdf.SetTextColor(0, 0, 0)

	for _, ct := range result.CategoryTotals {
		pdf.CellFormat(catColW[0], 7, ct.CategoryName, "1", 0, "L", false, 0, "")
		pdf.CellFormat(catColW[1], 7, fmt.Sprintf("%.2f", ct.TotalLaborHours), "1", 0, "C", false, 0, "")
		pdf.CellFormat(catColW[2], 7, fmt.Sprintf("%.3f", ct.GlobalCoeff), "1", 0, "C", false, 0, "")
		pdf.CellFormat(catColW[3], 7, fmt.Sprintf("%.2f", ct.AdjustedHours), "1", 0, "C", false, 0, "")
		pdf.CellFormat(catColW[4], 7, fmt.Sprintf("%.2f", ct.AdjustedDays), "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
	}

	// ── Гранд-тотал ──────────────────────────────────────────────────────────

	pdf.Ln(3)
	pdf.SetFont(fontFamily, "B", 12)
	pdf.SetFillColor(232, 240, 254)
	pdf.SetTextColor(0, 0, 0)

	pdf.CellFormat(125, 10, "ИТОГО ТРУДОЗАТРАТЫ:", "1", 0, "R", true, 0, "")
	pdf.CellFormat(40, 10, fmt.Sprintf("%.2f чел.-ч", result.GrandTotal), "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 10, fmt.Sprintf("%.2f чел.-дн.", result.GrandTotal/8), "1", 0, "C", true, 0, "")
	pdf.Ln(-1)

	// ── Запись в буфер ───────────────────────────────────────────────────────

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to write pdf buffer: %w", err)
	}
	return &buf, nil
}

// ─── Вспомогательные функции ────────────────────────────────────────────────

// fullBorder возвращает полную рамку ячейки заданного цвета.
func fullBorder(color string) []excelize.Border {
	sides := []string{"left", "top", "right", "bottom"}
	borders := make([]excelize.Border, len(sides))
	for i, side := range sides {
		borders[i] = excelize.Border{Type: side, Style: 1, Color: color}
	}
	return borders
}

// findLaborFont ищет файл обычного шрифта с поддержкой кириллицы.
func findLaborFont() string {
	candidates := []string{
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		"/usr/share/fonts/TTF/DejaVuSans.ttf",
		"fonts/DejaVuSans.ttf",
		"C:\\Windows\\Fonts\\arial.ttf",
		"/System/Library/Fonts/Supplemental/Arial.ttf",
	}
	return firstExistingPath(candidates)
}

// findLaborFontBold ищет файл жирного шрифта с поддержкой кириллицы.
func findLaborFontBold() string {
	candidates := []string{
		"/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf",
		"/usr/share/fonts/TTF/DejaVuSans-Bold.ttf",
		"fonts/DejaVuSans-Bold.ttf",
		"C:\\Windows\\Fonts\\arialbd.ttf",
		"/System/Library/Fonts/Supplemental/Arial Bold.ttf",
	}
	return firstExistingPath(candidates)
}

// firstExistingPath возвращает абсолютный путь первого существующего файла из списка.
func firstExistingPath(paths []string) string {
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			abs, _ := filepath.Abs(p)
			return abs
		}
	}
	return ""
}
