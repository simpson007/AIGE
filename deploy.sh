#!/bin/bash

# ========================================
# AIGE 项目一键部署脚本（本地端）
# ========================================
# 使用方法: ./deploy.sh [commit_message]
# 示例: ./deploy.sh "修复MOD加载问题"
# ========================================

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量
SERVER_USER="root"
SERVER_HOST="101.43.42.250"
SERVER_PATH="/root/AIGE"
GIT_BRANCH="main"

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}ℹ ${NC}$1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_step() {
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}▶ $1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

# 检查是否在项目根目录
if [ ! -f "docker-compose.yml" ]; then
    print_error "错误：请在项目根目录运行此脚本"
    exit 1
fi

# 获取提交消息
COMMIT_MSG="${1:-'部署更新'}"

print_step "1. 检查 Git 状态"
if ! git status &> /dev/null; then
    print_error "错误：当前目录不是 Git 仓库"
    print_info "初始化 Git 仓库: git init"
    exit 1
fi

# 显示当前分支
CURRENT_BRANCH=$(git branch --show-current)
print_info "当前分支: ${CURRENT_BRANCH}"

# 检查是否有未提交的更改
if [[ -n $(git status -s) ]]; then
    print_info "发现未提交的更改:"
    git status -s

    print_step "2. 添加文件到暂存区"
    git add .
    print_success "所有更改已添加到暂存区"

    print_step "3. 提交更改"
    git commit -m "$COMMIT_MSG"
    print_success "提交成功: $COMMIT_MSG"
else
    print_info "没有需要提交的更改"
fi

print_step "4. 推送到远程仓库"
# 检查是否配置了远程仓库
if ! git remote get-url origin &> /dev/null; then
    print_warning "未配置远程仓库"
    print_info "请先配置远程仓库:"
    echo "  git remote add origin git@github.com:simpson007/guzhenren.git"
    echo ""
    read -p "是否跳过推送步骤？(y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
    SKIP_PUSH=true
else
    print_info "推送到远程分支: ${CURRENT_BRANCH}"
    git push origin ${CURRENT_BRANCH}
    print_success "推送成功"
    SKIP_PUSH=false
fi

print_step "5. 连接到服务器并部署"
print_info "连接到: ${SERVER_USER}@${SERVER_HOST}"

# 生成服务器端执行的命令
SERVER_COMMANDS=$(cat <<'EOF'
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() { echo -e "${BLUE}ℹ ${NC}$1"; }
print_success() { echo -e "${GREEN}✓${NC} $1"; }
print_error() { echo -e "${RED}✗${NC} $1"; }
print_step() {
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}▶ $1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

cd SERVER_PATH_PLACEHOLDER

print_step "服务器端部署开始"

print_info "当前目录: $(pwd)"

# 检查目录是否存在
if [ ! -d ".git" ]; then
    print_error "错误：目录不是 Git 仓库"
    exit 1
fi

print_step "1. 备份数据库"
if [ -f "data/chat.db" ]; then
    BACKUP_FILE="data/chat.db.backup.$(date +%Y%m%d_%H%M%S)"
    cp data/chat.db "$BACKUP_FILE"
    print_success "数据库已备份到: $BACKUP_FILE"
else
    print_info "未找到数据库文件，跳过备份"
fi

print_step "2. 拉取最新代码"
print_info "切换到分支: BRANCH_PLACEHOLDER"
git fetch origin
git checkout BRANCH_PLACEHOLDER
git pull origin BRANCH_PLACEHOLDER
print_success "代码拉取成功"

print_step "3. 显示最新提交"
git log -1 --pretty=format:"%h - %an, %ar : %s"
echo ""

print_step "4. 停止现有容器"
docker-compose down
print_success "容器已停止"

print_step "5. 重新构建镜像"
print_info "清理旧镜像并重新构建..."
docker-compose build --no-cache
print_success "镜像构建完成"

print_step "6. 启动服务"
docker-compose up -d
print_success "服务已启动"

print_step "7. 等待服务启动..."
sleep 10

print_step "8. 检查容器状态"
docker-compose ps

print_step "9. 检查后端健康状态"
for i in {1..10}; do
    if curl -f -s http://localhost:8182/health > /dev/null; then
        print_success "后端服务健康检查通过"
        curl -s http://localhost:8182/health | jq '.'
        break
    else
        if [ $i -eq 10 ]; then
            print_error "后端服务启动失败"
            echo ""
            print_info "查看后端日志:"
            docker-compose logs --tail=50 backend
            exit 1
        fi
        print_info "等待后端启动... ($i/10)"
        sleep 3
    fi
done

print_step "10. 检查前端健康状态"
for i in {1..5}; do
    if curl -f -s http://localhost:3000 > /dev/null; then
        print_success "前端服务健康检查通过"
        break
    else
        if [ $i -eq 5 ]; then
            print_error "前端服务启动失败"
            echo ""
            print_info "查看前端日志:"
            docker-compose logs --tail=50 frontend
            exit 1
        fi
        print_info "等待前端启动... ($i/5)"
        sleep 2
    fi
done

print_step "11. 显示服务日志（最后 20 行）"
echo ""
print_info "【后端日志】"
docker-compose logs --tail=20 backend
echo ""
print_info "【前端日志】"
docker-compose logs --tail=20 frontend

echo ""
print_step "✅ 部署完成！"
print_success "应用已成功部署并运行"
print_info "访问地址: https://games.yushenjian.com"
echo ""
EOF
)

# 替换占位符
SERVER_COMMANDS="${SERVER_COMMANDS//SERVER_PATH_PLACEHOLDER/$SERVER_PATH}"
SERVER_COMMANDS="${SERVER_COMMANDS//BRANCH_PLACEHOLDER/$CURRENT_BRANCH}"

# 执行远程命令
if [ "$SKIP_PUSH" = false ]; then
    ssh -t ${SERVER_USER}@${SERVER_HOST} "bash -c '$SERVER_COMMANDS'"
else
    print_warning "跳过服务器部署（未推送到远程仓库）"
fi

echo ""
print_step "🎉 本地部署脚本执行完成"
print_success "项目已成功部署到服务器"
print_info "查看实时日志: ssh ${SERVER_USER}@${SERVER_HOST} 'cd ${SERVER_PATH} && docker-compose logs -f'"
echo ""
