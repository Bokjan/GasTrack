# ============================================
# GasTrack Backend — 多阶段构建
# ============================================

# --- 构建阶段 ---
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# 先复制依赖文件，利用 Docker 缓存层
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并编译
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o gastrack-server ./cmd/server/main.go

# --- 运行阶段 ---
FROM alpine:3.20

# 安装运行时必要依赖
RUN apk --no-cache add ca-certificates tzdata

# 创建非 root 用户
RUN addgroup -S gastrack && adduser -S gastrack -G gastrack

WORKDIR /app

# 从构建阶段复制二进制
COPY --from=builder /build/gastrack-server .

# 创建必要目录
RUN mkdir -p /app/logs /app/uploads && \
    chown -R gastrack:gastrack /app

# 切换到非 root 用户
USER gastrack

EXPOSE 8098

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8098/api/v1/health || exit 1

ENTRYPOINT ["./gastrack-server"]
