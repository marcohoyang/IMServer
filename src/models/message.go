package models

import (
	"encoding/json"
	"fmt"
	"time"

	im "github.com/hoyang/imserver/src/proto"
)

// MessageType 消息类型
const (
	MessageTypeUnknown = 0 // 未知消息
	MessageTypePrivate = 1 // 私聊消息
	MessageTypeGroup   = 2 // 群聊消息
)

// Message 消息模型
type Message struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	FromID      uint64         `gorm:"index;not null" json:"FormId"`      // 发送者ID
	ToID        uint64         `gorm:"index;not null" json:"TargetId"`    // 接收者ID（私聊为用户ID，群聊为群组ID）
	Type        im.MessageType `gorm:"not null" json:"Type"`              // 消息类型：1-私聊 2-群聊
	ContentType im.ContentType `gorm:"not null" json:"ContentType"`       // 消息内容类型：1-文本 2-图片 3-语音 4-视频 5-文件
	Content     []byte         `gorm:"type:blob;not null" json:"Content"` // 消息内容（二进制数据）
	CreatedAt   time.Time      `gorm:"not null" json:"created_at"`        // 创建时间
	UpdatedAt   time.Time      `gorm:"not null" json:"updated_at"`        // 更新时间
}

// UnreadMessage 未读消息记录（仅用于单聊）
type UnreadMessage struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"index;not null" json:"user_id"`    // 用户ID
	MessageID uint64    `gorm:"index;not null" json:"message_id"` // 消息ID
	CreatedAt time.Time `gorm:"not null" json:"created_at"`       // 创建时间
}

// TableName 指定表名
func (Message) TableName() string {
	return "messages"
}

// TableName 指定表名
func (UnreadMessage) TableName() string {
	return "unread_messages"
}

func MessageFromString(jsonStr string) (Message, error) {
	var msg Message
	err := json.Unmarshal([]byte(jsonStr), &msg)
	return msg, err
}

func (m *Message) String() string {
	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Sprintf("Message{error: %v}", err)
	}
	return string(data)
}
