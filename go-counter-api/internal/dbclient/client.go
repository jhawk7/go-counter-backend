package dbclient

import (
	"context"
	"fmt"
	"go-counter-api/internal/common"

	"github.com/redis/go-redis/v9"
)

const (
	key = "counter"
)

type IDBSvc interface {
	Get(context.Context, string) *redis.StringCmd
	Ping(context.Context) *redis.StatusCmd
}

type DBClient struct {
	svc IDBSvc
}

func InitDBClient(svc IDBSvc) (*DBClient, error) {
	if _, pingErr := svc.Ping(context.Background()).Result(); pingErr != nil {
		err := fmt.Errorf("failed to connect to redis instance; %v", pingErr)
		return nil, err
	}

	common.LogInfo("established connection to redis instance")

	return &DBClient{
		svc: svc,
	}, nil
}

func (c *DBClient) GetValue(ctx context.Context) (val int, err error) {
	res, resErr := c.svc.Get(ctx, key).Int()
	if resErr != nil {
		err = fmt.Errorf("failed to retrieve value by redis key: [key: %v] [error: %v]", key, resErr)
		return
	}

	common.LogInfo(fmt.Sprintf("successfully retrieved counter value: %v", res))

	val = res
	return
}
