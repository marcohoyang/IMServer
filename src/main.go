package main

import (
	"github.com/hoyang/imserver/src/router"
	rpcClient "github.com/hoyang/imserver/src/rpc"
	"github.com/hoyang/imserver/src/service"
	"github.com/hoyang/imserver/src/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	c := rpcClient.InitClientPool("localhost:50001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	redisDB := utils.CreateRedisConn("localhost:6379")
	server := service.NewUserService(c, redisDB)
	r := router.Router(server)
	r.Run()
}
