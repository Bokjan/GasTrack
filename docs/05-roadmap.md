# GasTrack 项目规划

> **最后更新**: 2026-03-31

## 1. 项目结构

```
GasTrack/
├── docs/                        # 设计文档（11 篇）
├── server/                      # Go 后端（独立 Go Module）
│   ├── cmd/server/main.go       # 入口：依赖注入、启动服务
│   ├── internal/
│   │   ├── config/              # 配置加载（Viper）
│   │   ├── router/router.go     # 路由注册（Go 1.22 ServeMux）
│   │   ├── middleware/          # 中间件（Auth/Logger/CORS/RateLimit/Recovery）
│   │   ├── handler/             # HTTP 处理器（auth/user/vehicle/fuel_record/stats/invite/export/reminder/notification/group）
│   │   ├── service/             # 业务逻辑层
│   │   ├── repository/          # 数据访问层
│   │   ├── model/               # 数据库模型（10 个：User/Vehicle/FuelRecord/RefreshToken/InviteCode/Reminder/Notification/Group/GroupMember/SharedVehicle）
│   │   ├── dto/                 # 请求/响应结构体
│   │   ├── database/            # 数据库连接 & AutoMigrate
│   │   └── pkg/                 # 内部工具（respond/decode/apperror/convert）
│   ├── config.yaml
│   ├── go.mod / go.sum
├── packages/                    # 前端 Monorepo (pnpm workspace)
│   ├── shared/                  # 共享代码 (@gastrack/shared)
│   │   └── src/
│   │       ├── types/           # TypeScript 类型定义
│   │       ├── api/             # API 调用层 (Axios)
│   │       ├── stores/          # 状态管理 (Zustand: auth/vehicle/theme)
│   │       ├── i18n/locales/    # zh-CN / en-US / ja-JP
│   │       ├── constants/       # 常量
│   │       └── utils/           # 工具函数
│   └── web/                     # React Web 应用 (@gastrack/web)
│       └── src/
│           ├── components/      # 通用组件（NotificationBell 等）
│           ├── pages/           # auth/dashboard/vehicle/record/stats/invite/reminder/settings/legal/group
│           ├── layouts/         # MainLayout
│           ├── hooks/           # useIsMobile 等
│           └── styles/          # global.css
├── docker-compose.yaml          # PostgreSQL（+ 可选 Mailpit）
├── pnpm-workspace.yaml / package.json
├── LICENSE (MIT) / README.md
```

## 2. 开发里程碑

### 第一期：MVP ✅（基本完成）

| 周次 | 任务 | 状态 |
|------|------|------|
| W1-W2 | 项目搭建 + 基础设施 | ✅ |
| W3 | 用户认证（注册/登录/JWT） | ✅ |
| W4 | 用户资料 + 多语言框架 + 邀请注册制 | ✅ |
| W5 | 车辆管理 CRUD（汽车/摩托/电动） | ✅ |
| W6-W7 | 加油/充电记录 CRUD + 油耗计算 | ✅ |
| W8 | 统计报表 + 深色模式 + 数据导出 CSV | ✅ |
| W9 | 多币种/单位 + 响应式适配 + 隐私政策 | ✅ |
| W10 | 通知/提醒系统 + 邀请码管理 | ✅ |

**第一期剩余**:

| 任务 | 优先级 | 说明 |
|------|--------|------|
| 后端 i18n 错误消息 | ⭐⭐ 中 | go-i18n TOML 翻译 |
| 忘记密码流程 | ⭐⭐ 中 | 后端邮件发送 + Token，DTO 已定义 |
| 记录列表筛选 UI | ⭐ 低 | 后端已支持，前端缺筛选控件 |

### 第二期：增强（P1）

| 任务 | 状态 |
|------|------|
| PWA 支持（离线 + 安装到桌面） | 🔲 |
| 多车辆对比图表 | 🔲 |
| 文件上传（车辆照片 + 头像） | 🔲 |
| ~~家庭群组管理（基础）~~ | ~~✅ 已完成~~ |
| ~~车辆共享标记~~ | ~~✅ 已完成（全栈：3 API + 权限 + 前端 UI）~~ |
| ~~群组油耗排行榜~~ | ~~✅ 已完成（全栈：API + 前端排行榜 Tab）~~ |
| ~~群组费用统计看板~~ | ~~✅ 已完成（全栈：API + 前端费用 Tab）~~ |
| ~~加油站推荐共享~~ | ~~✅ 已完成（全栈：API + 前端加油站 Tab）~~ |
| 更多语言（韩语/繁中/西/德/法） | 🔲 |
| 数据导出 PDF | 🔲 |

### 第三期：扩展（P2）

| 任务 | 状态 |
|------|------|
| 微信小程序（Taro） | 🔲 |
| 第三方登录（Google/Apple/微信） | 🔲 |
| 小票 OCR 识别 | 🔲 |
| 加油站地图（PostGIS） | 🔲 |
| 无障碍访问（WCAG 2.1 AA） | 🔲 |
| ~~汇率参考（只读展示）~~ | ~~✅ 已完成（frankfurter.app + 内存缓存 24h，设置页/仪表盘/统计/记录详情/记录列表全覆盖）~~ |

## 3. API 路由一览（V1）

> 54 条已注册路由 + 2 条待实现

