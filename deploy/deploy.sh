#!/bin/bash

# ============================================
# AIGE项目部署脚本
# 适用于CentOS/AlmaLinux + Docker + Nginx
# ============================================

set -e  # 遇到错误立即退出

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查是否为root用户
check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "请使用root用户或sudo权限运行此脚本"
        exit 1
    fi
}

# 检查Docker是否安装
check_docker() {
    log_info "检查Docker环境..."
    if ! command -v docker &> /dev/null; then
        log_error "Docker未安装，正在安装Docker..."
        install_docker
    else
        log_info "Docker已安装: $(docker --version)"
    fi

    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose未安装，正在安装..."
        install_docker_compose
    else
        log_info "Docker Compose已安装: $(docker-compose --version)"
    fi
}

# 安装Docker
install_docker() {
    log_info "开始安装Docker..."

    # 卸载旧版本
    yum remove -y docker docker-client docker-client-latest docker-common \
        docker-latest docker-latest-logrotate docker-logrotate docker-engine

    # 安装依赖
    yum install -y yum-utils device-mapper-persistent-data lvm2

    # 添加Docker仓库
    yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo

    # 安装Docker
    yum install -y docker-ce docker-ce-cli containerd.io

    # 启动Docker服务
    systemctl start docker
    systemctl enable docker

    log_info "Docker安装完成"
}

# 安装Docker Compose
install_docker_compose() {
    log_info "开始安装Docker Compose..."

    # 下载最新版本的Docker Compose
    COMPOSE_VERSION=$(curl -s https://api.github.com/repos/docker/compose/releases/latest | grep 'tag_name' | cut -d\" -f4)
    curl -L "https://github.com/docker/compose/releases/download/${COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose

    # 添加执行权限
    chmod +x /usr/local/bin/docker-compose

    log_info "Docker Compose安装完成"
}

# 检查.env文件
check_env_file() {
    log_info "检查环境变量配置..."
    if [ ! -f ".env" ]; then
        log_warn ".env文件不存在，从模板创建..."
        cp .env.example .env
        log_error "请编辑 .env 文件填入正确的配置信息，然后重新运行此脚本"
        exit 1
    else
        log_info ".env文件已存在"
    fi
}

# 创建必要的目录
create_directories() {
    log_info "创建必要的目录..."
    mkdir -p data
    mkdir -p logs
    chmod 755 data logs
    log_info "目录创建完成"
}

# 构建并启动Docker容器
start_containers() {
    log_info "构建并启动Docker容器..."

    # 停止旧容器
    docker-compose down

    # 构建镜像
    log_info "构建Docker镜像（这可能需要几分钟）..."
    docker-compose build --no-cache

    # 启动容器
    log_info "启动容器..."
    docker-compose up -d

    # 等待容器启动
    log_info "等待容器启动..."
    sleep 10

    # 检查容器状态
    docker-compose ps

    log_info "容器启动完成"
}

# 检查容器健康状态
check_health() {
    log_info "检查容器健康状态..."

    max_attempts=30
    attempt=0

    while [ $attempt -lt $max_attempts ]; do
        backend_health=$(docker inspect --format='{{.State.Health.Status}}' aige-backend 2>/dev/null || echo "not_found")
        frontend_health=$(docker inspect --format='{{.State.Health.Status}}' aige-frontend 2>/dev/null || echo "not_found")

        if [ "$backend_health" = "healthy" ] && [ "$frontend_health" = "healthy" ]; then
            log_info "所有容器健康检查通过"
            return 0
        fi

        log_warn "等待容器健康检查... (${attempt}/${max_attempts})"
        sleep 5
        attempt=$((attempt + 1))
    done

    log_error "容器健康检查超时，请检查日志"
    docker-compose logs
    exit 1
}

# 配置防火墙
configure_firewall() {
    log_info "配置防火墙..."

    if command -v firewall-cmd &> /dev/null; then
        # 开放HTTP和HTTPS端口
        firewall-cmd --permanent --add-service=http
        firewall-cmd --permanent --add-service=https
        firewall-cmd --reload
        log_info "防火墙配置完成"
    else
        log_warn "firewalld未安装，跳过防火墙配置"
    fi
}

# 显示部署信息
show_deployment_info() {
    echo ""
    log_info "============================================"
    log_info "AIGE项目部署完成！"
    log_info "============================================"
    echo ""
    log_info "访问地址:"
    log_info "  - HTTP:  http://$(hostname -I | awk '{print $1}')"
    log_info "  - HTTPS: https://your-domain.com (配置域名后)"
    echo ""
    log_info "容器状态:"
    docker-compose ps
    echo ""
    log_info "查看日志:"
    log_info "  docker-compose logs -f"
    echo ""
    log_info "停止服务:"
    log_info "  docker-compose down"
    echo ""
    log_info "重启服务:"
    log_info "  docker-compose restart"
    echo ""
    log_info "接下来的步骤:"
    log_info "  1. 配置Nginx反向代理（参考deploy/nginx/aige.conf）"
    log_info "  2. 配置域名和SSL证书（使用certbot）"
    log_info "  3. 访问应用并完成初始化设置"
    echo ""
}

# 主函数
main() {
    log_info "开始部署AIGE项目..."

    # 检查root权限
    # check_root  # 如果不想强制root，可以注释此行

    # 检查Docker环境
    check_docker

    # 检查环境变量
    check_env_file

    # 创建目录
    create_directories

    # 启动容器
    start_containers

    # 检查健康状态
    check_health

    # 配置防火墙
    configure_firewall

    # 显示部署信息
    show_deployment_info
}

# 执行主函数
main
