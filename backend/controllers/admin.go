package controllers

import (
	"AIGE/config"
	"AIGE/models"
	"AIGE/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UpdatePasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required"`
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	IsAdmin  bool   `json:"is_admin"`
}

type UpdateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	IsAdmin  *bool  `json:"is_admin"`
}

func GetUsers(c *gin.Context) {
	var users []models.User
	if err := config.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户列表失败"})
		return
	}

	var userList []UserInfo
	for _, user := range users {
		userList = append(userList, UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			IsAdmin:  user.IsAdmin,
		})
	}

	c.JSON(http.StatusOK, gin.H{"users": userList})
}

func GetUser(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		IsAdmin:  user.IsAdmin,
	})
}

func UpdateUserPassword(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 查找用户
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 加密新密码
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	// 更新密码
	if err := config.DB.Model(&user).Update("password", hashedPassword).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码更新成功"})
}

func DeleteUser(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 检查是否为管理员自己
	currentUserID, _ := c.Get("user_id")
	if uint(userID) == currentUserID.(uint) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能删除自己的账号"})
		return
	}

	// 查找用户
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 删除用户（物理删除）
	if err := config.DB.Unscoped().Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户删除成功"})
}

func ToggleUserAdmin(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 查找用户
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 切换管理员状态
	if err := config.DB.Model(&user).Update("is_admin", !user.IsAdmin).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "权限更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "权限更新成功"})
}

// CreateUser 管理员创建新用户
func CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 检查用户名是否已存在
	var existingUser models.User
	if err := config.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
		return
	}

	// 检查邮箱是否已存在
	if err := config.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "邮箱已存在"})
		return
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	// 创建用户
	user := models.User{
		Username: req.Username,
		Password: hashedPassword,
		Email:    req.Email,
		IsAdmin:  req.IsAdmin,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户创建失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "用户创建成功",
		"user": UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			IsAdmin:  user.IsAdmin,
		},
	})
}

// UpdateUser 管理员更新用户信息
func UpdateUser(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 查找用户
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 构建更新字段
	updates := make(map[string]interface{})

	// 更新用户名
	if req.Username != "" && req.Username != user.Username {
		// 检查新用户名是否已存在
		var existingUser models.User
		if err := config.DB.Where("username = ? AND id != ?", req.Username, userID).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
			return
		}
		updates["username"] = req.Username
	}

	// 更新邮箱
	if req.Email != "" && req.Email != user.Email {
		// 检查新邮箱是否已存在
		var existingUser models.User
		if err := config.DB.Where("email = ? AND id != ?", req.Email, userID).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "邮箱已存在"})
			return
		}
		updates["email"] = req.Email
	}

	// 更新管理员状态
	if req.IsAdmin != nil {
		updates["is_admin"] = *req.IsAdmin
	}

	// 执行更新
	if len(updates) > 0 {
		if err := config.DB.Model(&user).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "用户信息更新失败"})
			return
		}
	}

	// 重新获取更新后的用户信息
	config.DB.First(&user, userID)

	c.JSON(http.StatusOK, gin.H{
		"message": "用户信息更新成功",
		"user": UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			IsAdmin:  user.IsAdmin,
		},
	})
}