package main

import (
	"fmt"
	"log"
	"os"
	"time"

	grpc_server "github.com/hoyang/imserver/src/dbproxy/rpcserver"
	"github.com/hoyang/imserver/src/models"
	"github.com/hoyang/imserver/src/utils"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func createMysqlConn() *gorm.DB {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      true,
		})
	viper.SetConfigName("config") // 设置配置文件名（不带扩展名）
	viper.SetConfigType("yaml")   // 如果配置文件没有扩展名，则需要指定类型
	viper.AddConfigPath(".")      // 添加当前目录作为搜索路径

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Error reading config file, using defaults.", err)
	}

	port := viper.GetString("database.port")
	host := os.Getenv("MYSQL_HOST")
	if host == "" {
		host = viper.GetString("database.host") // 默认值，仅用于本地测试
	}
	user := viper.GetString("database.user")
	pass := viper.GetString("database.password")
	dbname := viper.GetString("database.dbname")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, pass, host, port, dbname)
	sqldb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: newLogger})
	if err != nil {
		log.Println("failed to connect database", err)
		panic("failed to connect database")
	}

	return sqldb
}

func main() {
	redisHost := os.Getenv("REDIS_CACHE_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	if redisHost == "" {
		redisHost = "localhost" // 默认值，仅用于本地测试
	}
	if redisPort == "" {
		redisPort = "6379"
	}
	redis := utils.CreateRedisConn(fmt.Sprintf("%s:%s", redisHost, redisPort))
	db := createMysqlConn()

	log.Println("Mysql 连接成功")

	db.AutoMigrate(&models.IMUser{}, &models.Contact{}, &models.Message{}, &models.UnreadMessage{})

	grpc_server.StartRpcServer(db, redis)
}
