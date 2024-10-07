package main

import (
	"context"
	"fmt"
	"net/http"

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

func InitConsumer(config *common.Config) {
	kconn, kErr := kafka.DialLeader(context.Background(), "tcp", config.KafkaUrl, config.KafkaTopic, 0)
	if kErr != nil {
		panic(fmt.Errorf("failed to establish connection to kafka; %v", kErr))
	}

	kConsumer = event.InitEventClient(kconn)
}

func main() {
	config := common.GetConfig()
	InitConsumer(config)
	InitDB(config)

	r := gin.Default()
	r.Use(common.ErrorHandler())
	r.Use()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// r.GET("/count", GetCount)
	// r.PUT("/count", ResetCount)
	// r.PATCH("/count", UpdateCount)
	go ReadStream()
	r.Run(":8888")
}

func ReadStream() {

}

// func GetCount(c *gin.Context) {
// 	val, err := db.GetValue(c)
// 	if err != nil {
// 		c.Error(err)
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"count": val,
// 	})
// }

// func ResetCount(c *gin.Context) {
// 	type Params struct {
// 		Counter int `json:"counter"`
// 	}

// 	var params Params
// 	if bindErr := c.ShouldBindJSON(&params); bindErr != nil {
// 		err := fmt.Errorf("missing counter param in request")
// 		c.Error(err)
// 		c.AbortWithStatusJSON(http.StatusBadRequest, err)
// 	}

// 	if err := db.SetValue(c, params.Counter); err != nil {
// 		c.Error(err)
// 		return
// 	}

// 	c.Status(http.StatusNoContent)
// }

// func UpdateCount(c *gin.Context) {
// 	if err := db.IncrementCount(c); err != nil {
// 		c.Error(err)
// 		return
// 	}

// 	c.Status(http.StatusNoContent)
// }
