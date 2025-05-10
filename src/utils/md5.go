package utils

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

// 小写
func Md5Enconde(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	tmpStr := h.Sum(nil)
	return hex.EncodeToString(tmpStr)
}

// 大写
func MD5Enconde(data string) string {
	return strings.ToUpper(Md5Enconde(data))
}

func MakePassword(plainPwd string, salt string) string {
	return MD5Enconde(plainPwd + salt)
}

func VaildPassword(plainPwd string, salt string, pwd string) bool {
	return MD5Enconde(plainPwd+salt) == pwd
}
