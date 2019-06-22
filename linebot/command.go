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
	// argsAllowed defines the number of args that this command is allowed
	// ex: guachi 儲值
	// then the value is 1
	argsAllowed int
	// execFunc is the execution function
	execFunc func(im *impl, args ...string) (*response, error)
}

var (
	// commands defines the allowed commands
	commands = map[string]command{
		// ex: 創建錢包 guachi
		"創建錢包": command{
			commandIndex: 0,
			argsAllowed:  1,
			execFunc:     createWallet,
		},
		// ex: guachi 儲值 100
		"儲值": command{
			commandIndex: 1,
			argsAllowed:  2,
			execFunc:     depositMoney,
		},
		// ex: guachi 晚餐 花費 100
		"花費": command{
			commandIndex: 1,
			argsAllowed:  3,
			execFunc:     spendMoney,
		},
	}
)

func (im *impl) procCommand(text string) (*response, error) {
	texts := strings.Split(text, " ")
	if len(texts) == 0 {
		return nil, ErrCommandNotExist
	}

	found := false
	targetCommand := command{}
	args := []string{}
	for i, text := range texts {
		// check that if text fits the command name
		// also, the command name should be in the right place
		if command, ok := commands[text]; ok && i == command.commandIndex {
			found = true
			targetCommand = command
			continue
		}
		args = append(args, text)
	}

	// (1) valid command name is not found
	// (2) lack of arguments
	if !found || (found && len(args) != targetCommand.argsAllowed) {
		return nil, ErrCommandNotExist
	}
	return targetCommand.execFunc(im, args...)
}

func createWallet(im *impl, args ...string) (*response, error) {
	userID := args[0]
	if err := im.wallet.Create(userID); err != nil && err != wallet.ErrWalletExist {
		logrus.WithField("err", err).Error("wallet.CreateWallet failed in createWallet")
		return nil, err
	} else if err == wallet.ErrWalletExist {
		return &response{
			text: linebot.NewTextMessage("錢包已經存在囉 ~"),
		}, nil
	}

	return &response{
		text: linebot.NewTextMessage("創建 " + userID + " 的錢包成功 !!!"),
	}, nil
}

func emptyWallet(im *impl, args ...string) (*response, error) {
	userID := args[0]
	if err := im.wallet.Empty(userID); err != nil && err != wallet.ErrWalletNotExist {
		logrus.WithField("err", err).Error("wallet.EmptyWallet failed in emptyWallet")
		return nil, err
	} else if err == wallet.ErrWalletNotExist {
		return &response{
			text: linebot.NewTextMessage("錢包不存在，請先建立錢包唷 ~"),
		}, nil
	}

	return &response{
		text: linebot.NewTextMessage("已清空 " + userID + " 的錢包 !!!"),
	}, nil
}

func depositMoney(im *impl, args ...string) (*response, error) {
	userID := args[0]
	amount, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil || (err == nil && amount < int64(0)) {
		return nil, ErrInvalidArgument
	}

	if err := im.wallet.Deposit(userID, amount); err != nil && err != wallet.ErrWalletNotExist {
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

func spendMoney(im *impl, args ...string) (*response, error) {
	userID := args[0]
	reason := args[1]
	amount, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil || (err == nil && amount < int64(0)) {
		return nil, ErrInvalidArgument
	}

	if err := im.wallet.Spend(userID, amount, reason); err != nil && err != wallet.ErrWalletNotExist {
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
