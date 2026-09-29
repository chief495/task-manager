package report

import "github.com/gin-gonic/gin"

func (h *ReportHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/reports", h.CreateReport)
	router.GET("/reports/:id", h.GetReport)
	router.PUT("/reports/:id", h.UpdateReport)
	router.DELETE("/reports/:id", h.DeleteReport)
	router.GET("/reports", h.ListReports)
	router.GET("/reports/by-user", h.ListReportsByUser)
	router.GET("/reports/by-project", h.ListReportsByProject)
}
