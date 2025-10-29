package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// 定义命令行参数
	username := flag.String("username", "", "用户名（必填）")
	password := flag.String("password", "", "密码（必填）")
	email := flag.String("email", "", "邮箱（选填，默认为 username@example.com）")
	isAdmin := flag.Bool("admin", false, "是否为管理员（默认为普通用户）")
	dbPath := flag.String("db", "../chat.db", "数据库文件路径")

	flag.Parse()

	// 验证必填参数
	if *username == "" || *password == "" {
		fmt.Println("❌ 错误: 用户名和密码不能为空")
		fmt.Println("\n使用方法:")
		fmt.Println("  go run add_user.go -username=用户名 -password=密码 [-email=邮箱] [-admin]")
		fmt.Println("\n示例:")
		fmt.Println("  # 添加普通用户")
		fmt.Println("  go run add_user.go -username=zhangsan -password=123456")
		fmt.Println("\n  # 添加管理员")
		fmt.Println("  go run add_user.go -username=lisi -password=123456 -admin")
		fmt.Println("\n  # 指定邮箱")
		fmt.Println("  go run add_user.go -username=wangwu -password=123456 -email=wangwu@qq.com")
		return
	}

	// 设置默认邮箱
	emailValue := *email
	if emailValue == "" {
		emailValue = *username + "@example.com"
	}

	// 生成密码哈希
	fmt.Println("🔐 正在加密密码...")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("❌ 密码加密失败:", err)
	}

	// 连接数据库
	fmt.Println("📂 正在连接数据库...")
	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		log.Fatal("❌ 数据库连接失败:", err)
	}
	defer db.Close()

	// 检查用户名是否已存在
	var existingCount int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ? AND deleted_at IS NULL", *username).Scan(&existingCount)
	if err != nil {
		log.Fatal("❌ 查询用户失败:", err)
	}
	if existingCount > 0 {
		log.Fatal("❌ 用户名已存在:", *username)
	}

	// 插入新用户
	fmt.Println("✨ 正在创建用户...")
	now := time.Now().Format("2006-01-02 15:04:05.999999-07:00")

	adminValue := 0
	if *isAdmin {
		adminValue = 1
	}

	_, err = db.Exec(`
		INSERT INTO users (username, password, email, is_admin, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, *username, string(hashedPassword), emailValue, adminValue, now, now)

	if err != nil {
		log.Fatal("❌ 用户创建失败:", err)
	}

	// 查询新创建的用户ID
	var userID int
	err = db.QueryRow("SELECT id FROM users WHERE username = ? AND deleted_at IS NULL", *username).Scan(&userID)
	if err != nil {
		log.Fatal("❌ 获取用户ID失败:", err)
	}

	// 显示成功信息
	fmt.Println("\n✅ 用户创建成功！")
	fmt.Println("==========================================")
	fmt.Printf("用户ID:   %d\n", userID)
	fmt.Printf("用户名:   %s\n", *username)
	fmt.Printf("密码:     %s\n", *password)
	fmt.Printf("邮箱:     %s\n", emailValue)
	if *isAdmin {
		fmt.Println("权限:     管理员 ⭐")
	} else {
		fmt.Println("权限:     普通用户")
	}
	fmt.Println("==========================================")
}
