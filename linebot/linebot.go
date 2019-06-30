package linebot

import (
	"encoding/json"
	"net/http"

	"github.com/line/line-bot-sdk-go/linebot"
)

var (
	// ErrInvalidSignature ...
	ErrInvalidSignature = linebot.ErrInvalidSignature
)

// postbackReceiver is a model that will be json.Marshal to json string
// and be placed in `data` of `linebot.NewPostbackAction`
// ex: linebot.NewPostbackAction(label, data, text, displayText string) *PostbackAction
//
// when a user triggers a corresponding eventPostback, lintbot will send the data back to the server
// then, we cau use this data to recognize what to do next
//
// ex:
// postbackReceiver {
// 	CommandName: "歷史紀錄",
// 	UserID: "guachi",
// 	TimeRange: &timeRange{
// 		StartTime: int64(12345678),
// 		EndTime: int64(12345679),
// 	}
// }
// when we get this receiver, we could know that
// we should get histories of guachi's wallet beteween
// timestamp(12345678) to timestamp(12345679)
type postbackReceiver struct {
	CommandName string     `json:"command"`
	UserID      string     `json:"userID"`
	TimeRange   *timeRange `json:"timeRange"`
}

type timeRange struct {
	StartTime int64 `json:"startTime"`
	EndTime   int64 `json:"endTime"`
}

func getPostbackReceiver(commandName, userID string) *postbackReceiver {
	return &postbackReceiver{
		CommandName: commandName,
		UserID:      userID,
	}
}

func (receiver *postbackReceiver) withTimeRange(st, et int64) *postbackReceiver {
	receiver.TimeRange = &timeRange{
		StartTime: st,
		EndTime:   et,
	}
	return receiver
}

func (receiver *postbackReceiver) toJSONBytes() []byte {
	bytes, _ := json.Marshal(receiver)
	return bytes
}

// Linebot ...
type Linebot interface {
	// ParseLinebotCallback parses the callback from line and do corresponding logic
	ParseLinebotCallback(w http.ResponseWriter, r *http.Request) error
}
