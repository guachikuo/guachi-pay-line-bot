package linebot

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sirupsen/logrus"

	"github.com/andy/guachi-pay-line-bot/wallet"
)

const (
	helpCommandName = "!help"
	timeTemplate    = "2016/01/01"
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
	// helpDesc describe how this command works, and it will display when the user calls `!help`
	helpDesc string
}

type commandName string

const (
	commandCreateWallet   commandName = "創建錢包"
	commandEmptyWallet    commandName = "清空錢包"
	commandGetBalance     commandName = "查詢餘額"
	commandGetBalanceLogs commandName = "歷史紀錄"
	commandDepositMoney   commandName = "+"
	commandSpendMoney     commandName = "-"
)

var (
	// ErrCommandNotExist occurs when the command is invalid
	ErrCommandNotExist = fmt.Errorf("the command doesn't exist")
	// ErrInvalidArgument occurs when the caller gives invalid argument
	ErrInvalidArgument = fmt.Errorf("invalid argument is found")

	// commandDisplayedInHelp defines what commands could be displayed in `!help` and their orders
	commandDisplayedInHelp = []commandName{
		commandCreateWallet,
		commandEmptyWallet,
		commandGetBalance,
		commandGetBalanceLogs,
		commandDepositMoney,
		commandSpendMoney,
	}

	// commands defines the allowed commands
	commands = map[commandName]command{
		// ex: guachi 創建錢包
		commandCreateWallet: command{
			commandIndex: 1,
			argsAllowed:  1,
			execFunc:     createWallet,
			helpDesc:     "<錢包名稱> 創建錢包\nex: guachi 創建錢包",
		},

		// ex: guachi 清空錢包
		commandEmptyWallet: command{
			commandIndex: 1,
			argsAllowed:  1,
			execFunc:     emptyBalance,
			helpDesc:     "<錢包名稱> 清空錢包\nex: guachi 清空錢包",
		},

		// ex: 查詢餘額 guachi
		commandGetBalance: command{
			commandIndex: 0,
			argsAllowed:  1,
			execFunc:     getBalance,
			helpDesc:     "查詢餘額 <錢包名稱>\nex: 查詢餘額 guachi",
		},

		// ex: 歷史紀錄 guachi 2019/05/20 2019/05/21
		commandGetBalanceLogs: command{
			commandIndex: 0,
			argsAllowed:  2,
			execFunc:     getBalanceLogs,
			helpDesc:     "歷史紀錄 <錢包名稱> <開始時間> [結束時間] \n時間格式: 2019/05/20\nex: 歷史紀錄 guachi 2019/05/20 2019/05/21",
		},

		// ex: guachi 中樂透 + 100
		commandDepositMoney: command{
			commandIndex: 2,
			argsAllowed:  3,
			execFunc:     depositMoney,
			helpDesc:     "<錢包名稱> <原因> + <多少錢>\nex: guachi 中樂透 + 100",
		},

		// ex: guachi 晚餐 - 100
		commandSpendMoney: command{
			commandIndex: 2,
			argsAllowed:  3,
			execFunc:     spendMoney,
			helpDesc:     "<錢包名稱> <原因> - <多少錢>\nex: guachi 晚餐 - 100",
		},
	}
)

func getCommands() string {
	text := "<欄位1> : 欄位必填\n[欄位2] : 欄位選填\n"
	for i, command := range commandDisplayedInHelp {
		text += strconv.FormatInt(int64(i+1), 10) + ". " + commands[command].helpDesc + "\n"
		if i != len(commandDisplayedInHelp)-1 {
			text += "\n"
		}
	}
	return text
}

func getWalletNotFoundResponse() *response {
	return &response{
		text: linebot.NewTextMessage("錢包不存在，請先建立錢包"),
	}
}

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
		if command, ok := commands[commandName(text)]; ok && i == command.commandIndex {
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
		logrus.WithField("err", err).Error("wallet.Create failed in createWallet")
		return nil, err
	} else if err == wallet.ErrWalletExist {
		return &response{
			text: linebot.NewTextMessage("錢包已經存在囉"),
		}, nil
	}

	return &response{
		text: linebot.NewTextMessage("創建 " + userID + " 的錢包成功"),
	}, nil
}

func emptyBalance(im *impl, args ...string) (*response, error) {
	userID := args[0]
	if err := im.wallet.EmptyBalance(userID); err == wallet.ErrWalletNotFound {
		return getWalletNotFoundResponse(), nil
	} else if err != nil {
		logrus.WithField("err", err).Error("wallet.EmptyBalance failed in emptyBalance")
		return nil, err
	}

	return &response{
		text: linebot.NewTextMessage("已清空 " + userID + " 的錢包"),
	}, nil
}

