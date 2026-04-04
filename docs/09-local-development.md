# GasTrack 本地开发环境调试指南

> **更新日期**: 2026-03-31

---

## 1. 环境要求

| 工具 | 版本要求 | 用途 |
|------|---------|------|
| Go | ≥ 1.23.0 | 后端编译运行 |
| Node.js | ≥ 18.0.0 | 前端构建 |
| pnpm | ≥ 8.0.0 | 前端包管理 |
| Docker | ≥ 24.0 | PostgreSQL 容器 |
| Docker Compose | ≥ 2.0 (V2) | 容器编排 |
| Git | 任意 | 版本控制 |

### 快速检查

```bash
go version          # go1.23.x
node -v             # v18.x 或更高
pnpm -v             # 8.x 或更高
docker --version    # Docker 24.x
docker compose version  # Docker Compose v2.x
```

---

## 2. 项目结构

```
GasTrack/
├── docs/                    # 设计文档
├── server/                  # Go 后端
│   ├── cmd/server/main.go  # 入口
│   ├── internal/           # 业务代码
│   ├── config.yaml         # 配置文件
│   └── go.mod / go.sum
├── packages/               # 前端 Monorepo
│   ├── shared/             # 共享类型/API/状态管理/i18n
│   └── web/                # React Web 应用
├── docker-compose.yaml     # PostgreSQL 容器
├── package.json            # 根 package.json
├── pnpm-workspace.yaml
├── .env.example            # 前端环境变量示例
├── LICENSE                 # MIT
└── README.md
```

---

## 3. 首次搭建步骤

### Step 1: 克隆项目

```bash
git clone <repo-url> GasTrack
cd GasTrack
```

### Step 2: 启动数据库

```bash
# 启动 PostgreSQL 容器（后台运行）
docker compose up -d

# 验证数据库已就绪
docker compose ps
# 应该看到 gastrack-postgres 状态为 healthy

# 如需查看日志
docker compose logs postgres
```

### Step 3: 安装前端依赖

```bash
# 在项目根目录执行
pnpm install
```

### Step 4: 安装后端依赖

```bash
cd server
go mod download
cd ..
```

### Step 5: 启动后端服务

```bash
cd server
go run cmd/server/main.go
```

启动成功后会看到：
```
{"level":"info","msg":"starting GasTrack server","host":"0.0.0.0","port":8098}
{"level":"info","msg":"database connected","host":"localhost","port":5432,"dbname":"gastrack"}
{"level":"info","msg":"HTTP server listening","addr":"0.0.0.0:8098"}
```

> 首次启动时 GORM AutoMigrate 会自动建表。

### Step 6: 启动前端开发服务器

```bash
# 另开一个终端，在项目根目录执行
pnpm dev
```

前端将在 `http://localhost:3000` 启动。

### Step 7: 访问应用

打开浏览器访问：**http://localhost:3000**

---

## 4. 服务端口一览

| 服务 | 端口 | URL |
|------|------|-----|
| 前端 (Vite Dev Server) | 3000 | http://localhost:3000 |
| 后端 (Go HTTP Server) | 8098 | http://localhost:8098 |
| PostgreSQL | 5432 | `localhost:5432` |
| Mailpit Web UI（可选） | 8025 | http://localhost:8025 |
| Mailpit SMTP（可选） | 1025 | - |

---

## 5. 日常开发流程

### 5.1 启动服务（每天开始工作时）

```bash
# Terminal 1: 确保数据库在运行
docker compose up -d

# Terminal 2: 启动后端
cd server && go run cmd/server/main.go

# Terminal 3: 启动前端
cd GasTrack && pnpm dev
```

### 5.2 后端热重载（推荐 Air）

