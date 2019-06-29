package wallet

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/andy/guachi-pay-line-bot/db"
)

const (
	// UsersWallet related statements
	checkIfWalletExists = `SELECT 1 FROM "UsersWallet" WHERE "userID" = $1`
	createWallet        = `
		INSERT INTO "UsersWallet" ("userID", "balance")
		VALUES ($1, $2);
	`
	patchWallet       = `UPDATE "UsersWallet" SET balance = $1 WHERE "userID" = $2;`
	atomicPatchWallet = `UPDATE "UsersWallet" SET balance = balance + $1 WHERE "userID" = $2;`
	getBalance        = `SELECT balance FROM "UsersWallet" WHERE "userID" = $1`

	// UsersWalletLog related statements
	insertWalletLog = `
		INSERT INTO "UsersWalletLog" ("userID", reason, amount, timestamp)
			VALUES ($1, $2, $3, $4);
	`
	deleteAllWalletLogs = `DELETE FROM "UsersWalletLog" WHERE "userID" = $1`
	getWalletLogs       = `
		SELECT 
			reason, amount, timestamp 
		FROM 
			"UsersWalletLog 
		WHERE 
			"userID" = $1 AND timestamp >= $2 AND timestamp <= $3 
		ORDER BY
			timestamp
	`
)

type impl struct {
	db *sql.DB
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

func execRollBack(tx *sql.Tx) {
	if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
		logrus.WithField("err", err).Error("tx.Rollback() failed in Deposit")
	}
	return
}

func (im *impl) Create(userID string) error {
	result, err := im.db.Exec(checkIfWalletExists, userID)
	if err != nil {
		logrus.WithField("err", err).Error("im.db.Exec(checkIfWalletExists) failed in Create")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logrus.WithField("err", err).Error("result.RowsAffected failed in Create")
		return err
	}

	if rowsAffected == int64(1) {
		return ErrWalletExist
	}

	result, err = im.db.Exec(createWallet, userID, int64(0))
	if err != nil {
		logrus.WithField("err", err).Error("im.db.Exec(createWallet) failed in Create")
		return err
	}

	if rowsAffected, err = result.RowsAffected(); err != nil || err == nil && rowsAffected == int64(0) {
		logrus.WithField("err", err).Error("result.RowsAffected failed in Create")
		return err
	}
	return nil
}

func (im *impl) EmptyBalance(userID string) error {
	tx, err := im.db.Begin()
	if err != nil {
		logrus.WithField("err", err).Error("im.db.Begin failed in EmptyBalance")
		return err
	}
	defer execRollBack(tx)

	result, err := tx.Exec(patchWallet, int64(0), userID)
	if err != nil {
		logrus.WithField("err", err).Error("tx.Exec(patchWallet) failed in EmptyBalance")
		return err
	}

	if rowsAffected, err := result.RowsAffected(); err != nil {
		logrus.WithField("err", err).Error("result.RowsAffected failed in EmptyBalance")
		return err
	} else if rowsAffected == int64(0) {
		return ErrWalletNotFound
	}

	if _, err := tx.Exec(deleteAllWalletLogs, userID); err != nil {
		logrus.WithField("err", err).Error("tx.Exec(patchWallet) failed in EmptyBalance")
		return err
	}

	if err := tx.Commit(); err != nil {
		logrus.WithField("err", err).Error("tx.Commit() failed in EmptyBalance")
		return err
	}
	return nil
}

func (im *impl) GetBalance(userID string) (int64, error) {
	balance := int64(0)
	if err := im.db.QueryRow(getBalance, userID).Scan(
		&balance,
	); err != nil && err != sql.ErrNoRows {
		logrus.WithField("err", err).Error("im.db.Query(getBalance) failed in GetBalance")
		return int64(0), err
	} else if err == sql.ErrNoRows {
		return int64(0), ErrWalletNotFound
	}
	return balance, nil
}

func convertToTimestampStr(timestamp int64) string {
	temp := time.Unix(timestamp, 0)
	// change to Asia/Taipei zone
	location, _ := time.LoadLocation("Asia/Taipei")
	temp.In(location)
	return fmt.Sprintf("%d/%d/%d %d:%d", temp.Year(), temp.Month(), temp.Day(), temp.Hour(), temp.Minute())
}

