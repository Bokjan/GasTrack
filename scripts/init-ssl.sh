#!/bin/bash
# ============================================
# GasTrack SSL 初始化脚本
# ============================================
# 解决「首次部署无证书 → Nginx 无法启动 → Certbot 无法验证」的鸡蛋问题
#
# 使用方式:
#   chmod +x scripts/init-ssl.sh
#   DOMAIN=yourdomain.com EMAIL=you@email.com ./scripts/init-ssl.sh
# ============================================

set -euo pipefail

# ---------- 参数检查 ----------
DOMAIN="${DOMAIN:?请设置 DOMAIN 环境变量，例如: DOMAIN=gas.example.com}"
EMAIL="${EMAIL:?请设置 EMAIL 环境变量，用于 Lets Encrypt 通知}"
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.prod.yaml}"
ENV_FILE="${ENV_FILE:-.env.production}"

SSL_DIR="./ssl"
CERTBOT_WWW="./certbot/www"

echo "🔧 GasTrack SSL 初始化"
echo "   域名: ${DOMAIN}"
echo "   邮箱: ${EMAIL}"
echo ""

# ---------- 1. 创建目录 ----------
mkdir -p "${SSL_DIR}" "${CERTBOT_WWW}"

# ---------- 2. 生成临时自签名证书（让 Nginx 能先启动） ----------
if [ ! -f "${SSL_DIR}/fullchain.pem" ]; then
    echo "📝 生成临时自签名证书..."
    openssl req -x509 -nodes -newkey rsa:2048 \
        -days 1 \
        -keyout "${SSL_DIR}/privkey.pem" \
        -out "${SSL_DIR}/fullchain.pem" \
        -subj "/CN=${DOMAIN}" \
        2>/dev/null
    echo "   ✅ 临时证书已生成"
else
    echo "   ⏭️  证书文件已存在，跳过"
fi

# ---------- 3. 启动服务（Nginx 使用临时证书） ----------
echo ""
echo "🚀 启动服务..."
docker compose -f "${COMPOSE_FILE}" --env-file "${ENV_FILE}" up -d
echo "   等待 Nginx 启动..."
sleep 5

# ---------- 4. 验证 HTTP 可达 ----------
echo ""
echo "🔍 验证 HTTP 80 端口..."
if curl -sf -o /dev/null "http://${DOMAIN}/.well-known/acme-challenge/" 2>/dev/null || \
   curl -sf -o /dev/null "http://localhost/.well-known/acme-challenge/" 2>/dev/null; then
    echo "   ✅ HTTP 验证路径可达"
else
    echo "   ⚠️  无法访问验证路径，Certbot 可能会失败"
    echo "   请确保域名 DNS 已指向本服务器，且 80/443 端口已开放"
fi

# ---------- 5. 使用 Certbot 申请真实证书 ----------
echo ""
echo "📜 申请 Let's Encrypt 证书..."
docker run --rm \
    -v "${PWD}/${SSL_DIR}:/etc/letsencrypt/live/${DOMAIN}" \
    -v "${PWD}/${CERTBOT_WWW}:/var/www/certbot" \
    -v certbot-etc:/etc/letsencrypt \
    -v certbot-var:/var/lib/letsencrypt \
    certbot/certbot certonly \
        --webroot \
        --webroot-path=/var/www/certbot \
        --email "${EMAIL}" \
        --agree-tos \
        --no-eff-email \
        -d "${DOMAIN}" \
        --force-renewal

# ---------- 6. 复制证书到挂载目录 ----------
echo ""
echo "📋 复制证书..."
# certbot 输出在 /etc/letsencrypt/live/${DOMAIN}/
# 由于我们挂载了 certbot-etc volume，需要从中提取
docker run --rm \
    -v certbot-etc:/etc/letsencrypt:ro \
    -v "${PWD}/${SSL_DIR}:/output" \
    alpine sh -c "cp /etc/letsencrypt/live/${DOMAIN}/fullchain.pem /output/ && cp /etc/letsencrypt/live/${DOMAIN}/privkey.pem /output/"

echo "   ✅ 证书已复制到 ${SSL_DIR}/"

# ---------- 7. 重载 Nginx ----------
echo ""
echo "🔄 重载 Nginx..."
docker exec gastrack-nginx nginx -s reload
echo "   ✅ Nginx 已重载"

# ---------- 8. 验证 HTTPS ----------
echo ""
echo "🔍 验证 HTTPS..."
sleep 2
if curl -sf -o /dev/null "https://${DOMAIN}/api/v1/health"; then
    echo "   ✅ HTTPS 工作正常！"
else
    echo "   ⚠️  HTTPS 验证失败，请手动检查"
fi

echo ""
echo "============================================"
echo "🎉 SSL 初始化完成！"
echo ""
echo "站点地址: https://${DOMAIN}"
echo ""
echo "📌 别忘了设置证书自动续期（crontab -e）:"
echo "   0 3 * * * cd '${PWD}' && docker run --rm -v certbot-etc:/etc/letsencrypt -v certbot-var:/var/lib/letsencrypt -v '${PWD}/${CERTBOT_WWW}':/var/www/certbot certbot/certbot renew --quiet && docker exec gastrack-nginx nginx -s reload"
echo "============================================"
