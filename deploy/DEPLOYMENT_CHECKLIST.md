# 🎯 AIGE项目部署配置清单

本文档包含你的服务器专属配置信息和详细的部署步骤。

## ✅ 已配置信息汇总

### 服务器信息
- **服务器IP**: `101.43.42.250`
- **SSH端口**: `22`
- **操作系统**: CentOS/AlmaLinux

### 域名配置
- **主域名**: `yushenjian.com`
- **项目域名**: `games.yushenjian.com` （二级域名）
- **SSL邮箱**: `993418465@qq.com`

### 安全配置
- **JWT密钥**: `kmJta+6rigatndQ9IDn2l5Wn9m2NVwh3+z8Zt53568w=`
- **AI服务**: Anthropic Claude
- **API密钥**: 已配置（来自 https://q.quuvv.cn）

### 端口配置
- **前端容器端口**: `3000`
- **后端容器端口**: `8182`
- **Nginx HTTP**: `80`
- **Nginx HTTPS**: `443`

---

## 📋 部署前准备清单

在开始部署前，请确认以下事项：

- [ ] DNS已配置：`games.yushenjian.com` → `101.43.42.250`
- [ ] 腾讯云安全组已开放端口：22, 80, 443
- [ ] 服务器防火墙已开放端口：80, 443
- [ ] 已通过SSH连接到服务器
- [ ] 服务器有足够的磁盘空间（至少20GB）

### 验证DNS解析

在本地执行以下命令，确认DNS已生效：

```bash
ping games.yushenjian.com
# 应该返回 101.43.42.250
```

或使用：

```bash
nslookup games.yushenjian.com
# 应该显示 A 记录指向 101.43.42.250
```

---

## 🚀 部署步骤（详细版）

### 步骤1: 连接到服务器

```bash
ssh root@101.43.42.250
```

如果使用密钥登录：

```bash
ssh -i /path/to/your/key.pem root@101.43.42.250
```

### 步骤2: 上传项目文件

**方式A: 使用Git（推荐）**

```bash
cd /opt
git clone <你的仓库地址> AIGE
cd AIGE
```

**方式B: 使用SCP上传**

在本地执行：

```bash
# 打包项目（排除不必要的文件）
tar -czf AIGE.tar.gz --exclude='node_modules' --exclude='*.db' --exclude='.git' AIGE/

# 上传到服务器
scp AIGE.tar.gz root@101.43.42.250:/opt/

# 在服务器上解压
ssh root@101.43.42.250
cd /opt
tar -xzf AIGE.tar.gz
cd AIGE
```

### 步骤3: 验证配置文件

所有配置文件已经预配置好，检查以下文件：

```bash
# 检查环境变量文件
cat .env

# 检查Docker编排文件
cat docker-compose.yml

# 检查Nginx配置
cat deploy/nginx/aige.conf
```

**重要**: `.env` 文件已包含你的真实API密钥，请确保文件权限安全：

```bash
chmod 600 .env
```

### 步骤4: 安装Docker环境（如果未安装）

检查Docker是否已安装：

```bash
docker --version
docker compose --version
```

如果未安装，使用自动脚本安装：

```bash
bash deploy/deploy.sh
```

或者手动安装：

```bash
# 卸载旧版本
yum remove -y docker docker-client docker-common docker-latest

# 安装依赖
yum install -y yum-utils device-mapper-persistent-data lvm2

# 添加Docker仓库
yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo

# 安装Docker
yum install -y docker-ce docker-ce-cli containerd.io

# 启动Docker
systemctl start docker
systemctl enable docker

# 安装Docker Compose
COMPOSE_VERSION=$(curl -s https://api.github.com/repos/docker/compose/releases/latest | grep 'tag_name' | cut -d\" -f4)
curl -L "https://github.com/docker/compose/releases/download/${COMPOSE_VERSION}/docker compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker compose
chmod +x /usr/local/bin/docker compose
```

### 步骤5: 构建并启动容器

```bash
cd /opt/AIGE

# 构建Docker镜像（首次构建需要5-10分钟）
docker compose build

# 启动容器
docker compose up -d

# 查看容器状态
docker compose ps

# 查看启动日志
docker compose logs -f
```

等待所有容器状态变为 `healthy`。

### 步骤6: 配置防火墙

```bash
# 开放HTTP和HTTPS端口
firewall-cmd --permanent --add-service=http
firewall-cmd --permanent --add-service=https
firewall-cmd --reload

# 验证规则
firewall-cmd --list-all
```

### 步骤7: 验证服务运行

