# GasTrack 需求分析

> **最后更新**: 2026-03-31
>
> **状态说明**: ✅ 已完成 | 🔨 进行中 | 🔲 待实现

## 1. 项目背景与目标

GasTrack 是一款面向全球用户的油耗/电耗记录与分析系统：
- 记录每次加油/充电的详细信息（油量/电量、费用、里程等）
- 追踪车辆油耗/电耗趋势，辅助用车决策
- 多车辆管理（燃油车 & 电动车），满足家庭或车队需求
- 支持不同国家/地区的计量单位、货币和语言

## 2. 用户角色

| 角色 | 描述 |
|------|------|
| 普通用户 | 注册登录后记录加油数据、查看报表 |
| 家庭管理员 | 创建家庭/群组，邀请成员，查看群组内车辆数据 (P1) |
| 系统管理员 | 后台管理、用户管理、数据统计 (P2) |

## 3. 核心功能模块

### 3.1 用户与认证 (P0)
- ✅ 邮箱注册/登录
- ✅ 个人资料：昵称、时区（90 个 IANA 时区可搜索选择）、偏好语言、默认币种和计量单位
- ✅ 修改密码
- ✅ 账号注销（GDPR 合规数据删除）— *原定 P1，已提前完成*
- 🔲 忘记密码（邮件重置）— 后端 `ForgotPasswordRequest` DTO 已定义，Service/Handler 逻辑未实现
- 🔲 头像上传 (P1) — 需文件上传后端支持
- 🔲 第三方登录（Google / Apple / 微信）(P2)

### 3.2 车辆管理 (P0)
- ✅ 添加车辆：品牌、型号、年份、燃油/能源类型（汽油/柴油/混动/电动）
- ✅ 车辆类型：汽车(car)、摩托车(motorcycle)、其他(other)
- ✅ 燃油/混动车辆：油箱容量(L/gal)、排量(cc)
- ✅ 电动车辆：电池容量(kWh)，排量字段自动隐藏
- ✅ 编辑/删除车辆（软删除，支持 archived 过滤）
- ✅ 多车辆切换（支持默认车辆设置）
- ✅ 车牌号录入与展示（表单录入、列表/卡片显示，非必填，`maxLength=20`）
- 🔲 车辆照片上传 (P1) — 需文件上传后端支持
- 🔲 车辆里程校准 (P2)

### 3.3 加油/充电记录 (P0 - 核心功能)
- ✅ 记录加油/充电信息：日期时间、加油站/充电站、油量/电量、单价、总费用、当前里程
- ✅ 燃油车：加油量(L/gal)、加油站、是否加满、燃油标号选择
- ✅ 电动车：充电量(kWh)、充电站、是否充满
- ✅ 表单根据车辆能源类型自动切换 UI 语义和单位
- ✅ 加油量/单价/总费用任意两项自动计算第三项（基于编辑栈追踪用户操作顺序）
- ✅ 站点名称自动补全（基于用户历史加油/充电站名，按频次降序，最多 20 条）
- ✅ 记录的编辑/删除
- ✅ 记录列表：按时间排序，支持分页（page/page_size）
- ✅ 记录列表日期列 hover 显示精确时间（HH:mm），基于用户时区设置的时区感知格式化
- 🔲 记录列表筛选（按日期范围/站点）— 后端 API 已支持，前端未实现筛选 UI
- 🔲 支持手动输入和拍照识别加油小票 (P2)

### 3.4 油耗/电耗统计与报表 (P0)
- ✅ 燃油车油耗计算：L/100km、km/L 或 MPG（根据用户偏好）
- ✅ 电动车电耗计算：kWh/100km、km/kWh 或 mi/kWh
- ✅ 能耗趋势图：折线图展示油耗/电耗变化
- ✅ 费用统计：按月/按年维度切换，支持年份选择
- ✅ 费用趋势图：柱状图展示费用变化
- ✅ 往年同期对比：按月模式下自动叠加显示去年同期数据（灰色虚线/柱）
- ✅ 多维度统计卡片：费用、平均油耗/电耗、总里程、加油/充电次数
- ✅ 仪表盘按车辆维度独立展示统计（多车不混合汇总，油车/电车分别显示）
- ✅ 前端单位换算展示 — 后端 API 已按用户偏好完成全部换算（油耗、加油量、里程），前端使用 `convertFuelEfficiency` 做 Tooltip 多单位展示 + `litersToGallons` 处理 tank_capacity
- 🔲 多车辆对比图表 (P1)
- ✅ 数据导出 CSV/ZIP/JSON — GDPR 数据可携带权，`GET /api/v1/users/me/export?format=csv|zip|json&scope=basic|full`（详见 4.4.2）
- 🔲 数据导出 PDF (P2)

