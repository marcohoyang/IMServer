package grpc_server

import (
	"context"
	"errors"
	"log"
	"net"
	"time"

	"github.com/hoyang/imserver/src/conveter"
	"github.com/hoyang/imserver/src/models"
	im "github.com/hoyang/imserver/src/proto"
	"github.com/hoyang/imserver/src/utils"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
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
		log.Printf("listen failed %v\n", err)
	}
	rpcServer := grpc.NewServer()
	im.RegisterUserServiceServer(rpcServer, &server{db: db, redis: redis})
	log.Printf("server listening at %v\n", listen.Addr())
	if err := rpcServer.Serve(listen); err != nil {
		log.Printf("failed to serve: %v", err)
	}
}

func (s *server) PublishMsg(ctx context.Context, chanel string, msg string) {
	s.redis.Publish(ctx, chanel, msg)
}

func (s *server) CreateUser(ctx context.Context, user *im.IMUser) (*im.IMUser, error) {
	dbUser := conveter.ToDBIMUser(user)
	result := s.db.Create(dbUser)
	if result.Error != nil {
		// 处理错误
		log.Printf("创建用户失败: %v\n", result.Error)
		return nil, result.Error
	}
	log.Printf("创建用户成功, userId: %v \n", dbUser.ID)

	// 创建成功后，将用户信息存入缓存
	pbUser := conveter.ToPBIMUser(dbUser)
	userData, err := proto.Marshal(pbUser)
	if err != nil {
		log.Printf("序列化新创建用户失败: %v", err)
	} else {
		cacheKey := utils.UserCacheKey(uint64(dbUser.ID))
		s.redis.Set(ctx, cacheKey, userData, 5*time.Minute)
	}

	*user = *conveter.ToPBIMUser(dbUser)
	return pbUser, nil
}

func (s *server) UpdateUser(ctx context.Context, user *im.IMUser) (*im.IMUser, error) {
	dbUser := conveter.ToDBIMUser(user)
	result := s.db.Model(&dbUser).Updates(models.IMUser{Name: dbUser.Name, Password: dbUser.Password, Phone: dbUser.Phone, Email: dbUser.Email, Salt: dbUser.Salt})
	if result.Error != nil {
		// 处理错误
		log.Printf("更新用户失败: %v\n", result.Error)
		return nil, result.Error
	}
	log.Printf("更新用户成功, userId: %v\n", dbUser.ID)
	// 更新成功后，更新缓存或删除缓存（取决于业务需求）
	cacheKey := utils.UserCacheKey(uint64(dbUser.ID))
	pbUser := conveter.ToPBIMUser(dbUser)
	userData, err := proto.Marshal(pbUser)
	if err != nil {
		log.Printf("序列化更新后的用户失败: %v", err)
	} else {
		s.redis.Set(ctx, cacheKey, userData, 5*time.Minute)
	}
	return pbUser, nil
}

func (s *server) GetUser(ctx context.Context, req *im.UserRequest) (*im.IMUser, error) {
	// 检查请求参数
	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "用户名不能为空")
	}

	// 检查请求参数
	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "用户名不能为空")
	}

	// 先从缓存获取用户信息
	cacheKey := utils.UserCacheKey(req.Id)
	cachedUser, err := s.redis.Get(ctx, cacheKey).Result()

	if err == nil {
		// 缓存命中，反序列化并返回
		var pbUser im.IMUser
		if err := proto.Unmarshal([]byte(cachedUser), &pbUser); err != nil {
			log.Printf("反序列化用户缓存失败: %v", err)
			// 缓存数据损坏，继续从数据库查询
		} else {
			log.Printf("从缓存获取用户成功, userId: %d", req.Id)
			return &pbUser, nil
		}
	} else if err != redis.Nil {
		log.Printf("查询 Redis 缓存失败: %v", err)
		// 缓存查询错误，继续从数据库查询
	}
	var dbUser models.IMUser
	result := s.db.Where("name = ?", req.Name).First(&dbUser)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "用户 %s 不存在", req.Name)
		}
		// 其他错误
		log.Printf("查询数据库失败: %v\n", result.Error)
		return nil, status.Errorf(codes.Internal, "服务器内部错误")
	}

	log.Printf("查询用户成功, userId: %v\n", dbUser.ID)
	// 将查询结果存入缓存，设置合理的过期时间（如 5 分钟）
	pbUser := conveter.ToPBIMUser(&dbUser)
	userData, err := proto.Marshal(pbUser)
	if err != nil {
		log.Printf("序列化用户数据失败: %v", err)
	} else {
		s.redis.Set(ctx, cacheKey, userData, 5*time.Minute)
	}
	return pbUser, nil
}

