package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	lb "github.com/andy/guachi-pay-line-bot/linebot"
)

type handler struct {
	linebot lb.Linebot
}

// NewHandler ...
func NewHandler(route *gin.Engine, linebot lb.Linebot) {
	hd := handler{
		linebot: linebot,
	}
	route.POST("/callback", hd.handleCallback)
}

func (hd *handler) handleCallback(c *gin.Context) {
	if err := hd.linebot.ParseLinebotCallback(c.Writer, c.Request); err != nil {
		if err == lb.ErrInvalidSignature {
			c.JSON(http.StatusUnauthorized, gin.H{})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"errorMessage": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