func getBalance(im *impl, args ...string) (*response, error) {
	userID := args[0]
	balance, err := im.wallet.GetBalance(userID)
	if err == wallet.ErrWalletNotFound {
		return getWalletNotFoundResponse(), nil
	} else if err != nil {
		logrus.WithField("err", err).Error("wallet.GetBalance failed in getBalance")
		return nil, err
	}

	return &response{
		text: linebot.NewTextMessage("目前餘額 " + strconv.FormatInt(balance, 10) + "元"),
	}, nil
}

// format: 2019/05/20 12:00
func parseToTimestamp(timestampStr string) (int64, error) {
	location, _ := time.LoadLocation("Asia/Taipei")
	time, err := time.Parse(timeTemplate, timestampStr)
	if err != nil {
		logrus.WithField("err", err).Error("time.Parse failed in parseToTimestamp")
		return int64(0), err
	}
	return time.In(location).Unix(), nil
}

func getBalanceLogs(im *impl, args ...string) (*response, error) {
	userID := args[0]

	startTime, err := parseToTimestamp(args[1])
	if err != nil {
		logrus.WithField("err", err).Error("parseToTimestamp failed in getBalanceLogs")
		return nil, ErrInvalidArgument
	}

	options := []wallet.GetLogsOption{wallet.WithStartTime(startTime)}
	if len(args) > 2 {
		endTime, err := parseToTimestamp(args[2])
		if err != nil {
			logrus.WithField("err", err).Error("parseToTimestamp failed in getBalanceLogs")
			return nil, ErrInvalidArgument
		}
		options = append(options, wallet.WithEndTime(endTime))
	}

	balanceLogs, err := im.wallet.GetBalanceLogs(userID, options...)
	if err != nil {
		logrus.WithField("err", err).Error("wallet.GetBalanceLogs failed in getBalanceLogs")
		return nil, err
	}

	texts := ""
	for _, balanceLog := range balanceLogs {
		texts += balanceLog.Timestamp + " " + balanceLog.Reason + " " + strconv.FormatInt(balanceLog.Amount, 10) + "元\n"
	}

	return &response{
		text: linebot.NewTextMessage("歷史紀錄:\n" + texts),
	}, nil
}

func depositMoney(im *impl, args ...string) (*response, error) {
	userID := args[0]

	// get original balance first
	originalBalance, err := im.wallet.GetBalance(userID)
	if err == wallet.ErrWalletNotFound {
		return getWalletNotFoundResponse(), nil
	} else if err != nil {
		logrus.WithField("err", err).Error("wallet.GetBalance failed in depositMoney")
		return nil, err
	}

	reason := args[1]
	amount, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil || (err == nil && amount < int64(0)) {
		return nil, ErrInvalidArgument
	}

	if err := im.wallet.Deposit(userID, amount, reason); err == wallet.ErrWalletNotFound {
		return getWalletNotFoundResponse(), nil
	} else if err != nil {
		logrus.WithField("err", err).Error("wallet.Deposit failed in depositMoney")
		return nil, err
	}

	// get resulted balance after finish deposting
	// as we have check if wallet does exist above, we don't need to specially handle here
	resultedBalance, err := im.wallet.GetBalance(userID)
	if err != nil {
		logrus.WithField("err", err).Error("wallet.GetBalance failed in depositMoney")
		return nil, err
	}

	line1 := "上次餘額 " + strconv.FormatInt(originalBalance, 10) + "元"
	line2 := reason + " +" + args[2] + "元"
	line3 := "目前餘額 " + strconv.FormatInt(resultedBalance, 10) + "元"
	return &response{
		text: linebot.NewTextMessage(line1 + "\n" + line2 + "\n---\n" + line3),
	}, nil
}

func spendMoney(im *impl, args ...string) (*response, error) {
	userID := args[0]

	// get original balance first
	originalBalance, err := im.wallet.GetBalance(userID)
	if err == wallet.ErrWalletNotFound {
		return getWalletNotFoundResponse(), nil
	} else if err != nil {
		logrus.WithField("err", err).Error("wallet.GetBalance failed in spendMoney")
		return nil, err
	}

	reason := args[1]
	amount, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil || (err == nil && amount < int64(0)) {
		return nil, ErrInvalidArgument
	}

	if err := im.wallet.Spend(userID, amount, reason); err == wallet.ErrWalletNotFound {
		return getWalletNotFoundResponse(), nil
	} else if err != nil {
		logrus.WithField("err", err).Error("wallet.Spend failed in spendMoney")
		return nil, err
	}

	// get resulted balance after finish deposting
	// as we have check if wallet does exist above, we don't need to specially handle here
	resultedBalance, err := im.wallet.GetBalance(userID)
	if err != nil {
		logrus.WithField("err", err).Error("wallet.GetBalance failed in spendMoney")
		return nil, err
	}

	line1 := "上次餘額 " + strconv.FormatInt(originalBalance, 10) + "元"
	line2 := reason + " -" + args[2] + "元"
	line3 := "目前餘額 " + strconv.FormatInt(resultedBalance, 10) + "元"
	return &response{
		text: linebot.NewTextMessage(line1 + "\n" + line2 + "\n---\n" + line3),
	}, nil
}
