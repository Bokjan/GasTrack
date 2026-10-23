#!/bin/bash
# ============================================
# GasTrack SSL 自动续期 + 同步脚本
# ============================================
# 设计目标:
#   1. 调用 certbot renew 尝试续期（证书临近过期时才会真正签发新证）
#   2. 无论续期是否触发，每次都把 certbot 卷里「当前有效」的证书
#      同步到 nginx 挂载的 ./ssl/ 目录并重载 nginx
#      → 避免「卷里已续期、nginx 仍在用旧证」的断层
#   3. 自动探测 live 目录：优先匹配 .env.production 中的 DOMAIN
#      （兼容 certbot 生成的 <domain>-0001 等后缀），找不到再取首个有效目录
#
# 使用方式:
#   ./scripts/renew-ssl.sh
#
# crontab 示例（每天 04:07 执行）:
#   7 4 * * * /path/to/GasTrack/scripts/renew-ssl.sh
#
# 可配置环境变量:
#   ENV_FILE   环境变量文件（默认 ./.env.production，用于读取 DOMAIN）
#   DOMAIN     证书域名（默认从 ENV_FILE 读取）
#   LOG_FILE   日志文件（默认 ./backups/ssl-renew.log）
# ============================================

set -uo pipefail

# ---------- 路径与配置 ----------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
ENV_FILE="${ENV_FILE:-${PROJECT_ROOT}/.env.production}"
LOG_FILE="${LOG_FILE:-${PROJECT_ROOT}/backups/ssl-renew.log}"

mkdir -p "$(dirname "${LOG_FILE}")"

log() { echo "$(date '+%Y-%m-%d %H:%M:%S') $*" >> "${LOG_FILE}"; }

cd "${PROJECT_ROOT}" || { log "ERROR: 无法进入 ${PROJECT_ROOT}"; exit 1; }

if [ -z "${DOMAIN:-}" ] && [ -f "${ENV_FILE}" ]; then
  DOMAIN="$(grep -E '^DOMAIN=' "${ENV_FILE}" | tail -n1 | cut -d= -f2- | tr -d "\"' \r")"
fi
DOMAIN="${DOMAIN:-}"

# ---------- 1. 尝试续期 ----------
docker run --rm \
  -v certbot-etc:/etc/letsencrypt \
  -v certbot-var:/var/lib/letsencrypt \
  -v "${PROJECT_ROOT}/certbot/www:/var/www/certbot" \
  certbot/certbot renew --quiet >> "${LOG_FILE}" 2>&1
renew_rc=$?
if [ "${renew_rc}" -ne 0 ]; then
  log "WARN: certbot renew 返回非零 (rc=${renew_rc})，将仍尝试同步已有证书"
fi

# ---------- 2. 探测当前有效的 live 目录 ----------
# 原 lineage 的 renewal 配置损坏时，certbot 会新建 <DOMAIN>-0001 …
# → 优先取「renewal 配置非空」且编号最大的 <DOMAIN>[-NNNN]，否则取首个有效目录
LIVE=$(docker run --rm -e DOMAIN="${DOMAIN}" -v certbot-etc:/etc/letsencrypt:ro alpine sh -c '
  cd /etc/letsencrypt/live 2>/dev/null || exit 0
  pick=""
  if [ -n "$DOMAIN" ]; then
    for n in $(ls -d "$DOMAIN" "$DOMAIN"-[0-9]* 2>/dev/null | sort); do
      [ -e "$n/fullchain.pem" ] && [ -s "../renewal/$n.conf" ] && pick="$n"
    done
  fi
  if [ -z "$pick" ]; then
    for n in *; do
      [ -e "$n/fullchain.pem" ] && pick="$n" && break
    done
  fi
  [ -n "$pick" ] && echo "/etc/letsencrypt/live/$pick/"')

if [ -z "${LIVE}" ]; then
  log "ERROR: 未在 certbot 卷中找到有效的 fullchain.pem (DOMAIN=${DOMAIN:-未设置})"
  exit 1
fi

# ---------- 3. 复制证书到 nginx 挂载目录 ----------
docker run --rm -v certbot-etc:/etc/letsencrypt:ro -v "${PROJECT_ROOT}/ssl:/output" alpine sh -c \
  "cp \"${LIVE}fullchain.pem\" /output/ && cp \"${LIVE}privkey.pem\" /output/" >> "${LOG_FILE}" 2>&1
cp_rc=$?

if [ "${cp_rc}" -ne 0 ]; then
  log "ERROR: 证书复制失败 (live=${LIVE})"
  exit 1
fi

# ---------- 4. 重载 nginx ----------
if docker exec gastrack-nginx nginx -s reload >> "${LOG_FILE}" 2>&1; then
  log "OK: 证书已同步并重载 (live=${LIVE}, renew_rc=${renew_rc})"
else
  log "ERROR: nginx 重载失败 (live=${LIVE})"
  exit 1
fi
