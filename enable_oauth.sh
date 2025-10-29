#!/bin/bash

# 启用Linux.Do OAuth的脚本

# 设置数据库路径
DB_PATH="backend/chat.db"

# 检查数据库文件是否存在
if [ ! -f "$DB_PATH" ]; then
    echo "错误：数据库文件不存在：$DB_PATH"
    exit 1
fi

echo "正在启用Linux.Do OAuth..."

# 执行SQL命令启用OAuth
sqlite3 "$DB_PATH" << EOF
-- 启用Linux.Do OAuth
INSERT OR REPLACE INTO system_configs (key, value) VALUES ('oauth_linux_do_enabled', 'true');

-- 显示当前配置
SELECT key, value FROM system_configs WHERE key LIKE 'oauth_linux_do_%';
EOF

echo "Linux.Do OAuth已启用！"
echo ""
echo "注意：您还需要配置以下参数："
echo "1. Client ID"
echo "2. Client Secret"
echo "3. Redirect URL (默认: http://localhost:8000/auth/callback/linux-do)"
echo ""
echo "请通过管理界面或直接修改数据库来配置这些参数。"