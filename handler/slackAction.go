package handler

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type SlackEventType struct {
	Type      string `json:"type"`
	User      string `json:"user"`
	EventId   string `json:"event_id"`
	EventTime string `json:"event_ts"`
}

type SlackActionBody struct {
	Token     string          `json:"token" binding:"required"`
	Challenge string          `json:"challenge,omitempty"`
	Type      string          `json:"type" binding:"required"`
	Event     *SlackEventType `json:"event,omitempty"`
}

func ReactionBot(c *gin.Context) {
	var b SlackActionBody
	if err := c.ShouldBindJSON(&b); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if b.Challenge != "" {
		c.JSON(http.StatusOK, gin.H{
			"challenge": b.Challenge,
		})
		return
	}
	if b.Event == nil {
		c.Status(400)
		return
	}
	if b.Event.User == "ULJNGUBS8" {
		if err := reactToMessage(*b.Event); err != nil {
			if hub := sentrygin.GetHubFromContext(c); hub != nil {
				hub.CaptureException(err)
			}
		}
	}
}

type SlackReactionBody struct {
	Token     string `json:"token"`
	Channel   string `json:"channel"`
	Name      string `json:"name"`
	Timestamp string `json:"timestamp"`
}

func reactToMessage(e SlackEventType) error {
	postBody, err := json.Marshal(&SlackReactionBody{
		Token:     os.Getenv("SLACK_BOT_TOKEN"),
		Channel:   os.Getenv("CHANNEL_ID"),
		Name:      "fire",
		Timestamp: e.EventTime,
	})
	if err != nil {
		return errors.Wrap(err, "Failed to marshal")
	}
	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post("https://slack.com/api/reactions.add", "application/json", responseBody)
	if err != nil {
		return errors.Wrap(err, "Failed to send request")
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("Unexpected status")
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "Failed to read response body")
	}
	dst := &bytes.Buffer{}
	json.Indent(dst, body, "", "  ")
	log.Println(dst.String())
	return nil
}
