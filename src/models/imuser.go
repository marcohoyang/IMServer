package models

import (
	"time"

	"gorm.io/gorm"
)

type IMUser struct {
	gorm.Model
	Name          string     `json:"username" gorm:"type:varchar(255);unique" binding:"required,max=255"`
	Password      string     `json:"password" gorm:"not null" binding:"required,min=6"`
	Phone         *string    `json:"phone" gorm:"type:varchar(20);unique" binding:"omitempty,e164"`
	Email         *string    `json:"email" gorm:"type:varchar(255);unique;default:null" binding:"omitempty,email"`
	LoginTime     *time.Time `json:"login_time,omitempty"`
	LogoutTime    *time.Time `json:"logout_time,omitempty"`
	HeartbeatTime *time.Time `json:"heartbeat_time,omitempty"`
	ClientIp      string     `json:"client_ip,omitempty" gorm:"type:varchar(45)"`  // IPv6最长45字符
	ClientPort    string     `json:"client_port,omitempty" gorm:"type:varchar(5)"` // 端口范围0-65535
	Identity      string     `json:"identity,omitempty" gorm:"type:varchar(100)"`
	Device        string     `json:"device,omitempty" gorm:"type:varchar(100)"`
	IsLogout      bool       `json:"is_logout" gorm:"default:false"`
	Salt          string
}

func (table *IMUser) TableName() string {
	return "user_basic"
}
