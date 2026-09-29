package export

import "github.com/gin-gonic/gin"

func (h *ExportHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/export/tasks", h.ExportTasks)
	router.GET("/export/projects", h.ExportProjects)
	router.GET("/export/reports", h.ExportReports)
	router.GET("/export/timesheet", h.ExportTimesheet)
	router.POST("/export/labor/excel", h.ExportLaborExcel)
	router.POST("/export/labor/pdf", h.ExportLaborPDF)
}
