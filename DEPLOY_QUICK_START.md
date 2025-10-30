# 🚀 AIGE 部署速查表

## 初次设置（只需执行一次）

```bash
# 1. 配置 Git 远程仓库
git remote add origin https://github.com/yourusername/AIGE.git

# 2. 在服务器上克隆项目
ssh root@101.43.42.250
cd /root
git clone https://github.com/yourusername/AIGE.git
cd AIGE
cp .env.example .env
nano .env  # 配置 API 密钥

# 3. 配置 SSH 免密登录（推荐）
ssh-copy-id root@101.43.42.250

# 4. 编辑部署脚本配置
nano deploy.sh
# 修改: SERVER_USER, SERVER_HOST, SERVER_PATH
```

---

## 📦 一键部署

```bash
# 基本用法
./deploy.sh "提交信息"

# 示例
./deploy.sh "修复 MOD 加载问题"
./deploy.sh "添加新功能"
./deploy.sh "更新配置"
```

---

## 🔍 常用命令

### 查看状态
```bash
# 查看容器状态
ssh root@101.43.42.250 'cd /opt/AIGE && docker compose ps'

# 查看日志（实时）
ssh root@101.43.42.250 'cd /opt/AIGE && docker compose logs -f'

# 只看后端日志
ssh root@101.43.42.250 'cd /opt/AIGE && docker compose logs -f backend'

# 查看最近 100 行日志
ssh root@101.43.42.250 'cd /opt/AIGE && docker compose logs --tail=100'
```

### 重启服务
```bash
# 重启所有服务
ssh root@101.43.42.250 'cd /opt/AIGE && docker compose restart'

# 只重启后端
ssh root@101.43.42.250 'cd /opt/AIGE && docker compose restart backend'

# 只重启前端
ssh root@101.43.42.250 'cd /opt/AIGE && docker compose restart frontend'
```

### 停止/启动服务
```bash
# 停止
ssh root@101.43.42.250 'cd /opt/AIGE && docker compose down'

# 启动
ssh root@101.43.42.250 'cd /opt/AIGE && docker compose up -d'

# 重新构建并启动
ssh root@101.43.42.250 'cd /opt/AIGE && docker compose up -d --build'
```

---

## 🔧 健康检查

```bash
# 后端健康检查
curl http://101.43.42.250:8182/health

# 检查 MOD 列表
curl http://101.43.42.250:8182/api/game/mods

# 前端检查
curl http://101.43.42.250:3000
```

---

## 💾 数据备份与恢复

### 备份
```bash
# 自动备份（部署脚本会自动执行）
./deploy.sh

# 手动备份
ssh root@101.43.42.250 'cp /opt/AIGE/data/chat.db /opt/AIGE/backups/chat.db.$(date +%Y%m%d_%H%M%S)'
```

### 恢复
```bash
# 1. 查看备份
ssh root@101.43.42.250 'ls -lh /opt/AIGE/backups/'

# 2. 停止服务
ssh root@101.43.42.250 'cd /opt/AIGE && docker compose down'

# 3. 恢复备份（替换 YYYYMMDD_HHMMSS 为实际时间）
ssh root@101.43.42.250 'cp /opt/AIGE/backups/chat.db.YYYYMMDD_HHMMSS /opt/AIGE/data/chat.db'

# 4. 启动服务
ssh root@101.43.42.250 'cd /opt/AIGE && docker compose up -d'
```

---

## ⏪ 版本回滚

```bash
# 1. 查看提交历史
ssh root@101.43.42.250 'cd /opt/AIGE && git log --oneline -10'

# 2. 回滚到指定版本
ssh root@101.43.42.250 'cd /opt/AIGE && git reset --hard <commit-hash>'

# 3. 重新部署
ssh root@101.43.42.250 'cd /opt/AIGE && ./deploy/server-deploy.sh main'
```

---

## 🐛 问题排查

### MOD 加载失败
```bash
# 检查 MOD 目录
ssh root@101.43.42.250 'ls -la /opt/AIGE/mods'

# 检查容器内 MOD
ssh root@101.43.42.250 'docker exec aige-backend ls -la /app/mods'

# 查看后端日志中的 MOD 信息
ssh root@101.43.42.250 'cd /opt/AIGE && docker compose logs backend | grep -i mod'
```

### 端口被占用
```bash
# 查看端口占用
ssh root@101.43.42.250 'netstat -tulpn | grep -E "8182|3000"'

# 停止占用端口的进程
ssh root@101.43.42.250 'fuser -k 8182/tcp'
ssh root@101.43.42.250 'fuser -k 3000/tcp'
```

### Docker 问题
```bash
# 清理 Docker 资源
ssh root@101.43.42.250 'docker system prune -a'

# 重新构建（无缓存）
ssh root@101.43.42.250 'cd /opt/AIGE && docker compose build --no-cache'

# 查看容器详细信息
ssh root@101.43.42.250 'docker inspect aige-backend'
```

### 数据库锁定
```bash
# 重启后端服务
ssh root@101.43.42.250 'cd /opt/AIGE && docker compose restart backend'

# 如果还是不行，停止所有服务
ssh root@101.43.42.250 'cd /opt/AIGE && docker compose down && docker compose up -d'
```

---

## 📊 监控命令

```bash
# 查看资源使用情况
ssh root@101.43.42.250 'docker stats'

# 查看磁盘使用
ssh root@101.43.42.250 'df -h'

# 查看内存使用
ssh root@101.43.42.250 'free -h'

# 查看 Docker 磁盘使用
ssh root@101.43.42.250 'docker system df'
```

---

## 🔒 安全维护

```bash
# 更新环境变量
ssh root@101.43.42.250 'nano /opt/AIGE/.env'
ssh root@101.43.42.250 'cd /opt/AIGE && docker compose restart'

# 清理旧备份（保留最近 10 个）
ssh root@101.43.42.250 'cd /opt/AIGE/backups && ls -t chat.db.* | tail -n +11 | xargs rm'

# 更新系统
ssh root@101.43.42.250 'apt update && apt upgrade -y'
```

---

## 📁 文件结构

```
AIGE/
├── deploy.sh                    # 本地一键部署脚本
├── deploy/
│   └── server-deploy.sh         # 服务器端部署脚本
├── backend/                     # 后端代码
├── frontend/                    # 前端代码
├── mods/                        # 游戏 MOD
├── data/                        # 数据库文件
├── backups/                     # 备份文件
├── docker-compose.yml           # Docker 配置
├── .env                         # 环境变量（不提交）
├── DEPLOYMENT.md                # 详细部署文档
└── DEPLOY_QUICK_START.md        # 本速查表
```

---

## 🆘 快速救援

### 服务完全无法访问
```bash
ssh root@101.43.42.250
cd /opt/AIGE
docker compose down
docker compose up -d
docker compose logs -f
```

### 紧急回滚
```bash
ssh root@101.43.42.250
cd /opt/AIGE
git reset --hard HEAD~1
docker compose up -d --build
```

### 完全重新部署
```bash
ssh root@101.43.42.250
cd /opt/AIGE
docker compose down
docker system prune -a -f
git pull origin main
docker compose build --no-cache
docker compose up -d
```

---

## 📞 获取帮助

- 详细文档：`DEPLOYMENT.md`
- 项目配置：`CLAUDE.md`
- 问题反馈：GitHub Issues

---

**记得定期备份数据！** 💾
