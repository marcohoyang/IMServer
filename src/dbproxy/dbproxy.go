package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	init_grpc "github.com/hoyang/imserver/src/dbproxy/init"
	"github.com/hoyang/imserver/src/dbproxy/models"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type User struct {
	gorm.Model // 内嵌 gorm.Model，包含 ID, CreatedAt, UpdatedAt, DeletedAt
	Name       string
	Age        int
	Email      string `gorm:"type:varchar(255);uniqueIndex"`
	IsActive   bool
}

func createRedisConn() *redis.Client {
	// 创建 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis 地址
		Password: "",               // 密码（没有则留空）
		DB:       0,                // 默认数据库
	})
	pong, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("Redis 连接成功:", pong)

	return rdb
}

func createMysqlConn(logger logger.Interface) *gorm.DB {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("警告: 无法加载 .env 文件: %v", err)
	}

	// 获取环境变量（如果 .env 未加载，会尝试从系统环境变量获取）
	dbUser := os.Getenv("MYSQL_USER")
	dbPass := os.Getenv("MYSQL_PASSWORD")
	dbHost := os.Getenv("MYSQL_HOST")
	dbPort := os.Getenv("MYSQL_PORT")
	dbName := os.Getenv("MYSQL_DATABASE")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)
	sqldb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger})
	if err != nil {
		panic("failed to connect database")
	}

	return sqldb
}

var DB *gorm.DB
var RDB *redis.Client

func main() {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      true,
		})
	RDB = createRedisConn()
	DB = createMysqlConn(newLogger)

	fmt.Println("Mysql 连接成功:")

	DB.AutoMigrate(&User{}, &models.IMUser{})

	init_grpc.InitRpcServer(DB)
}
