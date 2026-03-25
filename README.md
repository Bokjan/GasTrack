# ⛽ GasTrack

**油耗/电耗记录与分析系统** — 记录每次加油/充电，追踪能耗趋势，辅助用车决策。

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## ✨ 功能特性

- 🚗 **多车辆管理** — 汽车、摩托车，燃油 & 电动车全覆盖
- ⛽ **加油/充电记录** — 日期、站点、量、单价、里程一站式录入
- 📊 **油耗/电耗统计** — 折线趋势图、费用汇总、多维分析
- 🌍 **多语言** — 中文简体 / English / 日本語
- 💱 **多币种 & 多计量体系** — CNY/USD/EUR/JPY/GBP/KRW · L/100km / km/L / MPG / kWh/100km
- 🔒 **JWT 认证** — Access Token + Refresh Token，安全可靠
- 📱 **响应式设计** — PC & 移动端自适应

## 🏗️ 技术栈

| 层级 | 技术 |
|------|------|
| **前端** | React 18 · TypeScript · Vite 5 · Ant Design 5 · ECharts · Zustand · i18next |
| **后端** | Go 1.22 · net/http 标准库 · GORM · JWT · Viper · Zap |
| **数据库** | PostgreSQL 16 |
| **包管理** | pnpm workspace (Monorepo) |
| **基础设施** | Docker Compose |

## 📁 项目结构

```
GasTrack/
├── packages/
│   ├── shared/            # 共享层：类型、常量、API、i18n、状态管理
│   └── web/               # Web 前端 (React + Vite)
├── server/                # Go 后端
│   ├── cmd/server/        # 入口 main.go
│   ├── internal/
│   │   ├── config/        # 配置加载
│   │   ├── database/      # 数据库连接 & 迁移
│   │   ├── handler/       # HTTP Handler
│   │   ├── middleware/     # 中间件（认证、CORS、限流、日志）
│   │   ├── model/         # GORM 数据模型
│   │   ├── dto/           # 请求/响应 DTO
│   │   ├── repository/    # 数据访问层
│   │   ├── service/       # 业务逻辑层
│   │   └── router/        # 路由注册
│   └── config.yaml        # 服务端配置
├── docs/                  # 项目文档（需求、架构、API、进度等）
├── docker-compose.yaml    # 开发环境依赖（PostgreSQL + Mailpit）
└── pnpm-workspace.yaml
```

## 🚀 快速开始

### 前置条件

- **Node.js** >= 18
- **pnpm** >= 8
- **Go** >= 1.22
- **Docker** & **Docker Compose**

### 1. 克隆仓库

```bash
git clone https://github.com/bokjan/GasTrack.git
cd GasTrack
```

### 2. 启动数据库

```bash
docker compose up -d
```

这将启动 PostgreSQL 16（端口 5432），默认凭据：

| 参数 | 值 |
|------|------|
| Host | localhost |
| Port | 5432 |
| User | gastrack |
| Password | gastrack |
| Database | gastrack |

### 3. 启动后端

```bash
cd server
go run ./cmd/server
```

后端默认监听 `http://localhost:8098`，首次启动会自动执行数据库迁移。

### 4. 启动前端

```bash
# 回到项目根目录
cd ..
pnpm install
pnpm dev
```

前端默认运行在 `http://localhost:5173`。

## ⚙️ 配置说明

后端配置文件位于 `server/config.yaml`：

```yaml
server:
  host: 0.0.0.0
  port: 8098

database:
  host: localhost
  port: 5432
  user: gastrack
  password: gastrack
  dbname: gastrack

jwt:
  secret: change-me-in-production-use-a-random-32-byte-string
  access_expiration: 15m
  refresh_expiration: 168h
```

可通过环境变量 `GASTRACK_CONFIG` 指定自定义配置文件路径。

## 📖 API 概览

| 模块 | 端点 | 说明 |
|------|------|------|
| 认证 | `POST /api/v1/auth/register` | 邮箱注册 |
| | `POST /api/v1/auth/login` | 登录 |
| | `POST /api/v1/auth/refresh` | 刷新 Token |
| 车辆 | `GET /api/v1/vehicles` | 车辆列表 |
| | `POST /api/v1/vehicles` | 添加车辆 |
| | `PUT /api/v1/vehicles/{id}` | 更新车辆 |
| | `DELETE /api/v1/vehicles/{id}` | 删除车辆 |
| 记录 | `GET /api/v1/vehicles/{vid}/records` | 加油/充电记录列表 |
| | `POST /api/v1/vehicles/{vid}/records` | 添加记录 |
| | `PUT /api/v1/records/{id}` | 更新记录 |
| | `DELETE /api/v1/records/{id}` | 删除记录 |
| 统计 | `GET /api/v1/stats/overview` | 统计总览 |

详细 API 文档见 [`docs/06-api-reference.md`](docs/06-api-reference.md)。

## 📚 文档

| 文档 | 说明 |
|------|------|
| [01-requirements.md](docs/01-requirements.md) | 需求分析 |
| [02-architecture.md](docs/02-architecture.md) | 系统架构 |
| [03-database.md](docs/03-database.md) | 数据库设计 |
| [04-tech-stack.md](docs/04-tech-stack.md) | 技术选型 |
| [05-roadmap.md](docs/05-roadmap.md) | 开发路线图 |
| [06-api-reference.md](docs/06-api-reference.md) | API 参考 |
| [07-database-setup.md](docs/07-database-setup.md) | 数据库部署 |
| [08-progress.md](docs/08-progress.md) | 开发进度 |
| [09-local-development.md](docs/09-local-development.md) | 本地开发指南 |

## 📜 License

[MIT](LICENSE) © Boyin Chen
