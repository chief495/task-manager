package export

import (
	"fmt"
	"net/http"
	"time"

	"github.com/eduardnesterov/task-manager/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type ExportHandler struct {
	exportService *ExportService
}

func NewExportHandler(exportService *ExportService) *ExportHandler {
	return &ExportHandler{exportService: exportService}
}

func sendExcelFile(c *gin.Context, f *excelize.File, prefix string) {
	filename := fmt.Sprintf("%s_%s.xlsx", prefix, time.Now().Format("2006-01-02"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")

	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write excel file"})
	}
}

func (h *ExportHandler) ExportTasks(c *gin.Context) {
	f, err := h.exportService.ExportTasks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	sendExcelFile(c, f, "tasks")
}

func (h *ExportHandler) ExportProjects(c *gin.Context) {
	f, err := h.exportService.ExportProjects()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	sendExcelFile(c, f, "projects")
}

func (h *ExportHandler) ExportReports(c *gin.Context) {
	f, err := h.exportService.ExportReports()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	sendExcelFile(c, f, "reports")
}

func (h *ExportHandler) ExportTimesheet(c *gin.Context) {
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	if dateFrom == "" || dateTo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date_from and date_to are required"})
		return
	}

	f, err := h.exportService.ExportTimesheet(dateFrom, dateTo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sendExcelFile(c, f, "timesheet")
}

// ExportLaborExcel принимает тело запроса с результатом расчёта и отдаёт Excel.
// POST /api/v1/export/labor/excel
func (h *ExportHandler) ExportLaborExcel(c *gin.Context) {
	var result models.LaborCalculationResult
	if err := c.ShouldBindJSON(&result); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	buf, err := h.exportService.ExportLaborExcel(&result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := fmt.Sprintf("labor_%s.xlsx", time.Now().Format("2006-01-02"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

// ExportLaborPDF принимает тело запроса с результатом расчёта и отдаёт PDF.
// POST /api/v1/export/labor/pdf
func (h *ExportHandler) ExportLaborPDF(c *gin.Context) {
	var result models.LaborCalculationResult
	if err := c.ShouldBindJSON(&result); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	buf, err := h.exportService.ExportLaborPDF(&result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := fmt.Sprintf("labor_%s.pdf", time.Now().Format("2006-01-02"))
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}
