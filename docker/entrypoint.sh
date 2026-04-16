#!/bin/sh
set -e

# 首次运行时（挂载的 html 目录为空），将镜像内置的前端初始化进去。
# 这样 volume 挂载后用户不需要手动 docker cp，后续可通过 API 热更新前端。
if [ -z "$(ls -A /usr/share/nginx/html 2>/dev/null)" ]; then
    echo "[entrypoint] html 目录为空，从镜像内置版本初始化..."
    cp -r /app/html-init/. /usr/share/nginx/html/
    echo "[entrypoint] 初始化完成"
fi

exec /usr/bin/supervisord -c /etc/supervisord.conf
