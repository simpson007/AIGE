#!/bin/bash

# ========================================
# AIGE 项目服务器端部署脚本
# ========================================
# 使用方法: ./server-deploy.sh [branch]
# 示例: ./server-deploy.sh main
# ========================================

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置变量
PROJECT_PATH="/root/AIGE"
GIT_BRANCH="${1:-main}"
BACKUP_DIR="$PROJECT_PATH/backups"

# 打印函数
print_info() { echo -e "${BLUE}ℹ ${NC}$1"; }
print_success() { echo -e "${GREEN}✓${NC} $1"; }
print_warning() { echo -e "${YELLOW}⚠${NC} $1"; }
print_error() { echo -e "${RED}✗${NC} $1"; }
print_step() {
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}▶ $1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

print_step "🚀 AIGE 服务器端部署开始"
print_info "项目路径: $PROJECT_PATH"
print_info "目标分支: $GIT_BRANCH"

# 检查是否在正确的目录
cd $PROJECT_PATH || {
    print_error "项目目录不存在: $PROJECT_PATH"
    exit 1
}

# 检查是否是 Git 仓库
if [ ! -d ".git" ]; then
    print_error "当前目录不是 Git 仓库"
    exit 1
fi

print_step "1. 创建备份"
mkdir -p "$BACKUP_DIR"

# 备份数据库
if [ -f "data/chat.db" ]; then
    BACKUP_FILE="$BACKUP_DIR/chat.db.$(date +%Y%m%d_%H%M%S)"
    cp data/chat.db "$BACKUP_FILE"
    print_success "数据库已备份: $BACKUP_FILE"

    # 保留最近 10 个备份
    ls -t "$BACKUP_DIR"/chat.db.* | tail -n +11 | xargs -r rm
    print_info "已清理旧备份，保留最近 10 个"
else
    print_warning "未找到数据库文件，跳过备份"
fi

print_step "2. 拉取最新代码"
print_info "当前分支: $(git branch --show-current)"

# 保存本地修改（如果有）
if [[ -n $(git status -s) ]]; then
    print_warning "发现本地修改，正在保存..."
    git stash save "Auto stash before deploy $(date +%Y%m%d_%H%M%S)"
fi

# 拉取代码
git fetch origin
git checkout $GIT_BRANCH
git pull origin $GIT_BRANCH

print_success "代码更新成功"

print_step "3. 显示最新提交"
git log -1 --stat --pretty=format:"%C(yellow)%h%C(reset) - %C(cyan)%an%C(reset), %C(green)%ar%C(reset) : %s" --color=always
echo ""
echo ""

print_step "4. 检查环境配置"
if [ ! -f ".env" ]; then
    print_error ".env 文件不存在"
    print_info "请创建 .env 文件并配置必要的环境变量"
    exit 1
fi
print_success ".env 配置文件存在"

print_step "5. 停止现有服务"
if docker-compose ps | grep -q "Up"; then
    print_info "正在停止容器..."
    docker-compose down
    print_success "容器已停止"
else
    print_info "没有运行中的容器"
fi

print_step "6. 清理 Docker 资源"
print_info "清理未使用的镜像..."
docker image prune -f
print_success "清理完成"

print_step "7. 重新构建镜像"
print_info "开始构建（无缓存）..."
docker-compose build --no-cache
print_success "镜像构建完成"

print_step "8. 启动服务"
docker-compose up -d
print_success "服务已启动"

print_step "9. 等待服务启动"
sleep 15

print_step "10. 检查容器状态"
docker-compose ps
echo ""

# 检查容器是否都在运行
RUNNING_COUNT=$(docker-compose ps | grep "Up" | wc -l)
if [ "$RUNNING_COUNT" -lt 2 ]; then
    print_error "部分容器未正常启动"
    print_info "查看日志："
    docker-compose logs --tail=50
    exit 1
fi

print_step "11. 后端健康检查"
for i in {1..15}; do
    if curl -f -s http://localhost:8182/health > /dev/null 2>&1; then
        print_success "后端服务运行正常"
        HEALTH_INFO=$(curl -s http://localhost:8182/health 2>/dev/null)
        if command -v jq &> /dev/null; then
            echo "$HEALTH_INFO" | jq '.'
        else
            echo "$HEALTH_INFO"
        fi
        break
    else
        if [ $i -eq 15 ]; then
            print_error "后端服务健康检查失败"
            print_info "查看后端日志："
            docker-compose logs --tail=100 backend
            exit 1
        fi
        print_info "等待后端启动... ($i/15)"
        sleep 3
    fi
done

print_step "12. 前端健康检查"
for i in {1..10}; do
    if curl -f -s -o /dev/null http://localhost:3000 2>&1; then
        print_success "前端服务运行正常"
        break
    else
        if [ $i -eq 10 ]; then
            print_error "前端服务健康检查失败"
            print_info "查看前端日志："
            docker-compose logs --tail=50 frontend
            exit 1
        fi
        print_info "等待前端启动... ($i/10)"
        sleep 2
    fi
done

print_step "13. 验证 MOD 加载"
print_info "检查游戏 MOD..."
MODS_RESPONSE=$(curl -s http://localhost:8182/api/game/mods 2>/dev/null || echo "[]")
MOD_COUNT=$(echo "$MODS_RESPONSE" | grep -o "game_id" | wc -l)

if [ "$MOD_COUNT" -gt 0 ]; then
    print_success "已加载 $MOD_COUNT 个游戏 MOD"
    if command -v jq &> /dev/null; then
        echo "$MODS_RESPONSE" | jq -r '.[] | "  - \(.game_id): \(.name)"'
    fi
else
    print_error "未检测到游戏 MOD"
    print_info "查看后端日志："
    docker-compose logs --tail=100 backend | grep -i "mod"
fi

print_step "14. 服务日志预览"
echo ""
print_info "【后端日志（最后 30 行）】"
docker-compose logs --tail=30 backend
echo ""
print_info "【前端日志（最后 20 行）】"
docker-compose logs --tail=20 frontend

print_step "✅ 部署完成"
echo ""
print_success "🎉 AIGE 项目部署成功！"
echo ""
print_info "服务信息："
echo "  • 前端地址: http://localhost:3000"
echo "  • 后端地址: http://localhost:8182"
echo "  • 健康检查: http://localhost:8182/health"
echo ""
print_info "常用命令："
echo "  • 查看日志: docker-compose logs -f"
echo "  • 重启服务: docker-compose restart"
echo "  • 停止服务: docker-compose down"
echo "  • 查看状态: docker-compose ps"
echo ""
print_info "备份位置: $BACKUP_DIR"
echo ""