### 3.5 多语言支持 (P0)
- ✅ 第一期：中文简体、英文、日语（前端已完成）
- ✅ 语言自动检测（浏览器/系统语言）+ 手动切换
- ✅ 语言偏好同步保存至后端用户设置（切换时自动调用 API 持久化）
- ✅ Ant Design 组件库内置文本（分页/日期选择器等）随语言联动切换
- 🔲 后端错误消息国际化（go-i18n TOML 翻译文件，目前错误消息英文硬编码）
- 🔲 第二期：韩语、繁体中文、西班牙语、德语、法语

### 3.6 多国家/地区与多币种 (P0)
- ✅ 支持主要国家/地区设置
- ✅ 燃油车计量单位（三种体系）：
  - 公制(欧标)：升(L)、公里(km)、L/100km（中国、欧洲等）
  - 公制(日标)：升(L)、公里(km)、km/L（日本等）
  - 英制：加仑(gal)、英里(mi)、MPG（美国、英国等）
- ✅ 电动车计量单位（三种体系）：
  - kWh/100km（欧标电耗）
  - km/kWh（日标电耗）
  - mi/kWh（英制电耗）
- ✅ 后端单位换算引擎（`server/internal/pkg/convert/`）
- ✅ 前端单位换算工具函数（`shared/src/utils/index.ts`）：L↔gal、km↔mi、油耗三体系互转
- ✅ 多币种支持：CNY、USD、EUR、JPY、GBP、KRW
- ✅ 加油表单支持选择燃油单位(L/gal/kWh)、货币、里程单位(km/mi)
- ✅ 设置页支持用户偏好单位系统(metric/imperial)、能耗单位、货币的配置与保存
- ✅ 前端根据用户偏好单位自动展示 — 后端 `fuelRecordToResponse` 按用户偏好转换（fuel_amount/odometer/trip_distance/fuel_efficiency/unit_price），Stats API 同步转换；前端列表/详情/卡片使用 record 级别字段（`record.fuel_unit`/`record.distance_unit`/`record.currency_code`）；GroupPage 全页面动态单位（`fuelUnit`/`distanceUnit`/`efficiencyUnit`）+ 动态货币（`formatCurrency` + `currency`）
- ✅ 汇率参考（只读展示）(P2) — 后端 `ExchangeRateService`（frankfurter.app + 内存缓存 24h TTL），`GET /api/v1/exchange-rates`；设置页参考换算币种 + 汇率表，仪表盘/统计/记录详情/记录列表全覆盖
- ✅ **群组页面货币换算** — 群组费用自动按用户偏好币种换算（`<ConvertedCost>` 组件 + "经换算"Tooltip），油耗/油量/距离单位按 metric/imperial 动态转换

### 3.7 深色模式 (P0) — *需求文档新增*
- ✅ 三种主题模式：浅色（Light）、深色（Dark）、跟随系统（System，默认值）
- ✅ 实时监听系统 `prefers-color-scheme` 偏好，自动切换
- ✅ 用户主题偏好持久化至 `localStorage`
- ✅ 全站 CSS 变量体系（`--gt-bg-body`、`--gt-bg-card`、`--gt-text-primary` 等）
- ✅ Ant Design ConfigProvider 组件级暗色 token 覆盖
- ✅ ECharts 图表暗色模式适配
- ✅ 设置页外观主题切换控件（Segmented 三选一）

