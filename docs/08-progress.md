# GasTrack 需求完成进度

> **更新日期**: 2026-03-31
>
> **当前阶段**: 第一期 MVP（基本完成）

---

## 1. 进度总览

| 模块 | 后端 | 前端 | 整体 |
|------|------|------|------|
| 项目基础设施 | ✅ | ✅ | ✅ |
| 用户认证（注册/登录/JWT） | ✅ | ✅ | ✅ |
| 邀请注册制 + 邀请码管理 | ✅ | ✅ | ✅ |
| 用户资料 + 设置 | ✅ | ✅ | ✅ |
| 车辆管理（含电动车） | ✅ | ✅ | ✅ |
| 加油/充电记录 | ✅ | ✅ | ✅ |
| 统计报表（月/年 + 同比） | ✅ | ✅ | ✅ |
| 多语言（zh-CN/en-US/ja-JP） | 🔲 后端 i18n | ✅ | 🔨 |
| 多币种/单位换算 | ✅ | ✅ | ✅ |
| 深色模式 | — | ✅ | ✅ |
| 移动端响应式适配 | — | ✅ | ✅ |
| 数据导出 CSV/ZIP/JSON（GDPR） | ✅ | ✅ | ✅ |
| 隐私政策 + 用户协议 | — | ✅ | ✅ |
| 保养提醒（里程/时间） | ✅ | ✅ | ✅ |
| 异常油耗预警 + 通知系统 | ✅ | ✅ | ✅ |
| 家庭群组管理（基础） | ✅ | ✅ | ✅ |
| 群组扩展（共享/排行/看板/加油站） | ✅ | ✅ | ✅ |
| 维修保养等开销记录 | ✅ | ✅ | ✅ |

**图例**: ✅ 完成 | 🔨 进行中 | 🔲 待实现

---

## 2. 后端进度

### 已完成模块

- **基础设施** ✅ — Go 1.22 + net/http, Viper 配置, Zap 日志 + Lumberjack 轮转, GORM + PostgreSQL, Docker Compose
- **中间件** ✅ — CORS, JWT Auth, 请求日志, Panic Recovery, Rate Limit (100 req/s), 中间件链
- **Auth** ✅ — 注册/登录/刷新/登出, Refresh Token Rotation (SELECT FOR UPDATE 原子消费), 邀请码注册制
- **User** ✅ — 资料 CRUD, 修改密码, 注销账号 (GDPR)
- **Vehicle** ✅ — CRUD (含电动车 battery_capacity), 默认车辆 (事务原子), 归档
- **FuelRecord** ✅ — CRUD, 分页, 站名建议, 油耗/电耗自动计算
- **Stats** ✅ — 车辆统计, 全局总览, 油耗趋势, 按月/年聚合 + 同比
- **Invite** ✅ — 邀请码 CRUD, GT-XXXXXX 格式, 并发安全消费
- **Export** ✅ — 数据导出 CSV/ZIP/JSON（scope=basic/full，10 个数据源，UTF-8 BOM，ZIP 多文件 + manifest.json）
- **Reminder** ✅ — 保养提醒 CRUD, 11 种保养类型, 3 种触发方式, 自动计算下次保养
- **Notification** ✅ — 通知 CRUD, 未读数, 标记已读, 异常油耗检测 (>30% 偏差), 保养到期检查, 邀请码使用通知
- **单位换算** ✅ — `pkg/convert/` 引擎, API 按用户偏好自动转换（含 unit_price 同步容量单位换算）
- **群组管理** ✅ — Group/GroupMember/SharedVehicle 模型, 19 条 API（基础 CRUD + 邀请码 + 权限管理 + 数据汇总 + 车辆共享 3 条 + 排行榜 + 费用看板 + 加油站推荐）
- **开销记录** ✅ — ExpenseRecord 模型, CRUD + 列表筛选 + 统计（按币种汇总/分类占比/月度趋势）+ 商家名建议 + 保养联动, 7 条 API

