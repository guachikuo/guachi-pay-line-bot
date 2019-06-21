package linebot

import (
	"fmt"
	"strings"

	"github.com/line/line-bot-sdk-go/linebot"

	"github.com/andy/guachi-pay-line-bot/wallet"
)

var (
	// ErrCommandNotExist occurs when the command is invalid
	ErrCommandNotExist = fmt.Errorf("the command doesn't exist")
)

type response struct {
	text *linebot.TextMessage
}

type command struct {
	// argsAllowed defines the number of args that this command is allowed
	argsAllowed int
	// execFunc is the execution function
	execFunc func(args ...string) (*response, error)
}

var (
	// commands defines the allowed commands
	commands = map[string]command{
		"創建錢包": command{
			argsAllowed: 1,
			execFunc:    initWallet,
		},
	}
)

func procCommand(text string) (*response, error) {
	texts := strings.Split(text, " ")
	if len(texts) == 0 {
		return nil, ErrCommandNotExist
	}
	commandName := texts[0]

	args := []string{}
	for i := 1; i < len(texts); i++ {
		args = append(args, texts[i])
	}

	command, ok := commands[commandName]
	if !ok || ok && len(args) != command.argsAllowed {
		return nil, ErrCommandNotExist
	}
	return command.execFunc(args...)
}

func initWallet(args ...string) (*response, error) {
	userID := args[0]
	if err := wallet.InitWallet(userID); err != nil {
		return nil, err
	}
	return &response{
		text: linebot.NewTextMessage("創建 " + userID + " 的錢包成功 !!!"),
	}, nil
}
