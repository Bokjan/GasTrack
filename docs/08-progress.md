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
| 数据导出 CSV（GDPR） | ✅ | ✅ | ✅ |
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
- **Export** ✅ — CSV 数据导出 (UTF-8 BOM, 流式写入)
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
  - 设置 (时区, 外观主题, 语言, 单位, 数据导出, 账号注销)
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
| ~~家庭群组~~ | ~~✅ 基础已完成：CRUD + 邀请 + 权限 + 数据汇总~~ |
| ~~车辆共享标记~~ | ~~✅ 已完成（全栈：3 API + 权限 + 前端 UI）~~ |
| ~~群组油耗排行榜~~ | ~~✅ 已完成（全栈：API + 前端排行榜 Tab）~~ |
| ~~群组费用统计看板~~ | ~~✅ 已完成（全栈：API + 前端费用 Tab）~~ |
| ~~加油站推荐共享~~ | ~~✅ 已完成（全栈：API + 前端加油站 Tab）~~ |
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
| ~~汇率参考~~ | ~~✅ 已完成（只读展示，frankfurter.app + 设置页/仪表盘/统计/记录详情/记录列表全覆盖）~~ |

---

## 5. 已知问题

> 当前无未修复的已知问题 🎉
>
> 历史已修复问题（21 项）已归档，详见 Git 历史。

---

## 6. 变更日志

> 仅记录功能级别变更摘要，详细实现细节见 Git commit 历史。

### 2026-03-31

- ✅ **维修保养等开销记录模块（全栈）** — 独立于燃油记录的车辆开销台账：
  - 后端：ExpenseRecord 模型 + CRUD + 列表筛选（分类/日期/关键词/金额区间）+ 统计（按币种汇总/分类占比/月度趋势/近30天大额）+ 商家名建议 + 保养提醒联动（完成记账后自动更新 Reminder 的 last_mileage/last_date 并重算 next 值），7 条 API
  - 前端：ExpenseListPage（分页+筛选+统计摘要）、ExpenseFormPage（分类/金额/商家/里程/提醒关联）、ExpenseDetailPage，侧边栏入口 + 车辆列表快捷入口
  - 10 种开销分类：maintenance/repair/insurance/parking/toll/car_wash/inspection/parts/fine/other
  - 共享车辆权限复用（verifyVehicleAccess）
  - 三语 i18n 支持（~60 翻译键）
- ✅ **群组页面单位/货币国际化全面修复** — GroupPage 15+ 处硬编码单位和货币符号消除：
  - 新增 `useExchangeRateStore` 汇率获取 + `formatConvertedCost` 智能换算（CNY→用户偏好币种）
  - 新增 `<ConvertedCost>` 组件：发生汇率换算时自动显示 `<InfoCircleOutlined>` + Tooltip "经换算"（三语 i18n：经换算/Converted/換算済み）
  - 新增 `convertFuel` / `convertDistance` / `convertEfficiency` 辅助函数
  - Overview Tab：total_cost 改用 `ConvertedCost`，total_fuel 改用 `fuelUnit`，avg_efficiency 改用 `efficiencyUnit`
  - Leaderboard Tab：group_avg 和 item.value 按 metric 类型动态转换（efficiency→`convertEfficiency`，distance→`convertDistance`，cost→`ConvertedCost`）
  - Expense Stats Tab：4 张统计卡片（`prefix:'¥'`→`ConvertedCost`，`suffix:'L'`→`fuelUnit`，`suffix:'km'`→`distanceUnit`，`suffix:'L/100km'`→`efficiencyUnit`）
  - 趋势表格：4 列 render 全部改为动态转换
  - 成员占比：`¥...L` 改为 `ConvertedCost` + `fuelUnit`
  - 加油站 Tab：内联 `user?.unit_system === 'imperial' ? 'gal' : 'L'` 简化为 `fuelUnit`，`user?.currency_code || 'CNY'` 简化为 `currency`
