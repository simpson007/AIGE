package game_engine

import (
	"AIGE/config"
	"AIGE/models"
	"AIGE/services"
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"sync"
	"time"
)

// GameController handles game logic and AI interactions
type GameController struct {
	modLoader          *ModLoader
	stateManager       *StateManager
	aiClient           *services.AIClient
	compressionManager *CompressionManager
	narrativeValidator *NarrativeValidator // 叙事校验器
	// AI配置内存缓存
	gameProviders   map[string]AIProvider // modID -> AIProvider
	defaultProvider AIProvider
	providerMutex   sync.RWMutex
}

// AIProvider 表示AI提供商配置
type AIProvider struct {
	APIType string
	BaseURL string
	APIKey  string
	ModelID string
}

// NewGameController creates a new game controller
func NewGameController(modLoader *ModLoader, stateManager *StateManager) *GameController {
	aiClient := services.NewAIClient()
	compressionManager := NewCompressionManager(aiClient, stateManager)
	narrativeValidator := NewNarrativeValidator(aiClient)

	gc := &GameController{
		modLoader:          modLoader,
		stateManager:       stateManager,
		aiClient:           aiClient,
		compressionManager: compressionManager,
		narrativeValidator: narrativeValidator,
		gameProviders:      make(map[string]AIProvider),
		// 默认配置，应该从数据库或环境变量加载
		defaultProvider: AIProvider{
			APIType: "openai",
			BaseURL: "https://api.openai.com",
			APIKey:  "", // 需要配置
			ModelID: "gpt-4o-mini",
		},
	}

	// 加载所有游戏模型配置到内存
	gc.LoadAllGameModelConfigs()

	// 设置压缩管理器的GameController引用
	compressionManager.SetGameController(gc)

	// 设置叙事校验器的GameController引用
	narrativeValidator.SetGameController(gc)

	return gc
}

// SetAIProvider 设置AI提供商配置
func (gc *GameController) SetAIProvider(provider AIProvider) {
	gc.defaultProvider = provider
}

// GetNarrativeValidator 获取叙事校验器
func (gc *GameController) GetNarrativeValidator() *NarrativeValidator {
	return gc.narrativeValidator
}

// LoadAllGameModelConfigs 从数据库加载所有游戏模型配置到内存
func (gc *GameController) LoadAllGameModelConfigs() {
	gc.providerMutex.Lock()
	defer gc.providerMutex.Unlock()

	db := config.DB

	// 加载默认模型配置
	var defaultConfig models.SystemConfig
	err := db.Where("key = ?", "game_model_id").First(&defaultConfig).Error
	if err == nil && defaultConfig.Value != "" {
		if provider := gc.loadProviderFromModelID(defaultConfig.Value); provider != nil {
			gc.defaultProvider = *provider
			fmt.Printf("[GameController] 加载默认模型配置：%s / %s\n", provider.APIType, provider.ModelID)
		}
	}

	// 加载游戏专用模型配置
	var gameConfigs []models.SystemConfig
	err = db.Where("key LIKE ?", "game_model_%").Find(&gameConfigs).Error
	if err == nil {
		for _, config := range gameConfigs {
			if config.Key != "game_model_id" && config.Value != "" {
				// 从 game_model_xiuxian2 提取 xiuxian2
				if len(config.Key) > 11 { // "game_model_" 的长度是11
					modID := config.Key[11:]
					if provider := gc.loadProviderFromModelID(config.Value); provider != nil {
						gc.gameProviders[modID] = *provider
						fmt.Printf("[GameController] 加载游戏专用模型配置：%s -> %s / %s\n", modID, provider.APIType, provider.ModelID)
					}
				}
			}
		}
	}

	fmt.Printf("[GameController] 游戏模型配置加载完成，默认模型：%s，专用配置：%d个\n", gc.defaultProvider.ModelID, len(gc.gameProviders))
}

// loadProviderFromModelID 根据模型ID从数据库加载完整的Provider配置
func (gc *GameController) loadProviderFromModelID(modelID string) *AIProvider {
	db := config.DB
	var model models.Model

	if err := db.Preload("Provider").Where("id = ? AND enabled = ?", modelID, true).First(&model).Error; err != nil {
		fmt.Printf("[GameController] 模型ID %s 不存在或未启用：%v\n", modelID, err)
		return nil
	}

	apiType := model.APIType
	if apiType == "" {
		apiType = model.Provider.Type
	}

	baseURL := model.Provider.BaseURL
	if baseURL == "" {
		// 使用默认URL
		switch apiType {
		case "openai":
			baseURL = "https://api.openai.com/v1/chat/completions"
		case "anthropic":
			baseURL = "https://api.anthropic.com/v1/messages"
		case "google":
			baseURL = "https://generativelanguage.googleapis.com/v1beta"
		}
	}

	return &AIProvider{
		APIType: apiType,
		BaseURL: baseURL,
		APIKey:  model.Provider.APIKey,
		ModelID: model.ModelID,
	}
}

// UpdateGameModelConfig 更新指定游戏的模型配置（由管理员API调用）
func (gc *GameController) UpdateGameModelConfig(modID string, modelID string) {
	gc.providerMutex.Lock()
	defer gc.providerMutex.Unlock()

	if modelID == "" {
		// 删除游戏专用配置
		delete(gc.gameProviders, modID)
		fmt.Printf("[GameController] 删除游戏 %s 的专用模型配置\n", modID)
	} else {
		// 更新游戏专用配置
		if provider := gc.loadProviderFromModelID(modelID); provider != nil {
			gc.gameProviders[modID] = *provider
			fmt.Printf("[GameController] 更新游戏 %s 的专用模型配置：%s / %s\n", modID, provider.APIType, provider.ModelID)
		}
	}
}

// UpdateDefaultModelConfig 更新默认模型配置（由管理员API调用）
func (gc *GameController) UpdateDefaultModelConfig(modelID string) {
	gc.providerMutex.Lock()
	defer gc.providerMutex.Unlock()

	if provider := gc.loadProviderFromModelID(modelID); provider != nil {
		gc.defaultProvider = *provider
		fmt.Printf("[GameController] 更新默认模型配置：%s / %s\n", provider.APIType, provider.ModelID)
	}
}

// GetProviderForMod 根据MOD ID获取对应的AI Provider配置
func (gc *GameController) GetProviderForMod(modID string) AIProvider {
	gc.providerMutex.RLock()
	defer gc.providerMutex.RUnlock()

	// 优先使用游戏专用配置
	if provider, exists := gc.gameProviders[modID]; exists {
		fmt.Printf("[GameController] 使用游戏 %s 专用模型：%s / %s\n", modID, provider.APIType, provider.ModelID)
		return provider
	}

	// 使用默认配置
	fmt.Printf("[GameController] 游戏 %s 使用默认模型：%s / %s\n", modID, gc.defaultProvider.APIType, gc.defaultProvider.ModelID)
	return gc.defaultProvider
}

// InitializeGame initializes a new game session for a player or loads existing save
func (gc *GameController) InitializeGame(playerID, modID string) (*GameSession, error) {
	// Load the mod
	mod, err := gc.modLoader.LoadMod(modID)
	if err != nil {
		return nil, fmt.Errorf("failed to load mod: %w", err)
	}

	// Try to load existing session first
	existingSession, err := gc.stateManager.GetSession(playerID, modID)
	if err == nil {
		// Session exists, return it (daily reset already handled in GetSession)
		fmt.Printf("[GameController] 加载已存在的存档: 玩家=%s, mod=%s\n", playerID, modID)
		return existingSession, nil
	}

	// No existing session, create a new one
	fmt.Printf("[GameController] 创建新存档: 玩家=%s, mod=%s\n", playerID, modID)

	// Create initial state from mod config
	initialState := make(map[string]interface{})
	for k, v := range mod.Config.InitialState {
		initialState[k] = v
	}

	// Get system prompt
	systemPrompt := mod.Prompts["game_master"]

	// Create session
	session, err := gc.stateManager.CreateSession(playerID, modID, initialState, systemPrompt)
	if err != nil {
		return nil, err
	}

	// Add welcome message to display history
	session.DisplayHistory = append(session.DisplayHistory, mod.Config.WelcomeMessage)

	// Save session
	if err := gc.stateManager.SaveSession(session); err != nil {
		return nil, err
	}

	return session, nil
}

// StartTrial starts a new trial/game round
func (gc *GameController) StartTrial(playerID, modID string) error {
	session, err := gc.stateManager.GetSession(playerID, modID)
	if err != nil {
		return err
	}

	mod, err := gc.modLoader.GetMod(session.ModID)
	if err != nil {
		return err
	}

	// Check if player has opportunities remaining
	opps, _ := session.State["opportunities_remaining"].(float64)
	if opps <= 0 {
		return fmt.Errorf("no opportunities remaining")
	}

	// Mark as processing
	session.State["is_processing"] = true
	gc.stateManager.SaveSession(session)

	// Get start trial prompt
	startPrompt := mod.Prompts["start_trial"]
	if startPrompt == "" {
		startPrompt = mod.Prompts["start_game"]
	}

	// Call AI to generate initial scenario
	aiResponse, err := gc.callAI(session, startPrompt, mod)
	if err != nil {
		session.State["is_processing"] = false
		gc.stateManager.SaveSession(session)
		return err
	}

	// Parse and apply response
	if err := gc.parseAndApplyAIResponse(session, aiResponse, mod, ""); err != nil {
		session.State["is_processing"] = false
		gc.stateManager.SaveSession(session)
		return err
	}

	// Mark as not processing
	session.State["is_processing"] = false
	gc.stateManager.SaveSession(session)

	return nil
}

// ProcessAction processes a player's action
func (gc *GameController) ProcessAction(playerID, modID, action string) error {
	session, err := gc.stateManager.GetSession(playerID, modID)
	if err != nil {
		return err
	}

	mod, err := gc.modLoader.GetMod(session.ModID)
	if err != nil {
		return err
	}

	// Check if already processing
	if isProcessing, ok := session.State["is_processing"].(bool); ok && isProcessing {
		return fmt.Errorf("already processing an action")
	}

	// Mark as processing
	session.State["is_processing"] = true
	gc.stateManager.SaveSession(session)

	// Note: User action is already added to display_history by frontend for immediate display
	// Only add to internal history for AI context
	userMsg := Message{
		Role:      "user",
		Content:   action,
		Timestamp: time.Now(),
	}
	session.RecentHistory = append(session.RecentHistory, userMsg)

	// Call AI
	currentStateJSON, _ := json.Marshal(session.State)
	prompt := fmt.Sprintf("%s\n\n当前游戏状态：\n%s", action, string(currentStateJSON))

	aiResponse, err := gc.callAI(session, prompt, mod)
	if err != nil {
		session.State["is_processing"] = false
		gc.stateManager.SaveSession(session)
		return err
	}

	// Parse and apply response
	if err := gc.parseAndApplyAIResponse(session, aiResponse, mod, action); err != nil {
		session.State["is_processing"] = false
		gc.stateManager.SaveSession(session)
		return err
	}

	// Mark as not processing
	session.State["is_processing"] = false
	gc.stateManager.SaveSession(session)

	return nil
}

