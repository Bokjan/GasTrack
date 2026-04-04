# GasTrack 需求完成进度

> **更新日期**: 2026-04-05
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
| 隐私政策 + 用户协议 + Footer | — | ✅ | ✅ |
| 保养提醒（里程/时间） | ✅ | ✅ | ✅ |
| 异常油耗预警 + 通知系统 | ✅ | ✅ | ✅ |
| 家庭群组管理（基础） | ✅ | ✅ | ✅ |
| 群组扩展（共享/排行/看板/加油站） | ✅ | ✅ | ✅ |
| 维修保养等开销记录 | ✅ | ✅ | ✅ |
| 后端单元测试 | ✅ | — | ✅ |
| PWA 支持（安装到桌面/离线缓存） | — | ✅ | ✅ |

**图例**: ✅ 完成 | 🔨 进行中 | 🔲 待实现

---

## 2. 后端进度

### 已完成模块

- **基础设施** ✅ — Go 1.22 + net/http, Viper 配置, Zap 日志 + Lumberjack 轮转, GORM + PostgreSQL, Docker Compose
- **中间件** ✅ — CORS, JWT Auth, 请求日志, Panic Recovery, Rate Limit (100 req/s), 中间件链
- **Auth** ✅ — 注册/登录/刷新/登出, Refresh Token Rotation (SELECT FOR UPDATE 原子消费), 邀请码注册制
- **User** ✅ — 资料 CRUD, 修改密码, 注销账号 (GDPR)
- **Vehicle** ✅ — CRUD (含电动车 battery_capacity), 默认车辆 (事务原子), 归档
- **FuelRecord** ✅ — CRUD, 分页, 站名建议, 油耗/电耗自动计算, `GetCostByCurrency` 按币种分组聚合费用
- **Stats** ✅ — 车辆统计, 全局总览（含开销汇总）, 油耗趋势, 按月/年聚合 + 同比, `costs_by_currency` 多币种费用明细, 开销按月/年聚合 + 同比 (`expense-period-stats` API)
- **Invite** ✅ — 邀请码 CRUD, GT-XXXXXX 格式, 并发安全消费
- **Export** ✅ — 数据导出 CSV/ZIP/JSON（scope=basic/full，10 个数据源，UTF-8 BOM，ZIP 多文件 + manifest.json）
- **Reminder** ✅ — 保养提醒 CRUD, 11 种保养类型, 3 种触发方式, 自动计算下次保养
- **Notification** ✅ — 通知 CRUD, 未读数, 标记已读, 异常油耗检测 (>30% 偏差), 保养到期检查, 邀请码使用通知
- **单位换算** ✅ — `pkg/convert/` 引擎, API 按用户偏好自动转换（含 unit_price 同步容量单位换算）
- **群组管理** ✅ — Group/GroupMember/SharedVehicle 模型, 19 条 API（基础 CRUD + 邀请码 + 权限管理 + 数据汇总 + 车辆共享 + 排行榜 + 费用看板 + 加油站推荐）
- **开销记录** ✅ — ExpenseRecord 模型, CRUD + 列表筛选 + 统计（按币种汇总/分类占比/月度趋势）+ 商家名建议 + 保养联动, 7 条 API
- **单元测试** ✅ — 71 个测试用例全部通过：Repository 接口提取（9 接口 ~114 方法）+ Service 接口提取（2 接口）+ go.uber.org/mock 生成 mock + 12 个 Service 全覆盖 + pkg 工具包（convert/apperror）覆盖

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
- **共享包** ✅ — Types (完全对齐后端 DTO), API 层 (Axios + 401 自动刷新), Zustand (auth/vehicle/theme), i18n (3 语), Constants, Utils (formatDateTime 时区感知, `sumConvertedCostsByCurrency` 多币种汇率换算)
- **页面** ✅:
  - 登录/注册 (邀请码实时校验)
  - 仪表盘 (按车辆分组统计, 多币种费用汇率换算, 加油费用+开销+综合总费用)
  - 车辆列表/表单 (含电动车适配)
  - 加油记录列表/表单/详情 (三值自动计算, 站点补全, 智能分析, Segmented Tab 切换开销)
  - 统计 (三 Tab: ⛽加油/💸开销/📊综合, 月/年维度 + 同比, ECharts, 多币种费用汇率换算)
  - 邀请码管理 (Table/卡片, 复制/启停/删除)
  - 保养提醒 (卡片式管理, 逾期标识)
  - 设置 (时区, 外观主题, 语言, 单位, 数据导出 CSV/ZIP/JSON + 范围/格式选择, 账号注销)
  - 隐私政策 / 用户协议