- ✅ **汇率换算扩展** — 记录详情页单价新增汇率 Tag 展示（带 /单位 后缀）；记录列表页单价+总价 Tooltip hover 换算（桌面端表格+移动端卡片）；引入 `useExchangeRateStore` + `getRateTooltip` 辅助函数
- 🔧 **单位切换 Bug 全面修复（L↔gal）** — 后端 `fuelRecordToResponse` 新增 `unit_price` 同步容量单位换算（单价 × 反向容量比率），修复偏好切换后单价与 fuel_unit 不一致的问题；前端 RecordListPage 所有列 render 改用 record 级别字段（`record.fuel_unit`/`record.distance_unit`/`record.currency_code`）替代全局 fuelUnit/distanceUnit；RecordDetailPage 智能分析区域改用 `record.fuel_unit`；GroupPage 加油站推荐消除硬编码 ¥ 和 /L（改用 `formatCurrency` + 动态单位）
- ✅ **家庭群组管理（基础）** — 全栈实现：Group/GroupMember 模型 + CRUD + 邀请码加入 (GF-XXXXXX) + 权限管理 (Owner/Admin/Member) + 数据汇总 Overview API + 前端群组详情页 (3 Tab) + 三语 i18n (~50 翻译键) + 11 条 API
- ✅ **共享车辆权限全面修复** — FuelRecord/Stats/Reminder/Vehicle 四个 Service 统一 `verifyVehicleAccess` 鉴权模式（先查 owner → 回退 shared），非车主只能编辑/删除自己创建的记录；goroutine 异步任务改用 `context.WithoutCancel` 避免 HTTP 请求结束后 context 被取消
- 🔧 **Bug 修复** — GroupPage "暂无群组" 提示始终显示（改为条件渲染）；数据汇总表格"群主"翻译错误→"车主"（新增 vehicleOwner 翻译键）；RecordFormPage 车辆名称显示 UUID（加载期间 Select value 置为 undefined）
- 📄 **群组功能扩展设计** — 新增 `docs/11-group-features-design.md` 详细设计文档，涵盖 4 个扩展功能：车辆共享标记 / 群组油耗排行榜 / 群组费用统计看板 / 加油站推荐共享（含数据模型/API 设计/SQL 查询/前端 UI/i18n/实施计划）

### 2026-03-30

- ✅ **通知与提醒系统** — 保养提醒 (Reminder) 全栈：11 种保养类型 + 3 种触发方式 + 自动计算下次保养；异常油耗预警：加油后异步检测 (goroutine)，偏差 >30% 自动生成通知；通知系统：NotificationBell 组件 + 60s 轮询 + 标记已读
- ✅ **邀请码使用通知** — 邀请码被消费后异步通知创建者，新增 `invite_used` 通知类型
- ✅ **移动端响应式适配** — 全站 10 个文件改动：useIsMobile Hook, Sider→Drawer, Table→卡片列表, 表单/统计/仪表盘自适应
- ✅ **邀请码管理页面** — `/invites` 独立管理页：列表/创建/复制/启停/删除，移动端卡片适配
- ✅ **邀请注册制** — 全栈实现：`invite_only`/`open`/`closed` 三种注册模式，GT-XXXXXX 格式邀请码，SELECT FOR UPDATE 并发安全
- ✅ **日志自动轮转** — Lumberjack 按大小切割 + 定期清理 + gzip 压缩
- ✅ **并发安全修复** — Refresh Token Rotation 原子消费, 默认车辆设置事务, 邮箱 unique violation 409

### 2026-03-30 (earlier)

- ✅ **GDPR 合规三件套** — 账号注销 + CSV 数据导出 (流式 UTF-8 BOM) + 隐私政策/用户协议页面 (三语)
- ✅ **多币种/单位换算** — 后端 `pkg/convert/` 引擎 + API 按用户偏好自动转换 (fuel_amount/odometer/efficiency)，前端 `convertFuelEfficiency` + `litersToGallons` 工具函数
- ✅ **加油记录详情页** — 基本信息 + 智能分析 (油耗评级/对比/利用率)，完整 EV 适配
- ✅ **统计页重写** — 按月/按年维度切换 + 往年同比对比 (灰色虚线/柱)
- ✅ **深色模式** — 三种主题 (Light/Dark/System) + CSS 变量体系 + ECharts 暗色 + Ant Design token 覆盖

### 2026-03-26

- ✅ **初始项目搭建** — Go 后端骨架 + 前端 Monorepo + Docker PostgreSQL
- ✅ **核心模块全栈实现** — Auth (JWT + Refresh Token), User, Vehicle (含电动车), FuelRecord (站点补全/燃油标号/三值自动计算), Stats (车辆统计/趋势/聚合)
- ✅ **i18n 全面修复** — 21 项 bug 修复 (硬编码中文/类型不匹配/Ant Design locale 联动等)
- ✅ **前后端 API 一致性审查** — 10 项问题修复
- ✅ **设置页时区选择器** — 90 个 IANA 时区，三语翻译
- ✅ **时区感知日期显示** — dayjs utc + timezone 插件
- ✅ **电动车全栈支持** — battery_capacity 字段 + 充电记录表单适配 + 电耗统计
