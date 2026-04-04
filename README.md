# ⛽ GasTrack

**车辆能耗与费用管理系统** — 记录加油/充电，追踪维保开销，分析能耗趋势，辅助用车决策。

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## ✨ 功能特性

### 核心功能
- 🚗 **多车辆管理** — 汽车、摩托车，燃油 & 电动车全覆盖；支持归档、默认车辆
- ⛽ **加油/充电记录** — 日期、站点、量、单价、里程一站式录入；三值自动计算、站点名自动补全
- 📊 **油耗/电耗统计** — 折线趋势图、费用汇总、多维分析、按月/年聚合 + 往年同比；三 Tab 视角（⛽加油/💸开销/📊综合）
- 🔧 **维修保养开销** — 独立开销台账（10 种分类），支持统计摘要、商家建议、保养提醒联动；统计已集成到 Dashboard 和 Stats 页

### 国际化 & 多端
- 🌍 **多语言** — 中文简体 / English / 日本語
- 💱 **多币种 & 多计量体系** — CNY/USD/EUR/JPY/GBP/KRW · L/100km / km/L / MPG / kWh/100km
- 💹 **汇率参考** — frankfurter.app 实时汇率，记录/统计/群组全覆盖
- 🌙 **深色模式** — Light / Dark / 跟随系统，CSS 变量 + ECharts 暗色适配
- 📱 **响应式设计** — PC & 移动端自适应（Table↔卡片、Sider↔Drawer）

### 社交 & 协作
- 👨‍👩‍👧‍👦 **家庭群组** — 创建群组、邀请码加入、三级角色权限（Owner/Admin/Member）
- 🚙 **车辆共享** — 群组内共享车辆，多人为同一辆车记录加油
- 🏆 **油耗排行榜** — 四维排行（油耗/费用/里程/频次），前三名奖牌标识
- 💰 **群组费用看板** — 统计卡片 + 趋势表 + 成员占比，环比变化一目了然
- ⛽ **加油站推荐** — 群组成员常去加油站聚合，含价格趋势和常客信息

### 系统能力
- 🔒 **JWT 认证** — Access Token + Refresh Token（Rotation），安全可靠
- 🎫 **邀请注册制** — 可配置注册策略（invite_only / open / closed）
- 🔔 **通知系统** — 异常油耗预警、保养到期提醒、邀请码使用通知
- 📤 **数据导出** — 支持 CSV / ZIP / JSON 三种格式，基础或完整范围可选（GDPR 数据可携带权）
- 📜 **隐私合规** — 隐私政策 + 用户协议 + 账号注销（GDPR 数据删除权）

## 🏗️ 技术栈

| 层级 | 技术 |
|------|------|
| **前端** | React 18 · TypeScript · Vite 5 · Ant Design 5 · ECharts · Zustand · i18next |
| **后端** | Go 1.23 · net/http 标准库 · GORM · JWT · Viper · Zap |
| **数据库** | PostgreSQL 16 |
| **包管理** | pnpm workspace (Monorepo) |
| **部署** | Docker Compose · Nginx（HTTPS + 反向代理）· Let's Encrypt |

## 📁 项目结构

