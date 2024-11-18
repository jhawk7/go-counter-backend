package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-counter-backend/shared/common"
	"github.com/go-counter-backend/shared/dbclient"
	"github.com/go-counter-backend/shared/mqtt_client"
	"github.com/redis/go-redis/v9"
)

var (
	db           *dbclient.DBClient
	config       *common.Config
	mqttConsumer *mqtt_client.MQTTClient
)

const (
	buffer = 10
	sleep  = 2
)

func InitDB() {
	opts := redis.Options{
		Addr:     config.RedisHost,
		Password: config.RedisPass,
		DB:       0,
	}

	svc := redis.NewClient(&opts)
	dbsvc, dbErr := dbclient.InitDBClient(svc)
	if dbErr != nil {
		panic(dbErr)
	}

	db = dbsvc
}

func InitConsumer() {
	mqttConn, mqttErr := mqtt_client.InitConn(config, "go-counter-consumer")
	if mqttErr != nil {
		panic(fmt.Errorf("failed to establish connection to mqtt; %v", mqttErr))
	}

	mqttConsumer = mqtt_client.InitClient(mqttConn)
}

func main() {
	config = common.GetConfig()
	InitDB()
	InitConsumer()

	r := gin.Default()
	r.Use(common.ErrorHandler())
	r.Use()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	go ReadStream()
	r.Run(":8889")
}

func ReadStream() {
	ctx := context.Background()
	ch := make(chan mqtt_client.Msg, buffer)
	go mqttConsumer.ReadMsg(ch)

	for m := range ch {
		switch m.Event {
		case "increment":
			err := db.IncrementCount(ctx)
			common.LogError(err, false)
		case "reset":
			err := db.SetValue(ctx, 0)
			common.LogError(err, false)
		case "decrement":
			err := db.DecrementCount(ctx)
			common.LogError(err, false)
		default:
			common.LogError(fmt.Errorf("unexpected event message %v", m.Event), false)
		}

		time.Sleep(sleep * time.Millisecond)
	}
}
