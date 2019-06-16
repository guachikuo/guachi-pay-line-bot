package main

import (
	// "fmt"
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

	logrus.Info("start serving https request")
	return
}

func handleCallback(c *gin.Context) {
	if err := parseLinebotCallback(c.Writer, c.Request); err != nil {
		if err == linebot.ErrInvalidSignature {
			c.JSON(http.StatusUnauthorized, gin.H{})
			return
		}

		logrus.WithField("err", err).Error("parseLinebotCallback failed in handleCallback")
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
		logrus.WithField("err", err).Error("linebot.New failed in initLinebot")
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
		logrus.WithField("err", err).Error("initLinebot failed in getLinebotService")
		return nil, err
	}
	return botSrv, nil
}

func parseLinebotCallback(w http.ResponseWriter, r *http.Request) error {
	botSrv, err := getLinebotService()
	if err != nil {
		logrus.WithField("err", err).Error("getLinebotService failed in parseLinebotCallback")
		return err
	}

	events, err := botSrv.ParseRequest(r)
	if err != nil {
		logrus.WithField("err", err).Error("ParseRequest failed in parseLinebotCallback")
		return err
	}

	for _, event := range events {
		if event.Type != linebot.EventTypeMessage {
			continue
		}

		message := linebot.NewTextMessage("你好")
		if _, err := botSrv.ReplyMessage(event.ReplyToken, message).Do(); err != nil {
			logrus.WithFields(logrus.Fields{
				"err":     err,
				"userID":  event.Source.UserID,
				"roomID":  event.Source.RoomID,
				"groupID": event.Source.GroupID,
			}).Error("ReplyMessage failed in parseLinebotCallback")
			return err
		}

		logrus.WithFields(logrus.Fields{
			"userID":  event.Source.UserID,
			"roomID":  event.Source.RoomID,
			"groupID": event.Source.GroupID,
		}).Info("message is sent successfully")
	}
	return nil
}
