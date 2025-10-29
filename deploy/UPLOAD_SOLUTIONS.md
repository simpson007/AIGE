# 🚀 项目上传到服务器的解决方案

遇到 `Permission denied` 错误？这里有多种解决方案！

## 🎯 服务器信息
- **IP地址**: 101.43.42.250
- **用户**: root
- **目标目录**: /opt

---

## 方案1：先上传到家目录再移动（最简单）

```bash
# 步骤1：打包项目（在本地执行）
cd /Users/yushenjian/Downloads
tar -czf AIGE.tar.gz \
    --exclude='node_modules' \
    --exclude='.git' \
    --exclude='*.db' \
    AIGE-main/

# 步骤2：上传到root家目录（这个通常有权限）
scp AIGE.tar.gz root@101.43.42.250:~/
# 输入密码或使用密钥

# 步骤3：SSH登录服务器
ssh root@101.43.42.250

# 步骤4：在服务器上移动文件
sudo mv ~/AIGE.tar.gz /opt/
cd /opt
tar -xzf AIGE.tar.gz
mv AIGE-main AIGE
cd AIGE

# 步骤5：开始部署
bash deploy/deploy.sh
```

---

## 方案2：使用SSH密钥（如果有密钥文件）

腾讯云通常会提供密钥文件（.pem格式）。

```bash
# 步骤1：设置密钥权限
chmod 400 /path/to/your-key.pem

# 步骤2：打包项目
cd /Users/yushenjian/Downloads
tar -czf AIGE.tar.gz --exclude='node_modules' --exclude='.git' AIGE-main/

# 步骤3：使用密钥上传
scp -i /path/to/your-key.pem AIGE.tar.gz root@101.43.42.250:~/

# 步骤4：使用密钥登录
ssh -i /path/to/your-key.pem root@101.43.42.250

# 步骤5：移动并部署
sudo mv ~/AIGE.tar.gz /opt/
cd /opt
tar -xzf AIGE.tar.gz
mv AIGE-main AIGE
cd AIGE
bash deploy/deploy.sh
```

---

## 方案3：使用SFTP上传（可视化工具）

如果不习惯命令行，可以使用图形化工具：

### 使用FileZilla（免费，跨平台）

1. 下载安装：https://filezilla-project.org/

2. 配置连接：
   - 主机：`sftp://101.43.42.250`
   - 用户名：`root`
   - 密码：`你的密码`
   - 端口：`22`

3. 连接后：
   - 左侧浏览本地文件
   - 右侧浏览服务器文件
   - 直接拖拽上传 `AIGE.tar.gz` 到 `/root` 目录

4. 然后SSH登录移动文件：
   ```bash
   ssh root@101.43.42.250
   sudo mv ~/AIGE.tar.gz /opt/
   cd /opt
   tar -xzf AIGE.tar.gz
   ```

---

## 方案4：直接在服务器上部署（推荐，无需上传）

如果你的项目在Git仓库中：

```bash
# 步骤1：SSH登录服务器
ssh root@101.43.42.250

# 步骤2：安装Git
yum install -y git

# 步骤3：克隆项目
cd /opt
git clone <你的Git仓库地址> AIGE
cd AIGE

# 步骤4：单独上传.env文件（从本地）
# 在另一个终端执行：
# scp /Users/yushenjian/Downloads/AIGE-main/.env root@101.43.42.250:/opt/AIGE/

# 或者在服务器上直接创建.env文件
cat > .env << 'EOF'
JWT_SECRET=kmJta+6rigatndQ9IDn2l5Wn9m2NVwh3+z8Zt53568w=
ANTHROPIC_API_KEY=sk-twq3MkhMBSyimbADinn7AReWdyBJYAerXd1C1fN3Qt5xOkRP
OPENAI_API_KEY=
GOOGLE_API_KEY=
GIN_MODE=release
DATABASE_PATH=/app/data/chat.db
EOF

chmod 600 .env

# 步骤5：开始部署
bash deploy/deploy.sh
```

---

## 方案5：使用自动上传脚本

我已经为你创建了自动化脚本：

```bash
cd /Users/yushenjian/Downloads/AIGE-main
bash deploy/upload-guide.sh
```

脚本会：
- 自动打包项目
- 提供多种上传方式选择
- 给出详细的操作指导

---

## 🔧 故障诊断

### 问题1：一直提示输入密码但总是失败

**原因**：可能需要使用SSH密钥认证

**解决**：
1. 登录腾讯云控制台
2. 找到"云服务器" → "SSH密钥"
3. 下载或创建密钥
4. 使用方案2的密钥方式上传

### 问题2：没有/opt目录的写权限

**解决**：先上传到 `~/`（家目录），再用sudo移动
```bash
scp AIGE.tar.gz root@101.43.42.250:~/
ssh root@101.43.42.250
sudo mv ~/AIGE.tar.gz /opt/
```

### 问题3：SSH连接超时

**原因**：防火墙或安全组未开放22端口

**解决**：
1. 登录腾讯云控制台
2. 进入"安全组"设置
3. 确保开放了22端口
4. 检查服务器firewall状态

### 问题4：root登录被禁用

**解决**：使用普通用户登录
```bash
# 假设有ubuntu用户
scp AIGE.tar.gz ubuntu@101.43.42.250:~/
ssh ubuntu@101.43.42.250
sudo mv ~/AIGE.tar.gz /opt/
```

---

## 📝 推荐流程（最简单）

基于你的情况，我推荐这个流程：

```bash
# 1. 打包项目
cd /Users/yushenjian/Downloads
tar -czf AIGE.tar.gz --exclude='node_modules' --exclude='.git' AIGE-main/

# 2. 上传到家目录（不需要/opt权限）
scp AIGE.tar.gz root@101.43.42.250:~/

# 3. 登录服务器
ssh root@101.43.42.250

# 4. 在服务器上操作
sudo mv ~/AIGE.tar.gz /opt/
cd /opt
tar -xzf AIGE.tar.gz
mv AIGE-main AIGE
cd AIGE
chmod 600 .env
bash deploy/deploy.sh
```

**如果第2步还是报错**，说明是SSH认证问题，请：
- 检查是否需要使用密钥文件（.pem）
- 检查密码是否正确
- 检查腾讯云控制台的SSH密钥配置

---

## 🆘 仍然无法解决？

请提供以下信息：

1. 执行 `ssh root@101.43.42.250` 能否成功登录？
2. 是否有SSH密钥文件（.pem格式）？
3. 完整的错误信息是什么？

根据这些信息，我可以给出更具体的解决方案。

---

## ✅ 验证上传成功

上传后，在服务器上检查：

```bash
# 登录服务器
ssh root@101.43.42.250

# 检查文件
ls -lh /opt/AIGE.tar.gz

# 解压并验证
cd /opt
tar -xzf AIGE.tar.gz
ls -la AIGE/

# 检查.env文件
cat AIGE/.env
```

如果能看到文件和配置，说明上传成功！

---

**提示**: 如果你能SSH登录服务器，最简单的方法是直接在服务器上使用Git克隆项目（方案4），然后手动创建.env文件。这样完全不需要上传！
