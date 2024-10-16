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
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

const (
	wsTimeout = time.Second * 30
)

var (
	db           *dbclient.DBClient
	config       *common.Config
	kProducer    *event.EventClient
	lastResponse time.Time
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

func InitProducer() {
	kconn, kErr := kafka.DialLeader(context.Background(), "tcp", config.KafkaHost, config.KafkaTopic, 0)
	if kErr != nil {
		panic(fmt.Errorf("failed to establish connection to kafka; %v", kErr))
	}

	kProducer = event.InitEventClient(kconn)
}

func main() {
	config = common.GetConfig()
	InitDB()
	InitProducer()

	r := gin.Default()
	r.Use(common.ErrorHandler())
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/ws", Websocket)
	r.GET("/count", GetCount)
	r.Run(":8888")
}

func GetCount(c *gin.Context) {
	val, err := db.GetValue(c)
	if err != nil {
		c.Error(err)
		return
	}

	c.Header("Access-Control-Allow-Origin", "*")
	c.JSON(http.StatusOK, gin.H{
		"count": val,
	})
}

func Websocket(c *gin.Context) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, connErr := upgrader.Upgrade(c.Writer, c.Request, nil)
	if connErr != nil {
		err := fmt.Errorf("failed to upgrade connection to websocket; %v", connErr)
		common.LogError(err, false)
		return
	}

	conn.SetPongHandler(wsPongHandler)
	go wsKeepAlive(conn)
	defer conn.Close()

	for {
		_, rawMsg, rErr := conn.ReadMessage()
		if rErr != nil {
			common.LogError(fmt.Errorf("failed to read incoming msg; [error: %v]", rErr), false)
			return
		}

		if wErr := kProducer.WriteMsg(rawMsg); wErr != nil {
			common.LogError(wErr, false)
		}
	}
}

func wsPongHandler(msg string) error {
	lastResponse = time.Now()
	return nil
}

func wsKeepAlive(c *websocket.Conn) {
	for {
		if err := c.WriteMessage(websocket.PingMessage, []byte("PING")); err != nil {
			common.LogError(fmt.Errorf("failed to ping ws client"), false)
		}

		time.Sleep(wsTimeout / 2)
		if time.Since(lastResponse) > wsTimeout {
			common.LogInfo(fmt.Sprintf("ws client has not responded within the time window; [last response: %v]; connection will now close", lastResponse))
			c.Close()
		}
	}
}