- **组件** ✅ — MainLayout (Sider/Drawer 自适应 + Footer), NotificationBell (60s 轮询, 手机端 Drawer/桌面端 Popover 双模式), InstallPrompt (底部浮动卡片, design tokens 主题适配), ProtectedRoute
- **群组页面** ✅ — `/groups` 群组列表 + 详情面板 (6 Tab) + 创建/加入/编辑弹窗；全面单位/货币国际化（汇率自动换算 + "经换算"提示组件）
- **开销记录页面** ✅ — `/vehicles/{id}/expenses` 开销列表（分页+筛选+统计摘要）+ 创建/编辑表单 + 详情页，10 种开销分类，保养提醒联动
- **深色模式** ✅ — 三种主题模式, CSS 变量体系, ECharts 暗色适配
- **响应式** ✅ — useIsMobile Hook, 全站 Table→卡片/Sider→Drawer 适配
- **PWA** ✅ — vite-plugin-pwa, Web App Manifest, Service Worker (Workbox), 预缓存 + 运行时缓存, 自动更新提示, 安装引导 (Android/Desktop + iOS Safari), GT 品牌图标

### 待实现

| 功能 | 优先级 | 说明 |
|------|--------|------|
| 记录列表筛选 UI | ⭐ 低 | 后端 API 已支持筛选参数 |

---

## 4. 待实现功能汇总

### P1 (第二期)

| 功能 | 说明 |
|------|------|
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
> 历史已修复问题（22 项）已归档，详见 Git 历史。

---

## 6. 变更日志

> 仅记录功能级别变更摘要，详细实现细节见 Git commit 历史。

### 2026-04-05

- 🚀 **开销记录入口优化** — 在加油记录页和开销记录页之间新增 `Segmented` Tab 切换器，从任一页面可一键切换到另一页面，解决开销记录入口隐蔽的问题；同时为车辆列表页的钱包图标按钮增加 `Tooltip` 提示
- 🚀 **统计页三 Tab 改造（加油/开销/综合）** — Stats 页新增三个独立视角：⛽ 加油（原有指标+图表不变）、💸 开销（总额/分类/月度趋势+饼图）、📊 综合（加油+开销堆叠柱状图+合并汇总卡片）；后端新增 `GET /vehicles/:id/expense-period-stats` 接口支持开销按月/年聚合+同比；Dashboard 统计卡片新增开销费用和综合总费用展示
- 🚀 **Dashboard 开销数据接入** — 后端 `GET /stats/overview` 响应新增 `total_expense_cost` / `expense_costs_by_currency` 字段（VehicleStats 和 OverviewStats 均扩展）；Dashboard 单车卡片从 4 张扩为 6 张（⛽加油费/💸开销/综合总费用/里程/油耗/记录数），多车全局概览从 2 张扩为 4 张

### 2026-04-03

- 🔧 **i18n 漏翻专项修复** — 修复设置页汇率参考卡片 2 处国际化问题（`ja-JP` 缺失 `exchangeRate.currencyColumnName`、汇率数值列标题改为走 i18n key），并补齐三语语言包差异项（`invite.maxUsesPlaceholder`、通知结构化 message/direction keys），同时去掉 `exchangeRateStore` 的英文硬编码兜底，错误态统一由页面翻译文案承接

### 2026-04-01

- ✅ **后端单元测试全面补齐** — 从零开始为 Go server 端补齐完整单元测试体系（71 个测试用例）：
  - **接口提取** — 新增 `repository/interfaces.go`（9 个 Repository 接口，~114 个方法）和 `service/interfaces.go`（InviteServicer + NotificationServicer 接口）
  - **依赖注入重构** — 全部 11 个 Service struct 从具体 Repository 指针改为接口类型，支持 mock 测试
  - **Mock 生成** — 使用 go.uber.org/mock (mockgen) 自动生成 repository/mock/ 和 service/mock/ 两组 mock 实现
  - **Service 层测试（44 个）** — 覆盖全部 12 个 Service：AuthService（Register/Login/RefreshToken/Logout）、UserService、VehicleService、FuelRecordService、StatsService、InviteService、ExportService、ReminderService、NotificationService（含异常油耗检测/保养提醒触发）、GroupService（CRUD/加入/权限校验）、ExpenseRecordService、ExchangeRateService（httptest 模拟外部 API）
  - **工具包测试（27 个）** — `pkg/convert`（单位换算函数 + 往返一致性）和 `pkg/apperror`（错误构造 + errors.Is/As 兼容）
