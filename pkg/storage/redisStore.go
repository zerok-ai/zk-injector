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
		fmt.Printf("Error caught while getting hash set version from redis %v.\n ", err)
		return nil, err
	}
	return data, nil
}

func (zkRedis *ImageStore) testRedisConn() {
	pong, err := zkRedis.redisClient.Ping().Result()
	if err != nil {
		fmt.Printf("Error sending PING command to Redis: %v", err)
	}

	fmt.Println("Redis PING response:", pong)
}

func GetNewImageStore(redisConfig config.RedisConfig) *ImageStore {
	readTimeout := time.Duration(redisConfig.ReadTimeout) * time.Second
	addr := fmt.Sprint(redisConfig.Host, ":", redisConfig.Port)
	fmt.Printf("Address for redis is %v.\n", addr)
	_redisClient := redis.NewClient(&redis.Options{
		Addr:        addr,
		Password:    "",
		DB:          redisConfig.DB,
		ReadTimeout: readTimeout,
	})

	imgRedis := &ImageStore{
		redisClient: _redisClient,
		hashSetName: hashSetName,
	}

	imgRedis.testRedisConn()

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
			fmt.Printf("Error while scan from redis %v\n", err)
			return err
		}

		for i := 0; i < len(data); i += 2 {
			key := data[i]
			value := data[i+1]
			fmt.Println(value)
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

func (zkRedis *ImageStore) GetString(key string) (*string, error) {
	output := zkRedis.redisClient.HGet(zkRedis.hashSetName, key)
	err := output.Err()
	if err != nil {
		return nil, err
	}
	value := output.Val()
	return &value, nil
}
