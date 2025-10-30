#!/bin/bash

# ========================================
# 在服务器上设置 Git 并更新项目
# ========================================

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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

# 配置变量（请根据实际情况修改）
SERVER_USER="root"
SERVER_HOST="101.43.42.250"
SERVER_PATH="/opt/AIGE"
GIT_REMOTE="git@github.com:simpson007/guzhenren.git"
GIT_BRANCH="main"

print_step "🚀 在服务器上设置 Git 仓库"

print_info "连接到服务器: ${SERVER_USER}@${SERVER_HOST}"

# 创建远程执行的脚本
ssh ${SERVER_USER}@${SERVER_HOST} bash << EOF
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() { echo -e "\${BLUE}ℹ \${NC}\$1"; }
print_success() { echo -e "\${GREEN}✓\${NC} \$1"; }
print_warning() { echo -e "\${YELLOW}⚠\${NC} \$1"; }
print_error() { echo -e "\${RED}✗\${NC} \$1"; }

print_step() {
    echo ""
    echo -e "\${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\${NC}"
    echo -e "\${BLUE}▶ \$1\${NC}"
    echo -e "\${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\${NC}"
}

# 进入项目目录
cd ${SERVER_PATH}

print_step "1. 检查项目状态"
print_info "当前目录: \$(pwd)"

# 检查是否是 Git 仓库
if [ -d ".git" ]; then
    print_success "已经是 Git 仓库"

    # 检查远程仓库配置
    if git remote get-url origin &> /dev/null; then
        CURRENT_REMOTE=\$(git remote get-url origin)
        print_info "当前远程仓库: \$CURRENT_REMOTE"

        if [ "\$CURRENT_REMOTE" != "${GIT_REMOTE}" ]; then
            print_warning "远程仓库地址不同，更新中..."
            git remote set-url origin ${GIT_REMOTE}
            print_success "远程仓库已更新"
        fi
    else
        print_warning "未配置远程仓库，添加中..."
        git remote add origin ${GIT_REMOTE}
        print_success "远程仓库已添加"
    fi
else
    print_warning "不是 Git 仓库，初始化中..."

    print_step "2. 备份重要文件"
    # 备份 .env 和数据库
    if [ -f ".env" ]; then
        cp .env .env.backup
        print_success ".env 已备份"
    fi

    if [ -f "data/chat.db" ]; then
        mkdir -p backups
        cp data/chat.db backups/chat.db.before_git_\$(date +%Y%m%d_%H%M%S)
        print_success "数据库已备份"
    fi

    print_step "3. 初始化 Git 仓库"
    git init
    git remote add origin ${GIT_REMOTE}
    print_success "Git 仓库已初始化"
fi

print_step "4. 配置 Git"
git config pull.rebase false
print_success "Git 配置完成"

print_step "5. 拉取远程代码"
print_info "从远程仓库拉取最新代码..."

# 方案A：如果服务器代码很旧或有冲突，强制覆盖
read -p "是否要强制覆盖本地代码？这会丢失本地未提交的修改！(y/n) " -n 1 -r
echo
if [[ \$REPLY =~ ^[Yy]\$ ]]; then
    print_warning "强制覆盖模式..."

    # 保存 .env 和数据
    [ -f ".env" ] && cp .env /tmp/.env.backup
    [ -d "data" ] && cp -r data /tmp/data.backup

    # 强制重置到远程版本
    git fetch origin ${GIT_BRANCH}
    git reset --hard origin/${GIT_BRANCH}

    # 恢复 .env 和数据
    [ -f "/tmp/.env.backup" ] && cp /tmp/.env.backup .env
    [ -d "/tmp/data.backup" ] && cp -r /tmp/data.backup/* data/ 2>/dev/null || true

    print_success "代码已强制更新到最新版本"
else
    # 方案B：尝试合并
    print_info "尝试合并模式..."

    git fetch origin ${GIT_BRANCH}

    # 如果是新仓库，第一次拉取
    if ! git rev-parse HEAD &> /dev/null; then
        git checkout -b ${GIT_BRANCH}
        git branch --set-upstream-to=origin/${GIT_BRANCH} ${GIT_BRANCH}
        git pull origin ${GIT_BRANCH} --allow-unrelated-histories
    else
        # 已有提交，尝试合并
        if git pull origin ${GIT_BRANCH}; then
            print_success "代码合并成功"
        else
            print_error "代码合并失败，存在冲突"
            print_info "请手动解决冲突，或重新运行脚本选择强制覆盖模式"
            exit 1
        fi
    fi
fi

print_step "6. 检查更新后的文件"
git log -1 --stat
echo ""

print_step "7. 检查关键文件"
print_info "检查部署脚本..."
if [ -f "deploy/server-deploy.sh" ]; then
    print_success "部署脚本存在"
    chmod +x deploy/server-deploy.sh
else
    print_warning "未找到部署脚本"
fi

if [ -f "docker-compose.yml" ]; then
    print_success "docker-compose.yml 存在"
else
    print_error "docker-compose.yml 不存在"
fi

print_step "✅ Git 设置完成"
print_success "服务器项目已更新到最新版本"
print_info "远程仓库: ${GIT_REMOTE}"
print_info "当前分支: \$(git branch --show-current)"
echo ""
EOF

print_step "✅ 服务器 Git 设置完成"
print_success "现在可以使用 ./deploy.sh 进行一键部署了"
echo ""
