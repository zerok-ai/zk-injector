package zkredis

import (
	"fmt"
	"github.com/go-redis/redis"
	"time"
	"zerok-injector/utils-go/redis/config"
)

var (
	redisClient *redis.Client
)

func Init(redisConfig config.RedisConfig) {
	var readTimeout time.Duration = time.Duration(redisConfig.ReadTimeout) * time.Second
	redisClient = redis.NewClient(&redis.Options{
		Addr:        fmt.Sprint(redisConfig.Host, ":", redisConfig.Port),
		Password:    "",
		DB:          0,
		ReadTimeout: readTimeout,
	})
}

func SetString(key string, value string) error {
	return redisClient.Set(key, value, -1).Err()
}

func GetString(key string) (*string, error) {
	output := redisClient.Get(key)
	err := output.Err()
	if err != nil {
		if err == redis.Nil {
			return nil, err
		}
		return nil, err
	}
	value := output.Val()
	return &value, nil
}

func Delete(key string) error {
	output := redisClient.Del(key)
	err := output.Err()
	if err != nil {
		if err == redis.Nil {
			return err
		}
		return err
	}
	return nil
}

func SetStringWithExpiration(key string, value string, expiration time.Duration) error {
	return redisClient.Set(key, value, expiration).Err()
}
