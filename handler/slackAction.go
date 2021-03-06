package handler

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
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
	if b.Event.User == os.Getenv("USER_SLACK_ID") {
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

type SlackReactionResponse struct {
	Error string `json:"error"`
	Ok    bool   `json:"ok"`
}

func reactToMessage(e SlackEventType) error {
	postBody, err := json.Marshal(&SlackReactionBody{
		Token:     os.Getenv("SLACK_BOT_TOKEN"),
		Channel:   os.Getenv("CHANNEL_ID"),
		Name:      os.Getenv("EMOJI_NAME"),
		Timestamp: e.EventTime,
	})
	if err != nil {
		return errors.Wrap(err, "Failed to marshal")
	}
	req, err := http.NewRequest("POST", "https://slack.com/api/reactions.add", bytes.NewBuffer(postBody))
	if err != nil {
		return errors.Wrap(err, "Failed to post reaction to slack")
	}

	req.Header.Add("Authorization", "Bearer "+os.Getenv("SLACK_BOT_TOKEN"))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "Failed to send request")
	}
	if response.StatusCode != http.StatusOK {
		return errors.New("Unexpected status")
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return errors.Wrap(err, "Failed to read response body")
	}
	var result *SlackReactionResponse
	err = json.Unmarshal(body, result)
	if err != nil {
		return errors.Wrap(err, "Failed to decode json response")
	}
	if !result.Ok {
		return errors.New("Error from slack with content: " + result.Error)
	}
	return nil
}
