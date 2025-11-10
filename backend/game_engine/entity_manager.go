package game_engine

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// EntityType 实体类型
type EntityType string

const (
	EntityPlayer EntityType = "player"
	EntityNPC    EntityType = "npc"
	EntityItem   EntityType = "item"
	EntityPlace  EntityType = "place"
)

// Entity 核心实体结构
type Entity struct {
	ID         string                 `json:"id"`          // 唯一标识符
	Type       EntityType             `json:"type"`        // 实体类型
	Name       string                 `json:"name"`        // 名称
	Attributes map[string]interface{} `json:"attributes"`  // 属性集合
	Locked     bool                   `json:"locked"`      // 是否锁定（锁定后不可修改）
	CreatedAt  time.Time              `json:"created_at"`  // 创建时间
	UpdatedAt  time.Time              `json:"updated_at"`  // 更新时间
}

// EntityRegistry 实体注册表
type EntityRegistry struct {
	Entities      map[string]*Entity `json:"entities"`       // 所有实体
	PlayerEntity  *Entity            `json:"player_entity"`  // 玩家实体引用
	LockedFields  []string           `json:"locked_fields"`  // 锁定的字段列表
	Relationships map[string]string  `json:"relationships"`  // 实体关系映射
}

// EntityManager 实体管理器
type EntityManager struct {
	registries map[string]*EntityRegistry // playerID+modID -> registry
	mu         sync.RWMutex
	validator  *EntityValidator
}

// EntityValidator 实体验证器
type EntityValidator struct {
	rules map[string]ValidationRule
}

// ValidationRule 验证规则
type ValidationRule struct {
	Field       string
	Required    bool
	Immutable   bool
	Type        string
	ValidValues []interface{}
}

// NewEntityManager 创建实体管理器
func NewEntityManager() *EntityManager {
	return &EntityManager{
		registries: make(map[string]*EntityRegistry),
		validator:  NewEntityValidator(),
	}
}

// NewEntityValidator 创建验证器
func NewEntityValidator() *EntityValidator {
	return &EntityValidator{
		rules: map[string]ValidationRule{
			"player.gender": {
				Field:       "gender",
				Required:    true,
				Immutable:   true,
				Type:        "string",
				ValidValues: []interface{}{"男", "女", "male", "female"},
			},
			"player.name": {
				Field:     "name",
				Required:  true,
				Immutable: false,
				Type:      "string",
			},
			"npc.name": {
				Field:     "name",
				Required:  true,
				Immutable: false,
				Type:      "string",
			},
		},
	}
}

// GetOrCreateRegistry 获取或创建注册表
func (em *EntityManager) GetOrCreateRegistry(playerID, modID string) *EntityRegistry {
	em.mu.Lock()
	defer em.mu.Unlock()

	key := fmt.Sprintf("%s_%s", playerID, modID)
	if registry, exists := em.registries[key]; exists {
		return registry
	}

	registry := &EntityRegistry{
		Entities:      make(map[string]*Entity),
		Relationships: make(map[string]string),
		LockedFields: []string{
			"player.gender",
			"player.birthplace",
			"player.initial_talent",
		},
	}
	em.registries[key] = registry
	return registry
}

// RegisterEntity 注册实体
func (em *EntityManager) RegisterEntity(playerID, modID string, entity *Entity) error {
	registry := em.GetOrCreateRegistry(playerID, modID)

	// 验证实体
	if err := em.validator.ValidateEntity(entity); err != nil {
		return fmt.Errorf("entity validation failed: %v", err)
	}

	// 如果是玩家实体，特殊处理
	if entity.Type == EntityPlayer {
		if registry.PlayerEntity != nil && registry.PlayerEntity.ID != entity.ID {
			return fmt.Errorf("player entity already exists with ID: %s", registry.PlayerEntity.ID)
		}
		registry.PlayerEntity = entity
	}

	// 添加到注册表
	entity.UpdatedAt = time.Now()
	if entity.CreatedAt.IsZero() {
		entity.CreatedAt = time.Now()
	}
	registry.Entities[entity.ID] = entity

	log.Printf("Registered entity: %s (type: %s, name: %s)", entity.ID, entity.Type, entity.Name)
	return nil
}

// GetEntity 获取实体
func (em *EntityManager) GetEntity(playerID, modID, entityID string) (*Entity, error) {
	registry := em.GetOrCreateRegistry(playerID, modID)

	if entity, exists := registry.Entities[entityID]; exists {
		return entity, nil
	}

	return nil, fmt.Errorf("entity not found: %s", entityID)
}

