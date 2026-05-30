#!/bin/bash
# ============================================
# GasTrack 数据库 & 上传文件备份脚本
# ============================================
# 备份内容:
#   1. PostgreSQL 数据库（pg_dump 自定义格式 -Fc，自带压缩）
#   2. 上传文件目录（uploads volume：头像 / 车辆照片 / 加油小票）
#
# 使用方式:
#   chmod +x scripts/backup.sh
#   ./scripts/backup.sh                       # 使用默认配置
#   BACKUP_DIR=/opt/gastrack/backups ./scripts/backup.sh
#   SKIP_UPLOADS=1 ./scripts/backup.sh        # 只备份数据库
#
# 可配置环境变量（也会自动从 .env.production 读取 DB_*）:
#   CONTAINER        Postgres 容器名（默认 gastrack-postgres）
#   DB_USER          数据库用户（默认 gastrack）
#   DB_NAME          数据库名（默认 gastrack）
#   BACKUP_DIR       备份输出目录（默认 ./backups）
#   RETENTION_DAYS   备份保留天数（默认 30，0=不清理）
#   UPLOADS_VOLUME   上传文件 volume 名（默认自动探测）
#   SKIP_UPLOADS     设为 1 则跳过 uploads 备份
# ============================================

set -euo pipefail

# ---------- 路径与配置 ----------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
ENV_FILE="${ENV_FILE:-${PROJECT_ROOT}/.env.production}"

# 从 .env.production 读取 DB_* 配置（若存在）
if [ -f "${ENV_FILE}" ]; then
    # 仅导出 DB_ 开头的变量，避免污染环境
    while IFS='=' read -r key value; do
        case "${key}" in
            DB_USER|DB_NAME|DB_PASSWORD)
                # 去掉可能的引号
                value="${value%\"}"; value="${value#\"}"
                export "${key}=${value}"
                ;;
        esac
    done < <(grep -E '^(DB_USER|DB_NAME|DB_PASSWORD)=' "${ENV_FILE}" || true)
fi

CONTAINER="${CONTAINER:-gastrack-postgres}"
DB_USER="${DB_USER:-gastrack}"
DB_NAME="${DB_NAME:-gastrack}"
BACKUP_DIR="${BACKUP_DIR:-${PROJECT_ROOT}/backups}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"

TIMESTAMP="$(date +%Y%m%d_%H%M%S)"
DB_FILE="${BACKUP_DIR}/gastrack_db_${TIMESTAMP}.dump"
UPLOADS_FILE="${BACKUP_DIR}/gastrack_uploads_${TIMESTAMP}.tar.gz"

echo "🗄️  GasTrack 备份"
echo "   容器:     ${CONTAINER}"
echo "   数据库:   ${DB_NAME} (用户 ${DB_USER})"
echo "   输出目录: ${BACKUP_DIR}"
echo ""

# ---------- 前置检查 ----------
if ! docker ps --format '{{.Names}}' | grep -qx "${CONTAINER}"; then
    echo "❌ 容器 ${CONTAINER} 未运行，请先启动数据库" >&2
    exit 1
fi

mkdir -p "${BACKUP_DIR}"

# ---------- 1. 备份数据库 ----------
echo "📦 [1/2] 导出数据库..."
if docker exec "${CONTAINER}" pg_dump -U "${DB_USER}" -d "${DB_NAME}" -Fc > "${DB_FILE}"; then
    echo "   ✅ 数据库已备份 → ${DB_FILE} ($(du -h "${DB_FILE}" | cut -f1))"
else
    echo "   ❌ 数据库备份失败" >&2
    rm -f "${DB_FILE}"
    exit 1
fi

# ---------- 2. 备份上传文件 ----------
if [ "${SKIP_UPLOADS:-0}" != "1" ]; then
    echo ""
    echo "📦 [2/2] 备份上传文件..."

    # 自动探测 uploads volume 名（compose 会加项目前缀，如 gastrack_backend-uploads）
    UPLOADS_VOLUME="${UPLOADS_VOLUME:-$(docker volume ls --format '{{.Name}}' | grep -E 'backend-uploads$' | head -n1 || true)}"

    if [ -z "${UPLOADS_VOLUME}" ]; then
        echo "   ⏭️  未找到 backend-uploads volume，跳过（如需指定请设置 UPLOADS_VOLUME）"
    else
        if docker run --rm \
            -v "${UPLOADS_VOLUME}:/data:ro" \
            -v "${BACKUP_DIR}:/backup" \
            alpine tar czf "/backup/$(basename "${UPLOADS_FILE}")" -C /data . ; then
            echo "   ✅ 上传文件已备份 → ${UPLOADS_FILE} ($(du -h "${UPLOADS_FILE}" | cut -f1))"
        else
            echo "   ⚠️  上传文件备份失败（继续）"
            rm -f "${UPLOADS_FILE}"
        fi
    fi
else
    echo ""
    echo "⏭️  已跳过上传文件备份 (SKIP_UPLOADS=1)"
fi

# ---------- 3. 清理过期备份 ----------
if [ "${RETENTION_DAYS}" -gt 0 ] 2>/dev/null; then
    echo ""
    echo "🧹 清理 ${RETENTION_DAYS} 天前的旧备份..."
    find "${BACKUP_DIR}" -maxdepth 1 -name 'gastrack_db_*.dump' -mtime +"${RETENTION_DAYS}" -print -delete || true
    find "${BACKUP_DIR}" -maxdepth 1 -name 'gastrack_uploads_*.tar.gz' -mtime +"${RETENTION_DAYS}" -print -delete || true
fi

echo ""
echo "============================================"
echo "🎉 备份完成！"
echo ""
echo "恢复命令:"
echo "   ./scripts/restore.sh ${DB_FILE}"
echo "============================================"
