package utils

import "fmt"

// UserCacheKey 生成用户名的缓存键
func UserCacheKey(username string) string {
	return fmt.Sprintf("user:name:%s", username)
}

// UserIDCacheKey 生成用户ID的缓存键
func UserIDCacheKey(userID uint64) string {
	return fmt.Sprintf("user:id:%d", userID)
}

// 生成好友列表缓存键
func FriendsCacheKey(userID uint64) string {
	return fmt.Sprintf("friends:%d", userID)
}
