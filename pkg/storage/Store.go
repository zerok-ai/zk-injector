package storage

import (
	"fmt"
	"github.com/go-redis/redis"
	"time"
	"zerok-injector/utils-go/redis/config"
)

const (
	defaultExpiry time.Duration = time.Hour * 24 * 30
	hashTableName string        = "container_images"
)

type ImageStore struct {
	redisClient   *redis.Client
	hashTableName string
}

func GetNewImageStore(redisConfig config.RedisConfig) *ImageStore {
	readTimeout := time.Duration(redisConfig.ReadTimeout) * time.Second
	_redisClient := redis.NewClient(&redis.Options{
		Addr:        fmt.Sprint(redisConfig.Host, ":", redisConfig.Port),
		Password:    "",
		DB:          0,
		ReadTimeout: readTimeout,
	})

	//_redisClient.Expire(hashTableName, defaultExpiry)

	imgRedis := &ImageStore{
		redisClient:   _redisClient,
		hashTableName: hashTableName,
	}
	return imgRedis
}

func (zkRedis ImageStore) SetString(key string, value string) error {
	return zkRedis.redisClient.HSet(zkRedis.hashTableName, key, value).Err()
}

func (zkRedis ImageStore) GetString(key string) (*string, error) {
	output := zkRedis.redisClient.HGet(zkRedis.hashTableName, key)
	err := output.Err()
	if err != nil {
		return nil, err
	}
	value := output.Val()
	return &value, nil
}

func (zkRedis ImageStore) Delete(key string) error {
	return zkRedis.redisClient.HDel(zkRedis.hashTableName, key).Err()
}

func (zkRedis ImageStore) Length(key string) (int64, error) {
	// get the number of hash key-value pairs
	result, err := zkRedis.redisClient.HLen(zkRedis.hashTableName).Result()
	if err != nil {
		return 0, err
	}
	return result, err
}
