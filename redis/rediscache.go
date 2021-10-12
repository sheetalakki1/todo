package redis

import (
	"encoding/json"
	"fmt"
	"go_todo/model"
	"os"
	"time"

	"github.com/go-redis/redis"
)

type redisCache struct {
	host    string
	db      int
	expires time.Duration
}

func NewRedisCache(host string, db int, expires time.Duration) PostCache {
	return &redisCache{
		host:    host,
		db:      db,
		expires: expires,
	}
}

func (cache *redisCache) getClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       cache.db,
	})
}

func (cache *redisCache) Set(key string, value map[string]interface{}) {
	client := cache.getClient()

	json, err := json.Marshal(value)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	client.Set(key, json, cache.expires*time.Second)
}

func (cache *redisCache) Get(key string) *model.Todo {
	client := cache.getClient()

	val, err := client.Get(key).Result()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	post := model.Todo{}
	err = json.Unmarshal([]byte(val), &post)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return &post
}
