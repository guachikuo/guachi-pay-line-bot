package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sirupsen/logrus"

	"github.com/andy/guachi-pay-line-bot/wallet"
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

	logrus.Info("start serving https request")
	route.Run()
	return
}

func handleCallback(c *gin.Context) {
	if err := parseLinebotCallback(c.Writer, c.Request); err != nil {
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

		switch message := event.Message.(type) {
		case *linebot.TextMessage:
			texts := strings.Split(message.Text, " ")
			if !(len(texts) == 2 && texts[0] == "初始化") {
				temp := linebot.NewTextMessage("看不懂你在說什麼 !!!")
				if _, err := botSrv.ReplyMessage(event.ReplyToken, temp).Do(); err != nil {
					logrus.WithFields(logrus.Fields{
						"err":     err,
						"userID":  event.Source.UserID,
						"groupID": event.Source.GroupID,
					}).Error("ReplyMessage failed in parseLinebotCallback")
					return err
				}
				continue
			}

			userID := texts[1]
			if err := wallet.InitWallet(userID); err != nil {
				logrus.WithFields(logrus.Fields{
					"err":     err,
					"userID":  event.Source.UserID,
					"groupID": event.Source.GroupID,
				}).Info("message is sent successfully")
				return err
			}

			temp := linebot.NewTextMessage("初始化 " + userID + " 成功 !!!")
			if _, err := botSrv.ReplyMessage(event.ReplyToken, temp).Do(); err != nil {
				logrus.WithFields(logrus.Fields{
					"err":     err,
					"userID":  event.Source.UserID,
					"groupID": event.Source.GroupID,
				}).Error("ReplyMessage failed in parseLinebotCallback")
				return err
			}
		default:
			logrus.WithFields(logrus.Fields{
				"message": message,
			}).Info("not text mesaage type")
		}
	}
	return nil
}
