# GasTrack 技术选型

## 1. 技术栈总览

```
┌─────────────────────────────────────────────────┐
│  前端 (Web)         │  跨端 (小程序)              │
│  React 18 + TS      │  Taro 3 + React            │
│  Vite 5             │  (共享业务逻辑层)            │
│  Ant Design 5       │                             │
│  Ant Design Mobile  │                             │
│  react-i18next      │                             │
│  ECharts            │                             │
│  Zustand            │                             │
├─────────────────────┴─────────────────────────────┤
│  后端                                              │
│  Go 1.22+ │ net/http 标准库 │ GORM (ORM)            │
│  golang-jwt (认证) │ go-i18n (多语言)              │
│  go-cache (进程内缓存)                              │
├───────────────────────────────────────────────────┤
│  基础设施                                          │
│  PostgreSQL 16 │ MinIO/S3 (文件存储)               │
│  Docker + Docker Compose (开发环境)                 │
│  Nginx (反向代理 + 限流)                            │
│  GitHub Actions (CI/CD)                            │
└───────────────────────────────────────────────────┘
```

## 2. 前端技术选型详解

### 2.1 框架：React 18 + TypeScript

**选择理由：**
- 生态最丰富，组件库/工具库选择多
- TypeScript 提供类型安全，减少运行时错误
- Taro 3 支持 React，Web 和小程序可共享逻辑
- 函数组件 + Hooks 开发体验优秀

### 2.2 构建工具：Vite 5

- 开发环境秒启动（ESM 原生支持）
- HMR 极速热更新
- 生产构建基于 Rollup，体积优化好

### 2.3 UI 组件库：Ant Design 5 + Ant Design Mobile

- **Ant Design 5**：PC 端组件丰富，国际化支持完善
- **Ant Design Mobile**：移动端体验好，和 Ant Design 风格统一
- 使用响应式断点判断加载对应组件
- 支持 CSS-in-JS（主题定制方便）

### 2.4 状态管理：Zustand

- 轻量级（不到 1KB），API 简洁
- 天然支持 TypeScript
- 不需要 Provider 包裹

### 2.5 图表：ECharts

- 图表类型丰富（折线图、柱状图、饼图）
- 移动端触摸交互好
- 国际化支持完善
- 体积可按需引入

### 2.6 HTTP 客户端：Axios

- 拦截器机制（统一处理 Token 刷新、错误提示）
- 请求/响应类型推导

## 3. 后端技术选型详解

### 3.1 语言与 HTTP 层：Go + 标准库 net/http

**选择 Go 的理由：**
- 编译型语言，性能远超 Node.js
- 天然高并发（goroutine），适合 API 服务
- 单二进制部署，无需运行时环境
- 静态类型 + 编译检查，运行时错误少
- 内存占用低，服务器成本低
- Docker 镜像极小（基于 scratch/alpine 可到 10-20MB）

**选择标准库而非第三方框架的理由：**
- **Go 1.22 增强路由**：`net/http` 已原生支持 HTTP 方法匹配和路径参数
  ```go
  mux.HandleFunc("GET /api/v1/vehicles/{id}", h.GetVehicle)
  mux.HandleFunc("POST /api/v1/auth/login", h.Login)
  ```
- **零依赖**：不引入 Gin/Chi/Echo，减少供应链风险
- **长期稳定**：标准库由 Go 团队维护，向后兼容保证
- **中间件机制**：标准库的 `http.Handler` 接口天然支持中间件链式组合
  ```go
  // 中间件就是包装 http.Handler 的函数
  func AuthMiddleware(next http.Handler) http.Handler { ... }
  ```
- **性能无差异**：Gin 的性能优势主要来自 httprouter，对于本项目的并发量级差异可忽略
- **学习成本低**：标准库 API 稳定，Go 官方文档完善

**自行封装的薄工具层（约 200 行代码）：**
- JSON 响应辅助函数：`respond.JSON()`, `respond.Error()`
- 请求解析辅助函数：`decode.JSON()`, `decode.PathParam()`
- 中间件链组合：`middleware.Chain()`
- 统一错误处理：`AppError` 类型 + 错误中间件

### 3.2 ORM：GORM

- Go 社区最成熟的 ORM
- 支持 AutoMigrate（数据库迁移）
- 支持 PostgreSQL 高级特性（JSON/Array）
- 链式调用，API 友好
- 支持 Hooks（BeforeCreate/AfterUpdate 等）

### 3.3 认证：golang-jwt

- Go 标准的 JWT 库
- 支持多种签名算法（HS256/RS256）
- 轻量无依赖

### 3.4 配置管理：Viper

