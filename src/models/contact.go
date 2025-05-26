package models

import (
	"time"

	"gorm.io/gorm"
)

type Contact struct {
	UserID    uint64 `gorm:"primaryKey;index" json:"user_id"`
	FriendID  uint64 `gorm:"primaryKey;index" json:"friend_id"`
	Status    string `gorm:"default:'pending';not null" json:"status"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (s *Contact) TableName() string {
	return "user_friends"
}
