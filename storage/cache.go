package storage

import (
	"context"
	"encoding/json"
	goredis "github.com/redis/go-redis/v9"
	"github.com/yangwenz/model-webhook/utils"
	"time"
)

type Cache interface {
	GetKey(key string) (string, error)
	SetKey(key string, value interface{}, expiration time.Duration) error
}

type RedisClient struct {
	client *goredis.Client
}

type RedisClusterClient struct {
	client *goredis.ClusterClient
}

func NewRedisClient(config utils.Config) (Cache, error) {
	if config.RedisClusterMode {
		client := goredis.NewClusterClient(&goredis.ClusterOptions{
			Addrs:          []string{config.RedisAddress},
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
		return &RedisClusterClient{client: client}, nil
	} else {
		client := goredis.NewClient(&goredis.Options{
			Addr:     config.RedisAddress,
			Password: "",
			DB:       0, // use default DB
		})
		_, err := client.Ping(context.TODO()).Result()
		if err != nil {
			return nil, err
		}
		return &RedisClient{client: client}, nil
	}
}

func (client *RedisClusterClient) GetKey(key string) (string, error) {
	val, err := client.client.Get(context.TODO(), key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (client *RedisClusterClient) SetKey(key string, value interface{}, expiration time.Duration) error {
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

func (client *RedisClient) GetKey(key string) (string, error) {
	val, err := client.client.Get(context.TODO(), key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (client *RedisClient) SetKey(key string, value interface{}, expiration time.Duration) error {
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
