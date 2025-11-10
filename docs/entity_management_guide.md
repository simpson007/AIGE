# AIGE 实体管理系统使用指南

## 问题背景

在AI驱动的游戏中，特别是长对话场景下，容易出现上下文不一致的问题：
- 角色性别混淆（初始女性，后期变成男性）
- NPC名字变化（张三变成李四）
- 角色属性不一致

## 解决方案架构

### 1. 核心组件

#### EntityManager（实体管理器）
- **文件**: `backend/game_engine/entity_manager.go`
- **功能**: 管理所有游戏实体（玩家、NPC、物品、地点）
- **特性**:
  - 实体注册和追踪
  - 属性锁定机制
  - 一致性验证
  - 序列化/反序列化

#### EntityIntegration（实体集成）
- **文件**: `backend/game_engine/entity_integration.go`
- **功能**: 将实体管理集成到游戏流程
- **特性**:
  - 游戏开始时注册玩家实体
  - 从AI响应中提取实体信息
  - 验证响应一致性
  - 注入实体上下文

### 2. 数据结构

```go
// 实体结构
type Entity struct {
    ID         string                 // 唯一标识符
    Type       EntityType             // 实体类型（player/npc/item/place）
    Name       string                 // 名称
    Attributes map[string]interface{} // 属性集合
    Locked     bool                   // 是否锁定
    CreatedAt  time.Time             // 创建时间
    UpdatedAt  time.Time             // 更新时间
}

// 实体注册表
type EntityRegistry struct {
    Entities      map[string]*Entity // 所有实体
    PlayerEntity  *Entity            // 玩家实体引用
    LockedFields  []string          // 锁定的字段列表
    Relationships map[string]string  // 实体关系映射
}
```

### 3. 锁定字段

以下字段一旦设置就不能更改：
- `player.gender` - 玩家性别
- `player.birthplace` - 出身地
- `player.initial_talent` - 初始天赋

## 集成步骤

### 1. 数据库更新

在 `models/models.go` 的 `GameSave` 结构中添加了：
```go
EntityRegistry string `json:"entity_registry" gorm:"type:text"`
```

### 2. StateManager 更新

在 `state_manager.go` 中：
- 添加 `entityManager` 字段
- 在 `loadFromDB` 中反序列化实体注册表
- 在 `saveToDB` 中序列化实体注册表

### 3. GameController 集成

在 `game_controller.go` 的 `buildAIMessages` 方法中：
```go
// 注入实体上下文
if gc.stateManager.GetEntityManager() != nil {
    entityContext := gc.stateManager.GetEntityManager().BuildEntityContext(session.PlayerID, session.ModID)
    if entityContext != "" {
        messages = append(messages, services.Message{
            Role:    "system",
            Content: entityContext,
        })
    }
}
```

### 4. 压缩策略优化

在 `compression_manager.go` 中更新了压缩提示：
```go
- 【必须保留】角色的核心属性（姓名、性别、出身、身份、修为等）
- 【必须保留】重要NPC的姓名和关系
- 绝对不要改变角色的性别描述
- 绝对不要混淆角色的姓名
```

## 使用示例

### 1. 游戏开始时注册玩家

```go
// 在 ProcessActionStreamWithAttributes 中
if action == "start_trial" && customAttributes != nil {
    integration := NewEntityIntegration(gc)
    integration.ProcessGameStartWithEntities(playerID, modID, customAttributes)
}
```

### 2. 处理AI响应时验证一致性

```go
// 验证响应一致性
integration := NewEntityIntegration(gc)
isValid, violations := integration.ValidateResponseConsistency(playerID, modID, aiResponse)
if !isValid {
    fmt.Printf("警告：检测到实体不一致: %v\n", violations)
}
```

### 3. 实体上下文示例

发送给AI的实体上下文：
```
【核心实体信息】
以下信息是锁定的，必须严格保持一致，不得更改：

【玩家角色】
- ID: player_123
- 姓名: 林逸风
- gender: 男 [锁定]
- cultivation: 筑基期
- background: 散修出身

【NPC: 苏长老】
- position: 青云宗外门长老
- relationship: 玩家的引路人

请严格遵守以上标记为[锁定]的信息，确保描述的一致性。
```

## 配置建议

### 1. 启用实体管理

在初始化 `StateManager` 时自动创建 `EntityManager`：
```go
sm := &StateManager{
    sessions:      make(map[string]map[string]*GameSession),
    entityManager: NewEntityManager(),
}
```

### 2. 定期清理

清理过期实体（可选）：
```go
// 清理30天未更新的实体
deletedCount := entityManager.CleanupOldEntities(playerID, modID, 30)
```

### 3. 调试模式

启用详细日志：
```go
fmt.Printf("[实体管理] 注册实体: %s (类型: %s)\n", entity.ID, entity.Type)
fmt.Printf("[实体管理] 性别锁定: %v\n", entity.Attributes["gender"])
```

## 性能优化

### 1. 缓存策略
- 实体信息缓存在内存中
- 只在session保存时序列化到数据库

### 2. 压缩优化
- 实体注册表独立存储
- 不参与对话历史压缩

### 3. 验证频率
- 仅在关键操作时验证一致性
- 避免每次AI调用都验证

## 故障排查

### 问题1：性别仍然不一致
**检查**：
1. 确认实体已正确注册
2. 查看数据库中 `entity_registry` 字段
3. 检查AI消息中是否包含实体上下文

### 问题2：NPC名字变化
**解决**：
1. 提高NPC提取规则精度
2. 在压缩时明确保留NPC信息
3. 使用更强的锁定机制

### 问题3：页面重载后实体丢失
**检查**：
1. 确认 `loadFromDB` 正确反序列化
2. 验证数据库迁移已执行
3. 检查 `SaveSession` 调用时机

## 扩展建议

### 1. NLP集成
使用更高级的NLP库提取实体：
- 命名实体识别（NER）
- 关系提取
- 属性提取

### 2. 规则引擎
添加更复杂的验证规则：
- 境界升级合理性
- 时间线一致性
- 地理位置合理性

### 3. 可视化管理
创建管理界面：
- 实体关系图
- 属性修改历史
- 冲突检测报告

## 总结

这个实体管理系统通过以下机制解决上下文不一致问题：

1. **结构化存储**：将关键信息独立于对话历史存储
2. **锁定机制**：防止关键属性被意外修改
3. **主动注入**：在每次AI调用时注入实体上下文
4. **一致性验证**：检测并警告不一致情况
5. **压缩保护**：确保压缩时不丢失实体信息

系统设计为可选组件，不影响现有功能，可以渐进式采用。