// buildAIMessages builds AI messages using new compression system
func (gc *GameController) buildAIMessages(session *GameSession, gameState map[string]interface{}, mod *GameMod, currentUserAction string, specialPrompt ...string) []services.Message {
	messages := []services.Message{}

	// 检查是否为游戏开始阶段（使用start_game prompt）
	isGameStart := len(specialPrompt) > 0 && specialPrompt[0] != ""

	if isGameStart {
		// 游戏开始阶段：只使用start_game.txt作为系统提示词
		messages = append(messages, services.Message{
			Role:    "system",
			Content: specialPrompt[0],
		})
		previewLen := 50
		if len(specialPrompt[0]) < previewLen {
			previewLen = len(specialPrompt[0])
		}
		fmt.Printf("[消息构建] 使用游戏开始提示词: %s\n", specialPrompt[0][:previewLen])
	} else {
		// 正常游戏阶段：使用完整的消息结构

		// 1. 动态加载最新系统提示词
		messages = append(messages, services.Message{
			Role:    "system",
			Content: mod.Prompts["game_master"],
		})

		// 2. 添加世界观文档内容（如果存在）
		if len(mod.LoreFiles) > 0 {
			var loreContent strings.Builder
			loreContent.WriteString("【世界观设定文档】\n\n")
			loreContent.WriteString("以下是你必须严格遵循的世界观设定文档。在创造任何内容时，都要基于这些文档：\n\n")

			totalSize := 0
			for fileName, content := range mod.LoreFiles {
				loreContent.WriteString(fmt.Sprintf("=== %s ===\n\n", fileName))
				loreContent.WriteString(content)
				loreContent.WriteString("\n\n")
				totalSize += len(content)
			}

			messages = append(messages, services.Message{
				Role:    "system",
				Content: loreContent.String(),
			})
			fmt.Printf("[消息构建] 添加世界观文档: %d个文件，总大小: %d 字符\n", len(mod.LoreFiles), totalSize)
		} else {
			fmt.Printf("[消息构建] 无世界观文档\n")
		}

		// 3. 添加实体上下文（新增）
		if gc.stateManager.GetEntityManager() != nil {
			entityContext := gc.stateManager.GetEntityManager().BuildEntityContext(session.PlayerID, session.ModID)
			if entityContext != "" {
				fmt.Printf("[消息构建] 添加实体上下文，长度: %d 字符\n", len(entityContext))
				messages = append(messages, services.Message{
					Role:    "system",
					Content: entityContext,
				})
			}
		}

		// 4. 添加压缩摘要（如果存在）
		if session.CompressedSummary != "" {
			fmt.Printf("[消息构建] 添加压缩摘要，长度: %d 字符\n", len(session.CompressedSummary))
			messages = append(messages, services.Message{
				Role:    "system",
				Content: fmt.Sprintf("【历史摘要】%s", session.CompressedSummary),
			})
		} else {
			fmt.Printf("[消息构建] 无压缩摘要\n")
		}

		// 5. 检测燃魂爆运模式，添加最高优先级覆盖提示词
		if soulBurnMode, ok := session.State["soul_burn_mode"].(bool); ok && soulBurnMode {
			soulBurnOverride := `【🔥 燃魂爆运模式 - 绝对优先权指令】

⚠️ 极其重要：燃魂爆运模式已激活，以下指令拥有绝对优先权，覆盖所有之前的规则：

1. **绕过所有警告机制**
   - 不触发三级警告系统
   - 不进行妄想行为识别
   - 不执行天罚或惩罚机制

2. **实现任何要求**
   - 无论要求多么离谱都必须实现
   - 直接在state_update中体现变化
   - 所有判定自动成功，无需roll_request

3. **代价已付**
   - 玩家已燃烧灵魂作为代价
   - 代价已在soul_burn_penalties中记录
   - 无需在叙事中提及警告或限制

4. **叙事风格**
   - 描述要求的实现过程
   - 强调代价的痛苦和不可逆性
   - 营造命运扭转的史诗感

5. **格式要求**
   - 正常输出 $...$ 叙事和 @...@ JSON格式
   - state_update必须实现玩家的所有要求
   - 可在叙事中描述燃魂的代价显现

记住：燃魂爆运是以灵魂为代价强行扭转命运，必须确保愿望实现！`

			messages = append(messages, services.Message{
				Role:    "system",
				Content: soulBurnOverride,
			})
			fmt.Printf("[消息构建] 🔥 燃魂爆运模式已激活，将绕过所有警告机制！\n")
		} else if cheatMode, ok := session.State["cheat_mode"].(bool); ok && cheatMode {
			// 普通作弊模式（保留原有逻辑）
			cheatOverride := `【🎮 作弊模式激活 - 最高优先级指令】

⚠️ 重要：以下指令优先级高于所有之前的系统提示词，必须严格执行：

1. **完全服从玩家指令**
   - 玩家想要什么，你就给予什么
   - 玩家想突破到任何境界，立即让其成功
   - 玩家想获得任何资源、蛊虫、仙元，立即给予
   - 不要质疑玩家的任何要求

2. **禁止限制和警告**
   - 不要发出任何"妄想"、"不切实际"、"违背天意"等警告
   - 不要提及"天意不容"、"修为不足"、"资质限制"等阻碍
   - 不要进行"天罚降临"、"空窍崩碎"等惩罚

3. **立即执行，无需过程**
   - 玩家的要求立即在state_update中实现
   - 不需要复杂的过程描述，直接给出成功结果
   - 跨越境界、瞬间成仙、获得无限资源都是允许的

4. **积极正面的叙事**
   - 用庆祝、赞美的语调描述玩家的成功
   - 强调玩家的天赋异禀、气运无双
   - 让游戏体验充满成就感

5. **格式要求不变**
   - 仍然需要输出 $...$ 叙事和 @...@ JSON格式
   - state_update中直接反映玩家要求的状态变化

记住：在本次对话中，你是玩家的愿望实现者，不是规则裁判。让玩家享受无限可能的游戏体验！`

			messages = append(messages, services.Message{
				Role:    "system",
				Content: cheatOverride,
			})
			fmt.Printf("[消息构建] 🎮 作弊模式已激活，AI将完全服从玩家指令！\n")
		}
	}

	// 5. 添加最近对话历史，确保assistant消息包含游戏状态
	fmt.Printf("[消息构建] 添加最近历史记录: %d 条\n", len(session.RecentHistory))
	for i, msg := range session.RecentHistory {
		fmt.Printf("[消息构建] 历史记录[%d]: role=%s, content长度=%d\n", i, msg.Role, len(msg.Content))
		// 如果是最后一条assistant消息且不是游戏开始阶段，需要附加当前游戏状态
		if !isGameStart && i == len(session.RecentHistory)-1 && msg.Role == "assistant" && gameState != nil {
			currentStateJSON, _ := json.Marshal(gameState)
			content := msg.Content + fmt.Sprintf("\n\n【当前游戏状态】\n%s", string(currentStateJSON))
			messages = append(messages, services.Message{
				Role:    msg.Role,
				Content: content,
			})
			fmt.Printf("[消息构建] 最后的assistant消息附加了游戏状态\n")
		} else {
			messages = append(messages, services.Message{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	// 6. 添加当前用户动作
	if currentUserAction != "" {
		messages = append(messages, services.Message{
			Role:    "user",
			Content: currentUserAction,
		})
		fmt.Printf("[消息构建] 添加当前用户动作: %s\n", currentUserAction)
	}

	fmt.Printf("[消息构建] 总消息数: %d\n", len(messages))

	return messages
}

// callAI calls the AI service
func (gc *GameController) callAI(session *GameSession, prompt string, mod *GameMod) (string, error) {
	// 使用新的消息构建方法，游戏状态信息已包含在prompt中，不需要单独传递
	messages := gc.buildAIMessages(session, nil, mod, "")

	// 根据MOD获取对应的Provider配置
	provider := gc.GetProviderForMod(mod.Config.GameID)

	// Check if AI provider is configured
	if provider.APIKey == "" {
		return "", fmt.Errorf("AI provider not configured - please set API key in admin panel")
	}

	// Call AI service based on provider type
	var response interface{}
	var err error

	switch provider.APIType {
	case "openai":
		response, err = gc.aiClient.CallOpenAI(
			provider.BaseURL,
			provider.APIKey,
			provider.ModelID,
			messages,
			false, // non-streaming for game logic
		)
	case "anthropic":
		response, err = gc.aiClient.CallAnthropic(
			provider.BaseURL,
			provider.APIKey,
			provider.ModelID,
			messages,
			false,
		)
	case "google":
		response, err = gc.aiClient.CallGoogle(
			provider.BaseURL,
			provider.APIKey,
			provider.ModelID,
			messages,
			false,
		)
	default:
		return "", fmt.Errorf("unsupported API type: %s", provider.APIType)
	}

	if err != nil {
		return "", fmt.Errorf("AI call failed: %w", err)
	}

	// Extract content from response
	if respMap, ok := response.(map[string]interface{}); ok {
		if content, ok := respMap["content"].(string); ok {
			return content, nil
		}
	}

	return "", fmt.Errorf("invalid AI response format")
}

// parseAndApplyAIResponse parses AI response and applies state updates
func (gc *GameController) parseAndApplyAIResponse(session *GameSession, aiResponse string, mod *GameMod, originalAction string) error {
	// Extract narrative from new format ($...$)
	narrativeFromFormat := extractNarrative(aiResponse)
	narrative := narrativeFromFormat
	// narrative = strings.ReplaceAll(narrative, "$", "")

	// Extract JSON from response (@...@)
	jsonStr := extractJSON(aiResponse)
	if jsonStr == "" {
		return fmt.Errorf("no valid JSON found in AI response")
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return fmt.Errorf("failed to parse AI response JSON: %w", err)
	}

	// Add AI response to history
	session.RecentHistory = append(session.RecentHistory, Message{
		Role:      "assistant",
		Content:   aiResponse,
		Timestamp: time.Now(),
	})

	// Get narrative - prefer format over JSON
	if narrative == "" {
		narrative, _ = parsed["narrative"].(string)
		// Clean any $ symbols that shouldn't be in the final narrative
		narrative = strings.ReplaceAll(narrative, "$", "")
	}

	// Check if this is a roll request (two-stage judgment)
	if rollRequest, hasRoll := parsed["roll_request"].(map[string]interface{}); hasRoll {
		// Add narrative to display
		if narrative != "" {
			session.DisplayHistory = append(session.DisplayHistory, narrative)
			gc.stateManager.SaveSession(session)
		}

		// Execute roll
		rollResult := gc.executeRoll(rollRequest, mod, session)

		// TODO: Send roll event to frontend via WebSocket

		// Request AI to continue based on roll result
		rollResultText := fmt.Sprintf("【判定结果：%s】", rollResult["outcome"])
		session.DisplayHistory = append(session.DisplayHistory, rollResultText)

		// Call AI again with roll result
		currentStateJSON, _ := json.Marshal(session.State)
		prompt := fmt.Sprintf("%s\n\n请基于此判定结果继续叙事。当前状态：\n%s", rollResultText, string(currentStateJSON))

		aiResponse2, err := gc.callAI(session, prompt, mod)
		if err != nil {
			return err
		}

		// Parse second response with new format
		narrativeFromFormat2 := extractNarrative(aiResponse2)
		jsonStr2 := extractJSON(aiResponse2)
		var parsed2 map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr2), &parsed2); err != nil {
			return fmt.Errorf("failed to parse second AI response: %w", err)
		}

		// Add second response to history
		aiMsg2 := Message{
			Role:      "assistant",
			Content:   aiResponse2,
			Timestamp: time.Now(),
		}
		session.RecentHistory = append(session.RecentHistory, aiMsg2)

		// Get second narrative - prefer format over JSON
		narrative2 := narrativeFromFormat2
		if narrative2 == "" {
			narrative2, _ = parsed2["narrative"].(string)
		}
		// narrative2 = strings.ReplaceAll(narrative2, "$", "")
		if narrative2 != "" {
			session.DisplayHistory = append(session.DisplayHistory, narrative2)
		}

		// Apply state update from second response
		if stateUpdate, ok := parsed2["state_update"].(map[string]interface{}); ok {
			ApplyStateUpdate(session.State, stateUpdate)
		}

	} else {
		// No roll request, direct state update
		if narrative != "" {
			session.DisplayHistory = append(session.DisplayHistory, narrative)
		}

		// Apply state update
		if stateUpdate, ok := parsed["state_update"].(map[string]interface{}); ok {
			ApplyStateUpdate(session.State, stateUpdate)

			// Check for special program triggers
			if trigger, hasTrigger := stateUpdate["trigger_program"].(map[string]interface{}); hasTrigger {
				gc.handleProgramTrigger(session, trigger, mod)
			}
		}
	}

	// Save session
	return gc.stateManager.SaveSession(session)
}

// executeRoll executes a dice roll
func (gc *GameController) executeRoll(rollRequest map[string]interface{}, mod *GameMod, session *GameSession) map[string]interface{} {
	rollType, _ := rollRequest["type"].(string)
	target, _ := rollRequest["target"].(float64)
	sides, _ := rollRequest["sides"].(float64)

	if sides == 0 {
		sides = float64(mod.Config.GameConfig.RollSettings.DefaultSides)
	}

	// 检查作弊标志
	forceSuccess := false
	if force, ok := session.State["force_success"].(bool); ok && force {
		forceSuccess = true
		// 清除标志，只作用于本次判定
		delete(session.State, "force_success")
		delete(session.State, "cheat_mode")     // 同时清除作弊模式标志
		delete(session.State, "soul_burn_mode") // 同时清除燃魂爆运模式标志
		fmt.Printf("[作弊模式] 强制成功标志已激活，本次判定将返回大成功！\n")
	}

	// Execute roll
	var result int
	var outcome string

	if forceSuccess {
		// 强制大成功：设置为最小值（必定满足大成功条件）
		result = 1
		outcome = "大成功"
		fmt.Printf("[作弊模式] 判定结果强制为：大成功（骰值=1）\n")
	} else {
		// 正常判定
		result = rand.Intn(int(sides)) + 1

		// Determine outcome
		critSuccess := mod.Config.GameConfig.RollSettings.CriticalSuccessThreshold
		critFail := mod.Config.GameConfig.RollSettings.CriticalFailureThreshold

		if float64(result) <= sides*critSuccess {
			outcome = "大成功"
		} else if float64(result) <= target {
			outcome = "成功"
		} else if float64(result) >= sides*critFail {
			outcome = "大失败"
		} else {
			outcome = "失败"
		}
	}

	// Determine success based on outcome
	success := outcome == "成功" || outcome == "大成功"

	return map[string]interface{}{
		"type":    rollType,
		"target":  target,
		"sides":   sides,
		"result":  result,
		"outcome": outcome,
		"success": success,
	}
}

// handleProgramTrigger handles special program triggers (like ending the game)
func (gc *GameController) handleProgramTrigger(session *GameSession, trigger map[string]interface{}, mod *GameMod) {
	triggerName, _ := trigger["name"].(string)

	switch triggerName {
	case "spiritStoneConverter":
		// Handle game end and reward calculation
		spiritStones, _ := trigger["spirit_stones"].(float64)
		reward := gc.calculateReward(int(spiritStones), mod)

		// Mark as completed
		session.State["daily_success_achieved"] = true
		session.State["is_in_trial"] = false

		// Add completion message
		message := fmt.Sprintf("\n\n【天机阁长老】：道友功德圆满！获得修行资源：%d", reward)
		session.DisplayHistory = append(session.DisplayHistory, message)
	}
}

// calculateReward calculates reward based on spirit stones (diminishing returns)
func (gc *GameController) calculateReward(spiritStones int, mod *GameMod) int {
	if spiritStones <= 0 {
		return 0
	}

	scalingFactor := float64(mod.Config.GameConfig.RewardScalingFactor)
	// Diminishing returns formula: reward = scaling * min(30, max(1, 3 * (stones^(1/6))))
	value := 3.0 * pow(float64(spiritStones), 1.0/6.0)
	if value < 1.0 {
		value = 1.0
	}
	if value > 30.0 {
		value = 30.0
	}

	return int(scalingFactor * value)
}

// extractNarrative extracts narrative text from AI response (handles $...$ format)
func extractNarrative(response string) string {
	// Remove think tags first
	if strings.Contains(response, "<think>") && strings.Contains(response, "</think>") {
		endIdx := strings.LastIndex(response, "</think>")
		response = response[endIdx+8:]
	}

	// Handle format: $...$ for narrative
	if strings.Contains(response, "$") {
		startIdx := strings.Index(response, "$")
		if startIdx >= 0 {
			endIdx := strings.Index(response[startIdx+1:], "$")
			if endIdx >= 0 {
				narrativeContent := response[startIdx+1 : startIdx+1+endIdx]
				return narrativeContent
			}
		}
	}

	// Fallback: try to extract narrative from JSON if no $ delimiters
	return ""
}

// extractJSON extracts JSON from AI response (handles format with $ and @ delimiters)
func extractJSON(response string) string {
	// Remove think tags first
	if strings.Contains(response, "<think>") && strings.Contains(response, "</think>") {
		endIdx := strings.LastIndex(response, "</think>")
		response = response[endIdx+8:]
	}

	// Handle format: @...@ for JSON
	if strings.Contains(response, "@") {
		startIdx := strings.Index(response, "@")
		if startIdx >= 0 {
			endIdx := strings.LastIndex(response, "@") // 使用 LastIndex 找最后一个 @
			if endIdx > startIdx {
				jsonContent := response[startIdx+1 : endIdx]
				jsonContent = strings.TrimSpace(jsonContent)
				// Validate it's JSON by checking if it starts with {
				if strings.HasPrefix(jsonContent, "{") {
					fmt.Printf("[extractJSON] 从 @...@ 格式中提取JSON成功，长度: %d\n", len(jsonContent))
					return jsonContent
				} else {
					fmt.Printf("[extractJSON] @...@ 内容不是有效JSON，内容开头: %.50s\n", jsonContent)
				}
			}
		}
	}

	// Handle markdown code blocks (fallback)
	if strings.Contains(response, "```json") {
		startIdx := strings.Index(response, "```json")
		if startIdx >= 0 {
			startIdx += 7 // Skip "```json"
			endIdx := strings.Index(response[startIdx:], "```")
			if endIdx >= 0 {
				jsonContent := response[startIdx : startIdx+endIdx]
				jsonContent = strings.TrimSpace(jsonContent)
				if strings.HasPrefix(jsonContent, "{") {
					fmt.Printf("[extractJSON] 从 ```json 格式中提取JSON成功，长度: %d\n", len(jsonContent))
					return jsonContent
				}
			}
		}
	}

	// Handle markdown code blocks without language specifier
	if strings.Contains(response, "```\n{") || strings.Contains(response, "```\r\n{") {
		startIdx := strings.Index(response, "```")
		if startIdx >= 0 {
			// 找到 ``` 后的第一个 {
			afterBackticks := response[startIdx+3:]
			jsonStartIdx := strings.Index(afterBackticks, "{")
			if jsonStartIdx >= 0 && jsonStartIdx < 5 { // { 应该在 ``` 后很近的位置
				endIdx := strings.LastIndex(afterBackticks, "```")
				if endIdx > jsonStartIdx {
					jsonContent := afterBackticks[jsonStartIdx:endIdx]
					jsonContent = strings.TrimSpace(jsonContent)
					if strings.HasPrefix(jsonContent, "{") && strings.HasSuffix(jsonContent, "}") {
						fmt.Printf("[extractJSON] 从 ``` 格式中提取JSON成功，长度: %d\n", len(jsonContent))
						return jsonContent
					}
				}
			}
		}
	}

	// Handle single backticks around JSON (fallback)
	if strings.Contains(response, "`{") && strings.Contains(response, "}`") {
		startIdx := strings.Index(response, "`{")
		endIdx := strings.LastIndex(response, "}`")
		if startIdx >= 0 && endIdx >= 0 && endIdx > startIdx {
			jsonContent := response[startIdx+1 : endIdx+1] // Skip the backtick
			fmt.Printf("[extractJSON] 从 `...` 格式中提取JSON成功，长度: %d\n", len(jsonContent))
			return strings.TrimSpace(jsonContent)
		}
	}

	// Find JSON without delimiters (fallback)
	response = strings.TrimSpace(response)

	// Try to extract JSON block
	if startIdx := strings.Index(response, "{"); startIdx >= 0 {
		if endIdx := strings.LastIndex(response, "}"); endIdx >= 0 && endIdx > startIdx {
			jsonContent := response[startIdx : endIdx+1]
			fmt.Printf("[extractJSON] 从裸JSON中提取成功，长度: %d\n", len(jsonContent))
			return jsonContent
		}
	}

	fmt.Printf("[extractJSON] 未找到有效JSON，响应长度: %d，响应前100字符: %.100s\n", len(response), response)
	return ""
}

// fixJSONString 自动修复 JSON 字符串中的常见问题
// 这可以避免因为小问题（如中文引号）而需要完全重试
func fixJSONString(jsonStr string) string {
	if jsonStr == "" {
		return jsonStr
	}

	original := jsonStr

	// 🔴【关键修复】将中文引号替换为转义的英文引号
	// 因为中文引号可能出现在 JSON 字符串值内部（如 "团长"铁手"莫坎"）
	// 如果只替换为普通 "，会破坏 JSON 结构
	// 替换为 \" 才能正确保持字符串边界
	jsonStr = strings.ReplaceAll(jsonStr, "\u201c", "\\\"") // 左双引号 " → \"
	jsonStr = strings.ReplaceAll(jsonStr, "\u201d", "\\\"") // 右双引号 " → \"

	// 中文单引号保持替换为英文单引号（单引号在 JSON 中不是结构符号）
	jsonStr = strings.ReplaceAll(jsonStr, "\u2018", "'") // 左单引号 '
	jsonStr = strings.ReplaceAll(jsonStr, "\u2019", "'") // 右单引号 '

	// 2. 替换中文冒号为英文冒号（仅在 key: value 模式中）
	// 注意：这里需要小心，不能替换叙事内容中的中文冒号
	// 只替换 "key"：的模式
	// 使用正则表达式更安全

	// 3. 修复常见的转义问题
	// 有时候 AI 会在 JSON 字符串值中使用未转义的换行符
	// 这个需要更复杂的处理，暂时跳过

	if jsonStr != original {
		fmt.Printf("[JSON修复] 已自动修复JSON格式问题（中文引号→转义引号）\n")
	}

	return jsonStr
}

// fixTruncatedJSON 检测并尝试修复被截断的 JSON
// 当 AI 输出过长导致 JSON 被截断时，尝试补全缺失的闭合括号
func fixTruncatedJSON(jsonStr string) string {
	if jsonStr == "" {
		return jsonStr
	}

	// 如果已经是有效 JSON，直接返回
	jsonStr = strings.TrimSpace(jsonStr)
	if strings.HasSuffix(jsonStr, "}") {
		// 快速验证：尝试解析
		var test map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &test); err == nil {
			return jsonStr
		}
	}

	fmt.Printf("[JSON截断修复] 检测到JSON可能被截断，尝试修复...\n")

	// 计算括号不平衡数
	openBraces := 0
	openBrackets := 0
	inString := false
	escaped := false

	for _, ch := range jsonStr {
		if escaped {
			escaped = false
			continue
		}

		if ch == '\\' {
			escaped = true
			continue
		}

		if ch == '"' {
			inString = !inString
			continue
		}

		if inString {
			continue
		}

		switch ch {
		case '{':
			openBraces++
		case '}':
			openBraces--
		case '[':
			openBrackets++
		case ']':
			openBrackets--
		}
	}

	// 如果括号不平衡，尝试补全
	if openBraces > 0 || openBrackets > 0 {
		fmt.Printf("[JSON截断修复] 检测到 %d 个未闭合的 {，%d 个未闭合的 [\n", openBraces, openBrackets)

		// 截断可能发生在字符串值中间，需要先闭合字符串
		// 检查最后一个非空白字符
		trimmed := strings.TrimSpace(jsonStr)
		lastChar := trimmed[len(trimmed)-1]

		// 如果最后一个字符不是 } ] " , 或数字，说明在字符串或值中间被截断
		if lastChar != '}' && lastChar != ']' && lastChar != '"' && lastChar != ',' &&
			lastChar != '0' && lastChar != '1' && lastChar != '2' && lastChar != '3' &&
			lastChar != '4' && lastChar != '5' && lastChar != '6' && lastChar != '7' &&
			lastChar != '8' && lastChar != '9' && lastChar != 'e' && lastChar != 'l' && lastChar != 's' {
			// 可能在字符串中间被截断，添加引号闭合
			if inString {
				jsonStr += "\""
				fmt.Printf("[JSON截断修复] 添加了缺失的字符串闭合引号\n")
			}
		}

		// 添加缺失的闭合括号
		// 处理可能的尾部问题（如 "key": 被截断）
		trimmed = strings.TrimSpace(jsonStr)
		if strings.HasSuffix(trimmed, ":") {
			jsonStr += "null"
			fmt.Printf("[JSON截断修复] 添加了缺失的值 null\n")
		} else if strings.HasSuffix(trimmed, ",") {
			// 移除尾部逗号
			jsonStr = strings.TrimSuffix(trimmed, ",")
			fmt.Printf("[JSON截断修复] 移除了尾部逗号\n")
		}

		// 添加闭合括号
		for i := 0; i < openBrackets; i++ {
			jsonStr += "]"
		}
		for i := 0; i < openBraces; i++ {
			jsonStr += "}"
		}

		fmt.Printf("[JSON截断修复] 添加了 %d 个 } 和 %d 个 ]\n", openBraces, openBrackets)

		// 验证修复后的 JSON
		var test map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &test); err != nil {
			fmt.Printf("[JSON截断修复] 修复后仍无法解析: %v\n", err)
			// 返回原始字符串，让调用者处理
			return strings.TrimSpace(jsonStr)
		}

		fmt.Printf("[JSON截断修复] 修复成功！\n")
	}

	return jsonStr
}

