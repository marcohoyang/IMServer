package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hoyang/imserver/src/router"
	rpcClient "github.com/hoyang/imserver/src/rpc"
	"github.com/hoyang/imserver/src/service"
	"github.com/hoyang/imserver/src/utils"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func initClientPool() *rpcClient.ClientPool {
	// 连接dbproxy
	proxyHost := os.Getenv("DB_PROXY_HOST")
	proxyPORT := os.Getenv("DB_PROXY_PORT")
	if proxyHost == "" {
		proxyHost = "localhost" // 默认值，仅用于本地测试
	}
	if proxyPORT == "" {
		proxyPORT = "50001"
	}
	c := rpcClient.InitClientPool(fmt.Sprintf("%s:%s", proxyHost, proxyPORT), grpc.WithTransportCredentials(insecure.NewCredentials()))
	return c
}

func createRedisConn() *redis.Client {
	//连接redis-pubsub
	redisHost := os.Getenv("REDIS_PUBSUB_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	if redisHost == "" {
		redisHost = "localhost" // 默认值，仅用于本地测试
	}
	if redisPort == "" {
		redisPort = "6379"
	}
	redisDB := utils.CreateRedisConn(fmt.Sprintf("%s:%s", redisHost, redisPort))
	return redisDB
}

func main() {

	grpcClient := initClientPool()
	redisPubSub := createRedisConn()
	server := service.NewUserService(grpcClient, redisPubSub)
	r := router.Router(server)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	// 启动服务器
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server startup failed: %v", err)
		}
	}()

	// 监听系统信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down gin server...")

	// 创建超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 优雅关闭服务器
	log.Println("Shutting down gin server gracefully...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("gin Server shutdown complete")
}
