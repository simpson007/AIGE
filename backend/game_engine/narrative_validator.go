package game_engine

import (
	"AIGE/config"
	"AIGE/models"
	"AIGE/services"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// ValidatorConfig 校验器配置
type ValidatorConfig struct {
	Enabled              bool   `json:"enabled"`                // 是否启用校验
	UseRuleValidation    bool   `json:"use_rule_validation"`    // 是否启用规则校验（AI）
	UseConsistencyCheck  bool   `json:"use_consistency_check"`  // 是否启用一致性校验（AI）
	UseLogicCheck        bool   `json:"use_logic_check"`        // 是否启用逻辑一致性校验（AI）
	UseAutoCorrection    bool   `json:"use_auto_correction"`    // 是否启用自动修正（AI）
	ValidatorModelID     string `json:"validator_model_id"`     // 校验用模型ID
}

// NarrativeValidator 叙事校验器
// 使用AI校验和修正叙事内容
type NarrativeValidator struct {
	aiClient            *services.AIClient
	gameController      *GameController
	enabled             bool
	useRuleValidation   bool       // 规则校验（AI检测禁止词汇）
	useConsistencyCheck bool       // 一致性校验（AI检测判定结果与叙事是否一致）
	useLogicCheck       bool       // 逻辑一致性校验（AI检测叙事内部逻辑是否自洽）
	useAutoCorrection   bool       // 自动修正（AI修正问题叙事）
	validatorProvider   *AIProvider
	configMutex         sync.RWMutex
}

// ValidationResult 校验结果
type ValidationResult struct {
	IsValid       bool     `json:"is_valid"`
	Violations    []string `json:"violations"`
	Suggestions   []string `json:"suggestions"`
	CorrectedText string   `json:"corrected_text,omitempty"`
}

// AIValidationResponse AI校验响应结构
type AIValidationResponse struct {
	IsValid    bool     `json:"is_valid"`
	Violations []string `json:"violations"`
	Corrected  string   `json:"corrected,omitempty"`
}

// NewNarrativeValidator 创建叙事校验器
func NewNarrativeValidator(aiClient *services.AIClient) *NarrativeValidator {
	nv := &NarrativeValidator{
		aiClient:            aiClient,
		enabled:             true,
		useRuleValidation:   true,
		useConsistencyCheck: true,
		useLogicCheck:       true,
		useAutoCorrection:   true,
	}
	// 从数据库加载配置
	nv.LoadConfig()
	return nv
}

// LoadConfig 从数据库加载校验器配置
func (nv *NarrativeValidator) LoadConfig() {
	nv.configMutex.Lock()
	defer nv.configMutex.Unlock()

	db := config.DB
	if db == nil {
		fmt.Printf("[NarrativeValidator] 数据库未初始化，使用默认配置\n")
		return
	}

	// 加载启用状态
	var enabledConfig models.SystemConfig
	if err := db.Where("key = ?", "validator_enabled").First(&enabledConfig).Error; err == nil {
		nv.enabled = enabledConfig.Value == "true"
	}

	// 加载规则校验开关
	var ruleConfig models.SystemConfig
	if err := db.Where("key = ?", "validator_rule_check").First(&ruleConfig).Error; err == nil {
		nv.useRuleValidation = ruleConfig.Value == "true"
	}

	// 加载一致性校验开关
	var consistencyConfig models.SystemConfig
	if err := db.Where("key = ?", "validator_consistency_check").First(&consistencyConfig).Error; err == nil {
		nv.useConsistencyCheck = consistencyConfig.Value == "true"
	}

	// 加载逻辑一致性校验开关
	var logicConfig models.SystemConfig
	if err := db.Where("key = ?", "validator_logic_check").First(&logicConfig).Error; err == nil {
		nv.useLogicCheck = logicConfig.Value == "true"
	}

	// 加载自动修正开关
	var correctionConfig models.SystemConfig
	if err := db.Where("key = ?", "validator_auto_correction").First(&correctionConfig).Error; err == nil {
		nv.useAutoCorrection = correctionConfig.Value == "true"
	}

	// 加载校验模型配置
	var modelConfig models.SystemConfig
	if err := db.Where("key = ?", "validator_model_id").First(&modelConfig).Error; err == nil && modelConfig.Value != "" {
		provider := nv.loadProviderFromModelID(modelConfig.Value)
		if provider != nil {
			nv.validatorProvider = provider
			fmt.Printf("[NarrativeValidator] 加载校验模型: %s / %s\n", provider.APIType, provider.ModelID)
		}
	}

	fmt.Printf("[NarrativeValidator] 配置加载完成 - 启用: %v, 规则校验: %v, 一致性校验: %v, 逻辑校验: %v, 自动修正: %v\n",
		nv.enabled, nv.useRuleValidation, nv.useConsistencyCheck, nv.useLogicCheck, nv.useAutoCorrection)
}

// loadProviderFromModelID 根据模型ID加载Provider配置
func (nv *NarrativeValidator) loadProviderFromModelID(modelID string) *AIProvider {
	db := config.DB
	if db == nil {
		return nil
	}

	var model models.Model
	if err := db.Preload("Provider").Where("id = ? AND enabled = ?", modelID, true).First(&model).Error; err != nil {
		fmt.Printf("[NarrativeValidator] 模型ID %s 不存在或未启用: %v\n", modelID, err)
		return nil
	}

	apiType := model.APIType
	if apiType == "" {
		apiType = model.Provider.Type
	}

	baseURL := model.Provider.BaseURL
	if baseURL == "" {
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

// SetGameController 设置GameController引用
func (nv *NarrativeValidator) SetGameController(gc *GameController) {
	nv.gameController = gc
}

// SetEnabled 启用/禁用校验器
func (nv *NarrativeValidator) SetEnabled(enabled bool) {
	nv.configMutex.Lock()
	defer nv.configMutex.Unlock()
	nv.enabled = enabled
}

// UpdateConfig 更新校验器配置（由管理员API调用）
func (nv *NarrativeValidator) UpdateConfig(cfg ValidatorConfig) {
	nv.configMutex.Lock()
	defer nv.configMutex.Unlock()

	nv.enabled = cfg.Enabled
	nv.useRuleValidation = cfg.UseRuleValidation
	nv.useConsistencyCheck = cfg.UseConsistencyCheck
	nv.useLogicCheck = cfg.UseLogicCheck
	nv.useAutoCorrection = cfg.UseAutoCorrection

	if cfg.ValidatorModelID != "" {
		provider := nv.loadProviderFromModelID(cfg.ValidatorModelID)
		if provider != nil {
			nv.validatorProvider = provider
			fmt.Printf("[NarrativeValidator] 更新校验模型: %s / %s\n", provider.APIType, provider.ModelID)
		}
	} else {
		nv.validatorProvider = nil
		fmt.Printf("[NarrativeValidator] 清除校验模型配置，将使用默认轻量模型\n")
	}

	fmt.Printf("[NarrativeValidator] 配置已更新 - 启用: %v, 规则校验: %v, 一致性校验: %v, 逻辑校验: %v, 自动修正: %v\n",
		nv.enabled, nv.useRuleValidation, nv.useConsistencyCheck, nv.useLogicCheck, nv.useAutoCorrection)
}

// GetConfig 获取当前配置
func (nv *NarrativeValidator) GetConfig() ValidatorConfig {
	nv.configMutex.RLock()
	defer nv.configMutex.RUnlock()

	cfg := ValidatorConfig{
		Enabled:             nv.enabled,
		UseRuleValidation:   nv.useRuleValidation,
		UseConsistencyCheck: nv.useConsistencyCheck,
		UseLogicCheck:       nv.useLogicCheck,
		UseAutoCorrection:   nv.useAutoCorrection,
	}

	if nv.validatorProvider != nil {
		cfg.ValidatorModelID = nv.validatorProvider.ModelID
	}

	return cfg
}

// getProvider 获取校验用的AI Provider
func (nv *NarrativeValidator) getProvider() *AIProvider {
	nv.configMutex.RLock()
	validatorProvider := nv.validatorProvider
	nv.configMutex.RUnlock()

	if validatorProvider != nil {
		return validatorProvider
	}

	// 回退到游戏默认模型的轻量版本
	if nv.gameController != nil {
		provider := nv.gameController.defaultProvider
		// 自动选择轻量模型
		switch provider.APIType {
		case "openai":
			provider.ModelID = "gpt-4o-mini"
		case "anthropic":
			provider.ModelID = "claude-3-haiku-20240307"
		case "google":
			provider.ModelID = "gemini-1.5-flash"
		}
		return &provider
	}

	return nil
}

// ValidateNarrative 综合校验叙事内容（使用AI）
func (nv *NarrativeValidator) ValidateNarrative(narrative string, rollOutcome string, modID string) *ValidationResult {
	result := &ValidationResult{
		IsValid:     true,
		Violations:  []string{},
		Suggestions: []string{},
	}

	nv.configMutex.RLock()
	enabled := nv.enabled
	useRuleValidation := nv.useRuleValidation
	useConsistencyCheck := nv.useConsistencyCheck
	useLogicCheck := nv.useLogicCheck
	useAutoCorrection := nv.useAutoCorrection
	nv.configMutex.RUnlock()

	if !enabled {
		return result
	}

	provider := nv.getProvider()
	if provider == nil || provider.APIKey == "" {
		fmt.Printf("[NarrativeValidator] 无可用的AI Provider，跳过校验\n")
		return result
	}

	// 构建校验任务
	var tasks []string
	if useRuleValidation {
		tasks = append(tasks, "rule")
	}
	if useConsistencyCheck && rollOutcome != "" {
		tasks = append(tasks, "consistency")
	}
	if useLogicCheck {
		tasks = append(tasks, "logic")
	}

	if len(tasks) == 0 {
		return result
	}

	// 调用AI进行校验
	aiResult := nv.callAIValidation(narrative, rollOutcome, tasks, useAutoCorrection, provider)
	if aiResult != nil {
		result.IsValid = aiResult.IsValid
		result.Violations = aiResult.Violations
		if aiResult.Corrected != "" {
			result.CorrectedText = aiResult.Corrected
			result.Suggestions = append(result.Suggestions, "已自动修正叙事内容")
		}
	}

	return result
}

// callAIValidation 调用AI进行校验
func (nv *NarrativeValidator) callAIValidation(narrative string, rollOutcome string, tasks []string, autoCorrect bool, provider *AIProvider) *AIValidationResponse {
	// 构建校验提示词
	prompt := nv.buildValidationPrompt(narrative, rollOutcome, tasks, autoCorrect)

	messages := []services.Message{
		{Role: "user", Content: prompt},
	}

	var response interface{}
	var err error

	fmt.Printf("[NarrativeValidator] 调用AI校验 - 模型: %s/%s, 任务: %v\n", provider.APIType, provider.ModelID, tasks)

	switch provider.APIType {
	case "openai":
		response, err = nv.aiClient.CallOpenAI(
			provider.BaseURL,
			provider.APIKey,
			provider.ModelID,
			messages,
			false,
		)
	case "anthropic":
		response, err = nv.aiClient.CallAnthropic(
			provider.BaseURL,
			provider.APIKey,
			provider.ModelID,
			messages,
			false,
		)
	case "google":
		response, err = nv.aiClient.CallGoogle(
			provider.BaseURL,
			provider.APIKey,
			provider.ModelID,
			messages,
			false,
		)
	default:
		response, err = nv.aiClient.CallOpenAI(
			provider.BaseURL,
			provider.APIKey,
			provider.ModelID,
			messages,
			false,
		)
	}

	if err != nil {
		fmt.Printf("[NarrativeValidator] AI校验调用失败: %v\n", err)
		return nil
	}

	// 解析AI响应
	return nv.parseAIValidationResponse(response)
}

// buildValidationPrompt 构建校验提示词
func (nv *NarrativeValidator) buildValidationPrompt(narrative string, rollOutcome string, tasks []string, autoCorrect bool) string {
	var sb strings.Builder

	sb.WriteString("你是一个游戏叙事校验助手。请检查以下叙事内容是否存在问题。\n\n")
	sb.WriteString("【待校验叙事】\n")
	sb.WriteString(narrative)
	sb.WriteString("\n\n")

	// 规则校验任务
	for _, task := range tasks {
		if task == "rule" {
			sb.WriteString("【规则校验】\n")
			sb.WriteString("检查叙事中是否包含以下禁止出现的内容：\n")
			sb.WriteString("- '轮回转生'、'转生机缘'、'十转轮回'、'永堕轮回' 等轮回相关概念\n")
			sb.WriteString("- '机缘尚余X次'、'尚余X次机缘' 等机缘次数描述\n")
			sb.WriteString("- '第X世' 等转世次数描述\n")
			sb.WriteString("\n")
		}
		if task == "consistency" {
			sb.WriteString("【一致性校验】\n")
			sb.WriteString(fmt.Sprintf("判定结果为：%s\n", rollOutcome))
			sb.WriteString("检查叙事内容是否与判定结果一致：\n")
			sb.WriteString("- 如果判定结果是'成功'或'大成功'，叙事应该描述成功的场景\n")
			sb.WriteString("- 如果判定结果是'失败'或'大失败'，叙事应该描述失败的场景\n")
			sb.WriteString("- 叙事的整体基调应该与判定结果匹配\n")
			sb.WriteString("\n")
		}
		if task == "logic" {
			sb.WriteString("【逻辑一致性校验】\n")
			sb.WriteString("检查叙事内部是否存在逻辑矛盾或前后不一致：\n")
			sb.WriteString("1. 角色言行一致性：\n")
			sb.WriteString("   - 角色说的话与做的事是否矛盾（如：说要帮助却做出伤害行为）\n")
			sb.WriteString("   - 角色的态度是否前后一致（如：先友善后突然敌对，无合理转折）\n")
			sb.WriteString("2. 物品/能力效果一致性：\n")
			sb.WriteString("   - 同一物品/能力的效果描述是否自相矛盾\n")
			sb.WriteString("   - 例如：药物不能同时\"治愈伤势\"又\"加重病情\"\n")
			sb.WriteString("3. 因果逻辑一致性：\n")
			sb.WriteString("   - 事件的因果关系是否合理\n")
			sb.WriteString("   - 结果是否与前文描述的行动相符\n")
			sb.WriteString("4. 数值/状态一致性：\n")
			sb.WriteString("   - 描述的数值变化是否合理（如：受伤后不能说\"毫发无损\"）\n")
			sb.WriteString("\n")
		}
	}

	sb.WriteString("【输出要求】\n")
	sb.WriteString("请以JSON格式输出校验结果：\n")
	sb.WriteString("```json\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"is_valid\": true/false,\n")
	sb.WriteString("  \"violations\": [\"问题1\", \"问题2\", ...],\n")

	if autoCorrect {
		sb.WriteString("  \"corrected\": \"修正后的叙事内容（仅当is_valid为false时提供）\"\n")
	} else {
		sb.WriteString("  \"corrected\": \"\"\n")
	}

	sb.WriteString("}\n")
	sb.WriteString("```\n\n")

	if autoCorrect {
		sb.WriteString("如果发现问题，请在corrected字段中提供修正后的叙事，要求：\n")
		sb.WriteString("1. 保持原有的文风和世界观\n")
		sb.WriteString("2. 移除或替换禁止出现的内容\n")
		sb.WriteString("3. 确保叙事与判定结果一致\n")
		sb.WriteString("4. 修正逻辑矛盾，确保叙事内部自洽\n")
		sb.WriteString("5. 尽量保留原叙事的精彩部分\n")
	}

	sb.WriteString("\n只输出JSON，不要输出其他内容。")

	return sb.String()
}

// parseAIValidationResponse 解析AI校验响应
func (nv *NarrativeValidator) parseAIValidationResponse(response interface{}) *AIValidationResponse {
	respMap, ok := response.(map[string]interface{})
	if !ok {
		return nil
	}

	content, ok := respMap["content"].(string)
	if !ok || content == "" {
		return nil
	}

	// 提取JSON
	jsonStr := extractJSONFromResponse(content)
	if jsonStr == "" {
		fmt.Printf("[NarrativeValidator] 无法从AI响应中提取JSON: %s\n", content)
		return nil
	}

	var result AIValidationResponse
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		fmt.Printf("[NarrativeValidator] 解析AI响应JSON失败: %v, JSON: %s\n", err, jsonStr)
		return nil
	}

	fmt.Printf("[NarrativeValidator] AI校验结果 - 有效: %v, 问题数: %d\n", result.IsValid, len(result.Violations))
	return &result
}

// extractJSONFromResponse 从AI响应中提取JSON
func extractJSONFromResponse(content string) string {
	// 尝试提取 ```json ... ``` 格式
	if strings.Contains(content, "```json") {
		startIdx := strings.Index(content, "```json") + 7
		endIdx := strings.Index(content[startIdx:], "```")
		if endIdx > 0 {
			return strings.TrimSpace(content[startIdx : startIdx+endIdx])
		}
	}

	// 尝试提取 ``` ... ``` 格式
	if strings.Contains(content, "```") {
		startIdx := strings.Index(content, "```") + 3
		endIdx := strings.Index(content[startIdx:], "```")
		if endIdx > 0 {
			jsonContent := strings.TrimSpace(content[startIdx : startIdx+endIdx])
			if strings.HasPrefix(jsonContent, "{") {
				return jsonContent
			}
		}
	}

	// 尝试直接提取JSON对象
	if startIdx := strings.Index(content, "{"); startIdx >= 0 {
		if endIdx := strings.LastIndex(content, "}"); endIdx > startIdx {
			return content[startIdx : endIdx+1]
		}
	}

	return ""
}

// ValidateAndCorrectJSON 校验并修正JSON响应
func (nv *NarrativeValidator) ValidateAndCorrectJSON(jsonStr string, rollOutcome string) (string, *ValidationResult) {
	result := &ValidationResult{
		IsValid:    true,
		Violations: []string{},
	}

	nv.configMutex.RLock()
	enabled := nv.enabled
	nv.configMutex.RUnlock()

	if !enabled {
		return jsonStr, result
	}

	// 解析JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		result.IsValid = false
		result.Violations = append(result.Violations, "JSON解析失败")
		return jsonStr, result
	}

	// 提取叙事内容
	narrative, _ := parsed["narrative"].(string)
	if narrative == "" {
		return jsonStr, result
	}

	// 校验叙事
	validationResult := nv.ValidateNarrative(narrative, rollOutcome, "")

	if !validationResult.IsValid {
		result.IsValid = false
		result.Violations = validationResult.Violations

		// 如果有修正后的文本，替换原叙事
		if validationResult.CorrectedText != "" {
			parsed["narrative"] = validationResult.CorrectedText
			correctedJSON, err := json.Marshal(parsed)
			if err == nil {
				return string(correctedJSON), result
			}
		}
	}

	return jsonStr, result
}

// QuickFilter 快速过滤（替换禁止词汇）- 作为备用方案
func (nv *NarrativeValidator) QuickFilter(text string) string {
	replacements := map[string]string{
		"轮回转生": "蛊道修行",
		"转生机缘": "修行之路",
		"十转轮回": "蛊道征途",
		"永堕轮回": "身死道消",
	}

	result := text
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	// 使用正则替换"机缘尚余X次"等模式
	patterns := []struct {
		Pattern string
		Replace string
	}{
		{`机缘尚余[一二三四五六七八九十\d]+次`, "修行继续"},
		{`尚余[一二三四五六七八九十\d]+次机缘`, "修行继续"},
		{`第[一二三四五六七八九十\d]+世`, "此生"},
	}

	for _, p := range patterns {
		re := regexp.MustCompile(p.Pattern)
		result = re.ReplaceAllString(result, p.Replace)
	}

	return result
}
