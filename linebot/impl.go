package linebot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sirupsen/logrus"

	"github.com/andy/guachi-pay-line-bot/base"
	wl "github.com/andy/guachi-pay-line-bot/wallet"
)

type impl struct {
	linebot *linebot.Client
	wallet  wl.Wallet
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

func (im *impl) helpReply(replyToken, commandName string) error {
	_, err := im.linebot.ReplyMessage(replyToken, linebot.NewTextMessage(getHelpReplyText(commandName))).Do()
	return err
}

func (im *impl) errorReply(replyToken string) error {
	text := "系統錯誤，請重新試試 (moon apology)"
	_, err := im.linebot.ReplyMessage(replyToken, linebot.NewTextMessage(text)).Do()
	return err
}

func (im *impl) handleEventTypePostback(replyToken string, postback *linebot.Postback) error {
	postbackReceiver := postbackReceiver{}
	if err := json.Unmarshal([]byte(postback.Data), &postbackReceiver); err != nil {
		logrus.WithField("err", err).Error("json.Unmarshal failed in handleEventTypePostback")
		im.errorReply(replyToken)
		return err
	}

	userID := postbackReceiver.UserID
	commandName := postbackReceiver.CommandName
	text := fmt.Sprintf("%s %s", commandName, userID)
	if commandName == commandGetBalanceLogs && postbackReceiver.TimeRange != nil {
		startTimeStr := base.ParseToyyymmdd(postbackReceiver.TimeRange.StartTime)
		endTimeStr := base.ParseToyyymmdd(postbackReceiver.TimeRange.EndTime)

		text = fmt.Sprintf("%s %s %s %s", commandName, userID, startTimeStr, endTimeStr)
	}

	// modify to the valid message, and handle it
	response, err := im.procCommand(text)
	if err != nil {
		im.errorReply(replyToken)
		return err
	}

	// we will reply back accroding to the response after processing the command
	if _, err := im.linebot.ReplyMessage(replyToken, response.messages...).Do(); err != nil {
		logrus.WithField("err", err).Error("ReplyMessage failed in handleEventTypePostback")
		return err
	}
	return nil
}

func (im *impl) handleEventTypeMessage(replyToken string, messageInterface linebot.Message) error {
	switch message := messageInterface.(type) {
	case *linebot.TextMessage:
		texts := strings.Split(message.Text, " ")
		if len(texts) > 0 && texts[0] == helpCommandName {
			// we will reply back description of commands if user uses helpCommandName
			commandName := ""
			if len(texts) > 1 {
				commandName = texts[1]
			}
			if err := im.helpReply(replyToken, commandName); err != nil {
				logrus.WithField("err", err).Error("im.helpReply failed in handleEventTypeMessage")
				return err
			}
			return nil
		} else if len(texts) == 1 && im.wallet.IsWalletExist(texts[0]) {
			userID := texts[0]

			message := linebot.NewTemplateMessage("欲知詳情", linebot.NewCarouselTemplate(
				linebot.NewCarouselColumn("https://upload.cc/i1/2019/06/30/MRH0J9.jpg", "記帳/查詢", "選擇一個想做的事吧!",
					//linebot.NewMessageAction("記帳", "help 歷史紀錄"),
					linebot.NewPostbackAction("餘額查詢",
						string(getPostbackReceiver(commandGetBalance, userID).toJSONBytes()),
						"", "",
					),
					linebot.NewPostbackAction("歷史紀錄",
						string(getPostbackReceiver(commandGetBalanceLogs, userID).toJSONBytes()),
						"", "",
					),
				),
				linebot.NewCarouselColumn("https://upload.cc/i1/2019/06/30/41YH7A.jpeg", "錢包", "選擇一個想做的事吧!",
					linebot.NewPostbackAction("清空錢包",
						string(getPostbackReceiver(commandEmptyWallet, userID).toJSONBytes()),
						"", "",
					),
					linebot.NewPostbackAction("刪除錢包",
						string(getPostbackReceiver(commandDeleteWallet, userID).toJSONBytes()),
						"", "",
					),
				),
			))

			// we will reply back accroding to the response after processing the command
			if _, err := im.linebot.ReplyMessage(replyToken, message).Do(); err != nil {
				logrus.WithField("err", err).Error("ReplyMessage failed in handleEventTypeMessage")
				return err
			}
		}

		// get the message, and handle it
		response, err := im.procCommand(message.Text)
		if err != nil {
			im.helpReply(replyToken, "")
			return nil
		}

		// we will reply back accroding to the response after processing the command
		if _, err := im.linebot.ReplyMessage(replyToken, response.messages...).Do(); err != nil {
			logrus.WithField("err", err).Error("ReplyMessage failed in handleEventTypeMessage")
			return err
		}
	case *linebot.StickerMessage:
		stickerMessage := linebot.NewStickerMessage(getSticker())
		// we will reply back a sticker randomly if we get also a sticker
		if _, err := im.linebot.ReplyMessage(replyToken, stickerMessage).Do(); err != nil {
			logrus.WithField("err", err).Error("ReplyMessage failed in handleEventTypeMessage")
			return err
		}
	default:
		im.helpReply(replyToken, "")
	}
	return nil
}

// ParseLinebotCallback parses the callback from line and do corresponding logic
func (im *impl) ParseLinebotCallback(w http.ResponseWriter, r *http.Request) error {
	events, err := im.linebot.ParseRequest(r)
	if err != nil {
		logrus.WithField("err", err).Error("ParseRequest failed in ParseLinebotCallback")
		return err
	}

	for _, event := range events {
		switch event.Type {
		case linebot.EventTypeMessage:
			if err := im.handleEventTypeMessage(event.ReplyToken, event.Message); err != nil {
				logrus.WithField("err", err).Error("handleEventTypeMessage failed in ParseLinebotCallback")
			}
		case linebot.EventTypePostback:
			if err := im.handleEventTypePostback(event.ReplyToken, event.Postback); err != nil {
				logrus.WithField("err", err).Error("handleEventTypePostback failed in ParseLinebotCallback")
			}
		default:
		}
	}
	return nil
}