```bash
# 测试后端API
curl http://localhost:8182/health

# 测试前端
curl http://localhost:3000
```

如果返回正常响应，说明服务已启动成功。

### 步骤8: 配置Nginx反向代理

```bash
# 复制Nginx配置文件
cp /opt/AIGE/deploy/nginx/aige.conf /etc/nginx/conf.d/

# 测试Nginx配置
nginx -t

# 如果测试通过，重启Nginx
systemctl restart nginx
systemctl enable nginx
```

### 步骤9: 申请SSL证书

**确保DNS已解析到服务器IP后再执行此步骤！**

```bash
cd /opt/AIGE

# 运行SSL配置脚本（已预设域名和邮箱）
bash deploy/setup-ssl.sh
```

脚本会自动：
- 安装certbot工具
- 申请Let's Encrypt证书
- 配置Nginx HTTPS
- 设置证书自动续期

如果遇到问题，可以手动指定域名和邮箱：

```bash
DOMAIN=games.yushenjian.com EMAIL=993418465@qq.com bash deploy/setup-ssl.sh
```

### 步骤10: 验证部署

**HTTP访问（临时）**:
```bash
curl -I http://101.43.42.250
```

**HTTPS访问（推荐）**:
```bash
curl -I https://games.yushenjian.com
```

在浏览器中访问：
- HTTP: http://101.43.42.250（会自动重定向到HTTPS）
- HTTPS: https://games.yushenjian.com

---

## 🔧 常用管理命令

### 容器管理

```bash
# 查看所有容器状态
docker compose ps

# 查看实时日志
docker compose logs -f

# 查看特定服务日志
docker compose logs -f backend
docker compose logs -f frontend

# 重启所有服务
docker compose restart

# 重启特定服务
docker compose restart backend

# 停止所有服务
docker compose stop

# 启动所有服务
docker compose start

# 完全停止并删除容器
docker compose down

# 停止并删除所有数据（谨慎使用）
docker compose down -v
```

### 应用更新

```bash
cd /opt/AIGE

# 拉取最新代码
git pull

# 重新构建镜像
docker compose build --no-cache

# 重启容器
docker compose down
docker compose up -d

# 查看更新后的状态
docker compose ps
docker compose logs -f
```

### 数据备份

```bash
# 备份数据库文件
cd /opt/AIGE
tar -czf backup_$(date +%Y%m%d_%H%M%S).tar.gz data/

# 将备份文件下载到本地
scp root@101.43.42.250:/opt/AIGE/backup_*.tar.gz ./
```

### 系统监控

```bash
# 查看容器资源使用
docker stats

# 查看磁盘使用
df -h
docker system df

# 查看系统资源
top
htop  # 需要安装: yum install -y htop
```

---

## 🔍 故障排查

### 问题1: DNS未生效

**症状**: 无法通过域名访问，ping域名失败

**解决方案**:
```bash
# 检查DNS解析
nslookup games.yushenjian.com

# 如果未生效，等待DNS传播（通常需要10-60分钟）
# 临时可以先通过IP访问: http://101.43.42.250
```

### 问题2: 容器无法启动

**症状**: `docker compose ps` 显示容器Exit状态

**解决方案**:
```bash
# 查看详细日志
docker compose logs backend
docker compose logs frontend

# 检查端口占用
netstat -tulpn | grep -E '3000|8182'

# 重新构建
docker compose down
docker compose build --no-cache
docker compose up -d
```

### 问题3: SSL证书申请失败

**症状**: certbot报错，无法获取证书

**解决方案**:
```bash
# 确认DNS已解析
ping games.yushenjian.com

# 确认80端口未被占用
netstat -tulpn | grep :80

# 检查防火墙
firewall-cmd --list-all

# 手动申请证书
systemctl stop nginx
certbot certonly --standalone -d games.yushenjian.com --email 993418465@qq.com
systemctl start nginx
```

### 问题4: 无法访问应用

**症状**: 浏览器无法打开页面

**解决方案**:
```bash
# 1. 检查容器状态
docker compose ps

# 2. 检查Nginx状态
systemctl status nginx
nginx -t

# 3. 检查防火墙
firewall-cmd --list-all

# 4. 检查腾讯云安全组
# 登录腾讯云控制台，确认安全组已开放80、443端口

# 5. 查看Nginx日志
tail -f /var/log/nginx/aige-error.log
tail -f /var/log/nginx/aige-access.log
```

### 问题5: API请求失败

**症状**: 前端加载但API调用失败

