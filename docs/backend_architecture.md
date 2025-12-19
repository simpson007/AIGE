# AIGE 后端架构文档

## 项目概述

AIGE (AI Game Engine) 是一个基于 Go 语言开发的 AI 驱动游戏后端系统，使用 Gin 框架构建 RESTful API 和 WebSocket 服务，支持多种 AI 提供商（OpenAI、Anthropic、Google）进行游戏叙事生成。

## 技术栈

- **语言**: Go 1.24
- **Web框架**: Gin
- **数据库**: SQLite (通过 GORM)
- **认证**: JWT
- **实时通信**: WebSocket (gorilla/websocket)
- **密码加密**: bcrypt

---

## 目录结构

```
backend/
├── main.go                 # 程序入口
├── config/
│   ├── database.go         # 数据库配置
│   └── oauth.go            # OAuth配置
├── controllers/
│   ├── admin.go            # 管理员用户管理
│   ├── ai.go               # AI对话测试
│   ├── auth.go             # 用户认证
│   ├── chat.go             # 聊天记录管理
│   ├── config.go           # 系统配置
│   ├── game.go             # 游戏核心控制
│   ├── model.go            # AI模型管理
│   ├── oauth.go            # OAuth登录
│   └── provider.go         # AI提供商管理
├── game_engine/
│   ├── game_controller.go  # 游戏逻辑控制器
│   ├── state_manager.go    # 游戏状态管理
│   ├── mod_loader.go       # MOD加载器
│   ├── compression_manager.go # 历史压缩管理
│   ├── entity_manager.go   # 实体管理器
│   └── entity_integration.go # 实体集成
├── middleware/
│   └── auth.go             # 认证中间件
├── models/
│   ├── models.go           # 数据模型定义
│   └── migrate.go          # 数据库迁移
├── routes/
│   └── routes.go           # 路由定义
├── services/
│   ├── ai_client.go        # AI API客户端
│   └── model_fetcher.go    # 模型列表获取
└── utils/
    └── auth.go             # 认证工具函数
```

---

## 数据模型

### 1. User (用户)
```go
type User struct {
    ID            uint      // 主键
    Username      string    // 用户名 (唯一)
    Password      string    // 密码哈希
    Email         string    // 邮箱
    IsAdmin       bool      // 是否管理员
    OAuthProvider string    // OAuth提供商 (如 "linux-do")
    OAuthID       *string   // OAuth用户ID
    Avatar        string    // 头像URL
}
```

### 2. Provider (AI提供商)
```go
type Provider struct {
    ID             uint      // 主键
    Name           string    // 提供商名称
    Type           string    // 类型: openai/anthropic/google
    APIKey         string    // API密钥
    BaseURL        string    // API基础URL
    Enabled        bool      // 是否启用
    AllowCustomURL bool      // 是否允许自定义URL
    Models         []Model   // 关联的模型列表
}
```

### 3. Model (AI模型)
```go
type Model struct {
    ID           uint       // 主键
    ModelID      string     // 模型标识符 (如 gpt-4o)
    Name         string     // 显示名称
    ProviderID   uint       // 所属提供商ID
    Enabled      bool       // 是否启用
    APIType      string     // API类型 (可覆盖Provider的Type)
    Capabilities string     // 模型能力描述
    LastTested   *time.Time // 最后测试时间
    TestStatus   string     // 测试状态
}
```

### 4. GameSave (游戏存档)
```go
type GameSave struct {
    ID                uint      // 主键
    UserID            uint      // 用户ID
    ModID             string    // MOD标识
    SessionDate       string    // 会话日期
    State             string    // 游戏状态JSON
    RecentHistory     string    // 最近对话历史JSON
    CompressedSummary string    // 压缩摘要
    CompressionRound  int       // 压缩轮次
    DisplayHistory    string    // 显示历史JSON
    EntityRegistry    string    // 实体注册表JSON
}
```

### 5. SystemConfig (系统配置)
```go
type SystemConfig struct {
    ID    uint   // 主键
    Key   string // 配置键 (唯一)
    Value string // 配置值
}
```

---

## API 路由

### 公开路由 (无需认证)

| 方法 | 路径 | 功能 | 控制器 |
|------|------|------|--------|
| GET | `/health` | 健康检查 | inline |
| POST | `/api/auth/register` | 用户注册 | `auth.Register` |
| POST | `/api/auth/login` | 用户登录 | `auth.Login` |
| GET | `/api/auth/oauth/linux-do` | Linux.Do OAuth登录 | `oauth.LinuxDoLogin` |
| GET | `/api/auth/oauth/linux-do/callback` | OAuth回调 | `oauth.LinuxDoCallback` |

