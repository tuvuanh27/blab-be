package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var (
	ChainKey                  = "CHAIN"
	DifficultyKey             = "DIFFICULTY"
	TransactionPoolKey        = "TRANSACTION_POOL"
	ChannelSyncNodeKey        = "BLOCKCHAIN"
	ChannelSyncTransactionKey = "TRANSACTION"
	CurrentBlockCrawledKey    = "CURRENT_BLOCK_CRAWLED"
	HistoryTransactionsKey    = "HISTORY_TRANSACTIONS"
)

var Ctx = context.Background()

type IRedis interface {
	Get(key string) string
	Set(key string, value string)
	SetSet(key string, value string)
	GetSet(key string) []string
	Publish(channel string, message string)
	Subscribe(channel string) *redis.PubSub
}

type redisService struct {
	client *redis.Client
}

func NewRedisService(client *redis.Client) IRedis {
	return &redisService{
		client: client,
	}
}

func Connect() *redis.Client {
	// Connect to Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})

	// Ping the Redis server and check if any errors occurred
	_, err := redisClient.Ping(Ctx).Result()
	if err != nil {
		panic(err)
	}

	// Print the Redis server's status
	fmt.Println("Connected to Redis!")
	return redisClient
}

func (rs *redisService) Get(key string) string {
	val, err := rs.client.Get(Ctx, key).Result()
	if err != nil {
		return ""
	}
	return val
}

func (rs *redisService) Set(key string, value string) {
	err := rs.client.Set(Ctx, key, value, 0).Err()
	if err != nil {
		panic(err)
	}
}

func (rs *redisService) SetSet(key string, value string) {
	err := rs.client.SAdd(Ctx, key, value).Err()
	if err != nil {
		panic(err)
	}
}

func (rs *redisService) GetSet(key string) []string {
	val, err := rs.client.SMembers(Ctx, key).Result()
	if err != nil {
		return []string{}
	}
	return val
}

func (rs *redisService) Publish(channel string, message string) {
	err := rs.client.Publish(Ctx, channel, message).Err()
	if err != nil {
		panic(err)
	}
}

func (rs *redisService) Subscribe(channel string) *redis.PubSub {
	pubsub := rs.client.Subscribe(Ctx, channel)
	return pubsub
}

var RedisService = NewRedisService(Connect())
