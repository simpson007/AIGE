# 🚀 AIGE部署文件说明

本目录包含AIGE项目在腾讯云服务器上部署所需的所有配置文件和脚本。

## 📁 文件结构

```
deploy/
├── README.md              # 本文件
├── DEPLOYMENT.md          # 详细部署指南
├── deploy.sh              # 一键部署脚本
├── setup-ssl.sh           # SSL证书配置脚本
└── nginx/
    └── aige.conf          # Nginx反向代理配置
```

## 🎯 快速开始

### 方式一：自动化部署（推荐）

```bash
# 1. 上传项目到服务器
scp -r AIGE root@your-server-ip:/opt/

# 2. 连接到服务器
ssh root@your-server-ip

# 3. 配置环境变量
cd /opt/AIGE
cp .env.example .env
vim .env  # 填入你的API密钥

# 4. 运行部署脚本
bash deploy/deploy.sh
```

### 方式二：手动部署

参考 [DEPLOYMENT.md](./DEPLOYMENT.md) 中的详细步骤。

## 📋 部署前检查清单

在开始部署前，请确保：

- [ ] 服务器系统为CentOS/AlmaLinux（或其他RHEL兼容系统）
- [ ] 服务器有公网IP
- [ ] 已开放80、443端口（防火墙和安全组）
- [ ] 已准备好AI服务的API密钥（OpenAI/Anthropic/Google）
- [ ] 如需HTTPS，已准备域名并配置DNS解析

## 🔑 环境变量配置

在 `.env` 文件中必须配置以下项：

```bash
# JWT密钥（使用强随机字符串）
JWT_SECRET=your-very-strong-secret-key

# AI服务API密钥（至少配置一个）
OPENAI_API_KEY=sk-xxxxxxxxxxxx
ANTHROPIC_API_KEY=sk-ant-xxxxxxxx
GOOGLE_API_KEY=AIzaSyxxxxxxxxxx

# 运行模式
GIN_MODE=release
```

生成强随机JWT密钥：
```bash
openssl rand -base64 32
```

## 🔐 配置HTTPS（可选但推荐）

部署完成后，运行SSL配置脚本：

```bash
bash deploy/setup-ssl.sh
```

脚本会：
1. 安装certbot
2. 申请Let's Encrypt免费SSL证书
3. 配置Nginx支持HTTPS
4. 设置证书自动续期

## 📦 部署脚本说明

### deploy.sh

主要部署脚本，功能包括：
- 检查并安装Docker和Docker Compose
- 验证环境变量配置
- 创建必要的目录
- 构建Docker镜像
- 启动容器
- 配置防火墙
- 健康检查

### setup-ssl.sh

SSL证书配置脚本，功能包括：
- 安装certbot工具
- 申请Let's Encrypt证书
- 配置Nginx HTTPS
- 设置自动续期任务

## 🌐 Nginx配置说明

`nginx/aige.conf` 包含：
- HTTP到HTTPS重定向
- SSL/TLS安全配置
- 前端静态文件代理
- 后端API反向代理
- WebSocket连接支持
- 安全响应头配置

**使用前请替换配置中的域名**：
```bash
# 将 your-domain.com 替换为你的实际域名
sed -i 's/your-domain.com/example.com/g' nginx/aige.conf
```

## 🔧 常用命令

### 容器管理

```bash
# 查看容器状态
docker-compose ps

# 查看日志
docker-compose logs -f

# 重启服务
docker-compose restart

# 停止服务
docker-compose down

# 启动服务
docker-compose up -d
```

### 应用更新

```bash
# 拉取最新代码
git pull

# 重新构建
docker-compose build --no-cache

# 重启容器
docker-compose down && docker-compose up -d
```

### 数据备份

```bash
# 备份数据库
cp -r /opt/AIGE/data /opt/backups/aige_$(date +%Y%m%d)
```

## 🐛 故障排查

### 容器无法启动

```bash
# 查看详细日志
docker-compose logs backend
docker-compose logs frontend

# 检查端口占用
netstat -tulpn | grep -E '3000|8182'
```

### 无法访问应用

```bash
# 检查防火墙
firewall-cmd --list-all

# 检查容器健康状态
docker inspect aige-backend | grep Health
docker inspect aige-frontend | grep Health
```

### SSL证书问题

```bash
# 查看证书状态
certbot certificates

# 手动续期
certbot renew

# 测试续期
certbot renew --dry-run
```

## 📖 更多信息

- **完整部署指南**: [DEPLOYMENT.md](./DEPLOYMENT.md)
- **项目文档**: [CLAUDE.md](../CLAUDE.md)
- **Docker配置**: [docker-compose.yml](../docker-compose.yml)

## 🆘 需要帮助？

如果遇到问题：
1. 查看 [DEPLOYMENT.md](./DEPLOYMENT.md) 的故障排查章节
2. 检查容器日志: `docker-compose logs -f`
3. 提交GitHub Issue

---

**祝部署成功！** 🎉
