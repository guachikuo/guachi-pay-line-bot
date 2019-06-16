package wallet

import (
	"github.com/sirupsen/logrus"

	"github.com/andy/guachi-pay-line-bot/db"
)

const (
	checkIfWalletExists = `SELECT 1 FROM "UsersWallet" WHERE "userID" = $1`
	createWallet        = `
		INSERT INTO "UsersWallet" ("userID", "balance")
		VALUES ($1, $2);
	`
	patchWallet = `UPDATE "UsersWallet" SET balance = $1 WHERE "userID" = $2;`
)

// InitWallet initialize the user's wallet
// if wallet exists, set balance to 0, else create one
func InitWallet(userID string) error {
	rdsdbSrv, err := db.GetPostgresSrv()
	if err != nil {
		logrus.WithField("err", err).Error("db.GetPostgresSrv failed in initWallet")
		return err
	}

	result, err := rdsdbSrv.Exec(checkIfWalletExists, userID)
	if err != nil {
		logrus.WithField("err", err).Error("rdsdbSrv.Exec checkIfWalletExists failed in initWallet")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logrus.WithField("err", err).Error("result.RowsAffected failed in initWallet")
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
		logrus.WithField("err", err).Error("rdsdbSrv.Exec upsert wallet failed in initWallet")
		return err
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil || err == nil && rowsAffected == int64(0) {
		logrus.WithField("err", err).Error("result.RowsAffected failed in initWallet")
		return err
	}
	return nil
}
