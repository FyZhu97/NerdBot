package main

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"time"
)

var Connection *redis.Client

func InitRedis() {
	Connection = redis.NewClient(&redis.Options{
		Addr:     GlobalConfig.Redis.Address,
		Password: GlobalConfig.Redis.Password,
		DB:       GlobalConfig.Redis.Database,
	})
	pong, err := Connection.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}
	logrus.Info("initiate GoRedis Client success: ", pong)
}

func StoreRecord(key string, record *Record) error {
	// Convert the Record struct to JSON
	recordJSON, err := json.Marshal(record)
	if err != nil {
		return err
	}

	// Store the JSON in Redis
	err = Connection.Set(context.Background(), key, recordJSON, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func RetrieveOrDefaultRecord(key string) (*Record, error) {
	// Get the stored JSON from Redis
	recordJSON, err := Connection.Get(context.Background(), key).Bytes()
	if err == redis.Nil {
		ori := Record{
			Prompt:      GlobalConfig.AI.InitialPrompts,
			TotalTokens: 0,
			LastRequest: time.UnixMicro(0),
			Temperature: GlobalConfig.AI.DefaultTemperature,
		}
		return &ori, nil
	} else if err != nil {
		return nil, err
	}
	// Convert the JSON to a Record struct
	var record Record
	err = json.Unmarshal(recordJSON, &record)
	if err != nil {
		return nil, err
	}

	return &record, err
}
func DeleteRecord(key string) {
	Connection.Del(context.Background(), key)
}
