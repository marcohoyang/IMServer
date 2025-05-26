package utils

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// 自定义声明结构体，包含用户ID等信息
type Claims struct {
	UserID uint64 `json:"user_id"`
	jwt.RegisteredClaims
}

// 密钥（生产环境应从配置文件或环境变量获取）
var jwtKey = []byte("my-secret-key")

func GenerateToken(userID uint64) (string, error) {
	// 设置过期时间（例如24小时）
	expirationTime := time.Now().Add(24 * time.Hour)

	// 创建声明
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "im_server",
		},
	}

	// 创建 Token 对象，使用 HS256 算法
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 生成签名字符串
	return token.SignedString(jwtKey)
}

func JWTAuthMiddlewareForWS() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie("token")
		if err != nil {
			c.JSON(401, gin.H{"error": "未找到认证 Cookie"})
			return
		}

		// 验证 Token
		tokenStr := strings.TrimPrefix(cookie, "Bearer ")
		praseToken(c, tokenStr)
	}
}

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Header 中获取 Token
		tokenStr := c.GetHeader("Authorization")

		// 验证 Token 格式
		if tokenStr == "" {
			c.JSON(401, gin.H{"error": "请求头中缺少 Authorization 字段"})
			c.Abort()
			return
		}

		// 提取 Token（格式：Bearer <token>）
		if len(tokenStr) <= 7 || tokenStr[:7] != "Bearer " {
			c.JSON(401, gin.H{"error": "无效的 Authorization 格式"})
			c.Abort()
			return
		}
		tokenStr = tokenStr[7:]
		praseToken(c, tokenStr)
	}
}

func praseToken(c *gin.Context, tokenStr string) {
	// 解析 Token
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})

	// 验证 Token
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			c.JSON(401, gin.H{"error": "Token已过期"})
			c.Abort()
			return
		}
		c.JSON(401, gin.H{"error": "无效的Token"})
		c.Abort()
		return
	}

	if !token.Valid {
		c.JSON(401, gin.H{"error": "无效的Token"})
		c.Abort()
		return
	}

	// 将用户ID存入上下文，后续处理可直接获取
	log.Printf("set userID %v\n", claims.UserID)
	c.Set("user_id", claims.UserID)
	c.Next()
}