```
# 公开路由
POST   /api/v1/auth/register              # ✅ 注册
POST   /api/v1/auth/login                 # ✅ 登录
POST   /api/v1/auth/refresh               # ✅ 刷新 Token
GET    /api/v1/auth/registration-mode      # ✅ 查询注册模式
GET    /api/v1/invites/{code}              # ✅ 验证邀请码
GET    /api/v1/health                      # ✅ 健康检查

# 认证（需登录）
POST   /api/v1/auth/logout                # ✅ 登出
POST   /api/v1/auth/forgot-password       # 🔲 忘记密码

# 用户
GET    /api/v1/users/me                    # ✅ 获取资料
PATCH  /api/v1/users/me                    # ✅ 更新资料
PUT    /api/v1/users/me/password           # ✅ 修改密码
DELETE /api/v1/users/me                    # ✅ 注销账号
GET    /api/v1/users/me/export             # ✅ 数据导出 CSV

# 邀请码
POST   /api/v1/invites                     # ✅ 创建邀请码
GET    /api/v1/invites                     # ✅ 我的邀请码列表
PATCH  /api/v1/invites/{id}               # ✅ 更新邀请码
DELETE /api/v1/invites/{id}               # ✅ 删除邀请码

# 车辆
GET    /api/v1/vehicles                    # ✅ 车辆列表
POST   /api/v1/vehicles                   # ✅ 添加车辆
GET    /api/v1/vehicles/{id}              # ✅ 车辆详情
PATCH  /api/v1/vehicles/{id}              # ✅ 编辑车辆
DELETE /api/v1/vehicles/{id}              # ✅ 删除车辆

# 加油/充电记录
GET    /api/v1/vehicles/{id}/records       # ✅ 记录列表（分页）
POST   /api/v1/vehicles/{id}/records       # ✅ 添加记录
GET    /api/v1/vehicles/{id}/records/{rid} # ✅ 记录详情
PATCH  /api/v1/vehicles/{id}/records/{rid} # ✅ 编辑记录
DELETE /api/v1/vehicles/{id}/records/{rid} # ✅ 删除记录
GET    /api/v1/vehicles/{id}/stations      # ✅ 站名建议

# 统计
GET    /api/v1/vehicles/{id}/stats             # ✅ 车辆统计
GET    /api/v1/vehicles/{id}/efficiency-trend  # ✅ 油耗趋势
GET    /api/v1/vehicles/{id}/period-stats      # ✅ 按时段聚合（月/年 + 同比）
GET    /api/v1/stats/overview                  # ✅ 全局总览

# 保养提醒
GET    /api/v1/reminders                   # ✅ 提醒列表
POST   /api/v1/reminders                   # ✅ 创建提醒
GET    /api/v1/reminders/{id}             # ✅ 提醒详情
PATCH  /api/v1/reminders/{id}             # ✅ 更新提醒
DELETE /api/v1/reminders/{id}             # ✅ 删除提醒

# 通知
GET    /api/v1/notifications               # ✅ 通知列表
GET    /api/v1/notifications/unread-count  # ✅ 未读数
PATCH  /api/v1/notifications/{id}/read     # ✅ 标记已读
POST   /api/v1/notifications/read-all      # ✅ 全部已读
DELETE /api/v1/notifications/{id}          # ✅ 删除通知

# 群组管理（基础）
GET    /api/v1/groups                      # ✅ 我的群组列表
POST   /api/v1/groups                      # ✅ 创建群组
POST   /api/v1/groups/join                 # ✅ 通过邀请码加入
GET    /api/v1/groups/{id}                 # ✅ 群组详情
PATCH  /api/v1/groups/{id}                 # ✅ 更新群组信息
DELETE /api/v1/groups/{id}                 # ✅ 删除群组
POST   /api/v1/groups/{id}/regenerate-invite # ✅ 重新生成邀请码
POST   /api/v1/groups/{id}/leave           # ✅ 退出群组
GET    /api/v1/groups/{id}/overview        # ✅ 群组数据汇总
PATCH  /api/v1/groups/{id}/members/{uid}   # ✅ 更新成员角色
DELETE /api/v1/groups/{id}/members/{uid}   # ✅ 移除成员

# 群组扩展（✅ 已全部实现）
POST   /api/v1/groups/{id}/shared-vehicles        # ✅ 共享车辆到群组
DELETE /api/v1/groups/{id}/shared-vehicles/{vid}   # ✅ 取消车辆共享
GET    /api/v1/groups/{id}/shared-vehicles         # ✅ 获取群组共享车辆列表
GET    /api/v1/groups/{id}/leaderboard             # ✅ 群组油耗排行榜
GET    /api/v1/groups/{id}/expense-stats           # ✅ 群组费用统计看板
GET    /api/v1/groups/{id}/stations                # ✅ 加油站推荐共享

# 汇率参考
GET    /api/v1/exchange-rates              # ✅ 汇率查询

# 开销记录
GET    /api/v1/vehicles/{id}/expenses                # ✅ 开销列表（分页+筛选）
POST   /api/v1/vehicles/{id}/expenses                # ✅ 创建开销记录
GET    /api/v1/vehicles/{id}/expenses/{eid}          # ✅ 开销详情
PATCH  /api/v1/vehicles/{id}/expenses/{eid}          # ✅ 更新开销记录
DELETE /api/v1/vehicles/{id}/expenses/{eid}          # ✅ 删除开销记录
GET    /api/v1/vehicles/{id}/expense-stats           # ✅ 开销统计
GET    /api/v1/vehicles/{id}/expense-vendors         # ✅ 商家名建议

# 其他待实现
POST   /api/v1/upload/image               # 🔲 上传图片 (P1)
```
