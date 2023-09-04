package storage

import (
	"context"
	"encoding/json"
	goredis "github.com/redis/go-redis/v9"
	"time"
)

type Cache interface {
	getKey(key string, value interface{}) error
	setKey(key string, value interface{}, expiration time.Duration) error
}

type RedisClient struct {
	client *goredis.ClusterClient
}

func NewRedisClient(addr string) (Cache, error) {
	client := goredis.NewClusterClient(&goredis.ClusterOptions{
		Addrs:          []string{addr},
		Password:       "",
		MaxRetries:     10,
		ReadOnly:       false,
		RouteRandomly:  false,
		RouteByLatency: false,
	})
	_, err := client.Ping(context.TODO()).Result()
	if err != nil {
		return nil, err
	}
	return &RedisClient{client: client}, nil
}

func (client *RedisClient) getKey(key string, value interface{}) error {
	val, err := client.client.Get(context.TODO(), key).Result()
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(val), &value)
	if err != nil {
		return err
	}
	return nil
}

func (client *RedisClient) setKey(key string, value interface{}, expiration time.Duration) error {
	cacheEntry, err := json.Marshal(value)
	if err != nil {
		return err
	}
	err = client.client.Set(context.TODO(), key, cacheEntry, expiration).Err()
	if err != nil {
		return err
	}
	return nil
}
