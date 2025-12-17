# 🚀 服务器更新指南

## ✅ 已完成的全局修改

1. **Docker 命令**: `docker-compose` → `docker compose`（新版命令）
2. **项目路径**: `/root/AIGE` → `/opt/AIGE`
3. **Git 仓库**: `simpson007/FateBound` → `simpson007/AIGE`

---

## 📦 在服务器上更新代码

### 步骤 1：停止现有服务

```bash
cd /opt/AIGE
docker compose down
```

### 步骤 2：更新 Git 配置

```bash
cd /opt/AIGE

# 更新远程仓库地址
git remote set-url origin git@github.com:simpson007/AIGE.git

# 验证
git remote -v
```

### 步骤 3：拉取最新代码

```bash
cd /opt/AIGE

# 备份 .env
cp .env /tmp/.env.backup

# 拉取最新代码
git fetch origin main
git reset --hard origin/main

# 恢复 .env
cp /tmp/.env.backup .env

# 查看最新更改
git log -1 --stat
```

### 步骤 4：验证关键文件

```bash
cd /opt/AIGE

# 检查健康检查端点是否存在
echo "=== 检查健康检查端点 ==="
grep -A 3 "/health" backend/routes/routes.go

# 应该显示：
# r.GET("/health", func(c *gin.Context) {
#     c.JSON(200, gin.H{
#         "status": "ok",
```

### 步骤 5：重新构建并启动

```bash
cd /opt/AIGE

# 重新构建（无缓存）
docker compose build --no-cache

# 启动服务
docker compose up -d

# 等待服务启动
sleep 20

# 查看容器状态
docker compose ps
```

### 步骤 6：验证部署

```bash
cd /opt/AIGE

# 后端健康检查
echo "=== 后端健康检查 ==="
curl http://localhost:8182/health

# 应该返回：{"status":"ok","service":"AIGE Backend"}

# MOD 列表检查
echo ""
echo "=== MOD 列表 ==="
curl http://localhost:8182/api/game/mods

# 应该返回 FateBound 游戏信息

# 查看后端日志
echo ""
echo "=== 后端日志（最后 30 行）==="
docker compose logs --tail=30 backend
```

---

## 🎯 一键更新脚本

复制粘贴整段命令到服务器：

```bash
cd /opt/AIGE

echo "=========================================="
echo "🚀 AIGE 项目更新部署"
echo "=========================================="
echo ""

# 1. 停止服务
echo "1️⃣ 停止现有服务..."
docker compose down
echo "✓ 服务已停止"
echo ""

# 2. 更新 Git 配置
echo "2️⃣ 更新 Git 配置..."
git remote set-url origin git@github.com:simpson007/AIGE.git
echo "✓ Git 远程仓库已更新"
echo ""

# 3. 备份并拉取代码
echo "3️⃣ 拉取最新代码..."
cp .env /tmp/.env.backup.$(date +%Y%m%d_%H%M%S)
git fetch origin main
git reset --hard origin/main
cp /tmp/.env.backup.* .env
echo "✓ 代码已更新"
git log -1 --oneline
echo ""

# 4. 检查健康检查端点
echo "4️⃣ 检查健康检查端点..."
if grep -q "/health" backend/routes/routes.go; then
    echo "✓ 健康检查端点存在"
else
    echo "✗ 健康检查端点不存在（需要手动添加）"
fi
echo ""

# 5. 重新构建
echo "5️⃣ 重新构建镜像（可能需要几分钟）..."
docker compose build --no-cache
echo "✓ 镜像构建完成"
echo ""

# 6. 启动服务
echo "6️⃣ 启动服务..."
docker compose up -d
echo "✓ 服务已启动"
echo ""

# 7. 等待启动
echo "7️⃣ 等待服务启动（20秒）..."
sleep 20
echo ""

# 8. 验证部署
echo "8️⃣ 验证部署..."
echo ""
echo "容器状态："
docker compose ps
echo ""

echo "后端健康检查："
curl -s http://localhost:8182/health | jq . 2>/dev/null || curl -s http://localhost:8182/health
echo ""
echo ""

echo "MOD 加载状态："
MOD_COUNT=$(curl -s http://localhost:8182/api/game/mods | grep -o "game_id" | wc -l | tr -d ' ')
if [ "$MOD_COUNT" -gt 0 ]; then
    echo "✓ 已加载 $MOD_COUNT 个 MOD"
else
    echo "✗ 未检测到 MOD"
fi
echo ""

echo "=========================================="
echo "✅ 更新部署完成"
echo "=========================================="
echo ""
echo "访问地址："
echo "  - 前端: http://games.yushenjian.com"
echo "  - 后端: http://101.43.42.250:8182"
echo ""
echo "查看日志："
echo "  docker compose logs -f"
echo ""
```

---

## 🔧 如果健康检查还是 404

如果 `/health` 端点还是 404，说明代码没有正确更新。手动添加：

```bash
cd /opt/AIGE

# 备份文件
cp backend/routes/routes.go backend/routes/routes.go.backup

# 编辑文件
nano backend/routes/routes.go
```

在 `func SetupRoutes(r *gin.Engine) {` 后面添加：

```go
func SetupRoutes(r *gin.Engine) {
	// 健康检查接口（不需要认证）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"service": "AIGE Backend",
		})
	})

	// 公开路由
	auth := r.Group("/api/auth")
	{
		// ... 保持其他代码不变
```

保存后重新构建：

```bash
cd /opt/AIGE
docker compose down
docker compose build --no-cache backend
docker compose up -d
sleep 20
curl http://localhost:8182/health
```

---

## 📞 需要帮助？

如果遇到问题，收集以下信息：

```bash
cd /opt/AIGE

# 生成诊断报告
cat > /tmp/aige-diagnostic.txt << 'EOF'
========================================
AIGE 诊断报告
========================================

1. Git 状态
EOF

git status >> /tmp/aige-diagnostic.txt
git remote -v >> /tmp/aige-diagnostic.txt
git log -1 --stat >> /tmp/aige-diagnostic.txt

echo "" >> /tmp/aige-diagnostic.txt
echo "2. 容器状态" >> /tmp/aige-diagnostic.txt
docker compose ps >> /tmp/aige-diagnostic.txt

echo "" >> /tmp/aige-diagnostic.txt
echo "3. 后端日志（最后 100 行）" >> /tmp/aige-diagnostic.txt
docker compose logs --tail=100 backend >> /tmp/aige-diagnostic.txt

echo "" >> /tmp/aige-diagnostic.txt
echo "4. 健康检查端点" >> /tmp/aige-diagnostic.txt
grep -A 5 "/health" backend/routes/routes.go >> /tmp/aige-diagnostic.txt

cat /tmp/aige-diagnostic.txt
```

将输出发送给技术支持。

---

**更新完成后，记得使用 `./deploy.sh` 进行后续的一键部署！** 🚀