### 3.8 邀请注册制与邀请码管理 (P0)
- ✅ 邀请注册制：注册策略可配置（`invite_only` / `open` / `closed`）
- ✅ 注册时邀请码实时校验（debounce 500ms + ✅/❌ 状态反馈）
- ✅ 邀请码格式 `GT-XXXXXX`（6 位大写字母+数字，去除 I/O/0/1 避免混淆）
- ✅ 支持单次码（`max_uses=1`）和多次码（`max_uses=N`，0=不限）
- ✅ 支持过期时间设置（默认 30 天）和手动启用/停用
- ✅ 独立邀请码管理页面 `/invites`（侧边栏入口）：
  - 邀请码列表（状态/使用情况/过期时间/备注）
  - 创建邀请码弹窗（设置次数/过期时间/备注）
  - 一键复制邀请码到剪贴板
  - Switch 切换启用/停用
  - 删除（带二次确认）
- ✅ 并发安全：`SELECT FOR UPDATE` + 事务原子操作

### 3.9 移动端响应式适配 (P0) — *已完成*
- ✅ 全站移动端响应式适配，覆盖所有页面
- ✅ `useIsMobile()` Hook（基于 `matchMedia('max-width: 767px')`）
- ✅ MainLayout：移动端 Sider → Drawer 抽屉导航 + hamburger 菜单按钮
- ✅ RecordListPage：移动端 Table → 卡片列表
- ✅ RecordDetailPage：Descriptions/Tag/Insights 缩小适配
- ✅ StatsPage：筛选条件独立行、gutter 缩小
- ✅ DashboardPage：统计卡片 gutter 缩小
- ✅ InviteManagePage：移动端 Table → 卡片列表
- ✅ VehicleFormPage：表单全宽适配
- ✅ global.css：Card 内边距缩小、Statistic 字号/间距优化

### 3.10 家庭/群组管理 (P1) — *全部完成* ✅
- ✅ 创建家庭群组 — Group/GroupMember 模型 + Repository/Service/Handler 全链路，邀请码格式 `GF-XXXXXX`
- ✅ 邀请码邀请成员加入 — 通过群组邀请码加入，`SELECT FOR UPDATE` + 事务并发安全，成员上限检查
- ✅ 群组内车辆数据汇总查看 — Overview API 聚合所有成员车辆的加油记录（总费用/总油量/平均油耗）
- ✅ 成员权限管理（Owner/Admin/Member 三级角色）— Owner 可管理角色/移除成员，Admin 可移除普通成员
- ✅ 前端群组管理页面 `/groups` — 群组列表卡片 + 详情面板（群组信息/成员管理/数据汇总 Tabs）+ 创建/加入/编辑弹窗
- ✅ 三语 i18n 支持（zh-CN/en-US/ja-JP，~50 翻译键）
- ✅ 邀请码重新生成、退出群组、删除群组（群主）
- ✅ 19 条 API 路由：群组 CRUD(7) + 成员管理(3) + 数据汇总(1) + 车辆共享(3) + 排行榜(1) + 费用看板(1) + 加油站推荐(1) + 扩展预留(2)

#### 3.10.1 车辆共享标记 (P1) — *已完成* ✅
- ✅ 群组内车主可将自己的车标记为"共享"，群组其他成员可为共享车辆记录加油 — 前端 GroupPage 概览表格 Switch 切换共享状态
- ✅ 新增 `shared_vehicles` 关联表（group_id + vehicle_id 联合唯一约束）
- ✅ 加油表单车辆选择器支持显示共享车辆（`include_shared=true`，分组：我的车辆 / 共享车辆(来自XX群组)）
- ✅ 权限控制：后端 FuelRecord/Stats/Reminder/Vehicle 四个 Service 统一 `verifyVehicleAccess` 鉴权（先查 owner → 回退 shared），非车主只能编辑/删除自己创建的记录
- ✅ 3 条 API：`POST /groups/{id}/shared-vehicles`（共享）、`DELETE /groups/{id}/shared-vehicles/{vid}`（取消共享）、`GET /groups/{id}/shared-vehicles`（列表）

#### 3.10.2 群组油耗排行榜 (P1) — *已完成* ✅
- ✅ 四维排行：油耗（L/100km，ASC）、费用（DESC）、里程（DESC）、加油频次（DESC）
- ✅ 时间范围：本月/上月/近3月/今年
- ✅ 排行按"成员×车辆"粒度，至少 2 条记录才参与排行（`HAVING COUNT >= 2`）
- ✅ 前三名 🥇🥈🥉 标识，自己高亮（蓝色背景 + ✦ 徽章），显示相比群组平均值的差异百分比
- ✅ 前端按用户偏好动态转换：效率值用 `convertFuelEfficiency`、距离用 `convertDistance`、费用用 `ConvertedCost`（含汇率换算提示）
- ✅ 1 条 API：`GET /groups/{id}/leaderboard?metric=efficiency&period=current_month`

