package game_engine

import (
	"fmt"
	"strings"
	"time"
)

// EntityIntegration 实体集成功能
type EntityIntegration struct {
	controller *GameController
}

// NewEntityIntegration 创建实体集成
func NewEntityIntegration(gc *GameController) *EntityIntegration {
	return &EntityIntegration{
		controller: gc,
	}
}

// ProcessGameStartWithEntities 处理游戏开始时的实体注册
func (ei *EntityIntegration) ProcessGameStartWithEntities(
	playerID string,
	modID string,
	customAttributes map[string]interface{},
) error {
	if ei.controller.stateManager.GetEntityManager() == nil {
		return nil // 没有实体管理器，跳过
	}

	entityManager := ei.controller.stateManager.GetEntityManager()

	// 创建玩家实体
	playerEntity := &Entity{
		ID:         fmt.Sprintf("player_%s", playerID),
		Type:       EntityPlayer,
		Name:       "",
		Attributes: make(map[string]interface{}),
		Locked:     false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 处理自定义属性
	if name, ok := customAttributes["姓名"].(string); ok && name != "" {
		playerEntity.Name = name
		playerEntity.Attributes["name"] = name
	}
	if gender, ok := customAttributes["性别"].(string); ok && gender != "" {
		playerEntity.Attributes["gender"] = gender
	}
	if qualification, ok := customAttributes["资质"].(string); ok && qualification != "" {
		playerEntity.Attributes["qualification"] = qualification
	}
	if cultivation, ok := customAttributes["修为"].(string); ok && cultivation != "" {
		playerEntity.Attributes["cultivation"] = cultivation
	}
	if spiritStones, ok := customAttributes["元石"].(float64); ok && spiritStones > 0 {
		playerEntity.Attributes["spirit_stones"] = int(spiritStones)
	}
	if background, ok := customAttributes["出身"].(string); ok && background != "" {
		playerEntity.Attributes["background"] = background
	}

	// 注册实体
	if err := entityManager.RegisterEntity(playerID, modID, playerEntity); err != nil {
		return fmt.Errorf("failed to register player entity: %v", err)
	}

	fmt.Printf("[实体集成] 成功注册玩家实体: %s (性别: %v)\n", playerEntity.Name, playerEntity.Attributes["gender"])
	return nil
}

// ExtractAndRegisterEntitiesFromResponse 从AI响应中提取并注册实体
func (ei *EntityIntegration) ExtractAndRegisterEntitiesFromResponse(
	playerID string,
	modID string,
	aiResponse string,
	stateUpdate map[string]interface{},
) error {
	if ei.controller.stateManager.GetEntityManager() == nil {
		return nil
	}

	entityManager := ei.controller.stateManager.GetEntityManager()

	// 从state_update中提取current_life信息
	if currentLife, ok := stateUpdate["current_life"].(map[string]interface{}); ok {
		// 更新玩家实体
		playerEntity, err := entityManager.GetEntity(playerID, modID, fmt.Sprintf("player_%s", playerID))
		if err == nil {
			updates := make(map[string]interface{})

			// 更新基本属性
			if name, ok := currentLife["name"].(string); ok && name != "" {
				updates["name"] = name
				playerEntity.Name = name
			}
			if gender, ok := currentLife["gender"].(string); ok && gender != "" {
				updates["gender"] = gender
			}
			if cultivation, ok := currentLife["cultivation_level"].(string); ok {
				updates["cultivation"] = cultivation
			}

			// 应用更新
			if len(updates) > 0 {
				if err := entityManager.UpdateEntity(playerID, modID, playerEntity.ID, updates); err != nil {
					fmt.Printf("[实体集成] 警告：更新玩家实体失败: %v\n", err)
				}
			}
		} else {
			// 如果实体不存在，创建新的
			playerEntity = &Entity{
				ID:         fmt.Sprintf("player_%s", playerID),
				Type:       EntityPlayer,
				Name:       "",
				Attributes: make(map[string]interface{}),
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}

			if name, ok := currentLife["name"].(string); ok && name != "" {
				playerEntity.Name = name
				playerEntity.Attributes["name"] = name
			}
			if gender, ok := currentLife["gender"].(string); ok && gender != "" {
				playerEntity.Attributes["gender"] = gender
			}
			if cultivation, ok := currentLife["cultivation_level"].(string); ok {
				playerEntity.Attributes["cultivation"] = cultivation
			}

			entityManager.RegisterEntity(playerID, modID, playerEntity)
		}
	}

	// 提取NPC实体（简单的模式匹配）
	ei.extractNPCsFromText(playerID, modID, aiResponse)

	return nil
}

// extractNPCsFromText 从文本中提取NPC
func (ei *EntityIntegration) extractNPCsFromText(playerID, modID, text string) {
	entityManager := ei.controller.stateManager.GetEntityManager()

	// 简单的NPC名字提取规则
	patterns := []string{
		"一位名叫", "名为", "自称", "道号", "法号",
	}

	for _, pattern := range patterns {
		if idx := strings.Index(text, pattern); idx != -1 {
			// 提取名字（简化版本，实际应该更复杂）
			nameStart := idx + len(pattern)
			nameEnd := nameStart

			// 找到名字的结束位置（通常是"的"、"，"、"。"等）
			textRunes := []rune(text[nameStart:])
			for i := 0; i < len(textRunes) && i < 20; i++ {
				runeValue := textRunes[i]
				if runeValue == '的' || runeValue == '，' || runeValue == '。' ||
					runeValue == '！' || runeValue == '？' || runeValue == ' ' ||
					runeValue == '"' || runeValue == '「' || runeValue == '」' {
					nameEnd = nameStart + len(string(textRunes[:i]))
					break
				}
			}

			if nameEnd > nameStart {
				npcName := text[nameStart:nameEnd]
				npcName = strings.TrimSpace(npcName)

				if npcName != "" && len(npcName) < 10 {
					// 创建NPC实体
					npcEntity := &Entity{
						ID:         fmt.Sprintf("npc_%s_%d", npcName, time.Now().Unix()),
						Type:       EntityNPC,
						Name:       npcName,
						Attributes: make(map[string]interface{}),
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
					}

					// 尝试注册（如果已存在会失败，但不影响）
					entityManager.RegisterEntity(playerID, modID, npcEntity)
				}
			}
		}
	}
}

// ValidateResponseConsistency 验证AI响应的一致性
func (ei *EntityIntegration) ValidateResponseConsistency(
	playerID string,
	modID string,
	aiResponse string,
) (bool, []string) {
	if ei.controller.stateManager.GetEntityManager() == nil {
		return true, nil
	}

	entityManager := ei.controller.stateManager.GetEntityManager()
	violations := []string{}

	// 获取玩家实体
	playerEntity, err := entityManager.GetEntity(playerID, modID, fmt.Sprintf("player_%s", playerID))
	if err != nil {
		return true, nil // 没有玩家实体，跳过验证
	}

	// 检查性别一致性
	if gender, ok := playerEntity.Attributes["gender"].(string); ok {
		if gender == "男" || gender == "male" {
			// 检查是否有女性称谓
			femaleTerms := []string{"她", "女侠", "仙子", "姑娘", "女子", "女修"}
			for _, term := range femaleTerms {
				if strings.Contains(aiResponse, term) {
					// 排除一些特殊情况（如描述其他角色）
					if !ei.isDescribingOthers(aiResponse, term) {
						violations = append(violations,
							fmt.Sprintf("性别不一致：玩家是男性，但出现了'%s'", term))
					}
				}
			}
		} else if gender == "女" || gender == "female" {
			// 检查是否有男性称谓
			maleTerms := []string{"他", "公子", "少侠", "道友", "男子", "男修"}
			for _, term := range maleTerms {
				if strings.Contains(aiResponse, term) {
					// 特殊处理"他"，因为可能是"其他"等词
					if term == "他" && strings.Contains(aiResponse, "其他") {
						continue
					}
					if !ei.isDescribingOthers(aiResponse, term) {
						violations = append(violations,
							fmt.Sprintf("性别不一致：玩家是女性，但出现了'%s'", term))
					}
				}
			}
		}
	}

	// 检查名字一致性
	if name, ok := playerEntity.Attributes["name"].(string); ok && name != "" {
		// 检查是否AI生成了不同的名字
		if strings.Contains(aiResponse, "你叫") || strings.Contains(aiResponse, "你是") {
			// 这里可以添加更复杂的名字检测逻辑
		}
	}

	return len(violations) == 0, violations
}

// isDescribingOthers 判断是否在描述其他角色
func (ei *EntityIntegration) isDescribingOthers(text, term string) bool {
	// 简单的判断逻辑，检查是否在描述其他NPC
	// 可以根据上下文改进
	idx := strings.Index(text, term)
	if idx > 10 {
		// 检查前面是否有"看到"、"遇到"、"一位"等词
		prefix := text[idx-10 : idx]
		if strings.Contains(prefix, "看到") ||
			strings.Contains(prefix, "遇到") ||
			strings.Contains(prefix, "一位") ||
			strings.Contains(prefix, "那位") {
			return true
		}
	}
	return false
}

// InjectEntityContextIntoPrompt 向提示词注入实体上下文
func (ei *EntityIntegration) InjectEntityContextIntoPrompt(
	playerID string,
	modID string,
	basePrompt string,
) string {
	if ei.controller.stateManager.GetEntityManager() == nil {
		return basePrompt
	}

	entityContext := ei.controller.stateManager.GetEntityManager().BuildEntityContext(playerID, modID)
	if entityContext == "" {
		return basePrompt
	}

	// 将实体上下文插入到提示词中
	return basePrompt + "\n\n" + entityContext
}

// HandleEntityUpdate 处理实体更新事件
func (ei *EntityIntegration) HandleEntityUpdate(
	playerID string,
	modID string,
	entityType EntityType,
	entityName string,
	updates map[string]interface{},
) error {
	if ei.controller.stateManager.GetEntityManager() == nil {
		return nil
	}

	entityManager := ei.controller.stateManager.GetEntityManager()

	// 构建实体ID
	entityID := ""
	switch entityType {
	case EntityPlayer:
		entityID = fmt.Sprintf("player_%s", playerID)
	case EntityNPC:
		entityID = fmt.Sprintf("npc_%s_%s", entityName, playerID)
	default:
		entityID = fmt.Sprintf("%s_%s_%s", entityType, entityName, playerID)
	}

	// 尝试获取实体
	entity, err := entityManager.GetEntity(playerID, modID, entityID)
	if err != nil {
		// 实体不存在，创建新的
		entity = &Entity{
			ID:         entityID,
			Type:       entityType,
			Name:       entityName,
			Attributes: updates,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		return entityManager.RegisterEntity(playerID, modID, entity)
	}

	// 更新现有实体
	return entityManager.UpdateEntity(playerID, modID, entityID, updates)
}