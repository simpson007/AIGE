# AIGE 项目部署指南

## 📋 目录

- [快速开始](#快速开始)
- [初始化设置](#初始化设置)
- [一键部署](#一键部署)
- [手动部署](#手动部署)
- [常见问题](#常见问题)
- [回滚操作](#回滚操作)

---

## 🚀 快速开始

### 前置要求

**本地环境：**
- Git 已安装
- SSH 可以连接到服务器
- 已配置 SSH 密钥（推荐）

**服务器环境：**
- Ubuntu/CentOS Linux
- Docker 和 Docker Compose 已安装
- Git 已安装
- 端口 8182（后端）和 3000（前端）已开放

---

## 🔧 初始化设置

### 1. 配置 Git 仓库

#### 方案 A：使用 GitHub/GitLab（推荐）

```bash
# 在项目根目录
cd /Users/yushenjian/Downloads/AIGE-main

# 初始化 Git 仓库（如果还没有）
git init

# 添加远程仓库
git remote add origin https://github.com/yourusername/AIGE.git
# 或使用 GitLab
# git remote add origin https://gitlab.com/yourusername/AIGE.git

# 提交初始代码
git add .
git commit -m "Initial commit"
git branch -M main
git push -u origin main
```

#### 方案 B：使用服务器 Git 仓库

如果不想使用第三方 Git 服务，可以在服务器上创建裸仓库：

```bash
# 在服务器上创建裸仓库
ssh root@101.43.42.250
mkdir -p /root/git/AIGE.git
cd /root/git/AIGE.git
git init --bare
exit

# 在本地添加远程仓库
cd /Users/yushenjian/Downloads/AIGE-main
git remote add origin root@101.43.42.250:/root/git/AIGE.git
git add .
git commit -m "Initial commit"
git push -u origin main
```

### 2. 在服务器上克隆项目

```bash
# SSH 连接到服务器
ssh root@101.43.42.250

# 克隆项目到指定目录
cd /root
git clone https://github.com/yourusername/AIGE.git
# 或从服务器裸仓库克隆
# git clone /root/git/AIGE.git

# 进入项目目录
cd AIGE

# 复制环境变量配置
cp .env.example .env

# 编辑环境变量（配置 API 密钥等）
nano .env
```

### 3. 配置部署脚本

编辑本地 `deploy.sh` 文件，修改服务器配置：

```bash
# 打开 deploy.sh
nano deploy.sh

# 修改以下配置项：
SERVER_USER="root"              # 服务器用户名
SERVER_HOST="101.43.42.250"     # 服务器 IP 或域名
SERVER_PATH="/opt/AIGE"        # 服务器项目路径
GIT_BRANCH="main"               # 部署分支
```

### 4. 配置 SSH 免密登录（可选但推荐）

```bash
# 生成 SSH 密钥（如果还没有）
ssh-keygen -t rsa -b 4096 -C "your_email@example.com"

# 复制公钥到服务器
ssh-copy-id root@101.43.42.250

# 测试连接
ssh root@101.43.42.250
```

---

## 🎯 一键部署

### 基本用法

```bash
# 在项目根目录运行
./deploy.sh "提交信息"
```

### 示例

```bash
# 修复 bug
./deploy.sh "修复 MOD 加载问题"

# 添加新功能
./deploy.sh "添加游戏存档功能"

# 更新配置
./deploy.sh "更新 CORS 配置"

# 不指定消息（使用默认消息）
./deploy.sh
```

### 部署流程

脚本会自动执行以下步骤：

1. ✅ 检查 Git 状态
2. ✅ 添加所有更改到暂存区
3. ✅ 提交更改到本地仓库
4. ✅ 推送到远程仓库
5. ✅ SSH 连接到服务器
6. ✅ 备份数据库
7. ✅ 拉取最新代码
8. ✅ 停止现有容器
9. ✅ 重新构建镜像
10. ✅ 启动服务
11. ✅ 健康检查
12. ✅ 显示部署结果

---

## 🔨 手动部署

如果需要更精细的控制，可以手动执行部署步骤。

### 在服务器上手动部署

```bash
# 1. SSH 连接到服务器
ssh root@101.43.42.250

# 2. 进入项目目录
cd /opt/AIGE

# 3. 运行服务器端部署脚本
./deploy/server-deploy.sh main
```

### 分步部署

```bash
# SSH 到服务器
ssh root@101.43.42.250
cd /opt/AIGE

# 1. 备份数据库
cp data/chat.db data/chat.db.backup.$(date +%Y%m%d_%H%M%S)

# 2. 拉取代码
git pull origin main

# 3. 停止服务
docker compose down

# 4. 重新构建
docker compose build --no-cache

# 5. 启动服务
docker compose up -d

# 6. 查看日志
docker compose logs -f
```

---

## 📊 部署后验证

### 1. 检查容器状态

```bash
ssh root@101.43.42.250 'cd /opt/AIGE && docker compose ps'
```

预期输出：
```
NAME                IMAGE               STATUS
aige-backend        aige-backend        Up
aige-frontend       aige-frontend       Up
```

### 2. 检查后端健康

```bash
curl http://101.43.42.250:8182/health
```

预期输出：
```json
{
  "status": "ok",
  "service": "AIGE Backend"
}
```

### 3. 检查 MOD 加载

```bash
curl http://101.43.42.250:8182/api/game/mods
```

应该能看到 `guzhenren` 等游戏信息。

### 4. 访问前端

打开浏览器访问：
- HTTP: `http://games.yushenjian.com:3000`
- HTTPS: `https://games.yushenjian.com`（需配置 Nginx 反向代理）

---

## 🐛 常见问题

### 问题 1：SSH 连接失败

```bash
# 检查 SSH 配置
ssh -v root@101.43.42.250

# 解决方案：
# 1. 检查服务器 IP 是否正确
# 2. 检查防火墙是否开放 22 端口
# 3. 检查 SSH 密钥是否正确
```

### 问题 2：Git 推送失败

```bash
# 检查远程仓库配置
git remote -v

# 重新配置远程仓库
git remote set-url origin https://github.com/yourusername/AIGE.git

# 强制推送（谨慎使用）
git push -f origin main
```

### 问题 3：Docker 构建失败

```bash
# 查看详细错误
docker compose build --no-cache --progress=plain

# 清理 Docker 资源
docker system prune -a

# 重新构建
docker compose build --no-cache
```

### 问题 4：MOD 加载失败

```bash
# 检查 MOD 目录
ssh root@101.43.42.250 'ls -la /opt/AIGE/mods'

# 查看后端日志
ssh root@101.43.42.250 'cd /opt/AIGE && docker compose logs backend | grep -i mod'

# 检查容器内 MOD 路径
ssh root@101.43.42.250 'docker exec aige-backend ls -la /app/mods'
```

### 问题 5：端口被占用

```bash
# 检查端口占用
ssh root@101.43.42.250 'netstat -tulpn | grep -E "8182|3000"'

# 停止占用端口的进程
ssh root@101.43.42.250 'fuser -k 8182/tcp'

# 重新启动
ssh root@101.43.42.250 'cd /opt/AIGE && docker compose up -d'
```

### 问题 6：数据库丢失

```bash
# 检查备份
ssh root@101.43.42.250 'ls -lh /opt/AIGE/backups/'

# 恢复最新备份
ssh root@101.43.42.250 'cd /opt/AIGE && cp backups/chat.db.YYYYMMDD_HHMMSS data/chat.db'

# 重启服务
ssh root@101.43.42.250 'cd /opt/AIGE && docker compose restart backend'
```

---

## ⏪ 回滚操作

### 快速回滚到上一个版本

```bash
# 在服务器上执行
ssh root@101.43.42.250
cd /opt/AIGE

# 查看提交历史
git log --oneline -10

# 回滚到指定版本
git reset --hard <commit-hash>

# 重新部署
./deploy/server-deploy.sh main
```

### 回滚到特定版本

```bash
# 方法 1：使用 git revert（推荐）
git revert <commit-hash>
git push origin main
./deploy.sh "回滚到正常版本"

# 方法 2：使用 git reset（谨慎）
git reset --hard <commit-hash>
git push -f origin main
ssh root@101.43.42.250 'cd /opt/AIGE && git pull origin main && docker compose up -d --build'
```

### 恢复数据库备份

```bash
ssh root@101.43.42.250
cd /opt/AIGE

# 查看可用备份
ls -lh backups/

# 停止服务
docker compose down

# 恢复备份
cp backups/chat.db.20241029_140000 data/chat.db

# 启动服务
docker compose up -d
```

---

## 📝 查看日志

### 实时查看所有日志

```bash
ssh root@101.43.42.250 'cd /opt/AIGE && docker compose logs -f'
```

### 只看后端日志

```bash
ssh root@101.43.42.250 'cd /opt/AIGE && docker compose logs -f backend'
```

### 只看前端日志

```bash
ssh root@101.43.42.250 'cd /opt/AIGE && docker compose logs -f frontend'
```

### 查看最近的日志

```bash
# 最近 100 行
ssh root@101.43.42.250 'cd /opt/AIGE && docker compose logs --tail=100'

# 最近 50 行后端日志
ssh root@101.43.42.250 'cd /opt/AIGE && docker compose logs --tail=50 backend'
```

---

## 🔧 维护命令

### 重启服务

```bash
# 重启所有服务
ssh root@101.43.42.250 'cd /opt/AIGE && docker compose restart'

# 只重启后端
ssh root@101.43.42.250 'cd /opt/AIGE && docker compose restart backend'
```

### 停止服务

```bash
ssh root@101.43.42.250 'cd /opt/AIGE && docker compose down'
```

### 清理资源

```bash
# 清理未使用的 Docker 资源
ssh root@101.43.42.250 'docker system prune -a'

# 清理旧备份（保留最近 10 个）
ssh root@101.43.42.250 'cd /opt/AIGE/backups && ls -t chat.db.* | tail -n +11 | xargs rm'
```

### 更新环境变量

```bash
# 编辑 .env 文件
ssh root@101.43.42.250 'nano /opt/AIGE/.env'

# 重新启动服务使配置生效
ssh root@101.43.42.250 'cd /opt/AIGE && docker compose restart'
```

---

## 🔐 安全建议

1. **永远不要提交 .env 文件到 Git**
   - .env 已在 .gitignore 中
   - 包含敏感的 API 密钥

2. **定期备份数据库**
   - 自动备份：deploy 脚本会自动备份
   - 手动备份：`cp data/chat.db backups/chat.db.manual.$(date +%Y%m%d)`

3. **使用 SSH 密钥而非密码**
   - 更安全
   - 更方便

4. **限制服务器访问**
   - 配置防火墙
   - 只开放必要的端口

5. **定期更新依赖**
   - 更新 Docker 镜像
   - 更新 npm 包
   - 更新 Go 模块

---

## 📞 支持

如遇到问题，请查看：
1. 本文档的「常见问题」部分
2. 项目日志：`docker compose logs`
3. GitHub Issues

---

**Happy Deploying! 🎉**
