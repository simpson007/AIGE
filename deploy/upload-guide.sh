#!/bin/bash

# AIGE项目上传指南脚本
# 帮助用户将项目上传到服务器

set -e

# 配置信息
SERVER_IP="101.43.42.250"
SERVER_USER="root"
TARGET_DIR="/opt"

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}AIGE项目上传助手${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 方案选择
echo "请选择上传方式："
echo "1. 使用SCP上传（需要SSH密码或密钥）"
echo "2. 直接在服务器上通过Git克隆（推荐）"
echo "3. 手动上传到/root目录再移动"
echo ""
read -p "请输入选项 (1/2/3): " choice

case $choice in
    1)
        echo ""
        echo -e "${YELLOW}方式1: 使用SCP上传${NC}"
        echo ""

        # 检查是否有密钥文件
        read -p "是否使用SSH密钥文件？(y/n): " use_key

        if [ "$use_key" = "y" ]; then
            read -p "请输入密钥文件路径: " key_path

            echo ""
            echo "正在打包项目..."
            cd "$(dirname "$0")/.."
            tar -czf AIGE.tar.gz \
                --exclude='node_modules' \
                --exclude='.git' \
                --exclude='*.db' \
                --exclude='data' \
                .

            echo "正在上传到服务器..."
            scp -i "$key_path" AIGE.tar.gz ${SERVER_USER}@${SERVER_IP}:~/

            echo ""
            echo -e "${GREEN}上传完成！${NC}"
            echo ""
            echo "接下来请SSH登录到服务器："
            echo "  ssh -i $key_path ${SERVER_USER}@${SERVER_IP}"
            echo ""
            echo "然后执行："
            echo "  sudo mv ~/AIGE.tar.gz ${TARGET_DIR}/"
            echo "  cd ${TARGET_DIR}"
            echo "  tar -xzf AIGE.tar.gz"
            echo "  mv AIGE AIGE-main 或重命名为AIGE"
            echo "  cd AIGE"
            echo "  bash deploy/deploy.sh"
        else
            echo ""
            echo "正在打包项目..."
            cd "$(dirname "$0")/.."
            tar -czf AIGE.tar.gz \
                --exclude='node_modules' \
                --exclude='.git' \
                --exclude='*.db' \
                --exclude='data' \
                .

            echo "正在上传到服务器（请输入SSH密码）..."
            scp AIGE.tar.gz ${SERVER_USER}@${SERVER_IP}:~/

            echo ""
            echo -e "${GREEN}上传完成！${NC}"
            echo ""
            echo "接下来请SSH登录到服务器："
            echo "  ssh ${SERVER_USER}@${SERVER_IP}"
            echo ""
            echo "然后执行："
            echo "  sudo mv ~/AIGE.tar.gz ${TARGET_DIR}/"
            echo "  cd ${TARGET_DIR}"
            echo "  tar -xzf AIGE.tar.gz"
            echo "  cd AIGE"
            echo "  bash deploy/deploy.sh"
        fi
        ;;

    2)
        echo ""
        echo -e "${YELLOW}方式2: 通过Git克隆（推荐）${NC}"
        echo ""
        echo "步骤如下："
        echo ""
        echo "1. 首先将项目推送到Git仓库（GitHub/Gitee等）"
        echo ""
        echo "2. SSH登录到服务器："
        echo "   ssh ${SERVER_USER}@${SERVER_IP}"
        echo ""
        echo "3. 在服务器上安装Git（如果未安装）："
        echo "   yum install -y git"
        echo ""
        echo "4. 克隆项目："
        echo "   cd ${TARGET_DIR}"
        echo "   git clone <你的仓库地址> AIGE"
        echo ""
        echo "5. 复制本地的.env文件到服务器："
        echo "   从本地上传.env文件："
        echo "   scp .env ${SERVER_USER}@${SERVER_IP}:${TARGET_DIR}/AIGE/"
        echo ""
        echo "6. 开始部署："
        echo "   cd ${TARGET_DIR}/AIGE"
        echo "   bash deploy/deploy.sh"
        echo ""
        echo -e "${GREEN}注意：.env文件包含敏感信息，不要提交到Git仓库！${NC}"
        ;;

    3)
        echo ""
        echo -e "${YELLOW}方式3: 手动上传${NC}"
        echo ""
        echo "如果无法直接上传到/opt/，可以先上传到/root目录："
        echo ""
        echo "1. 打包项目："
        cd "$(dirname "$0")/.."
        tar -czf AIGE.tar.gz \
            --exclude='node_modules' \
            --exclude='.git' \
            --exclude='*.db' \
            --exclude='data' \
            .
        echo "   ✓ 已打包为 AIGE.tar.gz"
        echo ""
        echo "2. 上传到服务器家目录（请输入密码）："
        echo "   scp AIGE.tar.gz ${SERVER_USER}@${SERVER_IP}:~/"
        echo ""
        echo "3. SSH登录服务器："
        echo "   ssh ${SERVER_USER}@${SERVER_IP}"
        echo ""
        echo "4. 在服务器上执行："
        echo "   sudo mv ~/AIGE.tar.gz ${TARGET_DIR}/"
        echo "   cd ${TARGET_DIR}"
        echo "   tar -xzf AIGE.tar.gz"
        echo "   cd AIGE"
        echo "   bash deploy/deploy.sh"
        echo ""
        echo "现在执行上传命令："
        scp AIGE.tar.gz ${SERVER_USER}@${SERVER_IP}:~/

        echo ""
        echo -e "${GREEN}上传完成！请SSH登录服务器继续操作。${NC}"
        ;;

    *)
        echo ""
        echo -e "${RED}无效选项${NC}"
        exit 1
        ;;
esac

echo ""
echo -e "${GREEN}========================================${NC}"