### 用户路由 (需要认证)

| 方法 | 路径 | 功能 | 控制器 |
|------|------|------|--------|
| GET | `/api/profile` | 获取用户信息 | `auth.GetProfile` |
| GET | `/api/game/mods` | 获取可用MOD列表 | `game.GetAvailableMods` |
| POST | `/api/game/init` | 初始化游戏 | `game.InitializeGame` |
| GET | `/api/game/ws` | WebSocket游戏连接 | `game.GameWebSocket` |
| GET | `/api/game/state` | 获取游戏状态 | `game.GetGameState` |
| DELETE | `/api/game/reset` | 重置游戏 | `game.ResetGame` |
| POST | `/api/game/save` | 手动保存 | `game.ManualSaveGame` |
| POST | `/api/game/restart-opportunities` | 重启机缘 | `game.RestartOpportunities` |

### 管理员路由 (需要管理员权限)

#### 用户管理
| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/api/admin/users` | 获取用户列表 |
| POST | `/api/admin/users` | 创建用户 |
| GET | `/api/admin/users/:id` | 获取单个用户 |
| PUT | `/api/admin/users/:id` | 更新用户 |
| PUT | `/api/admin/users/:id/password` | 修改密码 |
| DELETE | `/api/admin/users/:id` | 删除用户 |
| PUT | `/api/admin/users/:id/toggle-admin` | 切换管理员权限 |

#### AI提供商管理
| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/api/admin/providers` | 获取提供商列表 |
| POST | `/api/admin/providers` | 创建提供商 |
| GET | `/api/admin/providers/:id` | 获取单个提供商 |
| PUT | `/api/admin/providers/:id` | 更新提供商 |
| DELETE | `/api/admin/providers/:id` | 删除提供商 |
| PUT | `/api/admin/providers/:id/toggle` | 启用/禁用提供商 |
| GET | `/api/admin/providers/:id/models/available` | 获取可用模型列表 |
| GET | `/api/admin/providers/:id/test` | 测试连接 |

#### AI模型管理
| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/api/admin/models` | 获取模型列表 |
| POST | `/api/admin/models` | 创建模型 |
| GET | `/api/admin/models/:id` | 获取单个模型 |
| PUT | `/api/admin/models/:id` | 更新模型 |
| DELETE | `/api/admin/models/:id` | 删除模型 |
| PUT | `/api/admin/models/:id/toggle` | 启用/禁用模型 |
| POST | `/api/admin/models/:id/test` | 测试模型 |
| PUT | `/api/admin/models/:id/capabilities` | 更新模型能力 |

#### 系统配置
| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/api/admin/config` | 获取所有配置 |
| GET | `/api/admin/config/:key` | 获取单个配置 |
| POST | `/api/admin/config` | 设置配置 |
| POST | `/api/admin/config/batch` | 批量设置配置 |

#### 游戏配置
| 方法 | 路径 | 功能 |
|------|------|------|
| POST | `/api/admin/game/reload-config` | 重载游戏配置 |
| GET | `/api/admin/game/model-config` | 获取游戏模型配置 |
| POST | `/api/admin/game/model-config` | 保存游戏模型配置 |

#### OAuth配置
| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/api/admin/oauth/config` | 获取OAuth配置 |
| POST | `/api/admin/oauth/config` | 保存OAuth配置 |

#### 聊天记录管理
| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/api/admin/chats` | 获取所有聊天记录 |
| GET | `/api/admin/chats/:id` | 获取单条记录 |
| PUT | `/api/admin/chats/:id` | 更新记录 |
| DELETE | `/api/admin/chats/:id` | 删除记录 |
| DELETE | `/api/admin/chats/user/:user_id` | 删除用户所有记录 |
| GET | `/api/admin/chats/stats` | 获取统计信息 |
| GET | `/api/admin/chats/export` | 导出记录 |

---

## 核心模块详解

### 1. 游戏引擎 (game_engine)

#### GameController
游戏核心控制器，负责：
- 初始化游戏会话
- 处理玩家动作
- 调用AI生成叙事
- 解析AI响应并更新状态
- 执行判定（骰子系统）

