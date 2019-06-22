package wallet

import (
	"database/sql"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/andy/guachi-pay-line-bot/db"
)

const (
	checkIfWalletExists = `SELECT 1 FROM "UsersWallet" WHERE "userID" = $1`
	createWallet        = `
		INSERT INTO "UsersWallet" ("userID", "balance")
		VALUES ($1, $2);
	`
	patchWallet       = `UPDATE "UsersWallet" SET balance = $1 WHERE "userID" = $2;`
	atomicPatchWallet = `UPDATE "UsersWallet" SET balance = balance + $1 WHERE "userID" = $2;`
)

var (
	// ErrWalletNotExist occurs when trying to do operation to the non-exist wallet
	ErrWalletNotExist = fmt.Errorf("the wallet doesn't exist")
	// ErrWalletExist occurs when trying to create a wallet that has already been created
	ErrWalletExist = fmt.Errorf("the wallet has already been created")
)

type impl struct {
	db *sql.DB
}

// Wallet ...
type Wallet interface {
	// Create creates a new wallet for user
	Create(userID string) error
	// Empty will empty user's wallet
	Empty(userID string) error
	// Deposit will deposit `amount` NTD to user's wallet
	Deposit(userID string, amount int64) error
	// Spend will spend `amount` NTD from user's wallet
	Spend(userID string, amount int64, reason string) error
}

// NewWallet creates a new Wallet interface
func NewWallet() (Wallet, error) {
	dbSrv, err := db.NewPostgresSrv()
	if err != nil {
		return nil, fmt.Errorf("db.GetPostgresSrv failed in NewWallet")
	}

	return &impl{
		db: dbSrv,
	}, nil
}

// Create creates a new wallet for user
func (im *impl) Create(userID string) error {
	result, err := im.db.Exec(checkIfWalletExists, userID)
	if err != nil {
		logrus.WithField("err", err).Error("im.db.Exec checkIfWalletExists failed in InitWallet")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logrus.WithField("err", err).Error("result.RowsAffected failed in InitWallet")
		return err
	}

	if rowsAffected == int64(1) {
		return ErrWalletExist
	}

	result, err = im.db.Exec(createWallet, userID, int64(0))
	if err != nil {
		logrus.WithField("err", err).Error("im.db.Exec upsert wallet failed in InitWallet")
		return err
	}

	if rowsAffected, err = result.RowsAffected(); err != nil || err == nil && rowsAffected == int64(0) {
		logrus.WithField("err", err).Error("result.RowsAffected failed in InitWallet")
		return err
	}
	return nil
}

// Empty will empty user's wallet
func (im *impl) Empty(userID string) error {
	result, err := im.db.Exec(patchWallet, int64(0), userID)
	if err != nil {
		logrus.WithField("err", err).Error("im.db.Exec upsert wallet failed in InitWallet")
		return err
	}

	if rowsAffected, err := result.RowsAffected(); err != nil {
		logrus.WithField("err", err).Error("result.RowsAffected failed in InitWallet")
		return err
	} else if rowsAffected == int64(0) {
		return ErrWalletNotExist
	}
	return nil
}

// Deposit will deposit `amount` NTD to user's wallet
func (im *impl) Deposit(userID string, amount int64) error {
	result, err := im.db.Exec(atomicPatchWallet, amount, userID)
	if err != nil {
		logrus.WithField("err", err).Error("im.db.Exec upsert wallet failed in AtomicPatchWallet")
		return err
	}

	if rowsAffected, err := result.RowsAffected(); err != nil {
		logrus.WithField("err", err).Error("result.RowsAffected failed in AtomicPatchWallet")
		return err
	} else if err == nil && rowsAffected == int64(0) {
		return ErrWalletNotExist
	}
	return nil
}

// Spend will spend `amount` NTD from user's wallet
func (im *impl) Spend(userID string, amount int64, reason string) error {
	result, err := im.db.Exec(atomicPatchWallet, -1*amount, userID)
	if err != nil {
		logrus.WithField("err", err).Error("im.db.Exec upsert wallet failed in AtomicPatchWallet")
		return err
	}

	if rowsAffected, err := result.RowsAffected(); err != nil {
		logrus.WithField("err", err).Error("result.RowsAffected failed in AtomicPatchWallet")
		return err
	} else if err == nil && rowsAffected == int64(0) {
		return ErrWalletNotExist
	}
	return nil
}
