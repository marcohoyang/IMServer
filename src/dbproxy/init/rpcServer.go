package grpc_server

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/hoyang/imserver/src/conveter"
	"github.com/hoyang/imserver/src/dbproxy/models"
	im "github.com/hoyang/imserver/src/proto"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type server struct {
	im.UnimplementedUserServiceServer
	db    *gorm.DB
	redis *redis.Client
}

func InitRpcServer(db *gorm.DB, redis *redis.Client) {
	listen, err := net.Listen("tcp", ":50001")
	if err != nil {
		fmt.Printf("listen failed %v\n", err)
	}
	rpcServer := grpc.NewServer()
	im.RegisterUserServiceServer(rpcServer, &server{db: db, redis: redis})
	fmt.Printf("server listening at %v\n", listen.Addr())
	if err := rpcServer.Serve(listen); err != nil {
		fmt.Printf("failed to serve: %v", err)
	}
}

func (s *server) PublishMsg(ctx context.Context, chanel string, msg string) {
	s.redis.Publish(ctx, chanel, msg)
}

func (s *server) CreateUser(ctx context.Context, user *im.IMUser) (*im.IMUser, error) {
	fmt.Println("call CreateUser")
	dbUser := conveter.ToDBIMUser(user)
	result := s.db.Create(dbUser)
	if result.Error != nil {
		// 处理错误
		fmt.Printf("创建用户失败: %v\n", result.Error)
		return nil, result.Error
	}
	fmt.Printf("创建用户成功, userId: %v \n", dbUser.ID)

	*user = *conveter.ToPBIMUser(dbUser)
	return user, nil
}

func (s *server) UpdateUser(ctx context.Context, user *im.IMUser) (*im.IMUser, error) {
	dbUser := conveter.ToDBIMUser(user)
	result := s.db.Model(&dbUser).Updates(models.IMUser{Name: dbUser.Name, Password: dbUser.Password, Phone: dbUser.Phone, Email: dbUser.Email, Salt: dbUser.Salt})
	if result.Error != nil {
		// 处理错误
		fmt.Printf("更新用户失败: %v\n", result.Error)
		return nil, result.Error
	}
	fmt.Printf("更新用户成功, userId: %v\n", dbUser.ID)

	*user = *conveter.ToPBIMUser(dbUser)
	return user, nil
}

func (s *server) GetUser(ctx context.Context, req *im.UserRequest) (*im.IMUser, error) {
	// 检查请求参数
	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "用户名不能为空")
	}
	var dbUser models.IMUser
	result := s.db.Where("name = ?", req.Name).First(&dbUser)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "用户 %s 不存在", req.Name)
		}
		// 其他错误
		fmt.Printf("查询数据库失败: %v\n", result.Error)
		return nil, status.Errorf(codes.Internal, "服务器内部错误")
	}

	fmt.Printf("查询用户成功, userId: %v\n", dbUser.ID)
	return conveter.ToPBIMUser(&dbUser), nil
}