**解决方案**:
```bash
# 检查后端日志
docker compose logs -f backend

# 检查API密钥配置
cat .env | grep API_KEY

# 进入容器调试
docker exec -it aige-backend sh
```

---

## 🔒 安全加固建议

部署完成后，建议执行以下安全加固措施：

### 1. 更改SSH端口

```bash
vim /etc/ssh/sshd_config
# 将 Port 22 改为其他端口，如 2222

systemctl restart sshd

# 更新防火墙规则
firewall-cmd --permanent --add-port=2222/tcp
firewall-cmd --permanent --remove-service=ssh
firewall-cmd --reload
```

### 2. 禁用root密码登录

```bash
vim /etc/ssh/sshd_config
# 设置以下选项
# PasswordAuthentication no
# PubkeyAuthentication yes

systemctl restart sshd
```

### 3. 安装fail2ban防护

```bash
yum install -y fail2ban
systemctl enable fail2ban
systemctl start fail2ban
```

### 4. 设置定期备份

```bash
# 创建备份脚本
cat > /usr/local/bin/backup-aige.sh <<'EOF'
#!/bin/bash
BACKUP_DIR="/opt/backups/aige"
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR
cd /opt/AIGE
tar -czf $BACKUP_DIR/backup_$DATE.tar.gz data/
# 保留最近7天的备份
find $BACKUP_DIR -name "backup_*" -mtime +7 -exec rm -f {} \;
EOF

chmod +x /usr/local/bin/backup-aige.sh

# 添加定时任务（每天凌晨2点备份）
echo "0 2 * * * /usr/local/bin/backup-aige.sh" | crontab -
```

### 5. 限制API访问频率

在Nginx配置中添加速率限制（可选）：

```nginx
# 在 http 块中添加
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;

# 在 location /api/ 块中添加
limit_req zone=api_limit burst=20 nodelay;
```

---

## 📊 性能监控

### 设置监控脚本

```bash
# 创建监控脚本
cat > /usr/local/bin/monitor-aige.sh <<'EOF'
#!/bin/bash
echo "=== AIGE服务监控 ==="
echo "时间: $(date)"
echo ""
echo "容器状态:"
docker compose -f /opt/AIGE/docker-compose.yml ps
echo ""
echo "资源使用:"
docker stats --no-stream
echo ""
echo "磁盘使用:"
df -h | grep -E 'Filesystem|/opt'
EOF

chmod +x /usr/local/bin/monitor-aige.sh

# 运行监控
/usr/local/bin/monitor-aige.sh
```

---

## 📞 技术支持

### 日志位置

- **应用日志**: `docker compose logs`
- **Nginx日志**: `/var/log/nginx/aige-*.log`
- **系统日志**: `/var/log/messages`

### 配置文件位置

- **环境变量**: `/opt/AIGE/.env`
- **Docker配置**: `/opt/AIGE/docker-compose.yml`
- **Nginx配置**: `/etc/nginx/conf.d/aige.conf`
- **SSL证书**: `/etc/letsencrypt/live/games.yushenjian.com/`

### 快速命令参考

```bash
# 进入项目目录
cd /opt/AIGE

# 查看服务状态
docker compose ps

# 查看日志
docker compose logs -f

# 重启服务
docker compose restart

# 更新应用
git pull && docker compose build && docker compose up -d

# 备份数据
tar -czf backup_$(date +%Y%m%d).tar.gz data/

# 查看SSL证书
certbot certificates

# 续期SSL证书
certbot renew
```

---

## ✅ 部署完成检查清单

部署完成后，请确认以下所有项目：

- [ ] 容器状态全部为 `Up (healthy)`
- [ ] HTTP自动重定向到HTTPS
- [ ] 可以通过 https://games.yushenjian.com 访问应用
- [ ] SSL证书有效期正常（90天）
- [ ] 自动续期任务已配置
- [ ] 防火墙规则正确
- [ ] 数据库文件可以正常读写
- [ ] API调用正常响应
- [ ] WebSocket连接正常
- [ ] 备份脚本已配置

---

## 🎉 恭喜！

你的AIGE项目已成功部署到生产环境！

**访问地址**: https://games.yushenjian.com

如有任何问题，请参考：
- 故障排查章节
- 详细部署文档: [DEPLOYMENT.md](./DEPLOYMENT.md)
- 项目文档: [CLAUDE.md](../CLAUDE.md)

---

**部署日期**: $(date +%Y-%m-%d)
**配置人员**: yushenjian
**服务器**: 101.43.42.250 (腾讯云)
