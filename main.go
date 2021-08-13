package main

import (
	"log"
	"os"

	"github.com/Minhphu0304/slack-reaction-bot/handler"
	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.New()
	router.Use(gin.Logger())

	router.GET("/", handler.HandleRootRequest)

	router.POST("/reaction-bot", handler.ReactionBot)

	router.Run(":" + port)
}
