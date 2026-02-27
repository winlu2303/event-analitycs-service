package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/event-analytics-service/internal/models"
	"github.com/yourusername/event-analytics-service/internal/service"
)

type StatsHandler struct {
	statsService *service.StatsService
}

func NewStatsHandler(statsService *service.StatsService) *StatsHandler {
	return &StatsHandler{
		statsService: statsService,
	}
}

func (h *StatsHandler) GetStatistics(c *gin.Context) {
	var req models.StatsRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Валидация обязательных полей
	if req.EventType == "" || req.StartDate == "" || req.EndDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "event_type, start_date, end_date are required"})
		return
	}

	stats, err := h.statsService.GetEventStatistics(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get statistics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statistics":   stats,
		"total_events": len(stats),
	})
}

func (h *StatsHandler) GetConversionRate(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	rate, err := h.statsService.CalculateConversionRate(c.Request.Context(), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate conversion rate"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"conversion_rate": rate,
		"period": gin.H{
			"start": startDate,
			"end":   endDate,
		},
	})
}
