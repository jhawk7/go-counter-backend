package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		for _, err := range c.Errors {
			log.Errorf("error: %v", err.Error())
		}

		if len(c.Errors) > 0 {
			c.JSON(http.StatusInternalServerError, nil)
		}
	}
}

func LogInfo(msg string) {
	log.Info(msg)
}

func LogError(err error, fatal bool) {
	log.Errorf("error: %v", err.Error())
	if fatal {
		panic(err)
	}
}
