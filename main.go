package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sirupsen/logrus"

	lb "github.com/andy/guachi-pay-line-bot/linebot"
)

func main() {
	// set the standard logger formatter
	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})

	gin.SetMode(gin.ReleaseMode)

	route := gin.Default()
	route.POST("/callback", handleCallback)

	logrus.Info("start serving https request")
	route.Run()
	return
}

func handleCallback(c *gin.Context) {
	if err := lb.ParseLinebotCallback(c.Writer, c.Request); err != nil {
		if err == linebot.ErrInvalidSignature {
			c.JSON(http.StatusUnauthorized, gin.H{})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"errorMessage": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
