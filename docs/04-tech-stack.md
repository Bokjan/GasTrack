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
│  Node.js 20 LTS + NestJS 10 + TypeScript          │
│  TypeORM (PostgreSQL)                              │
│  Passport.js (认证)                                │
│  nestjs-i18n (多语言)                               │
│  Bull (任务队列)                                    │
├───────────────────────────────────────────────────┤
│  基础设施                                          │
│  PostgreSQL 16 │ Redis 7 │ MinIO/S3 (文件存储)     │
│  Docker + Docker Compose (开发环境)                 │
│  Nginx (反向代理)                                   │
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

**对比排除：**
- Vue 3：生态略逊，Taro 对 Vue 3 支持不如 React 成熟
- Angular：学习曲线陡，不适合快速迭代

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
- 对比 Redux Toolkit 更简洁，够用即可

### 2.5 图表：ECharts

- 图表类型丰富（折线图、柱状图、饼图）
- 移动端触摸交互好
- 国际化支持完善
- 体积可按需引入

### 2.6 HTTP 客户端：Axios

- 拦截器机制（统一处理 Token 刷新、错误提示）
- 请求/响应类型推导

## 3. 后端技术选型详解

### 3.1 框架：NestJS 10

**选择理由：**
- 模块化架构天然适合后续拆分微服务
- 装饰器 + 依赖注入，代码组织清晰
- 内置支持：守卫(Guard)、管道(Pipe)、拦截器(Interceptor)
- TypeScript 原生支持
- 丰富的官方模块（Swagger/JWT/Throttle 等）

**对比排除：**
- Express：太轻量，缺少约束，大项目易混乱
- Fastify：性能好但生态不如 NestJS 丰富
- Go/Java：团队以 TS 全栈为主，减少语言切换成本

### 3.2 ORM：TypeORM

- 与 NestJS 深度集成
- 支持 Migration（数据库版本管理）
- 装饰器定义实体，与 NestJS 风格一致
- 支持 PostgreSQL 高级特性

### 3.3 认证：Passport.js + JWT

- 多策略支持（Local/JWT/Google/Apple/WeChat）
- NestJS 官方推荐方案
- 成熟稳定，社区活跃

### 3.4 文件存储：MinIO (自托管) / AWS S3 (云端)

- S3 兼容 API（代码不需改动即可切换）
- 开发环境用 MinIO（Docker 本地运行）
- 生产环境可选 S3 / 腾讯云 COS / 阿里云 OSS

### 3.5 任务队列：Bull (基于 Redis)

- 定时任务：统计数据预计算
- 异步任务：邮件发送、图片处理
- 可监控的任务面板

## 4. 跨端方案（微信小程序）

### 4.1 框架：Taro 3

**选择理由：**
- 支持 React 语法，与 Web 端代码复用度高
- 一套代码编译到微信/支付宝/百度小程序
- 可以共享：API 调用层、状态管理、工具函数、类型定义
- 京东出品，社区活跃

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

## 5. 开发工具与规范

| 类别 | 工具 | 说明 |
|------|------|------|
| 包管理 | pnpm | 高效磁盘利用，Monorepo 支持 |
| Monorepo | pnpm workspace | 管理多包项目 |
| 代码规范 | ESLint + Prettier | 统一代码风格 |
| Git 规范 | Commitlint + Husky | 约束提交信息 |
| API 文档 | Swagger (OpenAPI) | 自动生成 API 文档 |
| 测试 | Jest + Testing Library | 单元测试 + 组件测试 |
| E2E 测试 | Playwright | 端到端测试 (P1) |

## 6. 部署架构

### 6.1 开发环境
```bash
# Docker Compose 一键启动
docker-compose up -d
# 包含：PostgreSQL + Redis + MinIO + Mailpit(邮件测试)
```

### 6.2 生产环境（推荐方案）
- **云服务器**：2C4G 起步（轻量应用服务器即可）
- **容器部署**：Docker + Docker Compose
- **反向代理**：Nginx（SSL + 静态资源 + API 转发）
- **CI/CD**：GitHub Actions → 自动构建 → 部署
- **域名**：gastrack.app（示例）

### 6.3 生产环境（扩展方案）
当用户量增长后可升级：
- Kubernetes 编排
- 读写分离（PostgreSQL 主从）
- CDN 加速静态资源
- 监控告警（Prometheus + Grafana）