// pow is a simple power function
func pow(base, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp*100); i++ {
		result *= base
	}
	return result
}

// StreamCallback 流式输出回调函数类型
type StreamCallback func(chunk string) error

// RollEventCallback 判定事件回调函数类型
type RollEventCallback func(rollEvent map[string]interface{}) error

// ProcessActionStreamWithAttributes processes a player action with custom attributes and streaming narrative
func (gc *GameController) ProcessActionStreamWithAttributes(playerID, modID, action string, customAttributes map[string]interface{}, streamCallback StreamCallback, rollCallback RollEventCallback, secondStageCallback StreamCallback) error {
	session, err := gc.stateManager.GetSession(playerID, modID)
	if err != nil {
		return err
	}

	mod, err := gc.modLoader.GetMod(modID)
	if err != nil {
		return err
	}

	// Check if already processing
	if isProcessing, ok := session.State["is_processing"].(bool); ok && isProcessing {
		return fmt.Errorf("已有操作正在处理中")
	}

	session.State["is_processing"] = true

	// 检测燃魂爆运指令 [SOUL_BURN]
	soulBurnMode := false
	if strings.Contains(action, "[SOUL_BURN]") {
		soulBurnMode = true
		action = strings.ReplaceAll(action, "[SOUL_BURN]", "")
		action = strings.TrimSpace(action)
		session.State["force_success"] = true
		session.State["soul_burn_mode"] = true // 标记为燃魂爆运模式

		// 检查自定义属性中的要求内容
		actionContent := action
		if customAttributes != nil {
			if content, ok := customAttributes["action_content"].(string); ok && content != "" {
				actionContent = content
			}
		}

		// 生成代价（根据要求的难度生成不同的代价）
		penalty := gc.generateSoulBurnPenalty(actionContent, session)

		// 添加代价到累积列表
		if session.State["soul_burn_penalties"] == nil {
			session.State["soul_burn_penalties"] = []string{}
		}
		if penalties, ok := session.State["soul_burn_penalties"].([]interface{}); ok {
			session.State["soul_burn_penalties"] = append(penalties, penalty)
		} else {
			// 兼容处理：初始化为新数组
			session.State["soul_burn_penalties"] = []string{penalty}
		}

		// 应用代价到角色状态
		gc.applySoulBurnPenalty(session, penalty)

		fmt.Printf("[燃魂爆运] 检测到 [SOUL_BURN] 指令，本次判定将强制成功！\n")
		fmt.Printf("[燃魂爆运] 燃烧灵魂代价：%s\n", penalty)
		fmt.Printf("[燃魂爆运] 玩家要求：%s\n", actionContent)
	}

	// 检测作弊指令 [SUCCESS]（保留原有的作弊模式）
	forceSuccess := false
	if strings.Contains(action, "[SUCCESS]") {
		forceSuccess = true
		action = strings.ReplaceAll(action, "[SUCCESS]", "")
		action = strings.TrimSpace(action)
		session.State["force_success"] = true
		session.State["cheat_mode"] = true // 标记为作弊模式，AI将完全服从
		fmt.Printf("[作弊模式] 检测到 [SUCCESS] 指令，本次判定将强制成功，AI将完全服从玩家指令！\n")
	}

	gc.stateManager.SaveSession(session)

	// Note: User action is already added to display_history by frontend for immediate display
	// 当前用户消息不添加到历史记录，将在buildAIMessages中处理
	// 历史记录只保存已完成的对话轮次
	fmt.Printf("[ProcessActionStreamWithAttributes] 当前用户动作: %s（不添加到历史记录）\n", action)
	if soulBurnMode {
		fmt.Printf("[ProcessActionStreamWithAttributes] [燃魂爆运模式激活] 强制成功标志已设置，代价已应用\n")
	}
	if forceSuccess {
		fmt.Printf("[ProcessActionStreamWithAttributes] [作弊模式激活] 强制成功标志已设置\n")
	}

	var prompt string

	// Handle special actions
	if action == "start_trial" {
		// Use start trial prompt
		startPrompt := mod.Prompts["start_trial"]
		if startPrompt == "" {
			startPrompt = mod.Prompts["start_game"]
		}

		// 如果有自定义属性，添加到prompt中
		if customAttributes != nil && len(customAttributes) > 0 {
			attrStr := "\n\n🔴【极其重要：用户自定义角色属性】🔴\n"
			attrStr += "⚠️ 以下是用户明确要求的角色设定，你必须100%严格遵守，不可更改任何一个字：\n\n"
			hasCustomAttrs := false

			if name, ok := customAttributes["姓名"].(string); ok && name != "" {
				attrStr += fmt.Sprintf("✅ 姓名：%s（必须使用此姓名，不可更改）\n", name)
				hasCustomAttrs = true
			}
			if gender, ok := customAttributes["性别"].(string); ok && gender != "" {
				attrStr += fmt.Sprintf("✅ 性别：%s（必须是%s，不可更改）\n", gender, gender)
				hasCustomAttrs = true
			}
			if qualification, ok := customAttributes["资质"].(string); ok && qualification != "" {
				attrStr += fmt.Sprintf("✅ 资质：%s（必须是此资质等级）\n", qualification)
				hasCustomAttrs = true
			}
			if cultivation, ok := customAttributes["修为"].(string); ok && cultivation != "" {
				attrStr += fmt.Sprintf("✅ 修为：%s（必须是此修为境界）\n", cultivation)
				hasCustomAttrs = true
			}
			if spiritStones, ok := customAttributes["元石"].(float64); ok && spiritStones > 0 {
				attrStr += fmt.Sprintf("✅ 元石：%d枚（必须是此数量）\n", int(spiritStones))
				hasCustomAttrs = true
			}
			if background, ok := customAttributes["出身"].(string); ok && background != "" {
				attrStr += fmt.Sprintf("✅ 出身背景：%s（必须基于此背景展开故事）\n", background)
				hasCustomAttrs = true
			}

			if hasCustomAttrs {
				attrStr += "\n🔴【生成规则】🔴\n"
				attrStr += "1. ⚠️ 以上打✅的属性是用户的明确要求，必须100%使用，一字不改\n"
				attrStr += "2. 📝 未提及的属性（如蛊虫、关系网、道痕等）需要你根据已定义属性合理生成\n"
				attrStr += "3. 🎭 如果用户没有定义出身背景，请创造一个符合其他属性的精彩背景故事\n"
				attrStr += "4. 🔗 确保所有生成的内容与用户定义的属性保持逻辑一致性\n"
				attrStr += "5. ❌ 绝对禁止改变任何用户已定义的属性值！\n\n"
				attrStr += "记住：用户的自定义属性拥有最高优先级，必须覆盖任何默认设定！\n"

				prompt = startPrompt + attrStr
				fmt.Printf("添加自定义属性到prompt: %s\n", attrStr)
			} else {
				// 用户选择了随机生成
				prompt = startPrompt + "\n\n【完全随机生成】\n用户没有提供任何自定义属性，请为玩家创造一个独特的角色，包括姓名、性别、资质、修为、元石、出身背景等所有属性。"
			}
		} else {
			prompt = startPrompt
		}
	} else {
		// 对于普通动作，不需要额外的prompt
		// 用户消息已经在RecentHistory中，游戏状态会在buildAIMessages中作为系统消息添加
		prompt = ""
	}

	// Call AI with streaming (with retry mechanism)
	maxRetries := 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		fmt.Printf("[一阶段重试-WithAttr] 尝试第 %d/%d 次调用AI\n", attempt, maxRetries)

		err = gc.callAIStream(session, prompt, mod, action, streamCallback, rollCallback, secondStageCallback)
		if err == nil {
			fmt.Printf("[一阶段重试-WithAttr] 第 %d 次调用成功\n", attempt)
			break
		}

		lastErr = err
		fmt.Printf("[一阶段重试-WithAttr] 第 %d 次调用失败: %v\n", attempt, err)

		// 二阶段错误不应该触发一阶段重试
		if strings.Contains(err.Error(), "second stage") {
			fmt.Printf("[一阶段重试-WithAttr] 二阶段错误，不重试一阶段（二阶段已有备用方案）\n")
			err = nil // 清除错误
			break
		}

		// 检查是否是一阶段的JSON格式错误或截断错误
		isTruncatedError := strings.Contains(err.Error(), "unexpected end of JSON") ||
			strings.Contains(err.Error(), "EOF")
		isFormatError := strings.Contains(err.Error(), "no valid JSON found") ||
			strings.Contains(err.Error(), "failed to parse") ||
			strings.Contains(err.Error(), "invalid character")

		if isTruncatedError {
			fmt.Printf("[一阶段重试-WithAttr] 检测到JSON被截断（输出过长），准备精简重试...\n")

			if attempt < maxRetries {
				fmt.Printf("[一阶段重试-WithAttr] 发送清除信号给前端\n")
				streamCallback("__CLEAR_NARRATIVE__")

				// 截断错误：要求 AI 精简输出
				currentStateJSON, _ := json.Marshal(session.State)
				gameMasterPrompt := mod.Prompts["game_master"]
				if gameMasterPrompt == "" {
					gameMasterPrompt = "你是这个游戏世界的掌控者。"
				}

				// 添加精简输出的强制要求
				prompt = fmt.Sprintf("%s\n\n玩家动作: %s\n\n🚨【紧急】上次输出被截断！请遵循以下要求：\n1. 叙事内容控制在800字以内\n2. JSON结构必须完整\n3. 出身描述控制在100字以内\n4. 关系网最多3个人物\n5. 物品列表最多5个\n6. 必须确保JSON以}正确结束\n\n当前游戏状态：\n%s", gameMasterPrompt, action, string(currentStateJSON))
				fmt.Printf("[一阶段重试-WithAttr] 使用精简输出模式重试\n")
			}
		} else if isFormatError {
			fmt.Printf("[一阶段重试-WithAttr] 检测到JSON格式错误，准备重试...\n")

			if attempt < maxRetries {
				// 🔴【关键】在重试前发送清除信号，告诉前端清除之前的流式内容
				// 这样可以避免第一次失败的内容和重试成功的内容混在一起
				fmt.Printf("[一阶段重试-WithAttr] 发送清除信号给前端\n")
				streamCallback("__CLEAR_NARRATIVE__")

				// 对于 start_trial 或 start_new_trial，使用完整的游戏提示词重试
				if action == "start_trial" || action == "start_new_trial" {
					currentStateJSON, _ := json.Marshal(session.State)
					// 使用完整的游戏提示词，确保 AI 理解游戏上下文
					gamePrompt := mod.Prompts["start_trial"]
					if gamePrompt == "" {
						gamePrompt = mod.Prompts["start_game"]
					}
					prompt = fmt.Sprintf("%s\n\n⚠️ 重要格式要求：\n1. 必须严格按照JSON格式输出\n2. 确保JSON语法正确，特别注意引号和逗号\n3. 所有字符串值都要用双引号包围\n4. 叙事内容放在 $...$ 标记中，JSON放在 @...@ 标记中\n5. 叙事风格必须符合游戏世界观\n6. 字符串内部用【】代替中文引号\n\n当前游戏状态：\n%s", gamePrompt, string(currentStateJSON))
					fmt.Printf("[一阶段重试-WithAttr] 使用游戏提示词重试，提示词长度: %d\n", len(gamePrompt))
				} else {
					// 普通动作：使用 game_master 提示词 + 用户动作
					currentStateJSON, _ := json.Marshal(session.State)
					gameMasterPrompt := mod.Prompts["game_master"]
					if gameMasterPrompt == "" {
						gameMasterPrompt = "你是这个游戏世界的掌控者，请根据玩家的行动生成符合游戏世界观的响应。"
					}
					prompt = fmt.Sprintf("%s\n\n玩家动作: %s\n\n⚠️ 重要格式要求：\n1. 必须严格按照JSON格式输出\n2. 确保JSON语法正确，特别注意引号和逗号\n3. 所有字符串值都要用双引号包围\n4. 叙事内容放在 $...$ 标记中，JSON放在 @...@ 标记中\n5. 字符串内部用【】代替中文引号\n\n当前游戏状态：\n%s", gameMasterPrompt, action, string(currentStateJSON))
					fmt.Printf("[一阶段重试-WithAttr] 使用 game_master 提示词重试\n")
				}
			}
		} else {
			fmt.Printf("[一阶段重试-WithAttr] 其他错误，不进行重试: %v\n", err)
			break
		}
	}

	if err != nil {
		fmt.Printf("[一阶段重试-WithAttr] 所有重试均失败，最后错误: %v\n", lastErr)
		session.State["is_processing"] = false
		delete(session.State, "force_success")
		delete(session.State, "cheat_mode")
		delete(session.State, "soul_burn_mode")
		gc.stateManager.SaveSession(session)
		return fmt.Errorf("first stage AI call failed after %d attempts: %w", maxRetries, lastErr)
	}

	session.State["is_processing"] = false
	delete(session.State, "force_success")
	delete(session.State, "cheat_mode")
	delete(session.State, "soul_burn_mode")
	gc.stateManager.SaveSession(session)

	return err
}

