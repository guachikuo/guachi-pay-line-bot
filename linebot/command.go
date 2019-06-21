package linebot

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sirupsen/logrus"

	"github.com/andy/guachi-pay-line-bot/wallet"
)

var (
	// ErrCommandNotExist occurs when the command is invalid
	ErrCommandNotExist = fmt.Errorf("the command doesn't exist")
	// ErrInvalidArgument occurs when the caller gives invalid argument
	ErrInvalidArgument = fmt.Errorf("invalid argument is found")
)

type response struct {
	text *linebot.TextMessage
}

type command struct {
	// commandIndex defines the index that the main command name is placed at in strings.Split(text, " ")
	commandIndex int
	// argsAllowed defines the number of args that this command is allowed（including the main command)
	// ex: guachi 儲值
	// then the value is 2
	argsAllowed int
	// execFunc is the execution function
	execFunc func(args ...string) (*response, error)
}

var (
	// commands defines the allowed commands
	commands = map[string]command{
		"創建錢包": command{
			commandIndex: 0,
			argsAllowed:  2,
			execFunc:     initWallet,
		},
		"儲值": command{
			commandIndex: 1,
			argsAllowed:  3,
			execFunc:     deposit,
		},
		"花費": command{
			commandIndex: 1,
			argsAllowed:  3,
			execFunc:     spend,
		},
	}
)

func procCommand(text string) (*response, error) {
	texts := strings.Split(text, " ")
	if len(texts) == 0 {
		return nil, ErrCommandNotExist
	}

	found := false
	command := command{}
	args := []string{}
	for i, text := range texts {
		// check that if text fits the command name
		// also, the command name should be in the right place
		if command, ok := commands[text]; ok && i == command.commandIndex {
			found = true
			continue
		}
		args = append(args, text)
	}

	// (1) valid command name is not found
	// (2) lack of arguments
	if !found || (found && len(args) != command.argsAllowed) {
		return nil, ErrCommandNotExist
	}
	return command.execFunc(args...)
}

func initWallet(args ...string) (*response, error) {
	userID := args[0]
	if err := wallet.InitWallet(userID); err != nil {
		logrus.WithField("err", err).Error("wallet.InitWallet failed in initWallet")
		return nil, err
	}
	return &response{
		text: linebot.NewTextMessage("創建 " + userID + " 的錢包成功 !!!"),
	}, nil
}

func deposit(args ...string) (*response, error) {
	userID := args[0]
	moneyCount, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil || (err == nil && moneyCount < int64(0)) {
		return nil, ErrInvalidArgument
	}

	if err := wallet.AtomicPatchWallet(userID, moneyCount); err != nil && err != wallet.ErrWalletNotExist {
		logrus.WithField("err", err).Error("wallet.AtomicPatchWallet failed in deposit")
		return nil, err
	} else if err == wallet.ErrWalletNotExist {
		return &response{
			text: linebot.NewTextMessage("錢包不存在，請先建立錢包唷 ~"),
		}, nil
	}

	return &response{
		text: linebot.NewTextMessage(userID + " 儲值 " + args[1] + " 元成功"),
	}, nil
}

func spend(args ...string) (*response, error) {
	userID := args[0]
	moneyCount, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil || (err == nil && moneyCount < int64(0)) {
		return nil, ErrInvalidArgument
	}

	if err := wallet.AtomicPatchWallet(userID, -1*moneyCount); err != nil && err != wallet.ErrWalletNotExist {
		logrus.WithField("err", err).Error("wallet.AtomicPatchWallet failed in spend")
		return nil, err
	} else if err == wallet.ErrWalletNotExist {
		return &response{
			text: linebot.NewTextMessage("錢包不存在，請先建立錢包唷 ~"),
		}, nil
	}

	return &response{
		text: linebot.NewTextMessage("已紀錄 " + userID + " 花費了 " + args[1] + " 元"),
	}, nil
}
