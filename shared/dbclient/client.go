package dbclient

import (
	"context"
	"fmt"
	"github.com/go-counter-backend/shared/common"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	key = "counter"
)

type IDBSvc interface {
	Get(context.Context, string) *redis.StringCmd
	Set(context.Context, string, interface{}, time.Duration) *redis.StatusCmd
	Ping(context.Context) *redis.StatusCmd
	Incr(context.Context, string) *redis.IntCmd
	Decr(context.Context, string) *redis.IntCmd
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

	if _, kErr := svc.Get(context.Background(), key).Result(); kErr != nil {
		if kErr != redis.Nil {
			return nil, fmt.Errorf("failed to search redis instance for initial kv pair; [error: %v]", kErr)
		} else {
			svc.Set(context.Background(), key, 0, 0)
			common.LogInfo("initialized redis instance with counter:0 kv pair")
		}
	}

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

func (c *DBClient) SetValue(ctx context.Context, value int) (err error) {
	if setErr := c.svc.Set(ctx, key, value, 0).Err(); setErr != nil {
		err = fmt.Errorf("failed to set kv pair; [key: %v] [value: %v] [error: %v]", key, value, setErr)
		return
	}

	common.LogInfo(fmt.Sprintf("successfully reset counter: [%v:%v]", key, value))

	return
}

func (c *DBClient) IncrementCount(ctx context.Context) (err error) {
	val, incErr := c.svc.Incr(ctx, key).Result()
	if incErr != nil {
		err = fmt.Errorf("failed to increment counter; [error: %v]", incErr)
		return
	}

	common.LogInfo(fmt.Sprintf("counter incremented to %v", val))
	return
}

func (c *DBClient) DecrementCount(ctx context.Context) (err error) {
	val, decErr := c.svc.Decr(ctx, key).Result()
	if decErr != nil {
		err = fmt.Errorf("failed to decrement counter; [error: %v]", decErr)
		return
	}

	common.LogInfo(fmt.Sprintf("counter decremented to %v", val))
	return
}
