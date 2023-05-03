package storage

import (
	"fmt"
	"sync"
	"time"
	"zerok-injector/internal/config"
	"zerok-injector/pkg/utils"

	"github.com/go-redis/redis"
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

func (zkRedis *ImageStore) LoadAllData(imageRuntimeMap *sync.Map) error {
	var cursor uint64
	var data []string

	for {
		var err error
		//Getting 10 fields at once.
		data, cursor, err = zkRedis.redisClient.HScan(hashTableName, cursor, "*", 10).Result()
		if err != nil {
			return err
		}

		for i := 0; i < len(data); i += 2 {
			key := data[i]
			value := data[i+1]
			serializedValue, err := utils.FromJsonString(value)
			if err != nil {
				//TODO: Handle error
				fmt.Println(err)
			}
			imageRuntimeMap.Store(key, serializedValue)
		}

		if cursor == 0 {
			break
		}
	}
	return nil
}

// TODO: Add error handling here. Incase saving to redis fails.
func (zkRedis *ImageStore) SetString(key string, value string) error {
	return zkRedis.redisClient.HSet(zkRedis.hashTableName, key, value).Err()
}

// TODO: Add error handling here. Incase loading from redis fails.
func (zkRedis *ImageStore) GetString(key string) (*string, error) {
	output := zkRedis.redisClient.HGet(zkRedis.hashTableName, key)
	err := output.Err()
	if err != nil {
		return nil, err
	}
	value := output.Val()
	return &value, nil
}

func (zkRedis *ImageStore) Delete(key string) error {
	return zkRedis.redisClient.HDel(zkRedis.hashTableName, key).Err()
}

func (zkRedis *ImageStore) Length(key string) (int64, error) {
	// get the number of hash key-value pairs
	result, err := zkRedis.redisClient.HLen(zkRedis.hashTableName).Result()
	if err != nil {
		return 0, err
	}
	return result, err
}
