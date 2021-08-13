package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type SlackEventType struct {
	Type      string  `json:"type"`
	User      string  `json:"user"`
	EventId   string  `json:"event_id"`
	EventTime float64 `json:"event_time"`
}

type SlackActionBody struct {
	Token     string `json:"token"`
	Challenge string `json:"challenge,omitempty"`
	Type      string `json:"type"`
}

func ReactionBot(c *gin.Context) {
	var b SlackActionBody
	if err := c.ShouldBindJSON(&b); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if b.Challenge != "" {
		c.JSON(200, gin.H{
			"challenge": b.Challenge,
		})
		return
	}

}
