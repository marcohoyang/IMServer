package rpcClient

import (
	"fmt"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

type ClientPool struct {
	pool sync.Pool
}

func InitClientPool(target string, opts ...grpc.DialOption) *ClientPool {

	return &ClientPool{
		pool: sync.Pool{
			New: func() any {
				conn, err := grpc.NewClient(target, opts...)
				if err != nil {
					fmt.Printf("did not connect: %v", err)
				}
				return conn
			},
		},
	}
}

func (c *ClientPool) Get() *grpc.ClientConn {
	conn := c.pool.Get().(*grpc.ClientConn)
	if conn == nil || conn.GetState() == connectivity.Shutdown || conn.GetState() == connectivity.TransientFailure {
		if conn == nil {
			conn.Close()
		}
		conn = c.pool.New().(*grpc.ClientConn)
	}
	return conn
}

func (c *ClientPool) Put(conn *grpc.ClientConn) {
	if conn == nil {
		return
	}
	if conn.GetState() == connectivity.Shutdown || conn.GetState() == connectivity.TransientFailure {
		conn.Close()
		return
	}
	c.pool.Put(conn)
}