默认的 `go run` 不支持热重载，可以使用 [Air](https://github.com/air-verse/air)：

```bash
# 安装 Air
go install github.com/air-verse/air@latest

# 在 server/ 目录运行
cd server
air
```

Air 会监听文件变化并自动重新编译运行。

### 5.3 前端热更新

Vite 内置 HMR（Hot Module Replacement），保存文件后浏览器自动更新，无需手动操作。

---

## 6. API 调试

### 6.1 API 代理配置

前端 Vite 开发服务器已配置 API 代理（`packages/web/vite.config.ts`）：

```typescript
server: {
  port: 3000,
  proxy: {
    '/api': {
      target: 'http://localhost:8098',
      changeOrigin: true,
    },
  },
}
```

前端代码中所有 `/api/v1/...` 请求会被自动代理到后端 `http://localhost:8098`。

### 6.2 使用 curl 测试 API

**健康检查**
```bash
curl http://localhost:8098/api/v1/health
```

**注册用户**
```bash
curl -X POST http://localhost:8098/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123","nickname":"测试用户"}'
```

**登录**
```bash
curl -X POST http://localhost:8098/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
```

**携带 Token 请求**
```bash
# 从登录响应中获取 access_token
TOKEN="eyJhbGciOiJIUzI1NiIs..."

curl http://localhost:8098/api/v1/users/me \
  -H "Authorization: Bearer $TOKEN"
```

### 6.3 使用 Postman / Insomnia

导入以下环境变量：

| 变量 | 值 |
|------|-----|
| `base_url` | `http://localhost:8098/api/v1` |
| `access_token` | （登录后获取） |

---

## 7. 数据库调试

### 7.1 连接数据库

```bash
# 方式一：通过 Docker 容器
docker exec -it gastrack-postgres psql -U gastrack -d gastrack

# 方式二：直接使用 psql
psql -h localhost -p 5432 -U gastrack -d gastrack
# 密码: gastrack
```

### 7.2 常用查询

```sql
-- 查看所有表
\dt

-- 查看用户
SELECT id, email, nickname, status FROM users;

-- 查看车辆
SELECT id, name, vehicle_type, fuel_type FROM vehicles;

-- 查看加油记录
SELECT id, vehicle_id, fuel_amount, total_cost, refuel_date
FROM fuel_records ORDER BY refuel_date DESC LIMIT 10;

-- 查看 Refresh Token
SELECT id, user_id, expires_at FROM refresh_tokens;
```

### 7.3 重置数据库

```bash
# 方式一：删除并重建容器（丢失所有数据）
docker compose down -v
docker compose up -d

# 方式二：只清空表数据
docker exec -it gastrack-postgres psql -U gastrack -d gastrack -c "
  TRUNCATE notifications, reminders, fuel_records, vehicles, invite_codes, refresh_tokens, users CASCADE;
"
```

### 7.4 查看 GORM SQL 日志

在 `server/config.yaml` 中将日志级别调为 `debug`：

```yaml
log:
  level: debug
  format: console  # console 格式更易读
```

重启后端后，GORM 会打印所有执行的 SQL 语句。

---

## 8. TypeScript 类型检查

```bash
# 检查 shared 包
pnpm --filter @gastrack/shared exec tsc --noEmit

# 检查 web 包（包含 shared 的类型）
pnpm --filter @gastrack/web exec tsc --noEmit

# 检查所有包
pnpm type-check
```

---

## 9. 后端编译检查

```bash
cd server

# 编译检查（不生成二进制）
go build ./...

# 运行测试
go test ./...

# 代码格式化
gofmt -w .

# 代码检查（需安装 golangci-lint）
golangci-lint run
```

---

## 10. 常见问题排查

### Q1: 后端启动报 `failed to connect database`

**原因**: PostgreSQL 未启动或连接参数错误。

**排查步骤**:
```bash
# 1. 检查 Docker 容器状态
docker compose ps

# 2. 如果没有运行，启动它
docker compose up -d

# 3. 检查端口是否被占用
lsof -i :5432

# 4. 手动测试连接
psql -h localhost -p 5432 -U gastrack -d gastrack
```

### Q2: 前端报 CORS 错误

**原因**: 前端与后端不在同一域名/端口。

**排查步骤**:
1. 确认后端 CORS 配置包含前端地址：
   - `server/config.yaml` → `cors_origins` 包含 `http://localhost:3000`
2. 确认 Vite 代理配置正确（`vite.config.ts` 中 `/api` 代理到 `8098`）
3. 如果直接访问后端（不经过代理），需要后端 CORS 允许前端 Origin

### Q3: 前端报 401 Unauthorized

**原因**: Token 过期或未正确携带。

**排查步骤**:
1. 打开浏览器 DevTools → Application → Local Storage，检查 `access_token` 是否存在
2. Network 面板中检查请求 Header 是否包含 `Authorization: Bearer xxx`
3. Token 可能已过期（15 分钟有效期），前端应自动使用 Refresh Token 续期

### Q4: 后端端口被占用 (8098)

```bash
# 查看占用进程
lsof -i :8098

# 杀掉占用进程
kill -9 <PID>

# 或修改端口（config.yaml + vite.config.ts + api/client.ts）
```

### Q5: `pnpm install` 报错

```bash
# 清除缓存重试
pnpm store prune
rm -rf node_modules packages/*/node_modules
pnpm install
```

### Q6: Go 模块下载慢

```bash
# 设置 Go 模块代理（中国大陆）
export GOPROXY=https://goproxy.cn,direct

# 然后重新下载
cd server && go mod download
```

---

## 11. 调试技巧

### 11.1 后端 Debug（VS Code）

在 `.vscode/launch.json` 中添加：

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug GasTrack Server",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/server/cmd/server/main.go",
      "cwd": "${workspaceFolder}/server"
    }
  ]
}
```

### 11.2 前端 Debug（VS Code）

在 `.vscode/launch.json` 中添加：

```json
{
  "name": "Debug Chrome",
  "type": "chrome",
  "request": "launch",
  "url": "http://localhost:3000",
  "webRoot": "${workspaceFolder}/packages/web/src"
}
```

### 11.3 查看后端请求日志

后端中间件会自动记录每个请求的详细信息：

```
{"level":"info","msg":"HTTP request","method":"GET","path":"/api/v1/users/me","status":200,"duration":"2.5ms"}
```

### 11.4 前端 Axios 拦截器日志

前端 API 客户端 (`packages/shared/src/api/client.ts`) 中的请求/响应拦截器可以帮助调试：
- 请求拦截：自动添加 `Authorization` Header
- 响应拦截：401 时自动刷新 Token 并重试

在浏览器 DevTools → Network 面板中可以查看所有 API 请求和响应。

---

## 12. 可选工具

### 12.1 Mailpit（邮件测试）

```bash
# 启动 Mailpit（使用 full profile）
docker compose --profile full up -d

# 访问 Web UI
open http://localhost:8025
```

### 12.2 数据库 GUI

推荐工具：
- **DBeaver** (免费，跨平台)
- **pgAdmin** (PostgreSQL 官方)
- **TablePlus** (macOS，付费但好用)

连接参数：
```
Host: localhost
Port: 5432
User: gastrack
Password: gastrack
Database: gastrack
```

### 12.3 API 测试工具

- **Postman**: 功能完善的 API 测试
- **Insomnia**: 轻量级替代
- **HTTPie**: 命令行工具 (`http GET localhost:8098/api/v1/health`)
