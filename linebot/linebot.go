package linebot

import (
	"net/http"
	"os"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sirupsen/logrus"
)

var (
	botSrv *linebot.Client
)

type linebotSrv struct {
	*linebot.Client
}

func init() {
	if err := initLinebot(); err != nil {
		logrus.WithField("err", err).Fatal("initLinebot failed")
	}
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

func getLinebotService() (*linebotSrv, error) {
	if botSrv == nil {
		if err := initLinebot(); err != nil {
			logrus.WithField("err", err).Error("initLinebot failed in getLinebotService")
			return nil, err
		}
	}
	return &linebotSrv{botSrv}, nil
}

// ParseLinebotCallback parses the callback from line and do corresponding logic
func ParseLinebotCallback(w http.ResponseWriter, r *http.Request) error {
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
			response, err := procCommand(message.Text)
			if err != nil {
				botSrv.ErrReply(event.ReplyToken)
				continue
			}

			if _, err := botSrv.ReplyMessage(event.ReplyToken, response.text).Do(); err != nil {
				logrus.WithField("err", err).Error("ReplyMessage failed in parseLinebotCallback")
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

func (botSrv *linebotSrv) ErrReply(replyToken string) {
	errorText := linebot.NewTextMessage("看不懂你在說什麼，再說一遍")
	botSrv.ReplyMessage(replyToken, errorText).Do()
}
