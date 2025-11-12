package controllers

import (
	"AIGE/config"
	"AIGE/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ChatRecord 用于返回聊天记录的结构
type ChatRecord struct {
	ID                uint      `json:"id"`
	UserID            uint      `json:"user_id"`
	Username          string    `json:"username"`
	ModID             string    `json:"mod_id"`
	SessionDate       string    `json:"session_date"`
	State             string    `json:"state"`
	RecentHistory     string    `json:"recent_history"`
	CompressedSummary string    `json:"compressed_summary"`
	CompressionRound  int       `json:"compression_round"`
	DisplayHistory    string    `json:"display_history"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// GetAllChats 获取所有用户的聊天记录
func GetAllChats(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	userID := c.Query("user_id")
	modID := c.Query("mod_id")
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var total int64
	query := config.DB.Model(&models.GameSave{}).Where("deleted_at IS NULL")

	// 应用过滤条件
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if modID != "" {
		query = query.Where("mod_id = ?", modID)
	}
	if search != "" {
		query = query.Where("display_history LIKE ? OR recent_history LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// 获取总数
	query.Count(&total)

	// 获取分页数据
	var gameSaves []models.GameSave
	offset := (page - 1) * pageSize
	if err := query.Order("updated_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&gameSaves).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取聊天记录失败"})
		return
	}

	// 获取关联的用户信息
	var userIDs []uint
	for _, save := range gameSaves {
		userIDs = append(userIDs, save.UserID)
	}

	var users []models.User
	userMap := make(map[uint]string)
	if len(userIDs) > 0 {
		config.DB.Where("id IN ?", userIDs).Find(&users)
		for _, user := range users {
			userMap[user.ID] = user.Username
		}
	}

	// 构建响应数据
	var chatRecords []ChatRecord
	for _, save := range gameSaves {
		chatRecords = append(chatRecords, ChatRecord{
			ID:                save.ID,
			UserID:            save.UserID,
			Username:          userMap[save.UserID],
			ModID:             save.ModID,
			SessionDate:       save.SessionDate,
			State:             save.State,
			RecentHistory:     save.RecentHistory,
			CompressedSummary: save.CompressedSummary,
			CompressionRound:  save.CompressionRound,
			DisplayHistory:    save.DisplayHistory,
			CreatedAt:         save.CreatedAt,
			UpdatedAt:         save.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"records": chatRecords,
		"total":   total,
		"page":    page,
		"page_size": pageSize,
	})
}

// GetChat 获取单条聊天记录
func GetChat(c *gin.Context) {
	chatID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的聊天记录ID"})
		return
	}

	var gameSave models.GameSave
	if err := config.DB.First(&gameSave, chatID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "聊天记录不存在"})
		return
	}

	// 获取用户信息
	var user models.User
	config.DB.First(&user, gameSave.UserID)

	chatRecord := ChatRecord{
		ID:                gameSave.ID,
		UserID:            gameSave.UserID,
		Username:          user.Username,
		ModID:             gameSave.ModID,
		SessionDate:       gameSave.SessionDate,
		State:             gameSave.State,
		RecentHistory:     gameSave.RecentHistory,
		CompressedSummary: gameSave.CompressedSummary,
		CompressionRound:  gameSave.CompressionRound,
		DisplayHistory:    gameSave.DisplayHistory,
		CreatedAt:         gameSave.CreatedAt,
		UpdatedAt:         gameSave.UpdatedAt,
	}

	c.JSON(http.StatusOK, chatRecord)
}

// UpdateChat 更新聊天记录
func UpdateChat(c *gin.Context) {
	chatID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的聊天记录ID"})
		return
	}

	var gameSave models.GameSave
	if err := config.DB.First(&gameSave, chatID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "聊天记录不存在"})
		return
	}

	// 接收更新数据
	var updateData struct {
		State             string `json:"state"`
		RecentHistory     string `json:"recent_history"`
		CompressedSummary string `json:"compressed_summary"`
		DisplayHistory    string `json:"display_history"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 更新记录
	updates := map[string]interface{}{}
	if updateData.State != "" {
		updates["state"] = updateData.State
	}
	if updateData.RecentHistory != "" {
		updates["recent_history"] = updateData.RecentHistory
	}
	if updateData.CompressedSummary != "" {
		updates["compressed_summary"] = updateData.CompressedSummary
	}
	if updateData.DisplayHistory != "" {
		updates["display_history"] = updateData.DisplayHistory
	}

	if err := config.DB.Model(&gameSave).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新聊天记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "聊天记录更新成功"})
}

// DeleteChat 删除聊天记录
func DeleteChat(c *gin.Context) {
	chatID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的聊天记录ID"})
		return
	}

	var gameSave models.GameSave
	if err := config.DB.First(&gameSave, chatID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "聊天记录不存在"})
		return
	}

	// 软删除
	if err := config.DB.Delete(&gameSave).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除聊天记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "聊天记录删除成功"})
}

// DeleteUserChats 删除指定用户的所有聊天记录
func DeleteUserChats(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 检查用户是否存在
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 删除该用户的所有聊天记录
	if err := config.DB.Where("user_id = ?", userID).Delete(&models.GameSave{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除聊天记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户聊天记录已全部删除"})
}

// GetChatStats 获取聊天统计信息
func GetChatStats(c *gin.Context) {
	var stats struct {
		TotalChats   int64 `json:"total_chats"`
		ActiveUsers  int64 `json:"active_users"`
		TodayChats   int64 `json:"today_chats"`
		AverageChats float64 `json:"average_chats_per_user"`
	}

	// 总聊天数
	config.DB.Model(&models.GameSave{}).Count(&stats.TotalChats)

	// 活跃用户数（有聊天记录的用户）
	config.DB.Model(&models.GameSave{}).
		Select("COUNT(DISTINCT user_id)").
		Row().
		Scan(&stats.ActiveUsers)

	// 今日聊天数
	today := time.Now().Format("2006-01-02")
	config.DB.Model(&models.GameSave{}).
		Where("DATE(updated_at) = ?", today).
		Count(&stats.TodayChats)

	// 平均每用户聊天数
	if stats.ActiveUsers > 0 {
		stats.AverageChats = float64(stats.TotalChats) / float64(stats.ActiveUsers)
	}

	c.JSON(http.StatusOK, stats)
}

// ExportUserChats 导出用户聊天记录
func ExportUserChats(c *gin.Context) {
	userID := c.Query("user_id")
	format := c.DefaultQuery("format", "json")

	query := config.DB.Model(&models.GameSave{})
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	var gameSaves []models.GameSave
	if err := query.Order("created_at DESC").Find(&gameSaves).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取聊天记录失败"})
		return
	}

	// 根据格式导出
	switch format {
	case "json":
		c.Header("Content-Disposition", "attachment; filename=chat_export.json")
		c.JSON(http.StatusOK, gameSaves)
	case "csv":
		// TODO: 实现CSV导出
		c.JSON(http.StatusNotImplemented, gin.H{"error": "CSV导出功能尚未实现"})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的导出格式"})
	}
}