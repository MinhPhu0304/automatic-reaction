package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleRootRequest(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
	})
}
