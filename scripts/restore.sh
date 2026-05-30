#!/bin/bash
# ============================================
# GasTrack 数据库 & 上传文件恢复脚本
# ============================================
# ⚠️  恢复会覆盖当前数据库与上传文件，操作前请确认！
#
# 使用方式:
#   chmod +x scripts/restore.sh
#   ./scripts/restore.sh <数据库备份文件.dump>
#   ./scripts/restore.sh gastrack_db_20260530_030000.dump
#
#   # 同时恢复上传文件（指定 uploads 归档）:
#   ./scripts/restore.sh gastrack_db_xxx.dump gastrack_uploads_xxx.tar.gz
#
#   # 跳过确认（用于自动化，慎用）:
#   FORCE=1 ./scripts/restore.sh gastrack_db_xxx.dump
#
# 可配置环境变量（也会自动从 .env.production 读取 DB_*）:
#   CONTAINER        Postgres 容器名（默认 gastrack-postgres）
#   DB_USER          数据库用户（默认 gastrack）
#   DB_NAME          数据库名（默认 gastrack）
#   UPLOADS_VOLUME   上传文件 volume 名（默认自动探测）
# ============================================

set -euo pipefail

# ---------- 参数 ----------
if [ $# -lt 1 ]; then
    echo "用法: $0 <数据库备份文件.dump> [uploads归档.tar.gz]" >&2
    exit 1
fi

DB_FILE="$1"
UPLOADS_FILE="${2:-}"

if [ ! -f "${DB_FILE}" ]; then
    echo "❌ 数据库备份文件不存在: ${DB_FILE}" >&2
    exit 1
fi

# ---------- 路径与配置 ----------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
ENV_FILE="${ENV_FILE:-${PROJECT_ROOT}/.env.production}"

if [ -f "${ENV_FILE}" ]; then
    while IFS='=' read -r key value; do
        case "${key}" in
            DB_USER|DB_NAME|DB_PASSWORD)
                value="${value%\"}"; value="${value#\"}"
                export "${key}=${value}"
                ;;
        esac
    done < <(grep -E '^(DB_USER|DB_NAME|DB_PASSWORD)=' "${ENV_FILE}" || true)
fi

CONTAINER="${CONTAINER:-gastrack-postgres}"
DB_USER="${DB_USER:-gastrack}"
DB_NAME="${DB_NAME:-gastrack}"

echo "♻️  GasTrack 恢复"
echo "   容器:       ${CONTAINER}"
echo "   数据库:     ${DB_NAME} (用户 ${DB_USER})"
echo "   数据库备份: ${DB_FILE}"
echo "   上传文件:   ${UPLOADS_FILE:-（不恢复）}"
echo ""

# ---------- 前置检查 ----------
if ! docker ps --format '{{.Names}}' | grep -qx "${CONTAINER}"; then
    echo "❌ 容器 ${CONTAINER} 未运行，请先启动数据库" >&2
    exit 1
fi

# ---------- 二次确认 ----------
if [ "${FORCE:-0}" != "1" ]; then
    echo "⚠️  此操作将【覆盖】数据库 ${DB_NAME} 的现有数据！"
    read -r -p "    确认继续？请输入 yes: " CONFIRM
    if [ "${CONFIRM}" != "yes" ]; then
        echo "已取消。"
        exit 0
    fi
fi

# ---------- 1. 恢复数据库 ----------
echo ""
echo "📥 [1/2] 恢复数据库..."
# --clean --if-exists: 先删除已存在对象再重建，保证幂等恢复
if docker exec -i "${CONTAINER}" pg_restore -U "${DB_USER}" -d "${DB_NAME}" --clean --if-exists --no-owner < "${DB_FILE}"; then
    echo "   ✅ 数据库恢复完成"
else
    echo "   ⚠️  pg_restore 返回非零（部分忽略的 DROP 警告通常可接受，请检查上方输出）"
fi

# ---------- 2. 恢复上传文件 ----------
if [ -n "${UPLOADS_FILE}" ]; then
    echo ""
    echo "📥 [2/2] 恢复上传文件..."
    if [ ! -f "${UPLOADS_FILE}" ]; then
        echo "   ❌ 上传文件归档不存在: ${UPLOADS_FILE}" >&2
        exit 1
    fi

    UPLOADS_VOLUME="${UPLOADS_VOLUME:-$(docker volume ls --format '{{.Name}}' | grep -E 'backend-uploads$' | head -n1 || true)}"
    if [ -z "${UPLOADS_VOLUME}" ]; then
        echo "   ❌ 未找到 backend-uploads volume（请设置 UPLOADS_VOLUME）" >&2
        exit 1
    fi

    UPLOADS_DIR="$(cd "$(dirname "${UPLOADS_FILE}")" && pwd)"
    UPLOADS_BASE="$(basename "${UPLOADS_FILE}")"
    # 先清空再解包，保证与备份一致
    docker run --rm \
        -v "${UPLOADS_VOLUME}:/data" \
        -v "${UPLOADS_DIR}:/backup:ro" \
        alpine sh -c "rm -rf /data/* && tar xzf '/backup/${UPLOADS_BASE}' -C /data"
    echo "   ✅ 上传文件恢复完成 → volume ${UPLOADS_VOLUME}"
fi

echo ""
echo "============================================"
echo "🎉 恢复完成！建议重启后端容器以刷新连接:"
echo "   docker restart gastrack-backend"
echo "============================================"