- 🔧 **慢查询优化 & 缺失索引补充** — 全面扫描 repository 层 SQL 复杂度，修复 8 个慢查询问题：
  - **添加 3 个缺失索引** — `group_members(user_id)`（几乎所有群组查询通过 `gm.user_id` JOIN，复合主键无法走前缀索引）、`fuel_records(station_name)`（4 个加油站相关查询无索引）、`shared_vehicles(shared_by)`（按共享人查询无索引）
  - **重写 `GetGroupVehicleSummary` 关联子查询** — 将 `COALESCE((SELECT ... WHERE fr2.vehicle_id = v.id ...), ...)` 关联子查询改为 `LEFT JOIN LATERAL`，避免对每行车辆执行 O(N×M) 子查询
  - **`GetGroupExpenseByYear` 添加时间下界** — 原无 WHERE 日期限制导致全量扫描 `fuel_records` 表；新增最近 10 年的 cutoff 过滤
  - **`IsVehicleSharedToUser` COUNT→EXISTS** — `COUNT(*)` 改为 `SELECT EXISTS(... LIMIT 1)`，找到第一行即短路返回
  - **`ExistsByInviteCode`/`ExistsSharedVehicle`/`ExistsByEmail`/`ExistsByCode` COUNT→EXISTS** — 同上优化模式，4 处存在性检查全部改为 EXISTS 短路
  - **`ListSharedVehiclesForUser` DISTINCT→EXISTS** — 将 4 表 JOIN + `SELECT DISTINCT` 改为 `EXISTS` 子查询过滤，减少中间结果集大小
  - **`GetLeaderboard`/`GetGroupStationStats` SQL 拼接安全化** — `switch` + 字符串拼接 ORDER BY 改为白名单 `map[string]string` 映射模式
- 🔧 **Server 代码审查：性能优化与 Bug 修复** — 全面 review 后端 67 个 Go 文件，修复以下问题：
  - **GORM 共享查询状态 Bug（2 处）** — `fuel_record.go` 和 `expense_record.go` 的 `ListByVehicle` 方法中，同一 `*gorm.DB` 链先执行 `Count()` 再执行 `Find()` 导致内部状态污染、查询结果异常；修复为使用独立查询链分别执行 COUNT 和分页查询
  - **RateLimiter 内存泄漏 Bug** — `ratelimit.go` 中 IP-to-Limiter map 只增不减，长期运行下内存无限增长；修复：引入 `ipLimiter` 结构体追踪 `lastSeen` 时间戳，启动后台 goroutine 每 5 分钟清理 10 分钟无活动的条目
  - **N+1 数据库查询优化（8 处）** — 消除 `GroupService`（`buildGroupResponse`/`GetOverview`/`GetLeaderboard`/`GetExpenseStats`/`GetStationStats`/`ListSharedVehicles`）、`VehicleService.List`、`StatsService.GetOverview` 中的 N+1 查询；新增 `UserRepository.GetByIDs`、`VehicleRepository.GetByIDs`、`FuelRecordRepository.GetMultiVehicleStats`、`FuelRecordRepository.GetMultiVehicleCostByCurrency` 四个批量查询方法，将循环内逐条 SELECT 替换为 `WHERE ... IN` 批量查询
- ✅ **仪表盘/统计页多币种费用换算 Bug 修复（全栈）** — 后端 Stats API 原直接 `SUM(total_cost)` 不区分币种，多币种用户金额显示错误；修复：Repository 新增 `GetCostByCurrency` 按 `currency_code` 分组聚合；DTO 新增 `costs_by_currency` 字段；前端新增 `sumConvertedCostsByCurrency()` 工具函数按汇率换算后汇总；`DashboardPage` 和 `StatsPage` 费用展示均改用换算后金额
- 🔧 **通知面板手机端适配** — `NotificationBell` 手机端改用 `Drawer` 顶部滑出替代 `Popover`，解决竖屏溢出；桌面端保持 Popover；内容区 maxHeight 响应式
- 🎨 **PWA 安装引导 UI 重设计** — `InstallPrompt` 重构为底部浮动卡片，design tokens 适配深色/浅色主题，带 `slideUp`/`slideDown` CSS 动画
- 🔧 **法律页面滚动修复** — 隐私政策/用户协议页面添加 `scrollTo(0, 0)`，确保从设置页跳转后从顶部开始

### 2026-03-31

