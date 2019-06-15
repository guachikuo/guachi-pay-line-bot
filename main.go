package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sirupsen/logrus"
)

var (
	botSrv *linebot.Client
)

func main() {
	// set the standard logger formatter
	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})

	gin.SetMode(gin.ReleaseMode)

	route := gin.Default()
	route.POST("/callback", handleCallback)

	route.Run()
	return
}

func handleCallback(c *gin.Context) {
	if err := parseLinebotCallback(c.Writer, c.Request); err != nil {
		logrus.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"errorMessage": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func initLinebot() error {
	channelSecret := os.Getenv("channelSecret")
	channelAccessToken := os.Getenv("channelAccessToken")
	bot, err := linebot.New(channelSecret, channelAccessToken)
	if err != nil {
		logrus.Error(err)
		return err
	}

	botSrv = bot
	return nil
}

func getLinebotService() (*linebot.Client, error) {
	if botSrv != nil {
		return botSrv, nil
	}

	if err := initLinebot(); err != nil {
		logrus.Error(err)
		return nil, err
	}
	return botSrv, nil
}

func parseLinebotCallback(w http.ResponseWriter, r *http.Request) error {
	botSrv, err := getLinebotService()
	if err != nil {
		logrus.Error(err)
		return err
	}

	events, err := botSrv.ParseRequest(r)
	if err != nil {
		logrus.Error(err)
		return err
	}

	for _, event := range events {
		if event.Type != linebot.EventTypeMessage {
			continue
		}
		message := linebot.NewTextMessage("你好")
		if _, err := botSrv.PushMessage(event.Source.UserID, message).Do(); err != nil {
			logrus.Error(err)
			return err
		}
	}
	return nil
}