- 支持多种格式（YAML/TOML/ENV）
- 环境变量覆盖
- 热加载配置

### 3.5 日志：Zap

- Uber 出品的高性能结构化日志库
- 零内存分配的 JSON 日志
- 生产环境性能最佳

### 3.6 数据校验：go-playground/validator

- struct tag 声明式校验
- 在请求解析辅助函数中统一调用
- 自定义校验规则（如油量范围、里程递增等）

### 3.7 文件存储：MinIO (自托管) / AWS S3 (云端)

- S3 兼容 API（代码不需改动即可切换）
- 开发环境用 MinIO（Docker 本地运行）
- 生产环境可选 S3 / 腾讯云 COS / 阿里云 OSS

### 3.8 API 文档

两种可选方案：
- **手写 OpenAPI 3.0 YAML**：版本控制友好，前端可用 `openapi-typescript` 生成类型
- **go-swagger 注释生成**：代码注释自动生成文档（不依赖 Gin）
- 开发阶段可用 Swagger UI 在线调试

### 3.9 定时/异步任务

不引入额外消息队列，使用以下轻量方案：
- **定时任务**：`robfig/cron` 库（统计预计算、过期 Token 清理）
- **异步处理**：Go 原生 goroutine + channel（邮件发送等）
- 未来用户量增长后可引入 NATS/RabbitMQ

## 4. 为什么不需要 Redis？

| 需求 | Redis 方案 | 当前替代方案 |
|------|-----------|-------------|
| 数据缓存 | Redis GET/SET | `go-cache` 进程内缓存 |
| API 限流 | Redis INCR + EXPIRE | Nginx `limit_req` + Go `rate` 包 |
| 会话管理 | Redis 存 Session | JWT 无状态，Refresh Token 存 PostgreSQL |
| 任务队列 | Bull/BullMQ (基于Redis) | Go goroutine + channel |
| 分布式锁 | Redis SETNX | 单实例无需；PostgreSQL Advisory Lock 备用 |

**何时引入 Redis：**
- 多实例部署需要共享缓存时
- 用户量超过 5 万，数据库查询压力大时
- 需要发布/订阅(Pub/Sub)能力时

## 5. 跨端方案（微信小程序）

### 5.1 框架：Taro 3

**选择理由：**
- 支持 React 语法，与 Web 端代码复用度高
- 一套代码编译到微信/支付宝/百度小程序
- 可以共享：API 调用层、状态管理、工具函数、类型定义

**代码复用策略：**
```
packages/
├── shared/          # 共享代码（Monorepo）
│   ├── api/         # API 调用函数
│   ├── types/       # TypeScript 类型
│   ├── utils/       # 工具函数（单位换算等）
│   ├── stores/      # 状态管理
│   └── i18n/        # 翻译资源
├── web/             # Web 前端
│   └── src/
└── miniprogram/     # 小程序
    └── src/
```

## 6. 开发工具与规范

| 类别 | 工具 | 说明 |
|------|------|------|
| 前端包管理 | pnpm | 高效磁盘利用，Monorepo 支持 |
| 前端 Monorepo | pnpm workspace | 管理 Web/小程序/共享包 |
| 后端依赖管理 | Go Modules | Go 官方依赖管理 |
| 前端代码规范 | ESLint + Prettier | 统一代码风格 |
| 后端代码规范 | golangci-lint | Go 综合 linter |
| Git 规范 | Commitlint + Husky | 约束提交信息 |
| API 文档 | OpenAPI 3.0 YAML | 手写规范 + Swagger UI |
| 前端测试 | Jest + Testing Library | 单元测试 + 组件测试 |
| 后端测试 | Go testing + testify | Go 标准测试 |
| E2E 测试 | Playwright | 端到端测试 (P1) |

## 7. 部署架构

### 7.1 开发环境
```bash
# Docker Compose 一键启动依赖
docker compose up -d
# 包含：PostgreSQL
# 可选（需 --profile full）：Mailpit（邮件测试）
```

### 7.2 生产环境（推荐方案）
- **云服务器**：1C2G 即可起步（Go 内存占用极低）
- **部署方式**：单二进制 + systemd 或 Docker
- **反向代理**：Nginx（SSL + 静态资源 + API 转发 + 限流）
- **CI/CD**：GitHub Actions → 编译 → 部署
- Go 服务启动秒级，内存占用 ~20-50MB

### 7.3 生产环境（扩展方案）
当用户量增长后可升级：
- 引入 Redis 做分布式缓存
- Kubernetes 编排
- 读写分离（PostgreSQL 主从）
- CDN 加速静态资源
- 监控告警（Prometheus + Grafana）
