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

// postbackReceiver is a model that will be placed in data of `lintbot.Postback`
// and we will get the corresponding data if user triggers `postback` event
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
