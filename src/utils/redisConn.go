package utils

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

func CreateRedisConn(addr string) *redis.Client {
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

// 生成用户缓存键
func UserCacheKey(username string) string {
	return fmt.Sprintf("user:%s", username)
}

// 生成好友列表缓存键
func FriendsCacheKey(userID uint64) string {
	return fmt.Sprintf("friends:%d", userID)
}
