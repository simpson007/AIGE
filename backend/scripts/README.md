# 用户管理脚本

## 添加新用户 (add_user.go)

这个脚本用于快速向 AIGE 系统添加新用户。

### 使用方法

```bash
cd backend/scripts
go run add_user.go -username=用户名 -password=密码 [选项]
```

### 参数说明

| 参数 | 必填 | 说明 | 默认值 |
|------|------|------|--------|
| `-username` | ✅ | 用户名 | 无 |
| `-password` | ✅ | 密码 | 无 |
| `-email` | ❌ | 邮箱地址 | username@example.com |
| `-admin` | ❌ | 是否为管理员 | false（普通用户） |
| `-db` | ❌ | 数据库文件路径 | ../chat.db |

### 使用示例

#### 1. 添加普通用户
```bash
go run add_user.go -username=zhangsan -password=123456
```

#### 2. 添加管理员
```bash
go run add_user.go -username=lisi -password=admin888 -admin
```

#### 3. 指定邮箱
```bash
go run add_user.go -username=wangwu -password=123456 -email=wangwu@qq.com
```

#### 4. 完整示例（管理员 + 自定义邮箱）
```bash
go run add_user.go -username=boss -password=secret123 -email=boss@company.com -admin
```

### 输出示例

成功创建用户后，会显示如下信息：

```
🔐 正在加密密码...
📂 正在连接数据库...
✨ 正在创建用户...

✅ 用户创建成功！
==========================================
用户ID:   5
用户名:   zhangsan
密码:     123456
邮箱:     zhangsan@example.com
权限:     普通用户
==========================================
```

### 错误处理

脚本会自动处理以下错误情况：

- ❌ 用户名或密码为空
- ❌ 用户名已存在
- ❌ 数据库连接失败
- ❌ 密码加密失败

### 注意事项

1. **安全性**：密码会使用 bcrypt 自动加密，无需手动加密
2. **唯一性**：用户名必须唯一，重复的用户名会创建失败
3. **数据库路径**：默认使用 `../chat.db`，如需指定其他路径请使用 `-db` 参数
4. **权限建议**：谨慎创建管理员账号，建议只给受信任的用户管理员权限

### 快速参考

```bash
# 查看帮助
go run add_user.go

# 添加普通用户（最常用）
go run add_user.go -username=新用户名 -password=密码

# 添加管理员（谨慎使用）
go run add_user.go -username=管理员名 -password=密码 -admin
```

### 后续操作

创建用户后，用户可以使用以下方式登录：
- 前端登录页面：http://localhost:5173
- API 登录接口：`POST /api/v1/auth/login`

---

**维护者**: AIGE Team
**最后更新**: 2025-10-28