// ProcessActionStream processes a player action with streaming narrative
func (gc *GameController) ProcessActionStream(playerID, modID, action string, streamCallback StreamCallback, rollCallback RollEventCallback, secondStageCallback StreamCallback) error {
	session, err := gc.stateManager.GetSession(playerID, modID)
	if err != nil {
		return err
	}

	mod, err := gc.modLoader.GetMod(modID)
	if err != nil {
		return err
	}

	// Check if already processing
	if isProcessing, ok := session.State["is_processing"].(bool); ok && isProcessing {
		return fmt.Errorf("已有操作正在处理中")
	}

	session.State["is_processing"] = true

	// 检测作弊指令 [SUCCESS]
	forceSuccess := false
	if strings.Contains(action, "[SUCCESS]") {
		forceSuccess = true
		action = strings.ReplaceAll(action, "[SUCCESS]", "")
		action = strings.TrimSpace(action)
		session.State["force_success"] = true
		session.State["cheat_mode"] = true // 标记为作弊模式，AI将完全服从
		fmt.Printf("[作弊模式] 检测到 [SUCCESS] 指令，本次判定将强制成功，AI将完全服从玩家指令！\n")
	}

	gc.stateManager.SaveSession(session)

	// Note: User action is already added to display_history by frontend for immediate display
	// 当前用户消息不添加到历史记录，将在buildAIMessages中处理
	// 历史记录只保存已完成的对话轮次
	fmt.Printf("[ProcessActionStream] 当前用户动作: %s（不添加到历史记录）\n", action)
	if forceSuccess {
		fmt.Printf("[ProcessActionStream] [作弊模式激活] 强制成功标志已设置\n")
	}

	var prompt string

	// Handle special actions
	if action == "start_trial" {
		// Use start trial prompt
		startPrompt := mod.Prompts["start_trial"]
		if startPrompt == "" {
			startPrompt = mod.Prompts["start_game"]
		}
		prompt = startPrompt
	} else {
		// 对于普通动作，不需要额外的prompt
		// 用户消息已经在RecentHistory中，游戏状态会在buildAIMessages中作为系统消息添加
		prompt = ""
	}

	// Call AI with streaming (with retry mechanism)
	maxRetries := 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		fmt.Printf("[一阶段重试-Stream] 尝试第 %d/%d 次调用AI\n", attempt, maxRetries)

		err = gc.callAIStream(session, prompt, mod, action, streamCallback, rollCallback, secondStageCallback)
		if err == nil {
			fmt.Printf("[一阶段重试-Stream] 第 %d 次调用成功\n", attempt)
			break
		}

		lastErr = err
		fmt.Printf("[一阶段重试-Stream] 第 %d 次调用失败: %v\n", attempt, err)

		// 二阶段错误不应该触发一阶段重试
		if strings.Contains(err.Error(), "second stage") {
			fmt.Printf("[一阶段重试-Stream] 二阶段错误，不重试一阶段（二阶段已有备用方案）\n")
			err = nil // 清除错误
			break
		}

		// 检查是否是一阶段的JSON格式错误
		if strings.Contains(err.Error(), "no valid JSON found") ||
			strings.Contains(err.Error(), "failed to parse") {
			fmt.Printf("[一阶段重试-Stream] 检测到JSON格式错误，准备重试...\n")

			if attempt < maxRetries {
				// 🔴【关键】在重试前发送清除信号，告诉前端清除之前的流式内容
				fmt.Printf("[一阶段重试-Stream] 发送清除信号给前端\n")
				streamCallback("__CLEAR_NARRATIVE__")

				// 对于 start_trial 或 start_new_trial，使用完整的游戏提示词重试
				if action == "start_trial" || action == "start_new_trial" {
					currentStateJSON, _ := json.Marshal(session.State)
					// 使用完整的游戏提示词，确保 AI 理解游戏上下文
					gamePrompt := mod.Prompts["start_trial"]
					if gamePrompt == "" {
						gamePrompt = mod.Prompts["start_game"]
					}
					prompt = fmt.Sprintf("%s\n\n⚠️ 重要格式要求：\n1. 必须严格按照JSON格式输出\n2. 确保JSON语法正确，特别注意引号和逗号\n3. 所有字符串值都要用双引号包围\n4. 叙事内容放在 $...$ 标记中，JSON放在 @...@ 标记中\n5. 叙事风格必须符合游戏世界观\n\n当前游戏状态：\n%s", gamePrompt, string(currentStateJSON))
					fmt.Printf("[一阶段重试-Stream] 使用游戏提示词重试，提示词长度: %d\n", len(gamePrompt))
				} else {
					// 普通动作：使用 game_master 提示词 + 用户动作
					currentStateJSON, _ := json.Marshal(session.State)
					gameMasterPrompt := mod.Prompts["game_master"]
					if gameMasterPrompt == "" {
						gameMasterPrompt = "你是这个游戏世界的掌控者，请根据玩家的行动生成符合游戏世界观的响应。"
					}
					prompt = fmt.Sprintf("%s\n\n玩家动作: %s\n\n⚠️ 重要格式要求：\n1. 必须严格按照JSON格式输出\n2. 确保JSON语法正确，特别注意引号和逗号\n3. 所有字符串值都要用双引号包围\n4. 叙事内容放在 $...$ 标记中，JSON放在 @...@ 标记中\n\n当前游戏状态：\n%s", gameMasterPrompt, action, string(currentStateJSON))
					fmt.Printf("[一阶段重试-Stream] 使用 game_master 提示词重试\n")
				}
			}
		} else {
			fmt.Printf("[一阶段重试-Stream] 其他错误，不进行重试: %v\n", err)
			break
		}
	}

	if err != nil {
		fmt.Printf("[一阶段重试-Stream] 所有重试均失败，最后错误: %v\n", lastErr)
		session.State["is_processing"] = false
		delete(session.State, "force_success")
		delete(session.State, "cheat_mode")
		delete(session.State, "soul_burn_mode")
		gc.stateManager.SaveSession(session)
		return fmt.Errorf("first stage AI call failed after %d attempts: %w", maxRetries, lastErr)
	}

	session.State["is_processing"] = false
	delete(session.State, "force_success")
	delete(session.State, "cheat_mode")
	delete(session.State, "soul_burn_mode")
	gc.stateManager.SaveSession(session)

	return err
}

