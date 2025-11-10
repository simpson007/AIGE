package game_engine

import (
	"fmt"
	"testing"
	"time"
)

func TestEntityManager(t *testing.T) {
	// 创建实体管理器
	em := NewEntityManager()

	// 测试玩家实体注册
	playerEntity := &Entity{
		ID:   "player_123",
		Type: EntityPlayer,
		Name: "林逸风",
		Attributes: map[string]interface{}{
			"gender":     "男",
			"cultivation": "三转巅峰",
			"background": "散修出身",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 注册实体
	err := em.RegisterEntity("123", "xiuxian", playerEntity)
	if err != nil {
		t.Errorf("Failed to register player entity: %v", err)
	}

	// 获取实体
	retrieved, err := em.GetEntity("123", "xiuxian", "player_123")
	if err != nil {
		t.Errorf("Failed to get player entity: %v", err)
	}

	if retrieved.Name != "林逸风" {
		t.Errorf("Expected name '林逸风', got '%s'", retrieved.Name)
	}

	// 测试锁定字段更新（应该失败）
	err = em.UpdateEntity("123", "xiuxian", "player_123", map[string]interface{}{
		"gender": "女", // 尝试修改锁定字段
	})
	if err == nil {
		t.Errorf("Expected error when modifying locked field, but got none")
	}

	// 测试非锁定字段更新（应该成功）
	err = em.UpdateEntity("123", "xiuxian", "player_123", map[string]interface{}{
		"cultivation": "六转初阶",
	})
	if err != nil {
		t.Errorf("Failed to update non-locked field: %v", err)
	}

	// 测试序列化和反序列化
	serialized, err := em.SerializeRegistry("123", "xiuxian")
	if err != nil {
		t.Errorf("Failed to serialize registry: %v", err)
	}

	// 创建新的管理器并反序列化
	em2 := NewEntityManager()
	err = em2.DeserializeRegistry("123", "xiuxian", serialized)
	if err != nil {
		t.Errorf("Failed to deserialize registry: %v", err)
	}

	// 验证反序列化后的数据
	retrieved2, err := em2.GetEntity("123", "xiuxian", "player_123")
	if err != nil {
		t.Errorf("Failed to get entity after deserialization: %v", err)
	}

	if retrieved2.Name != "林逸风" {
		t.Errorf("Entity name mismatch after deserialization")
	}

	// 测试一致性验证
	aiResponse := "她走向了市场" // 错误的性别称谓
	err = em.ValidateConsistency("123", "xiuxian", aiResponse)
	if err == nil {
		t.Errorf("Expected consistency validation to fail, but it passed")
	}

	fmt.Println("All entity manager tests passed!")
}

func TestEntityIntegration(t *testing.T) {
	// 模拟游戏控制器
	stateManager := NewStateManager(false, 5*time.Minute)
	gc := &GameController{
		stateManager: stateManager,
	}

	integration := NewEntityIntegration(gc)

	// 测试游戏开始时的实体注册
	customAttrs := map[string]interface{}{
		"姓名":   "苏小小",
		"性别":   "女",
		"资质":   "甲等资质",
		"修为":   "一转中阶",
		"元石":   float64(100),
		"出身":   "南疆苏家",
	}

	err := integration.ProcessGameStartWithEntities("456", "xiuxian", customAttrs)
	if err != nil {
		t.Errorf("Failed to process game start with entities: %v", err)
	}

	// 获取注册的实体
	em := stateManager.GetEntityManager()
	playerEntity, err := em.GetEntity("456", "xiuxian", "player_456")
	if err != nil {
		t.Errorf("Failed to get player entity: %v", err)
	}

	if playerEntity.Name != "苏小小" {
		t.Errorf("Expected player name '苏小小', got '%s'", playerEntity.Name)
	}

	if gender, ok := playerEntity.Attributes["gender"].(string); !ok || gender != "女" {
		t.Errorf("Expected gender '女', got '%v'", playerEntity.Attributes["gender"])
	}

	// 测试一致性验证
	aiResponse1 := "苏小小缓缓走向丹药铺，她的步伐轻盈。" // 正确的性别
	isValid, violations := integration.ValidateResponseConsistency("456", "xiuxian", aiResponse1)
	if !isValid {
		t.Errorf("Valid response marked as invalid: %v", violations)
	}

	aiResponse2 := "苏小小看着眼前的灵草，他决定购买。" // 错误的性别
	isValid, violations = integration.ValidateResponseConsistency("456", "xiuxian", aiResponse2)
	if isValid {
		t.Errorf("Invalid response marked as valid")
	}

	fmt.Println("All entity integration tests passed!")
}

// 测试实体上下文构建
func TestEntityContextBuilding(t *testing.T) {
	em := NewEntityManager()

	// 注册玩家实体
	playerEntity := &Entity{
		ID:   "player_789",
		Type: EntityPlayer,
		Name: "叶无痕",
		Attributes: map[string]interface{}{
			"gender":     "男",
			"cultivation": "五转巅峰",
			"sect":       "灵源斋",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	em.RegisterEntity("789", "xiuxian", playerEntity)

	// 注册NPC实体
	npcEntity := &Entity{
		ID:   "npc_elder_wang",
		Type: EntityNPC,
		Name: "王长老",
		Attributes: map[string]interface{}{
			"position": "灵源斋长老",
			"relationship": "师父",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	em.RegisterEntity("789", "xiuxian", npcEntity)

	// 构建上下文
	context := em.BuildEntityContext("789", "xiuxian")

	// 验证上下文包含必要信息
	if context == "" {
		t.Errorf("Entity context is empty")
	}

	expectedStrings := []string{
		"叶无痕",
		"gender: 男",
		"王长老",
		"灵源斋",
	}

	for _, expected := range expectedStrings {
		if !contains(context, expected) {
			t.Errorf("Context missing expected string: %s", expected)
		}
	}

	fmt.Println("Entity context building test passed!")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}