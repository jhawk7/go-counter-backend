package main

import (
	"go-counter-consumer/internal/common"
	"go-counter-consumer/internal/dbclient"

	"github.com/redis/go-redis/v9"
)

var db *dbclient.DBClient

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

func main() {
	config := common.GetConfig()
	InitDB(config)

	// r := gin.Default()
	// r.Use(common.ErrorHandler())
	// r.Use()
	// r.GET("/ping", func(c *gin.Context) {
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"message": "pong",
	// 	})
	// })

	// r.GET("/count", GetCount)
	// r.PUT("/count", ResetCount)
	// r.PATCH("/count", UpdateCount)
	// r.Run(":8888")
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
