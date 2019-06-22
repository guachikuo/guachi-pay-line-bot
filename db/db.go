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

// NewPostgresSrv creates a PostgresSrv service
func NewPostgresSrv() (*sql.DB, error) {
	db, err := sql.Open(driverName, os.Getenv("DATABASE_URL"))
	if err != nil {
		logrus.WithField("err", err).Error("sql.Open failed in setupPostgresSrv")
		return nil, err
	}

	if err := db.Ping(); err != nil {
		logrus.WithField("err", err).Error("db.Ping failed in setupPostgresSrv")
		return nil, err
	}

	db.SetConnMaxLifetime(10 * time.Second)
	// use 2 * cpu thread for connection pool is enough
	db.SetMaxOpenConns(runtime.NumCPU() * 2)
	return db, nil
}