func (im *impl) GetBalanceLogs(userID string, options ...GetLogsOption) ([]*BalanceLog, error) {
	option := initOption(options...)

	startTime := option.startTime
	endTime := option.endTime
	if option.endTime == int64(0) {
		endTime = time.Now().Unix()
	}

	rows, err := im.db.Query(getWalletLogs, userID, startTime, endTime)
	if err != nil {
		logrus.WithField("err", err).Error("im.db.Query(getWalletLogs) failed in GetBalanceLogs")
		return nil, err
	}

	balanceLogs := []*BalanceLog{}
	for rows.Next() {
		amount := int64(0)
		reason := ""
		timestamp := int64(0)
		if err := rows.Scan(&amount, &reason, &timestamp); err != nil {
			logrus.WithField("err", err).Error("rows.Scan failed in GetBalanceLogs")
			return nil, err
		}

		balanceLogs = append(balanceLogs, &BalanceLog{
			Amount:    amount,
			Reason:    reason,
			Timestamp: convertToTimestampStr(timestamp),
		})
	}
	return balanceLogs, nil
}

func (im *impl) Deposit(userID string, amount int64, reason string) error {
	tx, err := im.db.Begin()
	if err != nil {
		logrus.WithField("err", err).Error("im.db.Begin failed in Deposit")
		return err
	}
	defer execRollBack(tx)

	result, err := tx.Exec(atomicPatchWallet, amount, userID)
	if err != nil {
		logrus.WithField("err", err).Error("tx.Exec(atomicPatchWallet) failed in Deposit")
		return err
	}

	if rowsAffected, err := result.RowsAffected(); err != nil {
		logrus.WithField("err", err).Error("result.RowsAffected failed in Deposit")
		return err
	} else if err == nil && rowsAffected == int64(0) {
		return ErrWalletNotFound
	}

	result, err = tx.Exec(insertWalletLog, userID, reason, amount, time.Now().Unix())
	if err != nil {
		logrus.WithField("err", err).Error("tx.Exec(insertWalletLog) failed in Deposit")
		return err
	}

	if rowsAffected, err := result.RowsAffected(); err != nil {
		logrus.WithField("err", err).Error("result.RowsAffected failed in Deposit")
		return err
	} else if err == nil && rowsAffected == int64(0) {
		return ErrWalletNotFound
	}

	if err := tx.Commit(); err != nil {
		logrus.WithField("err", err).Error("tx.Commit() failed in EmptyBalance")
		return err
	}
	return nil
}

func (im *impl) Spend(userID string, amount int64, reason string) error {
	tx, err := im.db.Begin()
	if err != nil {
		logrus.WithField("err", err).Error("im.db.Begin failed in Spend")
		return err
	}
	defer execRollBack(tx)

	result, err := im.db.Exec(atomicPatchWallet, -1*amount, userID)
	if err != nil {
		logrus.WithField("err", err).Error("im.db.Exec(atomicPatchWallet) failed in Spend")
		return err
	}

	if rowsAffected, err := result.RowsAffected(); err != nil {
		logrus.WithField("err", err).Error("result.RowsAffected failed in Spend")
		return err
	} else if err == nil && rowsAffected == int64(0) {
		return ErrWalletNotFound
	}

	result, err = tx.Exec(insertWalletLog, userID, reason, -1*amount, time.Now().Unix())
	if err != nil {
		logrus.WithField("err", err).Error("tx.Exec(insertWalletLog) failed in Spend")
		return err
	}

	if rowsAffected, err := result.RowsAffected(); err != nil {
		logrus.WithField("err", err).Error("result.RowsAffected failed in Spend")
		return err
	} else if err == nil && rowsAffected == int64(0) {
		return ErrWalletNotFound
	}

	if err := tx.Commit(); err != nil {
		logrus.WithField("err", err).Error("tx.Commit() failed in Spend")
		return err
	}
	return nil
}
