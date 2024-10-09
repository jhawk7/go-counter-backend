package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-counter-backend/shared/common"
	"github.com/go-counter-backend/shared/dbclient"
	"github.com/go-counter-backend/shared/event"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

var (
	db        *dbclient.DBClient
	config    *common.Config
	kConsumer *event.EventClient
)

const (
	buffer = 10
	sleep  = 2
)

func InitDB(config *common.Config) {
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

func InitConsumer(config *common.Config) {
	kconn, kErr := kafka.DialLeader(context.Background(), "tcp", config.KafkaHost, config.KafkaTopic, 0)
	if kErr != nil {
		panic(fmt.Errorf("failed to establish connection to kafka; %v", kErr))
	}

	kConsumer = event.InitEventClient(kconn)
}

func main() {
	config := common.GetConfig()
	InitDB(config)
	InitConsumer(config)

	r := gin.Default()
	r.Use(common.ErrorHandler())
	r.Use()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	go ReadStream()
	r.Run(":8888")
}

func ReadStream() {
	ctx := context.Background()
	ch := make(chan event.Msg, buffer)
	go kConsumer.ReadMsg(ch)

	for m := range ch {
		switch m.Event {
		case "increment":
			err := db.IncrementCount(ctx)
			common.LogError(err, false)
		case "reset":
			err := db.SetValue(ctx, 0)
			common.LogError(err, false)
		default:
			common.LogError(fmt.Errorf("unexpected event message %v", m.Event), false)
		}

		time.Sleep(sleep * time.Millisecond)
	}
}
