package grpc_server

import (
	"context"
	"time"

	"github.com/hoyang/imserver/src/models"
	pb "github.com/hoyang/imserver/src/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// MessageServiceImpl 消息服务实现
type MessageServiceImpl struct {
	pb.UnimplementedMessageServiceServer
	db *gorm.DB
}

// NewMessageService 创建消息服务实例
func NewMessageService(db *gorm.DB) *MessageServiceImpl {
	return &MessageServiceImpl{db: db}
}

// StoreMessage 存储消息
func (s *MessageServiceImpl) StoreMessage(ctx context.Context, req *pb.StoreMessageRequest) (*pb.StoreMessageResponse, error) {
	msg := req.Message

	// 开启事务
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 存储消息
		modelMsg := &models.Message{
			FromID:      msg.FromId,
			ToID:        msg.ToId,
			Type:        msg.Type,
			ContentType: msg.ContentType,
			Content:     msg.Content,
			CreatedAt:   msg.CreatedAt.AsTime(),
			UpdatedAt:   msg.UpdatedAt.AsTime(),
		}

		if err := tx.Create(modelMsg).Error; err != nil {
			return err
		}

		// 2. 如果是私聊消息，创建未读消息记录
		if msg.Type == models.MessageTypePrivate {
			unreadMsg := &models.UnreadMessage{
				UserID:    msg.ToId,
				MessageID: modelMsg.ID,
				CreatedAt: time.Now(),
			}
			if err := tx.Create(unreadMsg).Error; err != nil {
				return err
			}
		}

		msg.Id = modelMsg.ID
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &pb.StoreMessageResponse{
		MessageId: msg.Id,
	}, nil
}

// GetUnreadMessages 获取未读消息（仅用于单聊）
func (s *MessageServiceImpl) GetUnreadMessages(ctx context.Context, req *pb.GetUnreadMessagesRequest) (*pb.GetUnreadMessagesResponse, error) {
	var messages []*models.Message

	// 只查询私聊消息
	query := s.db.Model(&models.Message{}).
		Joins("JOIN unread_messages ON messages.id = unread_messages.message_id").
		Where("unread_messages.user_id = ? AND messages.type = ?", req.UserId, models.MessageTypePrivate)

	if req.LastMessageId > 0 {
		query = query.Where("messages.id > ?", req.LastMessageId)
	}

	err := query.Order("messages.id ASC").
		Limit(int(req.Limit)).
		Find(&messages).Error

	if err != nil {
		return nil, err
	}

	// 转换为 proto 消息
	protoMessages := make([]*pb.Message, len(messages))
	for i, msg := range messages {
		protoMessages[i] = convertToProtoMessage(msg)
	}

	return &pb.GetUnreadMessagesResponse{
		Messages: protoMessages,
	}, nil
}

// GetGroupMessages 获取群聊消息（分页）
func (s *MessageServiceImpl) GetGroupMessages(ctx context.Context, req *pb.GetGroupMessagesRequest) (*pb.GetGroupMessagesResponse, error) {
	var messages []*models.Message

	query := s.db.Model(&models.Message{}).
		Where("type = ? AND to_id = ?", models.MessageTypeGroup, req.GroupId)

	if req.LastMessageId > 0 {
		query = query.Where("id < ?", req.LastMessageId)
	}

	err := query.Order("id DESC").
		Limit(int(req.Limit)).
		Find(&messages).Error

	if err != nil {
		return nil, err
	}

	// 转换为 proto 消息
	protoMessages := make([]*pb.Message, len(messages))
	for i, msg := range messages {
		protoMessages[i] = convertToProtoMessage(msg)
	}

	return &pb.GetGroupMessagesResponse{
		Messages: protoMessages,
	}, nil
}

// convertToProtoMessage 将模型消息转换为 proto 消息
func convertToProtoMessage(msg *models.Message) *pb.Message {
	return &pb.Message{
		Id:          msg.ID,
		FromId:      msg.FromID,
		ToId:        msg.ToID,
		Type:        msg.Type,
		ContentType: msg.ContentType,
		Content:     msg.Content,
		CreatedAt:   timestamppb.New(msg.CreatedAt),
		UpdatedAt:   timestamppb.New(msg.UpdatedAt),
	}
}