```
GasTrack/
├── packages/
│   ├── shared/              # 共享层：类型、API、i18n、状态管理、工具函数
│   │   └── src/
│   │       ├── types/       # TypeScript 类型定义
│   │       ├── api/         # API 调用层 (Axios)
│   │       ├── stores/      # 状态管理 (Zustand: auth/vehicle/theme/exchangeRate)
│   │       ├── i18n/locales/# zh-CN / en-US / ja-JP
│   │       ├── constants/   # 常量
│   │       └── utils/       # 工具函数（单位换算、格式化等）
│   └── web/                 # Web 前端 (React + Vite)
│       └── src/
│           ├── components/  # 通用组件（NotificationBell 等）
│           ├── pages/       # 页面（auth/dashboard/vehicle/record/stats/
│           │                #       invite/reminder/settings/legal/group/expense）
│           ├── layouts/     # MainLayout (Sider/Drawer 自适应)
│           ├── hooks/       # useIsMobile 等
│           └── styles/      # global.css
├── server/                  # Go 后端
│   ├── cmd/server/          # 入口 main.go
│   ├── internal/
│   │   ├── config/          # 配置加载 (Viper)
│   │   ├── database/        # 数据库连接 & 迁移
│   │   ├── handler/         # HTTP Handler (11 个模块)
│   │   ├── middleware/      # 中间件（Auth/CORS/RateLimit/Logger/Recovery）
│   │   ├── model/           # GORM 数据模型 (11 张表)
│   │   ├── dto/             # 请求/响应 DTO
│   │   ├── repository/      # 数据访问层
│   │   ├── service/         # 业务逻辑层
│   │   ├── router/          # 路由注册 (61+ 条路由)
│   │   └── pkg/             # 内部工具（respond/decode/apperror/convert）
│   └── config.yaml          # 服务端配置
├── nginx/                   # Nginx 配置（HTTPS + HTTP-only 备选）
├── scripts/                 # 运维脚本（SSL 初始化等）
├── docs/                    # 项目文档（11 篇）
├── docker-compose.yaml      # 开发环境（PostgreSQL + Mailpit）
├── docker-compose.prod.yaml # 生产部署（PostgreSQL + Go + Nginx）
├── Dockerfile.web           # 前端多阶段构建
└── pnpm-workspace.yaml
```

## 🚀 快速开始

### 前置条件

- **Node.js** >= 18
- **pnpm** >= 8
- **Go** >= 1.23
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

后端默认监听 `http://localhost:8098`，首次启动会自动执行数据库迁移（GORM AutoMigrate）。

### 4. 启动前端

```bash
# 回到项目根目录
cd ..
pnpm install
pnpm dev
```

前端默认运行在 `http://localhost:3000`，API 请求自动代理到后端。

### 5. 访问应用

打开浏览器访问 **http://localhost:3000** 即可使用。

## 🐳 生产部署

使用 Docker Compose 一键部署（PostgreSQL + Go 后端 + Nginx + HTTPS）：

```bash
# 配置环境变量
cp .env.production.example .env.production
# 编辑 .env.production，填写 DB_PASSWORD、JWT_SECRET、DOMAIN、EMAIL

# 构建并启动
docker compose -f docker-compose.prod.yaml --env-file .env.production up -d --build

# 验证
curl https://your-domain.com/api/v1/health
```

详见 [`docs/10-deployment.md`](docs/10-deployment.md)。

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

可通过环境变量 `GASTRACK_CONFIG` 指定自定义配置文件路径，或使用 `GASTRACK_` 前缀覆盖任意配置项。

## 📖 API 概览

> 项目共 **61+ 条** RESTful API，以下为主要模块。完整文档见 [`docs/06-api-reference.md`](docs/06-api-reference.md)。

| 模块 | 示例端点 | 说明 |
|------|---------|------|
| 认证 | `POST /api/v1/auth/register` | 注册/登录/刷新/登出 (4) |
| 用户 | `GET /api/v1/users/me` | 资料/密码/注销/导出 (5) |
| 邀请码 | `POST /api/v1/invites` | CRUD + 验证 (5) |
| 车辆 | `GET /api/v1/vehicles` | CRUD + 详情 (5) |
| 加油记录 | `GET /api/v1/vehicles/{id}/records` | CRUD + 站名建议 (6) |
| 统计 | `GET /api/v1/vehicles/{id}/stats` | 车辆统计/趋势/聚合/总览/开销聚合 (5) |
| 保养提醒 | `GET /api/v1/reminders` | CRUD (5) |
| 通知 | `GET /api/v1/notifications` | 列表/未读数/标记已读/删除 (5) |
| 群组 | `GET /api/v1/groups` | CRUD + 成员 + 共享 + 排行 + 看板 + 加油站 (17) |
| 开销记录 | `GET /api/v1/vehicles/{id}/expenses` | CRUD + 统计 + 商家建议 (7) |
| 汇率 | `GET /api/v1/exchange-rates` | 汇率查询 (1) |

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
| [10-deployment.md](docs/10-deployment.md) | 线上部署 |

## 📜 License

[MIT](LICENSE) © Boyin Chen
