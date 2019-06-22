package main

import (
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"github.com/andy/guachi-pay-line-bot/api"
	lb "github.com/andy/guachi-pay-line-bot/linebot"
	wl "github.com/andy/guachi-pay-line-bot/wallet"
)

func main() {
	// set the standard logger formatter
	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})

	wallet, err := wl.NewWallet()
	if err != nil {
		logrus.Fatal("NewWallet failed")
		return
	}

	linebot, err := lb.NewLinebot(wallet)
	if err != nil {
		logrus.Fatal("NewLinebot failed")
		return
	}

	gin.SetMode(gin.ReleaseMode)

	route := gin.Default()
	api.NewHandler(route, linebot)

	logrus.Info("start serving https request")
	route.Run()
	return
}
