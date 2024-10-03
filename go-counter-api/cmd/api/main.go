package main

import (
	"fmt"
	"go-counter-api/internal/common"
	"go-counter-api/internal/dbclient"
	"go-counter-api/internal/producer"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

var (
	db            *dbclient.DBClient
	config        *common.Config
	kProducer     *producer.ProducerClient
	SECRET_HEADER = "SECRET_HEADER"
)

func InitDB(config *common.Config) {
	opts := redis.Options{
		Addr:     config.Host,
		Password: config.Pass,
		DB:       0,
	}

	svc := redis.NewClient(&opts)
	dbsvc, dbErr := dbclient.InitDBClient(svc)
	if dbErr != nil {
		panic(dbErr)
	}

	db = dbsvc
}

func InitProducer(config *common.Config) {
	kconn, kErr := kafka.Dial("tcp", config.KafkaUrl)
	if kErr != nil {
		panic(fmt.Errorf("failed to establish connection to kafka; %v", kErr))
	}

	kProducer = producer.InitProducer(kconn)
}

func main() {
	config = common.GetConfig()
	InitDB(config)
	InitProducer(config)

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

	c.JSON(http.StatusOK, gin.H{
		"count": val,
	})
}

func Websocket(c *gin.Context) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			secret := r.Header.Get(SECRET_HEADER)
			return secret == config.Secret
		},
	}

	conn, connErr := upgrader.Upgrade(c.Writer, c.Request, nil)
	if connErr != nil {
		err := fmt.Errorf("failed to upgrade connection to websocket; %v", connErr)
		c.Error(err)
		return
	}

	defer conn.Close()

	for {
		_, rawMsg, rErr := conn.ReadMessage()
		if rErr != nil {
			c.Error(fmt.Errorf("failed to read incoming msg; [error: %v]", rErr))
			continue
		}

		if wErr := kProducer.WriteMsg(rawMsg); wErr != nil {
			c.Error(wErr)
		}
	}

}
