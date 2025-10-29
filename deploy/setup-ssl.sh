#!/bin/bash

# ============================================
# AIGE项目SSL证书配置脚本
# 使用Let's Encrypt免费SSL证书
# ============================================

set -e

# 默认配置（可以通过环境变量覆盖）
DOMAIN="${DOMAIN:-games.yushenjian.com}"
EMAIL="${EMAIL:-993418465@qq.com}"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

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

# 获取域名
get_domain() {
    if [ -z "$DOMAIN" ]; then
        read -p "请输入你的域名（例如: example.com）: " DOMAIN
    fi

    if [ -z "$DOMAIN" ]; then
        log_error "域名不能为空"
        exit 1
    fi

    log_info "使用域名: $DOMAIN"
}

# 获取邮箱
get_email() {
    if [ -z "$EMAIL" ]; then
        read -p "请输入你的邮箱地址（用于证书通知）: " EMAIL
    fi

    if [ -z "$EMAIL" ]; then
        log_error "邮箱地址不能为空"
        exit 1
    fi

    log_info "使用邮箱: $EMAIL"
}

# 安装certbot
install_certbot() {
    log_info "检查certbot..."

    if ! command -v certbot &> /dev/null; then
        log_info "安装certbot..."
        yum install -y epel-release
        yum install -y certbot python3-certbot-nginx
    else
        log_info "certbot已安装"
    fi
}

# 配置Nginx
configure_nginx() {
    log_info "配置Nginx..."

    # 检查Nginx配置文件是否存在
    NGINX_CONF="/etc/nginx/conf.d/aige.conf"

    if [ ! -f "$NGINX_CONF" ]; then
        log_warn "Nginx配置文件不存在，从模板复制..."

        # 获取脚本所在目录
        SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
        PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

        # 复制Nginx配置文件
        if [ -f "$SCRIPT_DIR/nginx/aige.conf" ]; then
            cp "$SCRIPT_DIR/nginx/aige.conf" "$NGINX_CONF"
        elif [ -f "$PROJECT_ROOT/deploy/nginx/aige.conf" ]; then
            cp "$PROJECT_ROOT/deploy/nginx/aige.conf" "$NGINX_CONF"
        else
            log_error "找不到Nginx配置模板文件"
            exit 1
        fi

        # 替换域名
        sed -i "s/your-domain.com/$DOMAIN/g" "$NGINX_CONF"

        log_info "Nginx配置文件已创建"
    fi

    # 测试Nginx配置
    nginx -t

    # 重启Nginx
    systemctl restart nginx
}

# 获取SSL证书
obtain_certificate() {
    log_info "申请SSL证书..."

    # 停止Nginx以便certbot使用80端口
    systemctl stop nginx

    # 使用standalone模式获取证书
    certbot certonly --standalone \
        --non-interactive \
        --agree-tos \
        --email "$EMAIL" \
        -d "$DOMAIN"

    # 启动Nginx
    systemctl start nginx

    log_info "SSL证书获取成功"
}

# 更新Nginx SSL配置
update_nginx_ssl() {
    log_info "更新Nginx SSL配置..."

    NGINX_CONF="/etc/nginx/conf.d/aige.conf"

    # 更新证书路径
    sed -i "s|ssl_certificate .*|ssl_certificate /etc/letsencrypt/live/$DOMAIN/fullchain.pem;|g" "$NGINX_CONF"
    sed -i "s|ssl_certificate_key .*|ssl_certificate_key /etc/letsencrypt/live/$DOMAIN/privkey.pem;|g" "$NGINX_CONF"

    # 测试配置
    nginx -t

    # 重新加载Nginx
    systemctl reload nginx

    log_info "Nginx SSL配置更新完成"
}

# 设置自动续期
setup_auto_renewal() {
    log_info "配置证书自动续期..."

    # 添加续期任务到crontab
    CRON_CMD="0 3 * * * certbot renew --quiet --post-hook 'systemctl reload nginx'"

    # 检查crontab中是否已存在
    if ! crontab -l 2>/dev/null | grep -q "certbot renew"; then
        (crontab -l 2>/dev/null; echo "$CRON_CMD") | crontab -
        log_info "已添加自动续期任务（每天凌晨3点检查）"
    else
        log_info "自动续期任务已存在"
    fi

    # 测试续期
    certbot renew --dry-run

    log_info "证书自动续期配置完成"
}

# 显示完成信息
show_completion_info() {
    echo ""
    log_info "============================================"
    log_info "SSL证书配置完成！"
    log_info "============================================"
    echo ""
    log_info "你的站点现在可以通过HTTPS访问："
    log_info "  https://$DOMAIN"
    echo ""
    log_info "证书信息："
    log_info "  位置: /etc/letsencrypt/live/$DOMAIN/"
    log_info "  有效期: 90天"
    log_info "  自动续期: 已配置（每天凌晨3点检查）"
    echo ""
    log_info "手动续期命令："
    log_info "  certbot renew"
    echo ""
    log_info "查看证书信息："
    log_info "  certbot certificates"
    echo ""
}

# 主函数
main() {
    log_info "开始配置SSL证书..."

    # 检查root权限
    check_root

    # 获取域名和邮箱
    get_domain
    get_email

    # 安装certbot
    install_certbot

    # 配置Nginx
    configure_nginx

    # 获取SSL证书
    obtain_certificate

    # 更新Nginx SSL配置
    update_nginx_ssl

    # 设置自动续期
    setup_auto_renewal

    # 显示完成信息
    show_completion_info
}

# 执行主函数
main
