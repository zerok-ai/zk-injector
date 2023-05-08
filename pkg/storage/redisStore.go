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
	defaultExpiry     time.Duration = time.Hour * 24 * 30
	hashSetName       string        = "zk_img_proc_map"
	hashSetVersionKey string        = "zk_img_proc_version"
)

type ImageStore struct {
	redisClient *redis.Client
	hashSetName string
}

func (zkRedis *ImageStore) GetHashSetVersion() (*string, error) {
	data, err := zkRedis.GetString(hashSetVersionKey)
	if err != nil {
		fmt.Println("Error caught while getting hash set version from redis.")
		return nil, err
	}
	return data, nil
}

func GetNewImageStore(redisConfig config.RedisConfig) *ImageStore {
	readTimeout := time.Duration(redisConfig.ReadTimeout) * time.Second
	_redisClient := redis.NewClient(&redis.Options{
		Addr:        fmt.Sprint(redisConfig.Host, ":", redisConfig.Port),
		Password:    "",
		DB:          redisConfig.DB,
		ReadTimeout: readTimeout,
	})

	//_redisClient.Expire(hashTableName, defaultExpiry)

	imgRedis := &ImageStore{
		redisClient: _redisClient,
		hashSetName: hashSetName,
	}
	return imgRedis
}

func (zkRedis *ImageStore) LoadAllData(imageRuntimeMap *sync.Map) error {
	var cursor uint64
	var data []string

	for {
		var err error
		//Getting 10 fields at once.
		data, cursor, err = zkRedis.redisClient.HScan(hashSetName, cursor, "*", 10).Result()
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
			fmt.Printf("Saving value %v for key %v\n", key, value)
			imageRuntimeMap.Store(key, serializedValue)
		}

		if cursor == 0 {
			break
		}
	}
	return nil
}

func (zkRedis *ImageStore) GetString(key string) (*string, error) {
	output := zkRedis.redisClient.HGet(zkRedis.hashSetName, key)
	err := output.Err()
	if err != nil {
		return nil, err
	}
	value := output.Val()
	return &value, nil
}
