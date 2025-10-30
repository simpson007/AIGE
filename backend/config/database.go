package config

import (
	"log"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	// 从环境变量读取数据库路径，如果未设置则使用默认值
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		// 检测是否在容器中运行
		if _, err := os.Stat("/app"); err == nil {
			dbPath = "/app/data/chat.db"
		} else {
			dbPath = "chat.db" // 开发环境默认路径
		}
	}

	log.Printf("使用数据库路径: %s\n", dbPath)

	// 确保数据库文件的父目录存在
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatal("创建数据库目录失败:", err)
	}
	log.Printf("数据库目录已就绪: %s\n", dbDir)

	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	log.Println("数据库连接成功")
}