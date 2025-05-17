package utils

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

func CreateRedisConn(addr string) *redis.Client {

	// 从环境变量获取 Redis 配置, docker使用
	redisHost := os.Getenv("REDIS_CACHE_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	if redisHost == "" {
		redisHost = "localhost" // 默认值，仅用于本地测试
	}
	if redisPort == "" {
		redisPort = "6379"
	}
	// 创建 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr, // Redis 地址
		Password: "",   // 密码（没有则留空）
		DB:       0,    // 默认数据库
	})
	pong, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}
	log.Println("Redis 连接成功:", pong)

	return rdb
}

func Publish(redis *redis.Client, ctx context.Context, channel string, msg string) {
	redis.Publish(ctx, channel, msg)
}

func Subscription(redis *redis.Client, ctx context.Context, channel string) (string, error) {
	pubsub := redis.Subscribe(ctx, channel)
	msg, err := pubsub.ReceiveMessage(ctx)
	log.Println(msg.String())
	return msg.Payload, err
}
