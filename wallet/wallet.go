package wallet

import (
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
)

// InitWallet initialize the user's wallet
// if wallet exists, set balance to 0, else create one
func InitWallet(userID string) error {
	rdsdbSrv, err := db.GetPostgresSrv()
	if err != nil {
		logrus.WithField("err", err).Error("db.GetPostgresSrv failed in InitWallet")
		return err
	}

	result, err := rdsdbSrv.Exec(checkIfWalletExists, userID)
	if err != nil {
		logrus.WithField("err", err).Error("rdsdbSrv.Exec checkIfWalletExists failed in InitWallet")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logrus.WithField("err", err).Error("result.RowsAffected failed in InitWallet")
		return err
	}

	sqlStmt := patchWallet
	args := []interface{}{int64(0), userID}
	if rowsAffected == int64(0) {
		sqlStmt = createWallet
		args = []interface{}{userID, int64(0)}
	}

	result, err = rdsdbSrv.Exec(sqlStmt, args...)
	if err != nil {
		logrus.WithField("err", err).Error("rdsdbSrv.Exec upsert wallet failed in InitWallet")
		return err
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil || err == nil && rowsAffected == int64(0) {
		logrus.WithField("err", err).Error("result.RowsAffected failed in InitWallet")
		return err
	}
	return nil
}

// AtomicPatchWallet will increase or decrease balance for `userID`'s wallet
func AtomicPatchWallet(userID string, moneyCount int64) error {
	rdsdbSrv, err := db.GetPostgresSrv()
	if err != nil {
		logrus.WithField("err", err).Error("db.GetPostgresSrv failed in AtomicPatchWallet")
		return err
	}

	result, err := rdsdbSrv.Exec(atomicPatchWallet, moneyCount, userID)
	if err != nil {
		logrus.WithField("err", err).Error("rdsdbSrv.Exec upsert wallet failed in AtomicPatchWallet")
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
