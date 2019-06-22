package linebot

import (
	"fmt"
	"net/http"
	"os"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sirupsen/logrus"

	wl "github.com/andy/guachi-pay-line-bot/wallet"
)

var (
	// ErrInvalidSignature ...
	ErrInvalidSignature = linebot.ErrInvalidSignature
)

type impl struct {
	linebot *linebot.Client
	wallet  wl.Wallet
}

// Linebot ...
type Linebot interface {
	// ParseLinebotCallback parses the callback from line and do corresponding logic
	ParseLinebotCallback(w http.ResponseWriter, r *http.Request) error
}

func initLinebot() (*linebot.Client, error) {
	channelSecret := os.Getenv("channelSecret")
	channelAccessToken := os.Getenv("channelAccessToken")
	bot, err := linebot.New(channelSecret, channelAccessToken)
	if err != nil {
		logrus.WithField("err", err).Error("linebot.New failed in initLinebot")
		if err == linebot.ErrInvalidSignature {
			return nil, ErrInvalidSignature
		}
		return nil, err
	}
	return bot, nil
}

// NewLinebot creates a new Linebot interface
func NewLinebot(
	wallet wl.Wallet,
) (Linebot, error) {
	linebot, err := initLinebot()
	if err != nil {
		return nil, fmt.Errorf("initLinebot failed in NewLinebot")
	}

	return &impl{
		linebot: linebot,
		wallet:  wallet,
	}, nil
}

func (im *impl) errReply(replyToken string) {
	errorText := linebot.NewTextMessage("看不懂你在說什麼，再說一遍")
	im.linebot.ReplyMessage(replyToken, errorText).Do()
}

// ParseLinebotCallback parses the callback from line and do corresponding logic
func (im *impl) ParseLinebotCallback(w http.ResponseWriter, r *http.Request) error {
	events, err := im.linebot.ParseRequest(r)
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
			// get the message, and handle it
			response, err := im.procCommand(message.Text)
			if err != nil {
				im.errReply(event.ReplyToken)
				continue
			}

			// we will reply back accroding to the response after processing the command
			if _, err := im.linebot.ReplyMessage(event.ReplyToken, response.text).Do(); err != nil {
				logrus.WithField("err", err).Error("ReplyMessage failed in parseLinebotCallback")
				return err
			}
		case *linebot.StickerMessage:
			stickerMessage := linebot.NewStickerMessage(getSticker())
			// we will reply back a sticker randomly if we get also a sticker
			if _, err := im.linebot.ReplyMessage(event.ReplyToken, stickerMessage).Do(); err != nil {
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