// UpdateEntity 更新实体（检查锁定字段）
func (em *EntityManager) UpdateEntity(playerID, modID, entityID string, updates map[string]interface{}) error {
	registry := em.GetOrCreateRegistry(playerID, modID)

	entity, exists := registry.Entities[entityID]
	if !exists {
		return fmt.Errorf("entity not found: %s", entityID)
	}

	// 检查是否尝试修改锁定的实体
	if entity.Locked {
		return fmt.Errorf("entity %s is locked and cannot be modified", entityID)
	}

	// 检查是否尝试修改锁定的字段
	for field := range updates {
		fullField := fmt.Sprintf("%s.%s", entity.Type, field)
		for _, lockedField := range registry.LockedFields {
			if fullField == lockedField {
				// 如果字段已经存在且值不同，则拒绝修改
				if oldValue, exists := entity.Attributes[field]; exists {
					if fmt.Sprintf("%v", oldValue) != fmt.Sprintf("%v", updates[field]) {
						return fmt.Errorf("cannot modify locked field: %s", field)
					}
				}
			}
		}
	}

	// 应用更新
	if entity.Attributes == nil {
		entity.Attributes = make(map[string]interface{})
	}
	for k, v := range updates {
		entity.Attributes[k] = v
	}
	entity.UpdatedAt = time.Now()

	log.Printf("Updated entity %s: %v", entityID, updates)
	return nil
}

// ExtractEntitiesFromText AI文本中提取实体信息
func (em *EntityManager) ExtractEntitiesFromText(text string, existingRegistry *EntityRegistry) map[string]*Entity {
	entities := make(map[string]*Entity)

	// 简单的实体提取规则（可以用更复杂的NLP）
	// patterns := map[string][]string{
	// 	"npc": {
	// 		"一位名叫(.+?)的",
	// 		"(.+?)说道",
	// 		"遇到了(.+?)，",
	// 		"见到(.+?)正在",
	// 	},
	// 	"place": {
	// 		"来到(.+?)，",
	// 		"在(.+?)中",
	// 		"前往(.+?)",
	// 		"位于(.+?)的",
	// 	},
	// }

	// 这里简化处理，实际应该用更复杂的提取逻辑
	// 可以集成NLP库或使用正则表达式

	return entities
}

// ValidateConsistency 验证实体一致性
func (em *EntityManager) ValidateConsistency(playerID, modID string, aiResponse string) error {
	registry := em.GetOrCreateRegistry(playerID, modID)

	if registry.PlayerEntity == nil {
		return nil // 还没有玩家实体，跳过验证
	}

	// 检查关键属性是否被违反
	violations := []string{}

	// 检查性别一致性
	if gender, exists := registry.PlayerEntity.Attributes["gender"]; exists {
		genderStr := fmt.Sprintf("%v", gender)
		// 检查是否有相反性别的描述
		if (genderStr == "男" || genderStr == "male") &&
		   (strings.Contains(aiResponse, "她") || strings.Contains(aiResponse, "女侠") ||
		    strings.Contains(aiResponse, "仙子") || strings.Contains(aiResponse, "姑娘")) {
			violations = append(violations, fmt.Sprintf("性别不一致：玩家是%s性，但描述中出现女性称谓", genderStr))
		}
		if (genderStr == "女" || genderStr == "female") &&
		   (strings.Contains(aiResponse, "他") && !strings.Contains(aiResponse, "她") ||
		    strings.Contains(aiResponse, "公子") || strings.Contains(aiResponse, "少侠")) {
			violations = append(violations, fmt.Sprintf("性别不一致：玩家是%s性，但描述中出现男性称谓", genderStr))
		}
	}

	// 检查名字一致性
	if name, exists := registry.PlayerEntity.Attributes["name"]; exists {
		nameStr := fmt.Sprintf("%v", name)
		// 这里可以添加更复杂的名字验证逻辑
		_ = nameStr
	}

	if len(violations) > 0 {
		return fmt.Errorf("实体一致性验证失败：%s", strings.Join(violations, "; "))
	}

	return nil
}

// SerializeRegistry 序列化注册表
func (em *EntityManager) SerializeRegistry(playerID, modID string) (string, error) {
	registry := em.GetOrCreateRegistry(playerID, modID)

	data, err := json.Marshal(registry)
	if err != nil {
		return "", fmt.Errorf("failed to serialize registry: %v", err)
	}

	return string(data), nil
}

