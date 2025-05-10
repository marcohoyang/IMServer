package conveter

import (
	"time"

	"github.com/hoyang/imserver/src/dbproxy/models"
	im "github.com/hoyang/imserver/src/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// ToPBIMUser 将数据库模型转换为 protobuf 消息
func ToPBIMUser(dbUser *models.IMUser) *im.IMUser {
	if dbUser == nil {
		return nil
	}

	return &im.IMUser{
		Id:            uint64(dbUser.ID),
		CreatedAt:     timeToProto(dbUser.CreatedAt),
		UpdatedAt:     timeToProto(dbUser.UpdatedAt),
		DeletedAt:     timeToProto(dbUser.DeletedAt.Time),
		Name:          dbUser.Name,
		Password:      dbUser.Password, // 注意：生产环境不应传输密码
		Phone:         dbUser.Phone,
		Email:         convertEmailToProto(dbUser.Email),
		LoginTime:     convertTimeToProto(dbUser.LoginTime),
		LogoutTime:    convertTimeToProto(dbUser.LogoutTime),
		HeartbeatTime: convertTimeToProto(dbUser.HeartbeatTime),
		ClientIp:      dbUser.ClientIp,
		ClientPort:    dbUser.ClientPort,
		Identity:      dbUser.Identity,
		Device:        dbUser.Device,
		IsLogout:      dbUser.IsLogout,
		Salt:          dbUser.Salt,
	}
}

// ToDBIMUser 将 protobuf 消息转换为数据库模型
func ToDBIMUser(pbUser *im.IMUser) *models.IMUser {
	if pbUser == nil {
		return nil
	}

	return &models.IMUser{
		Model: gorm.Model{
			ID:        uint(pbUser.GetId()),
			CreatedAt: protoToTime(pbUser.GetCreatedAt()),
			UpdatedAt: protoToTime(pbUser.GetUpdatedAt()),
			DeletedAt: gorm.DeletedAt{
				Time:  protoToTime(pbUser.GetDeletedAt()),
				Valid: pbUser.GetDeletedAt() != nil,
			},
		},
		Name:          pbUser.GetName(),
		Password:      pbUser.GetPassword(),
		Phone:         pbUser.GetPhone(),
		Email:         convertProtoToEmail(pbUser.GetEmail()),
		LoginTime:     convertProtoToTime(pbUser.GetLoginTime()),
		LogoutTime:    convertProtoToTime(pbUser.GetLogoutTime()),
		HeartbeatTime: convertProtoToTime(pbUser.GetHeartbeatTime()),
		ClientIp:      pbUser.GetClientIp(),
		ClientPort:    pbUser.GetClientPort(),
		Identity:      pbUser.GetIdentity(),
		Device:        pbUser.GetDevice(),
		IsLogout:      pbUser.GetIsLogout(),
		Salt:          pbUser.GetSalt(),
	}
}

func convertEmailToProto(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}

func convertProtoToEmail(str string) *string {
	if str == "" {
		return nil
	}
	return &str
}

// 辅助函数：time.Time 转换为 protobuf Timestamp
func timeToProto(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

func convertTimeToProto(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil // 或者返回一个默认时间
	}
	return timestamppb.New(*t)
}

// 辅助函数：protobuf Timestamp 转换为 time.Time
func protoToTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}

func convertProtoToTime(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}
