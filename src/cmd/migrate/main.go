package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"github.com/hoyang/imserver/src/mysql/migrator"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 解析命令行参数
	var (
		dsn            string
		migrationsPath string
		rollback       bool
	)

	flag.StringVar(&dsn, "dsn", "", "数据库连接字符串")
	flag.StringVar(&migrationsPath, "path", "src/mysql/migrations", "迁移文件目录")
	flag.BoolVar(&rollback, "rollback", false, "是否回滚最后一次迁移")
	flag.Parse()

	if dsn == "" {
		log.Fatal("请提供数据库连接字符串")
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 获取迁移文件目录的绝对路径
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		log.Fatalf("获取迁移文件目录绝对路径失败: %v", err)
	}

	// 创建迁移管理器
	m := migrator.NewMigrator(db)

	// 执行迁移或回滚
	if rollback {
		if err := m.Rollback(absPath); err != nil {
			log.Fatalf("回滚迁移失败: %v", err)
		}
		fmt.Println("回滚成功")
	} else {
		if err := m.Migrate(absPath); err != nil {
			log.Fatalf("执行迁移失败: %v", err)
		}
		fmt.Println("迁移成功")
	}
}