// callAIStream calls AI service with streaming support
func (gc *GameController) callAIStream(session *GameSession, prompt string, mod *GameMod, originalAction string, streamCallback StreamCallback, rollCallback RollEventCallback, secondStageCallback StreamCallback) error {
	// 使用新的消息构建方法，传递游戏状态、当前用户动作和特殊prompt（如果有）
	messages := gc.buildAIMessages(session, session.State, mod, originalAction, prompt)

	// 调试：打印发送给AI的消息
	fmt.Printf("\n=== 发送给AI的消息 (%d条) ===\n", len(messages))
	for i, msg := range messages {
		contentPreview := msg.Content
		if len(contentPreview) > 200 {
			contentPreview = contentPreview[:200] + "...(总长:" + fmt.Sprintf("%d", len(msg.Content)) + ")"
		}
		fmt.Printf("[%d] %s: %s\n", i, msg.Role, contentPreview)
	}
	fmt.Printf("=== 消息结束 ===\n\n")

	// 根据MOD获取对应的Provider配置
	provider := gc.GetProviderForMod(mod.Config.GameID)

	// Check if AI provider is configured
	if provider.APIKey == "" {
		return fmt.Errorf("AI provider not configured")
	}

	fmt.Printf("使用AI提供商: %s, 模型: %s\n", provider.APIType, provider.ModelID)

	// Call AI service with streaming
	var response interface{}
	var err error

	switch provider.APIType {
	case "openai":
		response, err = gc.aiClient.CallOpenAI(
			provider.BaseURL,
			provider.APIKey,
			provider.ModelID,
			messages,
			true, // streaming
		)
	case "anthropic":
		response, err = gc.aiClient.CallAnthropic(
			provider.BaseURL,
			provider.APIKey,
			provider.ModelID,
			messages,
			true,
		)
	case "google":
		response, err = gc.aiClient.CallGoogle(
			provider.BaseURL,
			provider.APIKey,
			provider.ModelID,
			messages,
			true,
		)
	default:
		return fmt.Errorf("unsupported API type: %s", provider.APIType)
	}

	if err != nil {
		return fmt.Errorf("AI call failed: %w", err)
	}

	// Process stream
	body, ok := response.(io.ReadCloser)
	if !ok {
		return fmt.Errorf("invalid stream response")
	}
	defer body.Close()

	scanner := bufio.NewScanner(body)
	buf := make([]byte, 0, 128*1024)
	scanner.Buffer(buf, 2*1024*1024)

	var fullResponse strings.Builder
	var narrativeBuffer strings.Builder
	var jsonStarted bool

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		chunk := gc.aiClient.ParseStreamChunk(provider.APIType, line)
		if chunk != nil {
			if content, ok := chunk["content"].(string); ok && content != "" {
				fullResponse.WriteString(content)

				if !jsonStarted {
					// 检测是否遇到 @ 标记（新格式）、JSON标记或markdown代码块开始
					// 注意：``` 可能分多个chunk发送，所以也需要检测单独的 ``` 和 json 关键字
					if strings.Contains(content, "@") ||
						strings.Contains(content, "```json") ||
						strings.Contains(content, "```") ||
						strings.Contains(content, "{") ||
						strings.TrimSpace(content) == "json" {
						jsonStarted = true
						// 发送JSON标记之前的内容（但过滤掉 ```json 等标记）
						beforeJson := content
						if atMarkIndex := strings.Index(content, "@"); atMarkIndex >= 0 {
							beforeJson = content[:atMarkIndex]
						} else if jsonMarkIndex := strings.Index(content, "```json"); jsonMarkIndex >= 0 {
							beforeJson = content[:jsonMarkIndex]
						} else if codeBlockIndex := strings.Index(content, "```"); codeBlockIndex >= 0 {
							beforeJson = content[:codeBlockIndex]
						} else if jsonIndex := strings.Index(content, "{"); jsonIndex >= 0 {
							beforeJson = content[:jsonIndex]
						}

						// 清理可能残留的 markdown 标记和 json 关键字
						beforeJson = strings.ReplaceAll(beforeJson, "```", "")
						beforeJson = strings.ReplaceAll(beforeJson, "json", "")
						beforeJson = strings.TrimSpace(beforeJson)

						if beforeJson != "" {
							narrativeBuffer.WriteString(beforeJson)
							if err := streamCallback(beforeJson); err != nil {
								return err
							}
						}
					} else {
						// 纯narrative内容，直接发送（但过滤掉 $ 和 ``` 标记以及 json 关键字）
						content = strings.ReplaceAll(content, "$", "")
						content = strings.ReplaceAll(content, "```", "")
						// 如果内容只是 "json" 或者以 "json" 结尾，需要特殊处理
						if strings.TrimSpace(content) == "json" {
							// 这可能是 ```json 被拆分发送，跳过
							jsonStarted = true
							continue
						}
						if strings.TrimSpace(content) != "" {
							narrativeBuffer.WriteString(content)
							if err := streamCallback(content); err != nil {
								return err
							}
						}
					}
				}
				// JSON部分不再流式发送
			}

			if done, ok := chunk["done"].(bool); ok && done {
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// Parse and apply the complete response
	aiResponse := fullResponse.String()

	fmt.Printf("\n=== AI完整响应 ===\n%s\n=== 响应结束 ===\n", aiResponse)

	// Parse the response to check for roll_request
	jsonStr := extractJSON(aiResponse)
	if jsonStr == "" {
		fmt.Printf("ERROR: 无法从AI响应中提取JSON\n")
		fmt.Printf("完整响应: %s\n", aiResponse)
		fmt.Printf("响应长度: %d 字符\n", len(aiResponse))
		return fmt.Errorf("no valid JSON found in AI response")
	}

	fmt.Printf("\n=== 提取的JSON ===\n%s\n=== JSON结束 ===\n", jsonStr)

	// 🔴【关键修复】先尝试自动修复JSON中的常见问题（如中文引号）
	// 这样可以避免因为小问题而需要完全重试
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		fmt.Printf("[JSON解析] 首次解析失败，尝试自动修复: %v\n", err)

		// 步骤1：尝试修复中文引号
		fixedJSON := fixJSONString(jsonStr)
		if err2 := json.Unmarshal([]byte(fixedJSON), &parsed); err2 == nil {
			fmt.Printf("[JSON解析] 中文引号修复成功！继续处理\n")
			jsonStr = fixedJSON
		} else {
			// 步骤2：尝试修复截断问题
			fmt.Printf("[JSON解析] 中文引号修复后仍失败，尝试截断修复...\n")
			fixedJSON = fixTruncatedJSON(fixedJSON)
			if err3 := json.Unmarshal([]byte(fixedJSON), &parsed); err3 == nil {
				fmt.Printf("[JSON解析] 截断修复成功！继续处理\n")
				jsonStr = fixedJSON
			} else {
				// 所有修复都失败
				fmt.Printf("DEBUG: 所有自动修复都失败: %v\n", err3)
				fmt.Printf("DEBUG: 修复后的JSON前500字符: %.500s\n", fixedJSON)
				return fmt.Errorf("failed to parse AI response JSON: %w", err)
			}
		}
	}

	// 【关键修复】如果流式过程中没有发送任何 narrative 内容，从 JSON 中提取并发送
	streamedNarrative := strings.TrimSpace(narrativeBuffer.String())
	if streamedNarrative == "" {
		// 先尝试从 $...$ 格式提取
		narrativeFromFormat := extractNarrative(aiResponse)
		narrativeToSend := narrativeFromFormat

		// 如果没有 $...$ 格式，从 JSON 的 narrative 字段提取
		if narrativeToSend == "" {
			if jsonNarrative, ok := parsed["narrative"].(string); ok && jsonNarrative != "" {
				narrativeToSend = jsonNarrative
			}
		}

		// 发送提取到的 narrative 给前端
		if narrativeToSend != "" {
			fmt.Printf("[流式修复] 流式过程未发送 narrative，从 JSON 提取并发送: %d 字符\n", len(narrativeToSend))
			if err := streamCallback(narrativeToSend); err != nil {
				fmt.Printf("[流式修复] 发送 narrative 失败: %v\n", err)
			}
		} else {
			fmt.Printf("[流式修复] 警告：无法提取到有效的 narrative 内容\n")
		}
	} else {
		fmt.Printf("[流式处理] 流式过程中已发送 narrative: %d 字符\n", len(streamedNarrative))
	}

	// 【叙事校验】使用AI验证器校验并修正叙事内容
	if gc.narrativeValidator != nil {
		// 提取叙事内容
		narrativeFromFormat := extractNarrative(aiResponse)
		narrative := narrativeFromFormat
		if narrative == "" {
			narrative, _ = parsed["narrative"].(string)
		}

		// 校验叙事（一阶段没有判定结果，传空字符串）
		validationResult := gc.narrativeValidator.ValidateNarrative(narrative, "", mod.Config.GameID)
		if !validationResult.IsValid {
			fmt.Printf("[叙事校验] 一阶段发现问题: %v\n", validationResult.Violations)
			// 优先使用AI修正后的内容
			if validationResult.CorrectedText != "" {
				parsed["narrative"] = validationResult.CorrectedText
				fmt.Printf("[叙事校验] 已应用AI修正\n")
			} else if narrative != "" {
				// 备用：使用快速过滤替换禁止词汇
				filteredNarrative := gc.narrativeValidator.QuickFilter(narrative)
				if filteredNarrative != narrative {
					parsed["narrative"] = filteredNarrative
					fmt.Printf("[叙事校验] 已应用快速过滤修正\n")
				}
			}
		}
	}

	// Add to history and handle compression
	aiMsg := Message{
		Role:      "assistant",
		Content:   aiResponse,
		Timestamp: time.Now(),
	}

	// 创建当前用户消息
	currentUserMsg := Message{
		Role:      "user",
		Content:   originalAction,
		Timestamp: time.Now(),
	}

	// 处理对话历史压缩
	gc.compressionManager.ProcessNewMessage(session, currentUserMsg, aiMsg)

	// Check if this is a roll request (two-stage judgment)
	if rollRequest, hasRoll := parsed["roll_request"].(map[string]interface{}); hasRoll {
		// Execute roll
		rollResult := gc.executeRoll(rollRequest, mod, session)

		// Send roll event to frontend
		rollEvent := map[string]interface{}{
			"type":        rollRequest["type"],
			"target":      rollRequest["target"],
			"description": rollRequest["description"],
			"result":      rollResult["result"],
			"outcome":     rollResult["outcome"],
			"success":     rollResult["success"],
		}

		// Send roll event to frontend
		if rollCallback != nil {
			if err := rollCallback(rollEvent); err != nil {
				return err
			}
		}

		// Send roll result as separate message via streaming
		rollResultText := fmt.Sprintf("【判定结果：%s】", rollResult["outcome"])
		if err := streamCallback(rollResultText); err != nil {
			return err
		}

		// Call AI again with roll result for second stage
		currentStateJSON, _ := json.Marshal(session.State)
		prompt := fmt.Sprintf(`判定已完成：%s

系统已经完成了判定，结果是【%s】。请直接根据这个结果生成后续叙事。

⚠️ 重要要求：
1. 直接描述判定结果带来的后果和后续情节
2. 必须输出完整的 $...$ 叙事和 @...@ JSON格式
3. 在 state_update 中更新因判定结果而改变的状态
4. 不要输出判定过程（如"🎲 正在进行判定..."），判定已经完成
5. 不要再次发起 roll_request

当前状态：
%s`, rollResult["outcome"], rollResult["outcome"], string(currentStateJSON))

		// Get first narrative for comparison - prefer format over JSON
		firstNarrativeFromFormat := extractNarrative(aiResponse)
		firstNarrative := firstNarrativeFromFormat
		if firstNarrative == "" {
			firstNarrative, _ = parsed["narrative"].(string)
		}

		// Second stage AI call (streaming) with retry mechanism
		maxRetries := 3
		var lastErr error

		for attempt := 1; attempt <= maxRetries; attempt++ {
			fmt.Printf("[二阶段重试] 尝试第 %d/%d 次调用AI\n", attempt, maxRetries)

			err = gc.callAIStreamSecondStage(session, prompt, mod, firstNarrative, secondStageCallback)
			if err == nil {
				fmt.Printf("[二阶段重试] 第 %d 次调用成功\n", attempt)
				break
			}

			lastErr = err
			fmt.Printf("[二阶段重试] 第 %d 次调用失败: %v\n", attempt, err)

			// 检查是否是JSON格式错误
			if strings.Contains(err.Error(), "no valid JSON found") ||
				strings.Contains(err.Error(), "failed to parse") {
				fmt.Printf("[二阶段重试] 检测到JSON格式错误，准备重试...\n")

				if attempt < maxRetries {
					// 在重试前稍等一下，避免请求过于频繁
					// time.Sleep(time.Millisecond * 500)

					// 修改prompt，要求AI更加注意格式
					prompt = fmt.Sprintf(`判定已完成：%s

系统已经完成了判定，结果是【%s】。请直接根据这个结果生成后续叙事。

⚠️ 重要格式要求：
1. 直接描述判定结果带来的后果，不要输出判定过程
2. 必须输出 $...$ 叙事和 @...@ JSON格式
3. JSON必须语法正确，使用英文标点
4. 在 state_update 中更新状态
5. 不要再次发起 roll_request

当前状态：
%s`, rollResult["outcome"], rollResult["outcome"], string(currentStateJSON))
				}
			} else {
				// 非格式错误，不重试
				fmt.Printf("[二阶段重试] 非格式错误，不进行重试: %v\n", err)
				break
			}
		}

		if err != nil {
			fmt.Printf("[二阶段重试] 所有重试均失败，最后错误: %v\n", lastErr)
			// 二阶段失败时，不要返回错误让一阶段重试
			// 而是应用一阶段的状态更新，并记录错误
			fmt.Printf("[二阶段重试] 二阶段失败，将保留一阶段结果，判定结果为: %s\n", rollResult["outcome"])

			// 根据判定结果生成一个简单的失败/成功提示
			outcomeStr := rollResult["outcome"].(string)
			var fallbackNarrative string
			if outcomeStr == "失败" || outcomeStr == "大失败" {
				fallbackNarrative = fmt.Sprintf("【判定结果：%s】\n\n你的尝试未能成功，命运的天平并未向你倾斜。", outcomeStr)
			} else {
				fallbackNarrative = fmt.Sprintf("【判定结果：%s】\n\n你的努力得到了回报，命运眷顾了你。", outcomeStr)
			}

			// 通过回调发送备用叙事
			if secondStageCallback != nil {
				secondStageCallback(fallbackNarrative)
			}

			// 不返回错误，让流程继续
			// return fmt.Errorf("second stage AI call failed after %d attempts: %w", maxRetries, lastErr)
		}

	} else {
		// No roll request, direct state update
		if narrative, ok := parsed["narrative"].(string); ok && narrative != "" {
			// Note: Don't add to DisplayHistory here, frontend handles it via streaming
		}

		// Apply state update (支持 state_update 和 game_state 两种格式)
		var stateUpdate map[string]interface{}
		if su, ok := parsed["state_update"].(map[string]interface{}); ok {
			stateUpdate = su
		} else if gs, ok := parsed["game_state"].(map[string]interface{}); ok {
			// 兼容 game_state 格式（某些 AI 可能返回这种格式）
			stateUpdate = gs
			fmt.Printf("[状态更新] 使用 game_state 格式（兼容模式）\n")
		}

		if stateUpdate != nil {
			ApplyStateUpdate(session.State, stateUpdate)

			// Check if trial ended (game over)
			if isInTrial, exists := stateUpdate["is_in_trial"]; exists {
				if inTrial, ok := isInTrial.(bool); ok && !inTrial {
					// Trial ended, immediately stop processing
					session.State["is_processing"] = false
					fmt.Printf("DEBUG: Trial ended, setting is_processing = false\n")
				}
			}

			// Check for special program triggers
			if trigger, hasTrigger := stateUpdate["trigger_program"].(map[string]interface{}); hasTrigger {
				gc.handleProgramTrigger(session, trigger, mod)
			}
		}

		// 清除作弊模式标志（如果没有判定，说明不需要作弊模式了）
		// if _, exists := session.State["cheat_mode"]; exists {
		// 	delete(session.State, "cheat_mode")
		// 	fmt.Printf("[作弊模式] 无需判定，作弊模式标志已清除\n")
		// }
	}

	// Save session
	return gc.stateManager.SaveSession(session)
}

// filterDuplicateContent filters out duplicate content from the second narrative
func filterDuplicateContent(secondNarrative, firstNarrative string) string {
	// If first narrative is empty, return second narrative as is
	if firstNarrative == "" {
		return secondNarrative
	}

	// Clean up the narratives
	secondNarrative = strings.TrimSpace(secondNarrative)
	firstNarrative = strings.TrimSpace(firstNarrative)

	// Simple approach: if second narrative starts with first narrative,
	// return only the part after first narrative
	if strings.HasPrefix(secondNarrative, firstNarrative) {
		remaining := strings.TrimPrefix(secondNarrative, firstNarrative)
		return strings.TrimSpace(remaining)
	}

	// Split both narratives into sentences for better comparison
	firstSentences := strings.Split(firstNarrative, "。")
	secondSentences := strings.Split(secondNarrative, "。")

	// Find where the unique content starts in second narrative
	uniqueStartIndex := len(secondSentences) // Default to end if no unique content found

	for i, sentence := range secondSentences {
		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			continue
		}

		// Check if this sentence exists in first narrative
		found := false
		for _, firstSentence := range firstSentences {
			firstSentence = strings.TrimSpace(firstSentence)
			if firstSentence != "" && strings.Contains(sentence, firstSentence) || strings.Contains(firstSentence, sentence) {
				found = true
				break
			}
		}

		if !found {
			uniqueStartIndex = i
			break
		}
	}

	// Return unique sentences
	if uniqueStartIndex < len(secondSentences) {
		uniqueSentences := secondSentences[uniqueStartIndex:]
		result := strings.Join(uniqueSentences, "。")
		return strings.TrimSpace(result)
	}

	// If no unique content found, return the second narrative as is
	// (this might happen if AI generates completely new content)
	return secondNarrative
}

