# GasTrack 项目规划

## 1. 项目结构（Monorepo）

```
GasTrack/
├── docs/                    # 设计文档
├── packages/
│   ├── shared/              # 共享代码
│   │   ├── types/           # TypeScript 类型定义
│   │   ├── utils/           # 工具函数（单位换算、格式化）
│   │   ├── api/             # API 调用层
│   │   ├── stores/          # 状态管理
│   │   ├── i18n/            # 国际化资源
│   │   └── constants/       # 常量（国家/币种/燃油类型）
│   ├── web/                 # Web 前端
│   │   ├── src/
│   │   │   ├── components/  # 通用组件
│   │   │   ├── pages/       # 页面
│   │   │   ├── layouts/     # 布局
│   │   │   ├── hooks/       # 自定义 Hooks
│   │   │   └── styles/      # 全局样式
│   │   ├── public/
│   │   └── vite.config.ts
│   ├── server/              # 后端服务
│   │   ├── src/
│   │   │   ├── modules/     # 业务模块
│   │   │   │   ├── auth/
│   │   │   │   ├── user/
│   │   │   │   ├── vehicle/
│   │   │   │   ├── fuel-record/
│   │   │   │   ├── stats/
│   │   │   │   ├── group/
│   │   │   │   └── upload/
│   │   │   ├── common/      # 公共模块（守卫/过滤器/管道）
│   │   │   ├── config/      # 配置
│   │   │   └── database/    # 数据库迁移与种子
│   │   └── nest-cli.json
│   └── miniprogram/         # 小程序（第二阶段）
├── docker/                  # Docker 配置
│   ├── docker-compose.yml
│   ├── docker-compose.prod.yml
│   └── nginx/
├── .github/                 # CI/CD
│   └── workflows/
├── pnpm-workspace.yaml
├── package.json
├── tsconfig.base.json
└── .env.example
```

## 2. 开发里程碑

### 第一期：MVP（8-10 周）

| 周次 | 任务 | 交付物 |
|------|------|--------|
| W1-W2 | 项目搭建 + 基础设施 | Monorepo 骨架、Docker 环境、CI/CD |
| W3 | 用户认证（邮箱注册/登录） | Auth API + 登录/注册页面 |
| W4 | 用户资料 + 多语言框架 | 个人设置页、中英文切换 |
| W5 | 车辆管理 CRUD | 车辆列表/添加/编辑页 |
| W6-W7 | 加油记录 CRUD + 油耗计算 | 加油记录页、记录详情页 |
| W8 | 统计报表 | 油耗趋势图、费用统计页 |
| W9 | 多币种/单位支持 + UI 打磨 | 单位换算、响应式适配 |
| W10 | 测试 + Bug 修复 + 部署 | 生产环境上线 |

### 第二期：增强（4-6 周）

| 任务 | 说明 |
|------|------|
| 第三方登录 | Google + Apple 登录 |
| 家庭群组 | 群组 CRUD + 邀请 + 数据汇总 |
| 数据导出 | CSV / PDF 导出 |
| PWA 支持 | 离线访问、安装到桌面 |
| 更多语言 | 日语、韩语、繁中等 |
| 多车对比 | 车辆油耗/费用对比图表 |

### 第三期：扩展（持续）

| 任务 | 说明 |
|------|------|
| 微信小程序 | Taro 开发小程序端 |
| 微信登录 | 小程序/公众号登录 |
| 小票 OCR | 拍照识别加油小票 |
| 保养提醒 | 基于里程/时间的提醒 |
| 加油站地图 | 附近加油站展示 |

## 3. API 路由规划（V1）

```
POST   /api/v1/auth/register        # 注册
POST   /api/v1/auth/login            # 登录
POST   /api/v1/auth/refresh          # 刷新 Token
POST   /api/v1/auth/logout           # 登出
POST   /api/v1/auth/forgot-password  # 忘记密码

GET    /api/v1/users/me              # 获取当前用户
PATCH  /api/v1/users/me              # 更新用户资料
DELETE /api/v1/users/me              # 注销账号

GET    /api/v1/vehicles              # 车辆列表
POST   /api/v1/vehicles              # 添加车辆
GET    /api/v1/vehicles/:id          # 车辆详情
PATCH  /api/v1/vehicles/:id          # 编辑车辆
DELETE /api/v1/vehicles/:id          # 删除车辆

GET    /api/v1/vehicles/:id/records       # 加油记录列表
POST   /api/v1/vehicles/:id/records       # 添加记录
GET    /api/v1/vehicles/:id/records/:rid  # 记录详情
PATCH  /api/v1/vehicles/:id/records/:rid  # 编辑记录
DELETE /api/v1/vehicles/:id/records/:rid  # 删除记录

GET    /api/v1/vehicles/:id/stats         # 车辆统计
GET    /api/v1/stats/overview             # 全局统计总览
GET    /api/v1/stats/expenses             # 费用统计

POST   /api/v1/upload/image              # 上传图片
```