### 待实现

| 功能 | 优先级 | 说明 |
|------|--------|------|
| 后端 i18n 错误消息 | ⭐⭐ 中 | go-i18n TOML 翻译文件 |
| 忘记密码 | ⭐⭐ 中 | 邮件发送 + Token (DTO 已定义) |
| 文件上传 | P1 | 车辆照片 + 头像 (OSS/本地) |

---

## 3. 前端进度

### 已完成模块

- **基础设施** ✅ — Monorepo (pnpm workspace), Vite 5, React 18 + TS, React Router 6, Ant Design 5
- **共享包** ✅ — Types (完全对齐后端 DTO), API 层 (Axios + 401 自动刷新), Zustand (auth/vehicle/theme), i18n (3 语), Constants, Utils (formatDateTime 时区感知)
- **页面** ✅:
  - 登录/注册 (邀请码实时校验)
  - 仪表盘 (按车辆分组统计)
  - 车辆列表/表单 (含电动车适配)
  - 加油记录列表/表单/详情 (三值自动计算, 站点补全, 智能分析)
  - 统计 (月/年维度 + 同比, ECharts)
  - 邀请码管理 (Table/卡片, 复制/启停/删除)
  - 保养提醒 (卡片式管理, 逾期标识)
  - 设置 (时区, 外观主题, 语言, 单位, 数据导出 CSV/ZIP/JSON + 范围/格式选择, 账号注销)
  - 隐私政策 / 用户协议
- **组件** ✅ — MainLayout (Sider/Drawer 自适应), NotificationBell (60s 轮询), ProtectedRoute
- **群组页面** ✅ — `/groups` 群组列表 + 详情面板 (6 Tab: 群组信息/成员管理/数据汇总+共享车辆/排行榜/费用看板/加油站推荐) + 创建/加入/编辑弹窗 + 100+ 翻译键；全面单位/货币国际化（15+ 处硬编码修复，汇率自动换算 + "经换算"提示组件）
- **开销记录页面** ✅ — `/vehicles/{id}/expenses` 开销列表（分页+筛选+统计摘要）+ 创建/编辑表单 + 详情页，10 种开销分类，保养提醒联动
- **深色模式** ✅ — 三种主题模式, CSS 变量体系, ECharts 暗色适配
- **响应式** ✅ — useIsMobile Hook, 全站 Table→卡片/Sider→Drawer 适配

### 待实现

| 功能 | 优先级 | 说明 |
|------|--------|------|
| 记录列表筛选 UI | ⭐ 低 | 后端 API 已支持筛选参数 |

---

## 4. 待实现功能汇总

### P1 (第二期)

| 功能 | 说明 |
|------|------|
| PWA 支持 | Service Worker + 离线 + 安装到桌面 |
| 多车对比图表 | 油耗/费用/里程对比 |
| 文件上传 | 车辆照片 + 头像 (OSS/本地) |
| 更多语言 | 韩语/繁中/西/德/法 |
| 数据导出 PDF | 带图表的可视化报告 |

### P2 (第三期)

| 功能 | 说明 |
|------|------|
| 微信小程序 (Taro) | 共享 shared 包 |
| 第三方登录 | Google / Apple / 微信 |
| 小票 OCR | 拍照识别加油小票 |
| 加油站地图 | PostGIS + 位置服务 |
| 无障碍访问 | WCAG 2.1 AA |

---

## 5. 已知问题

> 当前无未修复的已知问题 🎉
>
> 历史已修复问题（21 项）已归档，详见 Git 历史。

---

## 6. 变更日志

> 仅记录功能级别变更摘要，详细实现细节见 Git commit 历史。

### 2026-03-31

