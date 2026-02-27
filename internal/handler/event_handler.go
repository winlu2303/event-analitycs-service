package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/event-analytics-service/internal/models"
	"github.com/yourusername/event-analytics-service/internal/service"
)

type EventHandler struct {
	eventService *service.EventService
}

func NewEventHandler(eventService *service.EventService) *EventHandler {
	return &EventHandler{
		eventService: eventService,
	}
}

func (h *EventHandler) TrackEvent(c *gin.Context) {
	var event models.Event

	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Добавляем информацию из запроса
	event.UserAgent = c.GetHeader("User-Agent")
	event.IPAddress = c.ClientIP()

	if err := h.eventService.ProcessEvent(c.Request.Context(), &event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process event"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message":  "Event tracked successfully",
		"event_id": event.ID,
	})
}
