package wallet

import (
	"fmt"
)

var (
	// ErrWalletNotFound occurs when trying to do operation to the non-exist wallet
	ErrWalletNotFound = fmt.Errorf("the wallet doesn't exist")
	// ErrWalletExist occurs when trying to create a wallet that has already been created
	ErrWalletExist = fmt.Errorf("the wallet has already been created")
)

// BalanceLog ...
type BalanceLog struct {
	Amount int64
	Reason string
	// format: 2019/05/20 12:00:00
	Timestamp string
}

// Wallet ...
type Wallet interface {
	// Create creates a new wallet for user
	Create(userID string) error
	// Empty will empty user's balance to zero
	EmptyBalance(userID string) error
	// GetBalance will get balance of a user
	GetBalance(userID string) (int64, error)
	// GetBalanceLogs will get balanceLogs of a user
	GetBalanceLogs(userID string, options ...GetLogsOption) ([]*BalanceLog, error)
	// Deposit will deposit `amount` NTD to user's wallet
	Deposit(userID string, amount int64, reason string) error
	// Spend will spend `amount` NTD from user's wallet
	Spend(userID string, amount int64, reason string) error
}

type getLogsOption struct {
	startTime int64
	endTime   int64
}

// GetLogsOption define optional params of getting balance
type GetLogsOption func(*getLogsOption)

// WithStartTime means getting balance according to the startTime
func WithStartTime(startTime int64) GetLogsOption {
	return func(opt *getLogsOption) {
		opt.startTime = startTime
	}
}

// WithEndTime means getting balance according to the endTime
func WithEndTime(endTime int64) GetLogsOption {
	return func(opt *getLogsOption) {
		opt.endTime = endTime
	}
}

func initOption(options ...GetLogsOption) getLogsOption {
	opt := getLogsOption{}
	for _, f := range options {
		f(&opt)
	}
	return opt
}
