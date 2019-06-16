package db

import (
	"database/sql"
	"os"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	driverName = "postgres"
)

var (
	rdsdb *sql.DB
)

func init() {
	if err := setupPostgresSrv(); err != nil {
		logrus.WithField("err", err).Fatal("setupPostgresSrv failed")
	}
}

func setupPostgresSrv() error {
	db, err := sql.Open(driverName, os.Getenv("DATABASE_URL"))
	if err != nil {
		logrus.WithField("err", err).Error("sql.Open failed in setupPostgresSrv")
		return err
	}

	if err := db.Ping(); err != nil {
		logrus.WithField("err", err).Error("db.Ping failed in setupPostgresSrv")
		return err
	}

	db.SetConnMaxLifetime(10 * time.Second)
	// use 2 * cpu thread for connection pool is enough
	db.SetMaxOpenConns(runtime.NumCPU() * 2)

	rdsdb = db
	return nil
}

// GetPostgresSrv get the `Postgres` service
func GetPostgresSrv() (*sql.DB, error) {
	if rdsdb == nil {
		if err := setupPostgresSrv(); err != nil {
			logrus.WithField("err", err).Error("setupPostgresSrv failed in getPostgresSrv")
			return nil, err
		}
	}
	return rdsdb, nil
}
