package main

import (
	"fmt"
	"log"
	"os"
	"time"

	grpc_server "github.com/hoyang/imserver/src/dbproxy/rpcserver"
	"github.com/hoyang/imserver/src/models"
	"github.com/hoyang/imserver/src/utils"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func createMysqlConn(logger logger.Interface) *gorm.DB {
	err := godotenv.Load()
	if err != nil {
		log.Printf("警告: 无法加载 .env 文件: %v", err)
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

func main() {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      true,
		})
	redis := utils.CreateRedisConn("localhost:6379")
	db := createMysqlConn(newLogger)

	log.Println("Mysql 连接成功:")

	db.AutoMigrate(&models.IMUser{}, &models.Contact{})

	grpc_server.InitRpcServer(db, redis)
}
