package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/event-analytics-service/internal/models"
	"github.com/yourusername/event-analytics-service/internal/service"
)

type ExportHandler struct {
	exportService *service.ExportService
}

func NewExportHandler(exportService *service.ExportService) *ExportHandler {
	return &ExportHandler{
		exportService: exportService,
	}
}

func (h *ExportHandler) ExportCSV(c *gin.Context) {
	var req models.StatsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data, err := h.exportService.ExportToCSV(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "export failed"})
		return
	}

	filename := fmt.Sprintf("export_%s_%s.csv", req.EventType, time.Now().Format("20060102_150405"))

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "text/csv")
	c.Data(http.StatusOK, "text/csv", data)
}

func (h *ExportHandler) ExportJSON(c *gin.Context) {
	var req models.StatsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data, err := h.exportService.ExportToJSON(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "export failed"})
		return
	}

	filename := fmt.Sprintf("export_%s_%s.json", req.EventType, time.Now().Format("20060102_150405"))

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/json")
	c.Data(http.StatusOK, "application/json", data)
}
