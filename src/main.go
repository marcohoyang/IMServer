package main

import (
	"github.com/hoyang/imserver/src/router"
	rpcClient "github.com/hoyang/imserver/src/rpc"
	"github.com/hoyang/imserver/src/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	c := rpcClient.InitClientPool("localhost:50001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	server := service.NewUserService(c)
	r := router.Router(server)
	r.Run()
}