- ✅ **PWA 支持** — `vite-plugin-pwa` 集成：Web App Manifest、Workbox Service Worker（预缓存 + 运行时缓存）、自动更新提示、安装引导（Android/Desktop + iOS Safari）、三语 i18n
- 🎨 **品牌升级** — GasTrack GT 图标设计（蓝色渐变 + 白色字母），SVG 矢量 + 5 尺寸 PNG 自动生成，替换旧 Vite 默认图标
- 📝 **名称与描述更新** — 系统定位扩展为"车辆能耗与费用管理"，更新 `index.html`/`package.json`/`README.md`/三语 i18n
- 🔧 **移动端表单聚焦自动缩放修复** — iOS Safari input 聚焦放大问题，viewport `maximum-scale=1` + 全局表单 `font-size: 16px`
- 🔧 **全栈审计修复（11 项）** — 消除硬编码值/魔数/i18n 缺失（排行榜日期格式、通知结构化 key、汇率表头、邀请码 placeholder、货币 labels、convert 包常量等）
- 🔧 **Ant Design 弃用 API 修复** — `destroyOnClose` → `destroyOnHidden`（5 处）、`overlayInnerStyle` → `styles.body`、`bodyStyle`/`headerStyle` → `styles`
- 🔧 **群组多币种 Bug 修复（3 项）** — 费用看板 SQL 混合汇总、加油站价格未换算、成员费用占比除零；全部修复并新增前端 `sumConvertedCosts()` 辅助函数
- 🔧 **群组车辆汇总币种 Bug 修复** — `currency_code` 改为从 `fuel_records` 取实际入账币种，避免用户切换偏好后错误展示
- 🧹 **包管理清理** — 删除多余 `package-lock.json`，移除废弃依赖，`engines.pnpm` → `>=9.0.0`
- 🗑️ **文档清理** — 移除已完成的群组设计文档（757 行），同步清理引用
- ✅ **数据导出增强** — 后端从 3 扩展到 10 个数据源，支持 CSV/ZIP/JSON 三种格式 + basic/full 两种范围
- ✅ **维修保养开销记录模块（全栈）** — 后端 CRUD + 筛选 + 统计 + 保养联动（7 API），前端列表/表单/详情页，10 种分类
- ✅ **群组页面国际化修复** — 15+ 处硬编码消除，新增 `<ConvertedCost>` 汇率换算组件
- ✅ **汇率换算扩展** — 记录详情/列表页新增汇率 Tag + Tooltip hover 换算
- 🔧 **单位切换 Bug 修复** — 后端 unit_price 同步容量单位换算
- ✅ **群组功能扩展（全栈）** — 车辆共享（3 API）、排行榜（4 维度）、费用看板、加油站推荐
- ✅ **家庭群组管理（全栈）** — Group/GroupMember 模型 + CRUD + 邀请码 + 三级权限 + 数据汇总（11 API）

### 2026-03-30

- ✅ **通知与提醒系统** — 保养提醒（11 种类型 + 3 种触发）+ 异常油耗预警 + NotificationBell 组件（60s 轮询）+ 邀请码使用通知
- ✅ **移动端响应式适配** — 全站 useIsMobile Hook，Sider→Drawer、Table→卡片、表单/统计/仪表盘自适应
- ✅ **邀请码管理** — `/invites` 独立管理页：列表/创建/复制/启停/删除
- ✅ **邀请注册制** — `invite_only`/`open`/`closed` 三种模式，GT-XXXXXX 格式，并发安全
- ✅ **日志系统** — Zap + Lumberjack 轮转
- ✅ **并发安全修复** — Refresh Token Rotation 原子消费、默认车辆设置事务、邮箱 unique violation 409
- ✅ **GDPR 合规** — 账号注销 + 数据导出（流式 UTF-8 BOM）+ 隐私政策/用户协议（三语）
- ✅ **多币种/单位换算** — 后端 `pkg/convert/` 引擎 + API 自动转换
- ✅ **加油记录详情页** — 基本信息 + 智能分析（油耗评级/对比/利用率），EV 适配
- ✅ **统计页** — 按月/年维度切换 + 往年同比对比
- ✅ **深色模式** — Light/Dark/System + CSS 变量 + ECharts 暗色 + Ant Design token

### 2026-03-26

- ✅ **项目初始搭建** — Go 后端骨架 + 前端 Monorepo + Docker PostgreSQL
- ✅ **核心模块全栈实现** — Auth（JWT）、User、Vehicle（含电动车）、FuelRecord（三值计算）、Stats（趋势/聚合）
- ✅ **i18n 修复** — 21 项问题修复
- ✅ **前后端 API 一致性审查** — 10 项问题修复
- ✅ **设置页时区** — 90 个 IANA 时区 + 时区感知日期显示
- ✅ **电动车全栈支持** — battery_capacity 字段 + 充电记录适配 + 电耗统计