- ✅ **数据导出增强（P0~P2 全量实现）** — 后端 ExportService 从 3 个数据源扩展到 10 个（+开销记录/保养提醒/通知/邀请码/群组关系/共享车辆），handler 实现 CSV/ZIP/JSON 三种格式 + basic/full 两种范围；前端设置页增加范围+格式选择（Radio.Group）；三语 i18n 新增 7 键；API 文档/README 同步更新
- ✅ **维修保养开销记录模块（全栈）** — 独立车辆开销台账：后端 CRUD + 筛选 + 统计 + 商家建议 + 保养提醒联动（7 API），前端列表/表单/详情页，10 种分类，三语 i18n（~60 键）
- ✅ **群组页面国际化全面修复** — GroupPage 15+ 处硬编码单位/货币消除，新增 `<ConvertedCost>` 汇率换算组件 + 辅助函数（convertFuel/Distance/Efficiency），所有 Tab 按用户偏好动态转换
- ✅ **汇率换算扩展** — 记录详情/列表页新增汇率 Tag + Tooltip hover 换算
- 🔧 **单位切换 Bug 修复** — 后端 unit_price 同步容量单位换算，前端改用 record 级别字段
- ✅ **群组功能扩展（全栈）** — 车辆共享标记（3 API + 统一 `verifyVehicleAccess`）、排行榜（4 维度 + 时间范围）、费用看板（统计卡片 + 趋势 + 成员占比）、加油站推荐（聚合 + 筛选 + 价格趋势）
- ✅ **家庭群组管理（全栈）** — Group/GroupMember 模型 + CRUD + 邀请码 + 三级权限 + 数据汇总 + 前端 6 Tab 详情页（11 API）
- 🔧 **Bug 修复** — GroupPage 条件渲染、翻译错误、RecordFormPage UUID 显示

### 2026-03-30

- ✅ **通知与提醒系统** — 保养提醒（11 种类型 + 3 种触发）+ 异常油耗预警（>30% 偏差）+ NotificationBell 组件（60s 轮询 + 标记已读）+ 邀请码使用通知
- ✅ **移动端响应式适配** — 全站 useIsMobile Hook，Sider→Drawer、Table→卡片、表单/统计/仪表盘自适应
- ✅ **邀请码管理** — `/invites` 独立管理页：列表/创建/复制/启停/删除
- ✅ **邀请注册制** — `invite_only`/`open`/`closed` 三种模式，GT-XXXXXX 格式，SELECT FOR UPDATE 并发安全
- ✅ **日志系统** — Zap + Lumberjack 轮转（按大小切割 + gzip 压缩）
- ✅ **并发安全修复** — Refresh Token Rotation 原子消费、默认车辆设置事务、邮箱 unique violation 409

### 2026-03-30 (earlier)

- ✅ **GDPR 合规** — 账号注销 + CSV 数据导出（流式 UTF-8 BOM）+ 隐私政策/用户协议（三语）
- ✅ **多币种/单位换算** — 后端 `pkg/convert/` 引擎 + API 自动转换，前端工具函数
- ✅ **加油记录详情页** — 基本信息 + 智能分析（油耗评级/对比/利用率），EV 适配
- ✅ **统计页** — 按月/年维度切换 + 往年同比对比
- ✅ **深色模式** — Light/Dark/System + CSS 变量 + ECharts 暗色 + Ant Design token

### 2026-03-26

- ✅ **项目初始搭建** — Go 后端骨架 + 前端 Monorepo + Docker PostgreSQL
- ✅ **核心模块全栈实现** — Auth（JWT + Refresh Token）、User、Vehicle（含电动车）、FuelRecord（站点补全/三值计算）、Stats（趋势/聚合）
- ✅ **i18n 修复** — 21 项问题修复（硬编码中文/类型不匹配/Ant Design locale 联动等）
- ✅ **前后端 API 一致性审查** — 10 项问题修复
- ✅ **设置页时区** — 90 个 IANA 时区 + 时区感知日期显示（dayjs utc + timezone）
- ✅ **电动车全栈支持** — battery_capacity 字段 + 充电记录适配 + 电耗统计