#### 3.10.3 群组费用统计看板 (P1) — *已完成* ✅
- ✅ 顶部 4 张统计卡片（总费用/总加油量/总里程/平均油耗）+ 环比变化百分比（▲/▼ 百分比标识）
- ✅ 费用趋势表格：按月/按年维度，含 `by_member` 成员费用分解，支持上年同期对比
- ✅ 成员费用占比列表（`member_breakdown`，含百分比）
- ✅ 所有数值按用户偏好单位动态展示：费用含汇率换算 + "经换算"提示，油量/距离/效率按 metric/imperial 动态转换
- ✅ 1 条 API：`GET /groups/{id}/expense-stats?period=month&year=2026`

#### 3.10.4 加油站推荐共享 (P1) — *已完成* ✅
- ✅ 聚合群组成员加油记录中的加油站数据（站名/平均油价/最新油价/加油次数/常客/价格趋势↑↓→/燃油标号）
- ✅ 支持按燃油标号筛选、按频次/油价/日期排序
- ✅ 支持 3/6/12 个月数据范围选择（默认 6 个月）
- ✅ 1 条 API：`GET /groups/{id}/stations?fuel_grade=95&months=6&sort_by=frequency`

> 📄 详细设计文档见 [`11-group-features-design.md`](./11-group-features-design.md)

### 3.11 通知与提醒 (P2)
- ✅ 保养提醒（按里程/时间）— 后端 Reminder CRUD + 11 种保养类型（oil_change/tire_rotation/brake_pads 等）+ 三种触发方式（mileage/time/both）+ 自动计算下次保养时间；前端 `/reminders` 页面卡片式管理（创建/编辑/删除/启用禁用）+ 逾期标识
- ~~🔲 定期加油提醒~~
- ✅ 异常油耗预警 — 加油记录创建后异步检测（goroutine），当本次油耗偏离历史平均值 >30% 时自动生成通知；通知铃铛组件（Header NotificationBell）+ 60s 轮询未读数 + 标记已读/全部已读
- ✅ 邀请码使用通知 — 邀请码被消费时异步通知创建者（新增 `invite_used` 通知类型，`VehicleID` 可选）

## 4. 非功能性需求

### 4.1 性能
- 页面首屏加载 < 2s
- API 响应时间 < 200ms（P95）
- 支持 1 万并发用户

### 4.2 安全
- ✅ JWT Token + Refresh Token 认证（含 401 自动刷新队列、Refresh Token Rotation 原子消费）
- ✅ 密码 bcrypt 加密存储（cost=12）
- ✅ 接口限流（Rate Limiting）— IP 级别令牌桶，100 req/s
- ✅ Panic Recovery 中间件
- ✅ CORS 配置（可配置允许的 Origins）
- ✅ 数据校验（go-playground/validator）
- ✅ 日志系统（Zap 结构化日志 + Lumberjack 自动轮转、文件持久化、gzip 压缩）
- ✅ 优雅关闭（信号监听 + Shutdown 超时）
- ✅ 并发安全（Refresh Token Rotation、默认车辆设置、邀请码消费均使用 `SELECT FOR UPDATE` + 事务）
- 🔲 HTTPS 全链路加密（部署时配置）
- 🔲 SQL 注入/XSS 防护审计

### 4.3 可用性
- ✅ 响应式设计（PC + 移动端）— 全站移动端适配，Sider→Drawer、Table→卡片、表单/统计自适应
- 🔲 PWA 支持，可离线使用基本功能 (P1)
- 🔲 无障碍访问（WCAG 2.1 AA）(P2)

### 4.4 数据合规

> 本节涵盖 GDPR（欧盟通用数据保护条例）及其他数据保护法规的合规要求。
> GasTrack 面向全球用户，需满足数据删除权、数据可携带权、知情同意权等核心要求。

