package rpcClient

import (
	"context"
	"time"

	im "github.com/hoyang/imserver/src/proto"
	pb "github.com/hoyang/imserver/src/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MessageProxy 消息服务代理
type MessageProxy struct {
	client pb.MessageServiceClient
}

// NewMessageProxy 创建消息服务代理
func NewMessageProxy(conn *grpc.ClientConn) *MessageProxy {
	return &MessageProxy{
		client: pb.NewMessageServiceClient(conn),
	}
}

// StoreMessage 存储消息
func (p *MessageProxy) StoreMessage(fromID, toID uint64, msgType im.MessageType, content []byte) (uint64, error) {
	now := time.Now()
	msg := &pb.Message{
		FromId:    fromID,
		ToId:      toID,
		Type:      msgType,
		Content:   content,
		CreatedAt: timestamppb.New(now),
		UpdatedAt: timestamppb.New(now),
	}

	resp, err := p.client.StoreMessage(context.Background(), &pb.StoreMessageRequest{
		Message: msg,
	})
	if err != nil {
		return 0, err
	}

	return resp.MessageId, nil
}

// GetUnreadMessages 获取未读消息
func (p *MessageProxy) GetUnreadMessages(userID, lastMessageID uint64, limit int32) ([]*pb.Message, error) {
	resp, err := p.client.GetUnreadMessages(context.Background(), &pb.GetUnreadMessagesRequest{
		UserId:        userID,
		LastMessageId: lastMessageID,
		Limit:         limit,
	})
	if err != nil {
		return nil, err
	}

	return resp.Messages, nil
}

// GetGroupMessages 获取群聊消息
func (p *MessageProxy) GetGroupMessages(groupID, lastMessageID uint64, limit int32) ([]*pb.Message, error) {
	resp, err := p.client.GetGroupMessages(context.Background(), &pb.GetGroupMessagesRequest{
		GroupId:       groupID,
		LastMessageId: lastMessageID,
		Limit:         limit,
	})
	if err != nil {
		return nil, err
	}

	return resp.Messages, nil
}
