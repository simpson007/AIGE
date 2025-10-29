# 🔐 手动配置SSL证书指南

由于脚本路径问题，这里提供完整的手动配置步骤。

## 📋 配置信息
- **域名**: games.yushenjian.com
- **邮箱**: 993418465@qq.com
- **服务器IP**: 101.43.42.250

---

## 🚀 快速手动配置步骤

在服务器上依次执行以下命令：

### 步骤1：复制Nginx配置文件

```bash
# 确认你在项目目录
cd /opt/AIGE

# 复制Nginx配置到系统目录
sudo cp deploy/nginx/aige.conf /etc/nginx/conf.d/

# 验证文件已复制
ls -la /etc/nginx/conf.d/aige.conf
```

### 步骤2：测试Nginx配置

```bash
# 测试配置文件语法
sudo nginx -t

# 如果测试通过，重启Nginx
sudo systemctl restart nginx
sudo systemctl status nginx
```

### 步骤3：安装Certbot

```bash
# 安装EPEL仓库
sudo yum install -y epel-release

# 安装Certbot
sudo yum install -y certbot python3-certbot-nginx

# 验证安装
certbot --version
```

### 步骤4：申请SSL证书

**⚠️ 重要：确保DNS已生效！**

先验证DNS：
```bash
ping games.yushenjian.com
# 应该返回 101.43.42.250
```

DNS生效后，申请证书：

```bash
# 停止Nginx以便certbot使用80端口
sudo systemctl stop nginx

# 申请证书
sudo certbot certonly --standalone \
    --non-interactive \
    --agree-tos \
    --email 993418465@qq.com \
    -d games.yushenjian.com

# 启动Nginx
sudo systemctl start nginx
```

### 步骤5：验证证书

```bash
# 查看证书信息
sudo certbot certificates

# 应该显示：
# Certificate Name: games.yushenjian.com
# Domains: games.yushenjian.com
# Expiry Date: ...
# Certificate Path: /etc/letsencrypt/live/games.yushenjian.com/fullchain.pem
# Private Key Path: /etc/letsencrypt/live/games.yushenjian.com/privkey.pem
```

### 步骤6：测试Nginx配置并重载

```bash
# 测试配置
sudo nginx -t

# 重新加载Nginx
sudo systemctl reload nginx
```

### 步骤7：配置自动续期

```bash
# 添加自动续期任务
echo "0 3 * * * certbot renew --quiet --post-hook 'systemctl reload nginx'" | sudo crontab -

# 测试续期（不会真正续期）
sudo certbot renew --dry-run
```

### 步骤8：配置防火墙

```bash
# 开放HTTPS端口
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --reload

# 验证规则
sudo firewall-cmd --list-all
```

---

## ✅ 验证部署

### 1. 检查HTTP重定向

```bash
curl -I http://games.yushenjian.com
# 应该看到 301 重定向到 https://
```

### 2. 检查HTTPS访问

```bash
curl -I https://games.yushenjian.com
# 应该看到 200 OK
```

### 3. 在浏览器中访问

打开浏览器访问：**https://games.yushenjian.com**

应该看到：
- ✅ 绿色小锁图标
- ✅ 证书有效
- ✅ 应用正常加载

---

## 🔧 故障排查

### 问题1：certbot申请证书失败

**错误信息**：`Timeout during connect`

**解决方案**：
```bash
# 1. 确认DNS已生效
nslookup games.yushenjian.com

# 2. 确认80端口未被占用
sudo netstat -tulpn | grep :80

# 3. 确认防火墙开放了80端口
sudo firewall-cmd --list-all

# 4. 检查腾讯云安全组是否开放80端口
```

### 问题2：Nginx测试失败

**错误信息**：`nginx: [emerg] cannot load certificate`

**原因**：证书文件不存在

**解决方案**：
```bash
# 先注释掉SSL配置，只启用HTTP
sudo vim /etc/nginx/conf.d/aige.conf

# 找到 server { listen 443 ... } 块，整个注释掉
# 保存后重启Nginx
sudo nginx -t
sudo systemctl restart nginx

# 然后重新申请证书
sudo certbot certonly --standalone -d games.yushenjian.com --email 993418465@qq.com

# 证书获取后，取消注释SSL配置
sudo vim /etc/nginx/conf.d/aige.conf
# 取消注释 server { listen 443 ... } 块
sudo nginx -t
sudo systemctl reload nginx
```

### 问题3：DNS未生效

**解决方案**：
```bash
# 查看DNS记录
nslookup games.yushenjian.com

# 如果返回的不是 101.43.42.250，需要等待DNS传播
# 通常需要10-60分钟

# 临时解决：修改本地hosts文件测试
# 在本地电脑（不是服务器）执行：
# Mac/Linux: sudo vim /etc/hosts
# Windows: notepad C:\Windows\System32\drivers\etc\hosts
# 添加一行：
# 101.43.42.250 games.yushenjian.com
```

---

## 📝 完整的配置命令（一键复制）

如果要一次性执行所有命令：

```bash
# 复制Nginx配置
cd /opt/AIGE
sudo cp deploy/nginx/aige.conf /etc/nginx/conf.d/
sudo nginx -t
sudo systemctl restart nginx

# 安装Certbot
sudo yum install -y epel-release
sudo yum install -y certbot python3-certbot-nginx

# 申请SSL证书（确保DNS已生效）
sudo systemctl stop nginx
sudo certbot certonly --standalone \
    --non-interactive \
    --agree-tos \
    --email 993418465@qq.com \
    -d games.yushenjian.com
sudo systemctl start nginx

# 配置自动续期
echo "0 3 * * * certbot renew --quiet --post-hook 'systemctl reload nginx'" | sudo crontab -

# 配置防火墙
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --reload

# 验证
sudo certbot certificates
curl -I https://games.yushenjian.com
```

---

## 🎉 完成！

配置完成后：

- **HTTP访问**: http://games.yushenjian.com（自动重定向到HTTPS）
- **HTTPS访问**: https://games.yushenjian.com
- **证书有效期**: 90天（自动续期）
- **自动续期**: 每天凌晨3点检查

---

## 📞 如果仍有问题

请检查：
1. DNS是否已正确解析到 101.43.42.250
2. 腾讯云安全组是否开放了 80, 443 端口
3. 服务器防火墙状态：`sudo firewall-cmd --list-all`
4. Nginx错误日志：`sudo tail -f /var/log/nginx/aige-error.log`
5. Certbot日志：`sudo tail -f /var/log/letsencrypt/letsencrypt.log`