#### 4.4.1 数据删除权（Right to Erasure）✅
- ✅ 账号注销功能：用户可自助删除账户及所有关联数据
- ✅ 设置页 Popconfirm 二次确认，防误操作
- ✅ 注销后自动登出并跳转登录页

#### 4.4.2 数据可携带权（Right to Data Portability）✅
- ✅ **已全部实现（P0 基础版 + P1 完整数据版 + P2 JSON 导出）**
  - **后端 API**：`GET /api/v1/users/me/export`（需认证）
    - 支持两个查询参数：`format`（`csv` / `zip` / `json`，默认 `csv`）、`scope`（`basic` / `full`，默认 `basic`）
    - **scope=basic**（P0 基础版）：用户资料、车辆列表、全部加油/充电记录
    - **scope=full**（P1 完整版）：在基础版基础上增加开销记录、保养提醒、通知、邀请码、群组关系（成员身份 + 共享车辆）
    - **format=csv**：三段式单文件 CSV（UTF-8 BOM），仅在 scope=basic 时可用；scope=full 时自动升级为 zip
    - **format=zip**：多文件分类导出 ZIP 容器（profile.csv, vehicles.csv, fuel_records.csv, expense_records.csv, reminders.csv, notifications.csv, invite_codes.csv, groups.csv, manifest.json）
    - **format=json**：结构化 JSON（带 meta 元数据：export_time/version/scope/modules/record_counts），方便二次处理
    - 实现细节：流式写入（`encoding/csv` Writer / `archive/zip` Writer / `encoding/json` Encoder）+ UTF-8 BOM，兼容 Excel/Numbers/Google Sheets
    - 返回：二进制流 + `Content-Disposition: attachment`，文件名含时间戳（如 `gastrack-export-20260331-154100.zip`）
  - **前端触发**：设置页"数据与隐私"卡片
    - 导出范围选择：基础数据 / 完整数据（Radio.Group 按钮组）
    - 文件格式选择：CSV / ZIP / JSON（Radio.Group 按钮组，scope=full 时 CSV 禁用自动升级为 ZIP）
    - 每个选项附带 Hint 说明（三语 i18n 支持）
    - 调用 `userApi.exportData({ format, scope })`（Axios `responseType: 'blob'`）
    - 浏览器端 Blob URL + 动态 `<a>` 元素触发下载
    - 从 `Content-Disposition` 响应头提取文件名（含 regex 回退）
    - 提供 Loading 状态与成功/失败提示

  - **向后兼容**：无参数调用 `GET /users/me/export` 仍返回原版三段式 CSV，现有功能零影响

  - **边界与合规约束**：
    - 仅导出**当前用户自身数据**或与其直接相关的业务关系数据
    - 不导出密码哈希、Refresh Token 哈希、内部审计日志、风控信息等敏感内部字段
    - 群组相关导出仅包含当前用户的成员身份和共享车辆关系，不泄露其他成员隐私
    - 所有时间字段统一使用 ISO 8601 / RFC3339 格式，避免跨时区歧义
    - 字段名稳定，结构可预期

  - **验收标准**（均已满足）：
    - ✅ 用户可在 3 步内完成导出操作并成功下载文件
    - ✅ 导出文件可被 Excel、Numbers、Google Sheets 或常见压缩/解析工具正常打开
    - ✅ 导出结果覆盖全部业务模块（10 个数据源），且不包含密码、Token、内部日志等敏感字段
    - ✅ 字段结构稳定、时间格式统一、跨语言环境下不出现乱码

  - **后续可选扩展**（P2+）：
    - 导出历史摘要（最后导出时间、导出版本）
    - 大数据量异步导出任务模式（生成任务 → 后台打包 → 用户下载）
    - PDF 可视化报告（附加能力，不替代结构化原始数据导出）

#### 4.4.3 知情同意权（Right to be Informed）✅
- ✅ **隐私政策页面**（Privacy Policy）
  - 前端页面 `/privacy`（公开路由，无需登录）
  - 内容涵盖：收集的数据类型、数据用途、数据存储与安全、第三方共享、用户权利、localStorage 说明、联系方式
  - 支持中文/英文/日语三语版本（i18n `privacy.*` 键，根据当前语言自动切换）
  - 注册页底部添加"注册即表示同意《隐私政策》"文案及链接
