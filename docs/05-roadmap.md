# GasTrack 项目规划

## 1. 项目结构

```
GasTrack/
├── docs/                        # 设计文档
├── server/                      # Go 后端（独立 Go Module）
│   ├── cmd/
│   │   └── server/
│   │       └── main.go          # 入口：创建 mux、注册路由、启动服务
│   ├── internal/
│   │   ├── config/              # 配置加载（Viper）
│   │   ├── router/              # 路由注册（基于 net/http.ServeMux）
│   │   │   └── router.go        # 统一注册所有路由
│   │   ├── middleware/          # 中间件（认证/日志/CORS/限流/Recovery）
│   │   ├── handler/             # HTTP 处理器（按模块分）
│   │   │   ├── auth.go
│   │   │   ├── user.go
│   │   │   ├── vehicle.go
│   │   │   ├── fuel_record.go
│   │   │   └── stats.go
│   │   ├── service/             # 业务逻辑层
│   │   ├── repository/          # 数据访问层
│   │   ├── model/               # 数据库模型（GORM）
│   │   ├── dto/                 # 请求/响应结构体
│   │   ├── database/            # 数据库连接 & 迁移
│   │   └── pkg/                 # 内部工具
│   │       ├── respond/         # JSON 响应辅助 (respond.JSON/Error)
│   │       ├── decode/          # 请求解析辅助 (decode.JSON/PathParam)
│   │       ├── apperror/        # 统一错误类型
│   │       └── convert/         # 单位换算引擎
│   ├── config.yaml              # 服务端配置
│   ├── go.mod
│   └── go.sum
├── packages/                    # 前端 Monorepo (pnpm workspace)
│   ├── shared/                  # 共享代码 (@gastrack/shared)
│   │   └── src/
│   │       ├── types/           # TypeScript 类型定义
│   │       ├── api/             # API 调用层 (Axios)
│   │       ├── stores/          # 状态管理 (Zustand)
│   │       ├── i18n/            # 国际化（i18next + 翻译资源）
│   │       │   └── locales/     # zh-CN.json / en-US.json / ja-JP.json
│   │       ├── constants/       # 常量（燃油类型/车辆类型/单位/货币等）
│   │       └── utils/           # 工具函数（格式化等）
│   └── web/                     # React Web 应用 (@gastrack/web)
│       ├── src/
│       │   ├── components/      # 通用组件
│       │   ├── pages/           # 页面
│       │   ├── layouts/         # 布局
│       │   ├── hooks/           # 自定义 Hooks
│       │   └── styles/          # 全局样式
│       ├── public/
│       └── vite.config.ts
├── docker-compose.yaml          # PostgreSQL 容器（+ 可选 Mailpit）
├── pnpm-workspace.yaml
├── package.json
├── .env.example                 # 前端环境变量示例
├── LICENSE                      # MIT
└── README.md
```

## 2. 开发里程碑

### 第一期：MVP（8-10 周）

| 周次 | 任务 | 交付物 |
|------|------|--------|
| W1-W2 | 项目搭建 + 基础设施 | Go 后端骨架、前端 Monorepo、Docker 环境、CI/CD |
| W3 | 用户认证（邮箱注册/登录） | Auth API + 登录/注册页面 |
| W4 | 用户资料 + 多语言框架 | 个人设置页、中英日三语切换 |
| W5 | 车辆管理 CRUD（汽车+摩托车） | 车辆列表/添加/编辑页 |
| W6-W7 | 加油记录 CRUD + 油耗计算 | 加油记录页、记录详情页 |
| W8 | 统计报表 | 油耗趋势图、费用统计页 |
| W9 | 多币种/单位支持(含km/L) + UI 打磨 | 单位换算、响应式适配 |
| W10 | 测试 + Bug 修复 + 部署 | 生产环境上线 |

### 第二期：增强（4-6 周）

| 任务 | 说明 |
|------|------|
| 家庭群组 | 群组 CRUD + 邀请 + 数据汇总 |
| 数据导出 | CSV / PDF 导出 |
| PWA 支持 | 离线访问、安装到桌面 |
| 更多语言 | 韩语、繁中、西班牙语等 |
| 多车对比 | 车辆油耗/费用对比图表 |
| 第三方登录 | Google + Apple 登录 (P2，可选) |

### 第三期：扩展（持续）

| 任务 | 说明 |
|------|------|
| 微信小程序 | Taro 开发小程序端 |
| 微信登录 | 小程序/公众号登录 |
| 小票 OCR | 拍照识别加油小票 |
| 保养提醒 | 基于里程/时间的提醒 |
| 加油站地图 | 附近加油站展示 |

## 3. API 路由规划（V1）

> 以下为已实际注册的路由（✅）和计划中的路由（🔲）

```
# 认证（公开）
POST   /api/v1/auth/register        # ✅ 注册
POST   /api/v1/auth/login            # ✅ 登录
POST   /api/v1/auth/refresh          # ✅ 刷新 Token

# 认证（需登录）
POST   /api/v1/auth/logout           # ✅ 登出
POST   /api/v1/auth/forgot-password  # 🔲 忘记密码

# 用户
GET    /api/v1/users/me              # ✅ 获取当前用户
PATCH  /api/v1/users/me              # ✅ 更新用户资料
PUT    /api/v1/users/me/password     # ✅ 修改密码
DELETE /api/v1/users/me              # ✅ 注销账号

# 车辆
GET    /api/v1/vehicles              # ✅ 车辆列表
POST   /api/v1/vehicles              # ✅ 添加车辆
GET    /api/v1/vehicles/{id}         # ✅ 车辆详情
PATCH  /api/v1/vehicles/{id}         # ✅ 编辑车辆
DELETE /api/v1/vehicles/{id}         # ✅ 删除车辆

# 加油/充电记录
GET    /api/v1/vehicles/{id}/records       # ✅ 记录列表（分页）
POST   /api/v1/vehicles/{id}/records       # ✅ 添加记录
GET    /api/v1/vehicles/{id}/records/{rid} # ✅ 记录详情
PATCH  /api/v1/vehicles/{id}/records/{rid} # ✅ 编辑记录
DELETE /api/v1/vehicles/{id}/records/{rid} # ✅ 删除记录

# 统计
GET    /api/v1/vehicles/{id}/stats              # ✅ 车辆统计
GET    /api/v1/vehicles/{id}/efficiency-trend   # ✅ 油耗/电耗趋势
GET    /api/v1/stats/overview                   # ✅ 全局统计总览
GET    /api/v1/stats/expenses                   # 🔲 费用统计

# 健康检查
GET    /api/v1/health                # ✅ 健康检查

# 文件上传（P1）
POST   /api/v1/upload/image          # 🔲 上传图片
```
