# AIGE项目生产环境部署指南

本文档详细说明如何在腾讯云服务器（CentOS/AlmaLinux）上使用Docker部署AIGE项目。

## 📋 目录

- [系统要求](#系统要求)
- [部署准备](#部署准备)
- [快速部署](#快速部署)
- [详细步骤](#详细步骤)
- [配置域名和HTTPS](#配置域名和https)
- [运维管理](#运维管理)
- [故障排查](#故障排查)
- [安全建议](#安全建议)

## 🖥️ 系统要求

### 硬件要求
- **CPU**: 2核或以上
- **内存**: 4GB或以上（推荐8GB）
- **存储**: 20GB或以上可用空间
- **网络**: 公网IP，开放80、443、8182端口

### 软件要求
- **操作系统**: CentOS 7/8, AlmaLinux 8/9, 或其他RHEL兼容系统
- **Docker**: 20.10或更高版本
- **Docker Compose**: 1.29或更高版本
- **Nginx**: 1.20或更高版本（已安装）

## 📦 部署准备

### 1. 连接到服务器

```bash
ssh root@your-server-ip
```

### 2. 安装必要的工具

```bash
# 更新系统
yum update -y

# 安装基础工具
yum install -y git wget curl vim
```

### 3. 克隆项目代码

```bash
# 进入工作目录
cd /opt

# 克隆项目（或上传压缩包）
git clone https://github.com/your-repo/AIGE.git
# 或者使用scp上传
# scp -r /local/path/AIGE root@your-server-ip:/opt/

cd AIGE
```

### 4. 配置环境变量

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑环境变量文件
vim .env
```

**必须配置的项目**：
```bash
# JWT密钥（使用强随机字符串）
JWT_SECRET=$(openssl rand -base64 32)

# AI服务API密钥（至少配置一个）
OPENAI_API_KEY=sk-xxxxxxxxxxxxxxxxxxxx
ANTHROPIC_API_KEY=sk-ant-xxxxxxxxxxxxxxxx
GOOGLE_API_KEY=AIzaSyxxxxxxxxxxxxx

# 运行模式
GIN_MODE=release
```

## 🚀 快速部署

使用自动化部署脚本，一键完成部署：

```bash
cd /opt/AIGE
bash deploy/deploy.sh
```

脚本将自动完成：
- ✅ 检查并安装Docker和Docker Compose
- ✅ 创建必要的目录
- ✅ 构建Docker镜像
- ✅ 启动容器
- ✅ 配置防火墙

部署完成后，访问 `http://your-server-ip` 即可看到应用。

## 📝 详细步骤

如果你想手动部署或了解详细过程，请按以下步骤操作：

### 步骤1: 安装Docker

```bash
# 卸载旧版本
yum remove -y docker docker-client docker-client-latest docker-common \
    docker-latest docker-latest-logrotate docker-logrotate docker-engine

# 安装依赖
yum install -y yum-utils device-mapper-persistent-data lvm2

# 添加Docker仓库
yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo

# 安装Docker
yum install -y docker-ce docker-ce-cli containerd.io

# 启动Docker
systemctl start docker
systemctl enable docker

# 验证安装
docker --version
```

### 步骤2: 安装Docker Compose

```bash
# 下载最新版本
COMPOSE_VERSION=$(curl -s https://api.github.com/repos/docker/compose/releases/latest | grep 'tag_name' | cut -d\" -f4)
curl -L "https://github.com/docker/compose/releases/download/${COMPOSE_VERSION}/docker compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker compose

# 添加执行权限
chmod +x /usr/local/bin/docker compose

# 验证安装
docker compose --version
```

### 步骤3: 构建和启动容器

```bash
cd /opt/AIGE

# 构建镜像
docker compose build

# 启动容器（后台运行）
docker compose up -d

# 查看容器状态
docker compose ps

# 查看日志
docker compose logs -f
```

### 步骤4: 配置防火墙

```bash
# 开放HTTP和HTTPS端口
firewall-cmd --permanent --add-service=http
firewall-cmd --permanent --add-service=https
firewall-cmd --reload

# 验证规则
firewall-cmd --list-all
```

## 🌐 配置域名和HTTPS

### 前置条件

1. 已有域名并配置DNS解析到服务器IP
2. 域名已解析生效（使用 `ping your-domain.com` 验证）

### 使用自动化脚本

```bash
cd /opt/AIGE
bash deploy/setup-ssl.sh
```

脚本将提示你输入：
- 域名（例如：example.com）
- 邮箱地址（用于证书通知）

### 手动配置SSL

#### 1. 安装Certbot

```bash
yum install -y epel-release
yum install -y certbot python3-certbot-nginx
```

#### 2. 获取SSL证书

```bash
# 停止Nginx
systemctl stop nginx

# 获取证书
certbot certonly --standalone \
    --agree-tos \
    --email your-email@example.com \
    -d your-domain.com \
    -d www.your-domain.com

# 启动Nginx
systemctl start nginx
```

#### 3. 配置Nginx

```bash
# 复制配置文件
cp /opt/AIGE/deploy/nginx/aige.conf /etc/nginx/conf.d/

# 编辑配置文件，替换域名
vim /etc/nginx/conf.d/aige.conf
# 将所有 your-domain.com 替换为你的实际域名

# 测试配置
nginx -t

# 重新加载Nginx
systemctl reload nginx
```

#### 4. 设置证书自动续期

```bash
# 添加到crontab
echo "0 3 * * * certbot renew --quiet --post-hook 'systemctl reload nginx'" | crontab -

# 测试续期
certbot renew --dry-run
```

## 🔧 运维管理

### 容器管理

```bash
# 查看容器状态
docker compose ps

# 查看实时日志
docker compose logs -f

# 查看特定服务日志
docker compose logs -f backend
docker compose logs -f frontend

# 重启服务
docker compose restart

# 停止服务
docker compose stop

# 启动服务
docker compose start

# 完全停止并删除容器
docker compose down
```

### 更新应用

```bash
cd /opt/AIGE

# 拉取最新代码
git pull

# 重新构建镜像
docker compose build --no-cache

# 重启服务
docker compose down
docker compose up -d
```

### 数据备份

```bash
# 备份数据库
cd /opt/AIGE
cp -r data data_backup_$(date +%Y%m%d)

# 或创建自动备份脚本
cat > /usr/local/bin/backup-aige.sh <<'EOF'
#!/bin/bash
BACKUP_DIR="/opt/backups/aige"
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR
cp -r /opt/AIGE/data $BACKUP_DIR/data_$DATE
# 保留最近7天的备份
find $BACKUP_DIR -name "data_*" -mtime +7 -exec rm -rf {} \;
EOF

chmod +x /usr/local/bin/backup-aige.sh

# 添加定时任务（每天凌晨2点备份）
echo "0 2 * * * /usr/local/bin/backup-aige.sh" | crontab -
```

### 监控和日志

```bash
# 查看容器资源使用情况
docker stats

# 查看Docker系统信息
docker system info

# 查看磁盘使用
docker system df

# 清理未使用的资源
docker system prune -a
```

## 🔍 故障排查

### 容器无法启动

```bash
# 查看详细日志
docker compose logs backend
docker compose logs frontend

# 检查容器状态
docker compose ps

# 进入容器调试
docker exec -it aige-backend sh
docker exec -it aige-frontend sh
```

### 数据库连接问题

```bash
# 检查数据库文件权限
ls -la /opt/AIGE/data/

# 如果权限不对，修复权限
chown -R 1000:1000 /opt/AIGE/data/
```

### 端口冲突

```bash
# 检查端口占用
netstat -tulpn | grep -E '3000|8182'

# 修改docker-compose.yml中的端口映射
vim docker-compose.yml
```

### Nginx配置问题

```bash
# 测试Nginx配置
nginx -t

# 查看Nginx错误日志
tail -f /var/log/nginx/aige-error.log

# 查看Nginx访问日志
tail -f /var/log/nginx/aige-access.log
```

### SSL证书问题

```bash
# 查看证书信息
certbot certificates

# 手动续期证书
certbot renew

# 重新获取证书
certbot delete --cert-name your-domain.com
certbot certonly --standalone -d your-domain.com
```

## 🔒 安全建议

### 1. 服务器安全

```bash
# 更改SSH默认端口
vim /etc/ssh/sshd_config
# Port 22 改为 Port 2222
systemctl restart sshd

# 禁用root密码登录，使用SSH密钥
vim /etc/ssh/sshd_config
# PasswordAuthentication no
# PubkeyAuthentication yes

# 安装fail2ban防止暴力破解
yum install -y fail2ban
systemctl enable fail2ban
systemctl start fail2ban
```

### 2. 应用安全

- ✅ 使用强随机的JWT_SECRET
- ✅ 定期轮换API密钥
- ✅ 不要在公开仓库提交.env文件
- ✅ 限制数据库文件访问权限
- ✅ 启用HTTPS强制跳转
- ✅ 配置适当的CORS策略

### 3. Docker安全

```bash
# 限制Docker日志大小
vim /etc/docker/daemon.json
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}

systemctl restart docker
```

### 4. 定期更新

```bash
# 定期更新系统
yum update -y

# 更新Docker镜像
docker compose pull
docker compose up -d
```

## 📊 性能优化

### 1. Nginx优化

```nginx
# 在 /etc/nginx/nginx.conf 中添加
worker_processes auto;
worker_connections 2048;

# 启用gzip压缩
gzip on;
gzip_vary on;
gzip_min_length 1024;
gzip_types text/plain text/css application/json application/javascript;
```

### 2. Docker优化

```bash
# 限制容器资源
# 在 docker-compose.yml 中添加
services:
  backend:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
```

### 3. 数据库优化

```bash
# SQLite优化（在后端代码中配置）
# 启用WAL模式提高并发性能
# 定期执行VACUUM清理
```

## 🆘 获取帮助

遇到问题？
- 📖 查看项目文档: [CLAUDE.md](../CLAUDE.md)
- 🐛 提交Issue: GitHub Issues
- 💬 社区讨论: 项目讨论区

## 📜 更新日志

### v1.0.0 (2024-01-01)
- ✅ 初始部署文档
- ✅ Docker部署方案
- ✅ SSL证书配置
- ✅ 自动化部署脚本

---

**祝部署顺利！** 🎉