- ✅ **用户协议页面**（Terms of Service）
  - 前端页面 `/terms`（公开路由，无需登录）
  - 内容涵盖：接受条款、服务范围、账号管理、用户责任、免责声明、服务终止、协议变更、联系方式
  - 同样支持三语切换（i18n `terms.*` 键）
  - 注册页底部与隐私政策并列显示
- ✅ **Cookie/本地存储说明**
  - 在隐私政策页中合并说明 `access_token`/`refresh_token`、`locale`、`theme_mode` 等 localStorage 项

#### 4.4.4 数据最小化原则 ✅
- ✅ 车辆和记录数据仅存储业务必需字段
- ✅ 密码使用 bcrypt 单向哈希存储，不存明文
- ✅ Refresh Token 仅存 SHA-256 哈希，不存原始值

## 5. 优先级说明

| 优先级 | 含义 | 时间范围 |
|--------|------|----------|
| P0 | 第一版必须实现 | 第一期（8-10周） |
| P1 | 重要但可延后 | 第二期（4-6周） |
| P2 | 锦上添花 | 第三期及以后 |

## 6. 第一期 MVP 剩余工作

> 以下为完成第一期 MVP 闭环所需的剩余任务，按建议优先级排序。

| # | 任务 | 说明 | 优先级 |
|---|------|------|--------|
| ~~1~~ | ~~**响应式适配（移动端）**~~ | ~~✅ 已完成：全站 Sider→Drawer、Table→卡片、表单/统计自适应~~ | ~~⭐⭐⭐~~ |
| ~~2~~ | ~~**前端单位换算完善**~~ | ~~✅ 已完成：后端 API 已按用户偏好完成 fuel_amount/odometer/trip_distance/fuel_efficiency 全部换算，前端直接展示 + tank_capacity 使用 `litersToGallons` 转换~~ | ~~⭐⭐⭐ 高~~ |
| 3 | **后端 i18n 错误消息** | 引入 `go-i18n`，创建 zh-CN/en-US/ja-JP TOML 翻译文件，API 错误返回翻译后的 message | ⭐⭐ 中 |
| 4 | **忘记密码流程** | 后端邮件发送 + Token 验证（DTO 已定义），前端登录页"忘记密码"入口 + 重置页面 | ⭐⭐ 中 |
| 5 | **记录列表筛选 UI** | 后端 API 已支持筛选参数，前端添加日期范围和站点筛选控件 | ⭐ 低 |
| ~~6~~ | ~~**数据导出（CSV/ZIP/JSON）**~~ | ~~✅ 已完成：GDPR 数据可携带权，`GET /api/v1/users/me/export?format=csv|zip|json&scope=basic|full`，支持三种格式 + 两种范围，前端设置页范围/格式选择~~ | ~~⭐⭐ 中~~ |
| ~~7~~ | ~~**隐私政策与用户协议**~~ | ~~✅ 已完成：`/privacy` + `/terms` 静态页面，三语 i18n 支持，注册页同意链接~~ | ~~⭐ 低~~ |

## 7. 维修保养等开销记录模块

> 独立于燃油记录的车辆开销台账，管理非油费开销。

### 7.1 功能范围

- **开销记录 CRUD**：维护类别（maintenance/repair/insurance/parking/toll/car_wash/inspection/parts/fine/other）、标题、金额、币种、日期、商家、里程、备注与预留凭证字段
- **开销列表与筛选**：按车辆查看历史记录，支持分类、日期范围、关键词、金额区间筛选
- **开销统计摘要**：展示总开销（按币种）、分类占比、月度趋势、近30天大额支出
- **保养联动**：开销记录可通过 `reminder_id` 显式关联保养提醒，完成登记后同步更新 `last_mileage/last_date` 并重算 `next_mileage/next_date`
- **共享车辆权限**：延续现有共享车辆可见、非车主仅可修改或删除自己创建记录的规则

### 7.2 边界约束

- 不修改现有 `fuel_records` 和 `VehicleStatsResponse`，避免污染油耗计算
- `receipt_url` 仅字段预留，不在 MVP 中实现上传
- 开销统计独立输出，不纳入现有 Dashboard 或 StatsPage
- 多币种记录存原币种，统计按 `totals_by_currency` 分组返回