func (s *server) GetFriends(ctx context.Context, req *im.UserRequest) (*im.Friends, error) {
	// 先从缓存获取好友列表
	cacheKey := utils.FriendsCacheKey(req.Id)
	cachedFriends, err := s.redis.Get(ctx, cacheKey).Result()

	if err == nil {
		// 缓存命中，反序列化并返回
		var pbFriends im.Friends
		if err := proto.Unmarshal([]byte(cachedFriends), &pbFriends); err != nil {
			log.Printf("反序列化好友列表缓存失败: %v", err)
		} else {
			log.Printf("从缓存获取好友列表成功, userId: %d", req.Id)
			return &pbFriends, nil
		}
	} else if err != redis.Nil {
		log.Printf("查询 Redis 缓存失败: %v", err)
	}

	var friends []models.FriendView
	// 执行连表查询
	err = s.db.Table("contact_table uf").
		Select(`
            u.id,
            u.name as username,
            u.is_logout,
            uf.status,
            uf.created_at
        `).
		Joins("JOIN user_basic u ON uf.target_id = u.id").
		Where("uf.owner_id  = (?)", req.Id).
		//Where("uf.status = ?", "accepted"). // 只返回已接受的好友关系
		Order("uf.created_at DESC"). // 按创建时间排序
		Find(&friends).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &im.Friends{}, nil // 返回空列表而不是错误
		}
		return nil, err
	}
	friendsView := conveter.FriendViewsToProtos(friends)

	// 将查询结果存入缓存
	friendsData, err := proto.Marshal(friendsView)
	if err != nil {
		log.Printf("序列化好友列表失败: %v", err)
	} else {
		s.redis.Set(ctx, cacheKey, friendsData, 10*time.Minute)
	}

	return friendsView, nil
}

func (s *server) AddFriend(ctx context.Context, contact *im.Contact) (*im.AddResponse, error) {
	resp := im.AddResponse{Success: true}

	tx := s.db.Begin()
	if tx.Error != nil {
		resp.Success = false
		return &resp, tx.Error
	}

	userShip1 := models.Contact{OwnerId: uint(contact.OwnerId), TargetId: uint(contact.TargetId), Status: "accepted"}
	result := s.db.Create(&userShip1)
	// 检查插入是否成功

	if result.Error != nil {
		tx.Rollback()
		resp.Success = false
		return &resp, result.Error
	}

	userShip2 := models.Contact{OwnerId: uint(contact.TargetId), TargetId: uint(contact.OwnerId), Status: "accepted"}
	result = s.db.Create(&userShip2)
	// 检查插入是否成功

	if result.Error != nil {
		tx.Rollback()
		resp.Success = false
		return &resp, result.Error
	}
	if tx.Commit().Error != nil {
		resp.Success = false
		return &resp, result.Error
	}

	// 添加成功后，删除双方的好友列表缓存
	s.redis.Del(ctx, utils.FriendsCacheKey(contact.OwnerId))
	s.redis.Del(ctx, utils.FriendsCacheKey(contact.TargetId))

	return &resp, nil
}