// DeserializeRegistry 反序列化注册表
func (em *EntityManager) DeserializeRegistry(playerID, modID string, data string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	var registry EntityRegistry
	if err := json.Unmarshal([]byte(data), &registry); err != nil {
		return fmt.Errorf("failed to deserialize registry: %v", err)
	}

	key := fmt.Sprintf("%s_%s", playerID, modID)
	em.registries[key] = &registry
	return nil
}

// ValidateEntity 验证实体
func (ev *EntityValidator) ValidateEntity(entity *Entity) error {
	if entity.ID == "" {
		return fmt.Errorf("entity ID is required")
	}
	if entity.Name == "" {
		return fmt.Errorf("entity name is required")
	}
	if entity.Type == "" {
		return fmt.Errorf("entity type is required")
	}

	// 根据类型验证特定规则
	for field, value := range entity.Attributes {
		ruleKey := fmt.Sprintf("%s.%s", entity.Type, field)
		if rule, exists := ev.rules[ruleKey]; exists {
			// 检查必填
			if rule.Required && value == nil {
				return fmt.Errorf("field %s is required for %s", field, entity.Type)
			}
			// 检查有效值
			if len(rule.ValidValues) > 0 {
				valid := false
				for _, validValue := range rule.ValidValues {
					if fmt.Sprintf("%v", value) == fmt.Sprintf("%v", validValue) {
						valid = true
						break
					}
				}
				if !valid {
					return fmt.Errorf("invalid value for %s: %v", field, value)
				}
			}
		}
	}

	return nil
}

// BuildEntityContext 构建实体上下文提示
func (em *EntityManager) BuildEntityContext(playerID, modID string) string {
	registry := em.GetOrCreateRegistry(playerID, modID)

	if registry.PlayerEntity == nil && len(registry.Entities) == 0 {
		return ""
	}

	var contextBuilder strings.Builder
	contextBuilder.WriteString("\n【核心实体信息】\n")
	contextBuilder.WriteString("以下信息是锁定的，必须严格保持一致，不得更改：\n\n")

	// 玩家信息
	if registry.PlayerEntity != nil {
		contextBuilder.WriteString("【玩家角色】\n")
		contextBuilder.WriteString(fmt.Sprintf("- ID: %s\n", registry.PlayerEntity.ID))
		contextBuilder.WriteString(fmt.Sprintf("- 姓名: %s\n", registry.PlayerEntity.Name))

		for key, value := range registry.PlayerEntity.Attributes {
			// 标记锁定字段
			isLocked := false
			fullField := fmt.Sprintf("player.%s", key)
			for _, lockedField := range registry.LockedFields {
				if fullField == lockedField {
					isLocked = true
					break
				}
			}
			if isLocked {
				contextBuilder.WriteString(fmt.Sprintf("- %s: %v [锁定]\n", key, value))
			} else {
				contextBuilder.WriteString(fmt.Sprintf("- %s: %v\n", key, value))
			}
		}
		contextBuilder.WriteString("\n")
	}

	// 重要NPC信息
	npcCount := 0
	for _, entity := range registry.Entities {
		if entity.Type == EntityNPC && npcCount < 10 {
			contextBuilder.WriteString(fmt.Sprintf("【NPC: %s】\n", entity.Name))
			for key, value := range entity.Attributes {
				contextBuilder.WriteString(fmt.Sprintf("- %s: %v\n", key, value))
			}
			contextBuilder.WriteString("\n")
			npcCount++
		}
	}

	contextBuilder.WriteString("请严格遵守以上标记为[锁定]的信息，确保描述的一致性。\n")

	return contextBuilder.String()
}

// CleanupOldEntities 清理过期的实体（可选）
func (em *EntityManager) CleanupOldEntities(playerID, modID string, keepDays int) int {
	registry := em.GetOrCreateRegistry(playerID, modID)

	cutoffTime := time.Now().AddDate(0, 0, -keepDays)
	deletedCount := 0

	for id, entity := range registry.Entities {
		// 不删除玩家实体和锁定的实体
		if entity.Type == EntityPlayer || entity.Locked {
			continue
		}

		if entity.UpdatedAt.Before(cutoffTime) {
			delete(registry.Entities, id)
			deletedCount++
		}
	}

	return deletedCount
}