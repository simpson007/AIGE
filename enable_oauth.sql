-- 启用Linux.Do OAuth
INSERT OR REPLACE INTO system_configs (key, value) VALUES ('oauth_linux_do_enabled', 'true');

-- 配置OAuth参数（需要替换为你的实际值）
INSERT OR REPLACE INTO system_configs (key, value) VALUES ('oauth_linux_do_client_id', 'aP4oWNbAAczw5LhGQHyjcOZQ4DHxPUXm');
INSERT OR REPLACE INTO system_configs (key, value) VALUES ('oauth_linux_do_client_secret', 'd6FPKOMIm8uf8zKVs83bHcwqIOimBQzw');
INSERT OR REPLACE INTO system_configs (key, value) VALUES ('oauth_linux_do_redirect_url', 'http://localhost:8000/auth/callback/linux-do');

-- 可选：自定义OAuth端点（如果使用默认值可以不设置）
-- INSERT OR REPLACE INTO system_configs (key, value) VALUES ('oauth_linux_do_auth_url', 'https://connect.linux.do/oauth2/authorize');
-- INSERT OR REPLACE INTO system_configs (key, value) VALUES ('oauth_linux_do_token_url', 'https://connect.linux.do/oauth2/token');
-- INSERT OR REPLACE INTO system_configs (key, value) VALUES ('oauth_linux_do_user_info_url', 'https://connect.linux.do/api/user');