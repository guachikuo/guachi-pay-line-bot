package linebot

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sirupsen/logrus"

	"github.com/andy/guachi-pay-line-bot/base"
	"github.com/andy/guachi-pay-line-bot/wallet"
)

const (
	helpDescGeneral = `感謝您傳送訊息給 guachi pay 😀
	
您還沒有錢包嗎? 可以參考這個指令唷 ~
💰 新增錢包【錢包名稱】

如果已經有，請輸入您的【錢包名稱】，我們將為您服務 🙂

【想要自己輸入指令也可以! 請輸入以下指令來取得相關操作】
💵 錢包:
1. 清空錢包【錢包名稱】
2. 刪除錢包【錢包名稱】
	
🔎 查詢:
1. 查詢餘額【錢包名稱】
2. 歷史紀錄【錢包名稱】

📋 記帳:
1.【錢包名稱】【原因】+【多少元】
2.【錢包名稱】【原因】-【多少元】`
)

type response struct {
	messages []linebot.SendingMessage
}

type command struct {
	// commandIndex defines the index that the main command name is placed at in strings.Split(text, " ")
	commandIndex int
	// argsAllowed defines the number of args that this command is allowed
	// ex: guachi 儲值
	// then the value is 1
	argsAllowed int
	// optionalArgsAllowed defines the number of optional args that this command is allowed
	// ex: 歷史紀錄 guachi 2019/05/20 2019/06/20
	// because the second date is optional, so the value of `optionalArgsAllowed` should be 1
	optionalArgsAllowed int
	// execFunc is the execution function
	execFunc func(im *impl, args ...string) (*response, error)
	// helpDesc describe how this command works, and it will display when the user calls `!help`
	helpDesc string
}

const (
	commandHelp           = "help"
	commandCreateWallet   = "新增錢包"
	commandDeleteWallet   = "刪除錢包"
	commandEmptyWallet    = "清空錢包"
	commandGetBalance     = "查詢餘額"
	commandGetBalanceLogs = "歷史紀錄"
	commandDepositMoney1  = "儲值"
	commandDepositMoney2  = "+"
	commandSpendMoney1    = "花費"
	commandSpendMoney2    = "-"
)

var (
	// ErrCommandNotExist occurs when the command is invalid
	ErrCommandNotExist = fmt.Errorf("the command doesn't exist")
	// ErrInvalidArgument occurs when the caller gives invalid argument
	ErrInvalidArgument = fmt.Errorf("invalid argument is found")

	// commands defines the allowed commands
	commands = map[string]command{
		// ex: 新增錢包 guachi
		commandCreateWallet: command{
			commandIndex: 0,
			argsAllowed:  1,
			execFunc:     (*impl).createWallet,
			helpDesc:     "新增錢包【錢包名稱】\nex: 新增錢包 guachi",
		},

		// ex: 刪除錢包 guachi
		commandDeleteWallet: command{
			commandIndex: 0,
			argsAllowed:  1,
			execFunc:     (*impl).deleteWallet,
			helpDesc:     "刪除錢包【錢包名稱】\nex: 刪除錢包 guachi",
		},

		// ex: 清空錢包 guachi
		commandEmptyWallet: command{
			commandIndex: 0,
			argsAllowed:  1,
			execFunc:     (*impl).emptyBalance,
			helpDesc:     "請輸入:\n清空錢包【錢包名稱】\n\nex: 清空錢包 guachi",
		},

		// ex: 查詢餘額 guachi
		commandGetBalance: command{
			commandIndex: 0,
			argsAllowed:  1,
			execFunc:     (*impl).getBalance,
			helpDesc:     "請輸入:\n查詢餘額【錢包名稱】\n\nex: 查詢餘額 guachi",
		},

		// ex: 歷史紀錄 guachi
		// ex: 歷史紀錄 guachi 2019/05/20 2019/06/20
		commandGetBalanceLogs: command{
			commandIndex:        0,
			argsAllowed:         1,
			optionalArgsAllowed: 2,
			execFunc:            (*impl).getBalanceLogs,
			helpDesc:            "請輸入:\n歷史紀錄【錢包名稱】【起日】【迄日】\n\nex: 歷史紀錄 guachi 2019/05/20 2019/06/20",
		},

		// ex: guachi 中樂透 (儲值 or +) 100
		commandDepositMoney1: command{
			commandIndex: 2,
			argsAllowed:  3,
			execFunc:     (*impl).depositMoney,
			helpDesc:     "請輸入:\n【錢包名稱】【原因】+【多少錢】\n\nex: guachi 中樂透 + 100",
		},

		// ex: guachi 中樂透 (儲值 or +) 100
		commandDepositMoney2: command{
			commandIndex: 2,
			argsAllowed:  3,
			execFunc:     (*impl).depositMoney,
			helpDesc:     "請輸入:\n【錢包名稱】【原因】+【多少錢】\n\nex: guachi 中樂透 + 100",
		},

		// ex: guachi 晚餐 (花費 or -) 100
		commandSpendMoney1: command{
			commandIndex: 2,
			argsAllowed:  3,
			execFunc:     (*impl).spendMoney,
			helpDesc:     "請輸入:\n【錢包名稱】【原因】-【多少錢】\n\nex: guachi 晚餐 - 100",
		},

		// ex: guachi 晚餐 (花費 or -) 100
		commandSpendMoney2: command{
			commandIndex: 2,
			argsAllowed:  3,
			execFunc:     (*impl).spendMoney,
			helpDesc:     "請輸入:\n【錢包名稱】【原因】-【多少錢】\n\nex: guachi 晚餐 - 100",
		},
	}
)

