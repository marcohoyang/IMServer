package models

// FriendView 用于查询的好友视图模型
type FriendView struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Online   bool   `json:"online"`
}