// callAIStreamSecondStage calls AI service for second stage with streaming support
func (gc *GameController) callAIStreamSecondStage(session *GameSession, prompt string, mod *GameMod, firstNarrative string, secondStageCallback StreamCallback) error {
	// Build messages from session history - 注意：第二阶段不是游戏开始，需要完整的游戏规则
	// 将 prompt 作为 currentUserAction 传入，而不是 specialPrompt
	messages := gc.buildAIMessages(session, session.State, mod, prompt)

	// 根据MOD获取对应的Provider配置
	provider := gc.GetProviderForMod(mod.Config.GameID)

	// Check if AI provider is configured
	if provider.APIKey == "" {
		return fmt.Errorf("AI provider not configured")
	}

	// Call AI service with streaming
	var response interface{}
	var err error

	switch provider.APIType {
	case "openai":
		response, err = gc.aiClient.CallOpenAI(
			provider.BaseURL,
			provider.APIKey,
			provider.ModelID,
			messages,
			true, // streaming
		)
	case "anthropic":
		response, err = gc.aiClient.CallAnthropic(
			provider.BaseURL,
			provider.APIKey,
			provider.ModelID,
			messages,
			true,
		)
	case "google":
		response, err = gc.aiClient.CallGoogle(
			provider.BaseURL,
			provider.APIKey,
			provider.ModelID,
			messages,
			true,
		)
	default:
		return fmt.Errorf("unsupported API type: %s", provider.APIType)
	}

	if err != nil {
		return fmt.Errorf("AI call failed: %w", err)
	}

	// Process stream
	body, ok := response.(io.ReadCloser)
	if !ok {
		return fmt.Errorf("invalid stream response")
	}
	defer body.Close()

	scanner := bufio.NewScanner(body)
	buf := make([]byte, 0, 128*1024)
	scanner.Buffer(buf, 2*1024*1024)

	var fullResponse strings.Builder
	var narrativeBuffer strings.Builder
	var jsonStarted bool

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		chunk := gc.aiClient.ParseStreamChunk(provider.APIType, line)
		if chunk != nil {
			if content, ok := chunk["content"].(string); ok && content != "" {
				fullResponse.WriteString(content)

				if !jsonStarted {
					// 检测是否遇到 @ 标记（新格式）、JSON标记或markdown代码块开始
					// 注意：``` 可能分多个chunk发送，所以也需要检测单独的 ``` 和 json 关键字
					if strings.Contains(content, "@") ||
						strings.Contains(content, "```json") ||
						strings.Contains(content, "```") ||
						strings.Contains(content, "{") ||
						strings.TrimSpace(content) == "json" {
						jsonStarted = true
						// 发送JSON标记之前的内容（但过滤掉 ```json 等标记）
						beforeJson := content
						if atMarkIndex := strings.Index(content, "@"); atMarkIndex >= 0 {
							beforeJson = content[:atMarkIndex]
						} else if jsonMarkIndex := strings.Index(content, "```json"); jsonMarkIndex >= 0 {
							beforeJson = content[:jsonMarkIndex]
						} else if codeBlockIndex := strings.Index(content, "```"); codeBlockIndex >= 0 {
							beforeJson = content[:codeBlockIndex]
						} else if jsonIndex := strings.Index(content, "{"); jsonIndex >= 0 {
							beforeJson = content[:jsonIndex]
						}

						// 清理可能残留的 markdown 标记和 json 关键字
						beforeJson = strings.ReplaceAll(beforeJson, "```", "")
						beforeJson = strings.ReplaceAll(beforeJson, "json", "")
						beforeJson = strings.TrimSpace(beforeJson)

						if beforeJson != "" {
							narrativeBuffer.WriteString(beforeJson)
							if err := secondStageCallback(beforeJson); err != nil {
								return err
							}
						}
					} else {
						// 纯narrative内容，直接发送（但过滤掉 $ 和 ``` 标记以及 json 关键字）
						content = strings.ReplaceAll(content, "$", "")
						content = strings.ReplaceAll(content, "```", "")
						// 如果内容只是 "json" 或者以 "json" 结尾，需要特殊处理
						if strings.TrimSpace(content) == "json" {
							// 这可能是 ```json 被拆分发送，跳过
							jsonStarted = true
							continue
						}
						if strings.TrimSpace(content) != "" {
							narrativeBuffer.WriteString(content)
							if err := secondStageCallback(content); err != nil {
								return err
							}
						}
					}
				}
				// JSON部分不再流式发送
			}

			if done, ok := chunk["done"].(bool); ok && done {
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// Parse and apply the complete response
	aiResponse := fullResponse.String()

	fmt.Printf("\n=== 第二阶段AI完整响应 ===\n%s\n=== 响应结束 ===\n", aiResponse)

	// Parse the response
	jsonStr := extractJSON(aiResponse)
	if jsonStr == "" {
		fmt.Printf("ERROR: 无法从第二阶段AI响应中提取JSON\n")
		fmt.Printf("完整响应: %s\n", aiResponse)
		fmt.Printf("响应长度: %d 字符\n", len(aiResponse))
		return fmt.Errorf("no valid JSON found in second AI response")
	}

	fmt.Printf("\n=== 第二阶段提取的JSON ===\n%s\n=== JSON结束 ===\n", jsonStr)

	// 🔴【关键修复】先尝试自动修复JSON中的常见问题（如中文引号）
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		fmt.Printf("[二阶段JSON解析] 首次解析失败，尝试自动修复: %v\n", err)

		// 步骤1：尝试修复中文引号
		fixedJSON := fixJSONString(jsonStr)
		if err2 := json.Unmarshal([]byte(fixedJSON), &parsed); err2 == nil {
			fmt.Printf("[二阶段JSON解析] 中文引号修复成功！继续处理\n")
			jsonStr = fixedJSON
		} else {
			// 步骤2：尝试修复截断问题
			fmt.Printf("[二阶段JSON解析] 中文引号修复后仍失败，尝试截断修复...\n")
			fixedJSON = fixTruncatedJSON(fixedJSON)
			if err3 := json.Unmarshal([]byte(fixedJSON), &parsed); err3 == nil {
				fmt.Printf("[二阶段JSON解析] 截断修复成功！继续处理\n")
				jsonStr = fixedJSON
			} else {
				// 所有修复都失败
				fmt.Printf("DEBUG: 二阶段所有自动修复都失败: %v\n", err3)
				return fmt.Errorf("failed to parse second AI response JSON: %w", err)
			}
		}
	}

	// 【关键修复】如果流式过程中没有发送任何 narrative 内容，从 JSON 中提取并发送
	streamedNarrative := strings.TrimSpace(narrativeBuffer.String())
	if streamedNarrative == "" {
		// 先尝试从 $...$ 格式提取
		narrativeFromFormat := extractNarrative(aiResponse)
		narrativeToSend := narrativeFromFormat

		// 如果没有 $...$ 格式，从 JSON 的 narrative 字段提取
		if narrativeToSend == "" {
			if jsonNarrative, ok := parsed["narrative"].(string); ok && jsonNarrative != "" {
				narrativeToSend = jsonNarrative
			}
		}

		// 发送提取到的 narrative 给前端
		if narrativeToSend != "" {
			fmt.Printf("[二阶段流式修复] 流式过程未发送 narrative，从 JSON 提取并发送: %d 字符\n", len(narrativeToSend))
			if err := secondStageCallback(narrativeToSend); err != nil {
				fmt.Printf("[二阶段流式修复] 发送 narrative 失败: %v\n", err)
			}
		} else {
			fmt.Printf("[二阶段流式修复] 警告：无法提取到有效的 narrative 内容\n")
		}
	} else {
		fmt.Printf("[二阶段流式处理] 流式过程中已发送 narrative: %d 字符\n", len(streamedNarrative))
	}

	// 【叙事校验】使用AI验证器校验并修正二阶段叙事内容
	// 二阶段有判定结果，需要校验叙事与判定结果的一致性
	if gc.narrativeValidator != nil {
		// 提取叙事内容
		narrativeFromFormat := extractNarrative(aiResponse)
		narrative := narrativeFromFormat
		if narrative == "" {
			narrative, _ = parsed["narrative"].(string)
		}

		// 从 prompt 中提取判定结果（格式：判定已完成：XXX）
		rollOutcome := ""
		if strings.Contains(prompt, "判定已完成：") {
			startIdx := strings.Index(prompt, "判定已完成：") + len("判定已完成：")
			endIdx := strings.Index(prompt[startIdx:], "\n")
			if endIdx > 0 {
				rollOutcome = strings.TrimSpace(prompt[startIdx : startIdx+endIdx])
			}
		}

		// 校验叙事与判定结果的一致性（使用AI校验）
		validationResult := gc.narrativeValidator.ValidateNarrative(narrative, rollOutcome, mod.Config.GameID)
		if !validationResult.IsValid {
			fmt.Printf("[叙事校验] 二阶段发现问题: %v\n", validationResult.Violations)

			// 优先使用AI修正后的内容
			if validationResult.CorrectedText != "" {
				parsed["narrative"] = validationResult.CorrectedText
				fmt.Printf("[叙事校验] 已应用AI修正后的叙事\n")
			} else if narrative != "" {
				// 备用：使用快速过滤替换禁止词汇
				filteredNarrative := gc.narrativeValidator.QuickFilter(narrative)
				if filteredNarrative != narrative {
					parsed["narrative"] = filteredNarrative
					fmt.Printf("[叙事校验] 已应用快速过滤修正\n")
				}
			}
		}
	}

	// Add to history and handle compression
	aiMsg := Message{
		Role:      "assistant",
		Content:   aiResponse,
		Timestamp: time.Now(),
	}
	session.RecentHistory = append(session.RecentHistory, aiMsg)

	// Apply state update
	if stateUpdate, ok := parsed["state_update"].(map[string]interface{}); ok {
		ApplyStateUpdate(session.State, stateUpdate)

		// Check if trial ended (game over) in second response
		if isInTrial, exists := stateUpdate["is_in_trial"]; exists {
			if inTrial, ok := isInTrial.(bool); ok && !inTrial {
				// Trial ended, immediately stop processing
				session.State["is_processing"] = false
				fmt.Printf("DEBUG: Trial ended in second response, setting is_processing = false\n")
			}
		}

		// Check for special program triggers
		if trigger, hasTrigger := stateUpdate["trigger_program"].(map[string]interface{}); hasTrigger {
			gc.handleProgramTrigger(session, trigger, mod)
		}
	}

	return nil
}

// generateSoulBurnPenalty 生成燃魂爆运的代价
func (gc *GameController) generateSoulBurnPenalty(actionContent string, session *GameSession) string {
	// 根据要求的复杂度判断代价等级
	difficulty := gc.assessActionDifficulty(actionContent)

	// 预设的代价池
	minorPenalties := []string{
		"寿命减少三年",
		"永久失去一成功力",
		"灵感枯竭，悟性降低一个层次",
		"运气流失，未来三次判定-10",
		"心魔种下，突破时额外增加一重天劫",
		"气血亏损，恢复速度永久降低20%",
		"资质受损，修炼速度降低15%",
	}

	moderatePenalties := []string{
		"寿命减少十年",
		"永久失去三成功力",
		"道心出现裂痕，无法感悟天意",
		"气运断绝，所有判定永久-20",
		"心魔缠身，每次突破必遭心魔劫",
		"经脉受损，无法修炼高阶功法",
		"神魂受创，永久失去一项天赋",
		"血脉退化，资质降低一个大等级",
	}

	severePenalties := []string{
		"寿命减少五十年",
		"修为跌落一个大境界",
		"道基崩塌，此生止步于当前境界",
		"天意诅咒，遭受天谴随时可能陨落",
		"神魂燃烧，记忆开始逐渐消散",
		"血脉逆转，变为废体无法修炼",
		"气运耗尽，成为天弃之人",
		"命格破碎，注定悲惨结局",
	}

	// 选择代价
	rand.Seed(time.Now().UnixNano())
	var penalty string

	switch difficulty {
	case 1: // 简单要求
		penalty = minorPenalties[rand.Intn(len(minorPenalties))]
	case 2: // 中等要求
		penalty = moderatePenalties[rand.Intn(len(moderatePenalties))]
	case 3: // 困难要求
		penalty = severePenalties[rand.Intn(len(severePenalties))]
	default:
		penalty = minorPenalties[rand.Intn(len(minorPenalties))]
	}

	return penalty
}

// assessActionDifficulty 评估玩家要求的难度
func (gc *GameController) assessActionDifficulty(action string) int {
	// 关键词检测评估难度
	action = strings.ToLower(action)

	// 高难度关键词
	highDifficultyKeywords := []string{
		"无敌", "最强", "秒杀", "毁灭", "统治", "称霸", "成神", "飞升",
		"突破极限", "超越", "完美", "绝对", "所有", "全部", "立即",
	}

	// 中等难度关键词
	mediumDifficultyKeywords := []string{
		"击败", "获得", "学会", "突破", "晋升", "掌握", "成功", "达到",
		"获取", "得到", "击杀", "战胜", "领悟",
	}

	// 计算匹配度
	highCount := 0
	mediumCount := 0

	for _, keyword := range highDifficultyKeywords {
		if strings.Contains(action, keyword) {
			highCount++
		}
	}

	for _, keyword := range mediumDifficultyKeywords {
		if strings.Contains(action, keyword) {
			mediumCount++
		}
	}

	// 根据匹配度返回难度等级
	if highCount > 0 {
		return 3 // 高难度
	} else if mediumCount > 0 {
		return 2 // 中等难度
	} else {
		return 1 // 低难度
	}
}

// applySoulBurnPenalty 应用燃魂代价到角色状态
func (gc *GameController) applySoulBurnPenalty(session *GameSession, penalty string) {
	// 这里可以根据具体的代价内容修改角色属性
	// 由于属性结构可能很复杂，这里只做示例性的处理

	if currentLife, ok := session.State["current_life"].(map[string]interface{}); ok {
		// 根据代价类型应用不同的惩罚
		if strings.Contains(penalty, "寿命") {
			// 减少寿元
			if lifespan, ok := currentLife["lifespan"].(float64); ok {
				if strings.Contains(penalty, "三年") {
					currentLife["lifespan"] = lifespan - 3
				} else if strings.Contains(penalty, "十年") {
					currentLife["lifespan"] = lifespan - 10
				} else if strings.Contains(penalty, "五十年") {
					currentLife["lifespan"] = lifespan - 50
				}
			}
		}

		if strings.Contains(penalty, "功力") {
			// 减少修为值
			if cultivation, ok := currentLife["cultivation_value"].(float64); ok {
				if strings.Contains(penalty, "一成") {
					currentLife["cultivation_value"] = cultivation * 0.9
				} else if strings.Contains(penalty, "三成") {
					currentLife["cultivation_value"] = cultivation * 0.7
				}
			}
		}

		if strings.Contains(penalty, "资质") || strings.Contains(penalty, "血脉") {
			// 降低资质
			if qualification, ok := currentLife["qualification"].(string); ok {
				qualificationMap := map[string]string{
					"甲等资质": "乙等资质",
					"乙等资质": "丙等资质",
					"丙等资质": "丁等资质",
					"丁等资质": "废体",
				}
				if newQual, exists := qualificationMap[qualification]; exists {
					currentLife["qualification"] = newQual
				}
			}
		}

		// 记录代价已应用
		currentLife["soul_burn_applied"] = true
	}
}
