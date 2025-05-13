package models

import (
	"gorm.io/gorm"
)

type Contact struct {
	gorm.Model
	OwnerId  uint   `gorm:"index" json:"user_id"`
	TargetId uint   `gorm:"index" json:"friend_id"`
	Status   string `gorm:"default:'pending';not null" json:"status"`
}

func (s *Contact) TableName() string {
	return "contact_table"
}