func getWalletNotFoundResponse() *response {
	return &response{
		messages: []linebot.SendingMessage{
			linebot.NewTextMessage("錢包不存在，請先建立錢包"),
		},
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
		if command, ok := commands[text]; ok && i == command.commandIndex {
			found = true
			targetCommand = command
			continue
		}
		args = append(args, text)
	}

	// (1) valid command name is not found
	// (2) lack of arguments
	validArgs := len(args) == targetCommand.argsAllowed || len(args) == targetCommand.argsAllowed+targetCommand.optionalArgsAllowed
	if !found || (found && !validArgs) {
		return nil, ErrCommandNotExist
	}
	return targetCommand.execFunc(im, args...)
}

func (im *impl) createWallet(args ...string) (*response, error) {
	userID := args[0]
	if err := im.wallet.Create(userID); err != nil && err != wallet.ErrWalletExist {
		logrus.WithField("err", err).Error("wallet.Create failed in createWallet")
		return nil, err
	} else if err == wallet.ErrWalletExist {
		return &response{
			messages: []linebot.SendingMessage{
				linebot.NewTextMessage("錢包已經存在囉"),
			},
		}, nil
	}

	return &response{
		messages: []linebot.SendingMessage{
			linebot.NewTextMessage("建立 " + userID + " 的錢包成功"),
		},
	}, nil
}

func (im *impl) deleteWallet(args ...string) (*response, error) {
	userID := args[0]
	if err := im.wallet.Delete(userID); err == wallet.ErrWalletNotFound {
		return getWalletNotFoundResponse(), nil
	} else if err != nil {
		logrus.WithField("err", err).Error("wallet.Delete failed in deleteWallet")
		return nil, err
	}

	return &response{
		messages: []linebot.SendingMessage{
			linebot.NewTextMessage("刪除 " + userID + " 的錢包成功"),
		},
	}, nil
}

func (im *impl) emptyBalance(args ...string) (*response, error) {
	userID := args[0]
	if err := im.wallet.EmptyBalance(userID); err == wallet.ErrWalletNotFound {
		return getWalletNotFoundResponse(), nil
	} else if err != nil {
		logrus.WithField("err", err).Error("wallet.EmptyBalance failed in emptyBalance")
		return nil, err
	}

	return &response{
		messages: []linebot.SendingMessage{
			linebot.NewTextMessage("已清空 " + userID + " 的錢包"),
		},
	}, nil
}

