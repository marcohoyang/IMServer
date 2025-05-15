package main

import (
	"fmt"
	"os"

	"github.com/hoyang/imserver/src/router"
	rpcClient "github.com/hoyang/imserver/src/rpc"
	"github.com/hoyang/imserver/src/service"
	"github.com/hoyang/imserver/src/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	proxyHost := os.Getenv("DB_PROXY_HOST")
	proxyPORT := os.Getenv("DB_PROXY_PORT")
	if proxyHost == "" {
		proxyHost = "localhost" // 默认值，仅用于本地测试
	}
	if proxyPORT == "" {
		proxyPORT = "50001"
	}
	c := rpcClient.InitClientPool(fmt.Sprintf("%s:%s", proxyHost, proxyPORT), grpc.WithTransportCredentials(insecure.NewCredentials()))

	redisHost := os.Getenv("REDIS_PUBSUB_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	if redisHost == "" {
		redisHost = "localhost" // 默认值，仅用于本地测试
	}
	if redisPort == "" {
		redisPort = "6379"
	}
	redisDB := utils.CreateRedisConn(fmt.Sprintf("%s:%s", redisHost, redisPort))

	server := service.NewUserService(c, redisDB)
	r := router.Router(server)
	r.Run()
}