**关键方法：**
```go
// 初始化游戏
InitializeGame(playerID, modID string) (*GameSession, error)

// 处理玩家动作（流式）
ProcessActionStreamWithAttributes(playerID, modID, action string, 
    customAttributes map[string]interface{}, 
    streamCallback StreamCallback, 
    rollCallback RollEventCallback, 
    secondStageCallback StreamCallback) error

// 获取MOD对应的AI配置
GetProviderForMod(modID string) AIProvider
```

**AI响应格式：**
- 叙事内容：`$...$` 包裹
- JSON数据：`@...@` 包裹

#### StateManager
游戏状态管理器，负责：
- 会话的创建、读取、保存、删除
- 内存缓存 + 数据库持久化
- 自动保存机制
- 每日重置检查

**GameSession 结构：**
```go
type GameSession struct {
    PlayerID         string                 // 玩家ID
    ModID            string                 // MOD ID
    SessionDate      string                 // 会话日期
    State            map[string]interface{} // 游戏状态
    RecentHistory    []Message              // 最近对话历史
    CompressedSummary string                // 压缩摘要
    CompressionRound int                    // 压缩轮次
    DisplayHistory   []string               // 显示历史
}
```

#### ModLoader
MOD加载器，负责：
- 扫描和加载MOD目录
- 解析MOD配置文件
- 加载提示词文件
- 加载世界观文档

**MOD目录结构：**
```
mods/
└── ModName/
    ├── config.json      # MOD配置
    ├── prompts/
    │   ├── game_master.txt
    │   └── start_game.txt
    └── lore/            # 世界观文档
```

#### CompressionManager
历史压缩管理器，负责：
- 监控对话历史长度
- 触发压缩（每15轮）
- 调用AI生成摘要
- 合并新旧摘要

#### EntityManager
实体管理器，负责：
- 注册和管理游戏实体（玩家、NPC、物品、地点）
- 锁定关键属性（如性别）
- 验证实体一致性
- 构建实体上下文提示

### 2. AI服务 (services)

#### AIClient
统一的AI API客户端，支持：
- OpenAI API (及兼容API)
- Anthropic Claude API
- Google Gemini API

**支持功能：**
- 同步调用
- 流式调用
- 自动URL构建
- 响应解析

#### ModelFetcher
模型列表获取器，用于从各AI提供商获取可用模型列表。

### 3. 认证系统 (middleware/auth)

**JWT认证流程：**
1. 用户登录获取JWT Token
2. 请求时在Header中携带 `Authorization: Bearer <token>`
3. WebSocket连接通过query参数 `?token=<token>` 传递

**中间件：**
- `AuthMiddleware`: 验证JWT，提取用户信息
- `AdminMiddleware`: 验证管理员权限

---

## 特殊功能

### 1. 作弊模式
玩家可在动作中包含 `[SUCCESS]` 触发作弊模式：
- 强制判定成功
- AI完全服从玩家指令
- 绕过所有限制

### 2. 燃魂爆运模式
玩家可在动作中包含 `[SOUL_BURN]` 触发：
- 强制判定成功
- 实现任何要求
- 但会产生灵魂代价

### 3. 自定义角色属性
游戏开始时可传入自定义属性：
- 姓名、性别、资质、修为、元石、出身
- 这些属性会被锁定，AI必须遵守

---

## 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `DATABASE_PATH` | 数据库文件路径 | `chat.db` (开发) / `/app/data/chat.db` (Docker) |
| `MODS_PATH` | MOD目录路径 | `../mods` (开发) / `./mods` (Docker) |
| `ALLOWED_ORIGINS` | CORS允许的域名 | localhost相关 |

---

## 默认配置

- **默认管理员**: 用户名 `admin`，密码 `admin123`
- **JWT密钥**: `your-secret-key-change-this-in-production` (需修改)
- **服务端口**: 8182
- **自动保存间隔**: 5分钟
- **历史压缩阈值**: 15轮对话

---

## 修改建议

### 安全相关
1. **JWT密钥**: `utils/auth.go` 中的 `JWTSecret` 应从环境变量读取
2. **默认密码**: 生产环境应禁用默认管理员或强制修改密码
3. **CORS配置**: 生产环境应限制允许的域名

### 功能扩展
1. **数据库**: 可考虑支持 PostgreSQL/MySQL
2. **缓存**: 可添加 Redis 缓存层
3. **日志**: 可添加结构化日志系统
4. **监控**: 可添加 Prometheus 指标

### 代码优化
1. **错误处理**: 部分地方错误处理可以更细致
2. **配置管理**: 可使用 Viper 等配置库
3. **测试覆盖**: 可添加更多单元测试