func (im *impl) getBalance(args ...string) (*response, error) {
	userID := args[0]
	balance, err := im.wallet.GetBalance(userID)
	if err == wallet.ErrWalletNotFound {
		return getWalletNotFoundResponse(), nil
	} else if err != nil {
		logrus.WithField("err", err).Error("wallet.GetBalance failed in getBalance")
		return nil, err
	}

	return &response{
		messages: []linebot.SendingMessage{
			linebot.NewTextMessage("目前餘額 " + strconv.FormatInt(balance, 10) + "元"),
		},
	}, nil
}

func (im *impl) getBalanceLogsTemplateMessage(userID string) (*response, error) {
	if !im.wallet.IsWalletExist(userID) {
		return getWalletNotFoundResponse(), nil
	}

	location, _ := time.LoadLocation("Asia/Taipei")
	now := time.Now()
	now = now.In(location)
	todayStartTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location).Unix()
	todayEndTime := todayStartTime + int64(86400)

	postbackReceiver := getPostbackReceiver(commandGetBalanceLogs, userID)
	bytesToday := postbackReceiver.withTimeRange(todayStartTime, todayEndTime).toJSONBytes()
	bytesLast3Days := postbackReceiver.withTimeRange(todayStartTime-int64(2*86400), todayEndTime).toJSONBytes()
	bytesLast7Days := postbackReceiver.withTimeRange(todayStartTime-int64(6*86400), todayEndTime).toJSONBytes()

	return &response{
		messages: []linebot.SendingMessage{
			linebot.NewTemplateMessage("欲知詳情", linebot.NewButtonsTemplate(
				"https://upload.cc/i1/2019/06/30/msrwg8.jpg", "歷史紀錄", "請選擇你想要查詢的日期",
				linebot.NewPostbackAction("今天", string(bytesToday), "", ""),
				linebot.NewPostbackAction("近3天", string(bytesLast3Days), "", ""),
				linebot.NewPostbackAction("近7天", string(bytesLast7Days), "", ""),
				linebot.NewMessageAction("自訂", "help 歷史紀錄"),
			)),
		},
	}, nil
}

func (im *impl) getBalanceLogs(args ...string) (*response, error) {
	userID := args[0]

	// if the command looks like `歷史紀錄 guachi`
	// we will display `TemplateMessage` for users, and let them choose what to do next
	if len(args) == 1 {
		return im.getBalanceLogsTemplateMessage(userID)
	}

	startTime, err := base.ParseToTimestamp(args[1])
	if err != nil {
		logrus.WithField("err", err).Error("parseToTimestamp failed in getBalanceLogs")
		return nil, ErrInvalidArgument
	}
	endTime, err := base.ParseToTimestamp(args[2])
	if err != nil {
		logrus.WithField("err", err).Error("parseToTimestamp failed in getBalanceLogs")
		return nil, ErrInvalidArgument
	}

	options := []wallet.GetLogsOption{
		wallet.WithStartTime(startTime),
		wallet.WithEndTime(endTime),
	}

	balanceLogs, err := im.wallet.GetBalanceLogs(userID, options...)
	if err != nil {
		logrus.WithField("err", err).Error("wallet.GetBalanceLogs failed in getBalanceLogs")
		return nil, err
	}

	texts := ""
	for i, balanceLog := range balanceLogs {
		texts += balanceLog.Timestamp + " " + balanceLog.Reason + " " + strconv.FormatInt(balanceLog.Amount, 10) + "元"
		if i != len(balanceLogs)-1 {
			texts += "\n"
		}
	}

	return &response{
		messages: []linebot.SendingMessage{
			linebot.NewTextMessage("歷史紀錄 :\n" + texts),
		},
	}, nil
}

func (im *impl) depositMoney(args ...string) (*response, error) {
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
		messages: []linebot.SendingMessage{
			linebot.NewTextMessage(line1 + "\n" + line2 + "\n---\n" + line3),
		},
	}, nil
}

func (im *impl) spendMoney(args ...string) (*response, error) {
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
		messages: []linebot.SendingMessage{
			linebot.NewTextMessage(line1 + "\n" + line2 + "\n---\n" + line3),
		},
	}, nil
}
