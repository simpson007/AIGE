package main

import (
	"AIGE/config"
	"AIGE/models"
	"AIGE/routes"
	"AIGE/utils"
	"log"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化数据库
	config.InitDB()

	// 自动迁移
	models.AutoMigrate()

	// 创建默认管理员用户
	utils.CreateDefaultAdmin()

	// 设置路由
	r := gin.Default()

	// 配置CORS - 支持开发环境和生产环境
	allowedOrigins := []string{
		"http://localhost:8000",
		"http://localhost:5173",
		"http://localhost:3000",
	}

	// 从环境变量读取额外的允许域名
	if extraOrigins := os.Getenv("ALLOWED_ORIGINS"); extraOrigins != "" {
		origins := strings.Split(extraOrigins, ",")
		for _, origin := range origins {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				allowedOrigins = append(allowedOrigins, origin)
			}
		}
	}

	log.Printf("CORS 允许的域名: %v\n", allowedOrigins)

	r.Use(cors.New(cors.Config{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"Upgrade",
			"Connection",
			"Sec-WebSocket-Extensions",
			"Sec-WebSocket-Key",
			"Sec-WebSocket-Version",
		},
		AllowCredentials: true,
		ExposeHeaders:    []string{"Content-Length"},
		MaxAge:           86400,
	}))

	// 注册路由
	routes.SetupRoutes(r)

	log.Println("服务器启动在端口 :8182")
	r.Run(":8182")
}