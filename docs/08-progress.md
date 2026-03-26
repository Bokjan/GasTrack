# GasTrack 需求完成进度

> **更新日期**: 2026-03-26
>
> **当前阶段**: 第一期 MVP 开发中

---

## 1. 进度总览

| 模块 | 后端 | 前端 | 整体状态 |
|------|------|------|---------|
| 项目基础设施 | ✅ 完成 | ✅ 完成 | ✅ 完成 |
| 用户认证 | ✅ 完成 | ✅ 完成 | ✅ 完成 |
| 用户资料 | ✅ 完成 | ✅ 完成 | ✅ 完成 |
| 车辆管理 | ✅ 完成 | ✅ 完成 | ✅ 完成 |
| 加油记录 | ✅ 完成 | ✅ 完成 | ✅ 完成 |
| 统计报表 | ✅ 完成 | ✅ 完成 | ✅ 完成 |
| 多语言 | 🔲 待实现 | ✅ 完成 | 🔨 进行中 |
| 多币种/单位 | ✅ 完成 | ✅ 完成 | ✅ 完成 |
| 前后端 API 对齐 | - | ✅ 完成 | ✅ 完成 |

**图例**: ✅ 完成 | 🔨 进行中 | 🔲 待实现 | ❌ 已放弃

---

## 2. 后端进度详情

### 2.1 基础设施 ✅

| 任务 | 状态 | 文件/说明 |
|------|------|----------|
| Go 项目骨架搭建 | ✅ | `server/` 目录结构，Go 1.22 + net/http |
| 配置管理 (Viper) | ✅ | `internal/config/config.go` + `config.yaml` |
| 日志系统 (Zap) | ✅ | 结构化日志，支持 JSON/Console 格式 |
| 数据库连接 (GORM + PostgreSQL) | ✅ | `internal/database/database.go`，AutoMigrate |
| Docker Compose (PostgreSQL) | ✅ | `docker-compose.yaml`，PostgreSQL 16-alpine |
| 统一响应格式 | ✅ | `internal/pkg/respond/`，OK/Created/Paged/Error |
| 请求解析工具 | ✅ | `internal/pkg/decode/`，JSON/PathParam/Query |
| 错误处理机制 | ✅ | `internal/pkg/apperror/`，AppError 类型 |
| 路由注册 | ✅ | `internal/router/router.go`，Go 1.22 ServeMux |

### 2.2 中间件 ✅

| 中间件 | 状态 | 说明 |
|--------|------|------|
| CORS | ✅ | 支持配置允许的 Origins |
| JWT 认证 | ✅ | Bearer Token 解析，用户 ID 注入 Context |
| 请求日志 | ✅ | 记录 method/path/status/duration |
| Panic Recovery | ✅ | 捕获 panic，返回 500 |
| Rate Limit | ✅ | IP 级别令牌桶限流，100 req/s |
| 中间件链 | ✅ | `Chain()` 链式组合 |

### 2.3 认证模块 (Auth) ✅

| API | 方法 | 路径 | 状态 |
|-----|------|------|------|
| 注册 | POST | `/auth/register` | ✅ |
| 登录 | POST | `/auth/login` | ✅ |
| 刷新 Token | POST | `/auth/refresh` | ✅ |
| 登出 | POST | `/auth/logout` | ✅ |
| 忘记密码 | POST | `/auth/forgot-password` | 🔲 DTO 已定义，逻辑未实现 |

### 2.4 用户模块 (User) ✅

| API | 方法 | 路径 | 状态 |
|-----|------|------|------|
| 获取资料 | GET | `/users/me` | ✅ |
| 更新资料 | PATCH | `/users/me` | ✅ |
| 修改密码 | PUT | `/users/me/password` | ✅ |
| 注销账号 | DELETE | `/users/me` | ✅ |

### 2.5 车辆模块 (Vehicle) ✅

| API | 方法 | 路径 | 状态 |
|-----|------|------|------|
| 车辆列表 | GET | `/vehicles` | ✅ 支持 include_archived 过滤 |
| 添加车辆 | POST | `/vehicles` | ✅ 支持 car/motorcycle/other + electric |
| 车辆详情 | GET | `/vehicles/{id}` | ✅ |
| 编辑车辆 | PATCH | `/vehicles/{id}` | ✅ |
| 删除车辆 | DELETE | `/vehicles/{id}` | ✅ 软删除 |

### 2.6 加油记录模块 (FuelRecord) ✅

| API | 方法 | 路径 | 状态 |
|-----|------|------|------|
| 记录列表 | GET | `/vehicles/{id}/records` | ✅ 分页（page/page_size） |
| 添加记录 | POST | `/vehicles/{id}/records` | ✅ |
| 记录详情 | GET | `/vehicles/{id}/records/{rid}` | ✅ |
| 编辑记录 | PATCH | `/vehicles/{id}/records/{rid}` | ✅ |
| 删除记录 | DELETE | `/vehicles/{id}/records/{rid}` | ✅ |

### 2.7 统计模块 (Stats) ✅

| API | 方法 | 路径 | 状态 |
|-----|------|------|------|
| 车辆统计 | GET | `/vehicles/{id}/stats` | ✅ |
| 全局总览 | GET | `/stats/overview` | ✅ 含各车辆子统计 |
| 油耗趋势 | GET | `/vehicles/{id}/efficiency-trend` | ✅ 支持 limit 参数 |
| 费用统计 | GET | `/stats/expenses` | 🔲 DTO 已定义，Handler 未实现 |

### 2.8 其他

| 功能 | 状态 | 说明 |
|------|------|------|
| 健康检查 | ✅ | `GET /health` |
| 后端 i18n | 🔲 | go-i18n TOML 文件已规划，未接入 |
| 文件上传 | 🔲 | 路由/Handler 未实现 |
| 数据校验 | ✅ | go-playground/validator，struct tag 校验 |
| 优雅关闭 | ✅ | 信号监听 + Shutdown 超时 |

---

## 3. 前端进度详情

### 3.1 基础设施 ✅

| 任务 | 状态 | 文件/说明 |
|------|------|----------|
| Monorepo (pnpm workspace) | ✅ | `packages/shared` + `packages/web` |
| Vite 构建配置 | ✅ | 端口 3000，API 代理至 8098 |
| TypeScript 配置 | ✅ | 严格模式，路径别名 `@/` |
| React 18 + React Router 6 | ✅ | SPA 路由 |
| Ant Design 5 | ✅ | 组件库 |
| 环境变量 | ✅ | `.env.example` |

### 3.2 共享包 (@gastrack/shared) ✅

| 模块 | 状态 | 说明 |
|------|------|------|
| 类型定义 (`types/`) | ✅ | 与后端 DTO 完全对齐（2026-03-26 全面审查通过） |
| API 调用层 (`api/`) | ✅ | Axios 客户端 + 各模块 API 封装 |
| HTTP 客户端 (`api/client.ts`) | ✅ | baseURL、Token 注入、401 自动刷新队列 |
| 状态管理 (`stores/authStore.ts`) | ✅ | Zustand，登录/登出/Token 刷新 |
| 状态管理 (`stores/vehicleStore.ts`) | ✅ | 车辆列表/选中车辆 |
| 工具函数 (`utils/`) | ✅ | formatNumber/formatCurrency（含 null 防护） |
| 常量 (`constants/`) | ✅ | FUEL_TYPES（含 electric）/ VEHICLE_TYPES 等 |
| i18n 框架 | 🔨 | i18next 已安装，翻译资源待补充 |

### 3.3 页面组件 (@gastrack/web)

| 页面 | 路由 | 状态 | 说明 |
|------|------|------|------|
| 登录页 | `/login` | ✅ | 邮箱+密码表单 |
| 注册页 | `/register` | ✅ | 邮箱+密码+昵称表单 |
| 仪表盘 | `/dashboard` | ✅ | 全局统计总览（卡片+车辆列表） |
| 车辆列表 | `/vehicles` | ✅ | 车辆卡片列表 |
| 添加/编辑车辆 | `/vehicles/new`, `/vehicles/:id/edit` | ✅ | 车辆表单（含 electric 类型） |
| 加油记录列表 | `/vehicles/:id/records` | ✅ | 分页表格 |
| 添加/编辑记录 | `/vehicles/:id/records/new`, `.../edit` | ✅ | 加油表单（字段与后端完全对齐） |
| 统计页 | `/stats` | ✅ | 油耗趋势图 + 距离图 + 统计卡片 |
| 个人设置 | `/settings` | 🔨 | 基础框架，待完善 |

### 3.4 通用组件

| 组件 | 状态 | 说明 |
|------|------|------|
| MainLayout | ✅ | 侧边栏导航 + 用户信息 |
| ProtectedRoute | ✅ | 登录态路由守卫 |
| ECharts 图表 | ✅ | 油耗趋势折线图 + 距离折线图 |

### 3.5 前后端 API 对齐审查 ✅

> 2026-03-26 完成全面审查，以下为修复记录：

| # | 问题 | 涉及文件 | 修复 |
|---|------|---------|------|
| 1 | `form.setFieldsValues` 拼写错误 | `RecordFormPage.tsx` | → `setFieldsValue` |
| 2 | `FUEL_TYPES` 缺少 `electric` | `constants/index.ts` | 添加 electric 选项 |
| 3 | 缺少 `FuelEfficiencyTrendResponse` 类型 | `types/index.ts` | 新增接口定义 |
| 4 | `efficiencyTrend` 返回类型错误 | `api/index.ts` | 改为 `FuelEfficiencyTrendResponse` |
| 5 | StatsPage 直接用 trend 响应当数组 | `StatsPage.tsx` | 改为 `.items` + `efficiency_unit` |
| 6 | 硬编码 `L/100km` 单位 | `StatsPage.tsx` | 使用动态 `efficiencyUnit` |
| 7 | Auth 类型不匹配 | `types/`, `api/`, `stores/` | 全面重写对齐后端 DTO |
| 8 | User 字段名不匹配 | `types/`, 多个页面 | `currency` → `currency_code` 等 |
| 9 | FuelRecord 字段名不匹配 | `types/`, 多个页面 | 全面重写对齐后端 DTO |
| 10 | VehicleStats 字段名不匹配 | `types/`, `StatsPage.tsx` | 全面重写对齐后端 DTO |

---

## 4. 待实现功能

### 4.1 第一期剩余 (P0)

| 功能 | 前端 | 后端 | 优先级 |
|------|------|------|--------|
| 后端 i18n 错误消息 | - | 🔲 错误消息 i18n | 中 |
| 个人设置页完善 | 🔨 | ✅ API 已有 | 中 |
| 响应式适配（移动端） | 🔲 | - | 中 |
| 单位换算展示 | 🔲 前端换算 | ✅ 存储原始值 | 中 |

### 4.2 第二期 (P1)

| 功能 | 说明 | 状态 |
|------|------|------|
| 家庭群组 | 群组 CRUD + 邀请 | 🔲 DTO 已规划 |
| 数据导出 CSV/PDF | 前端触发，后端生成 | 🔲 |
| PWA 支持 | 离线访问 | 🔲 |
| 多车辆对比图表 | 油耗/费用对比 | 🔲 |
| 车辆照片上传 | 文件上传接口 | 🔲 |
| 费用统计 API | 按月/季度/年汇总 | 🔲 DTO 已定义 |
| 忘记密码 | 邮件重置 | 🔲 DTO 已定义 |

### 4.3 第三期 (P2)

| 功能 | 说明 | 状态 |
|------|------|------|
| 微信小程序 | Taro 3 开发 | 🔲 |
| 小票 OCR | 拍照识别加油小票 | 🔲 |
| 保养提醒 | 基于里程/时间 | 🔲 |
| 加油站地图 | PostGIS + 位置服务 | 🔲 |
| 第三方登录 | Google / Apple / 微信 | 🔲 |

---

## 5. 已知问题

| # | 问题 | 严重度 | 状态 |
|---|------|--------|------|
| 1 | ~~后端 `Paged()` 响应格式为 `{data, meta: {page, page_size, total, total_pages}}`，前端 `PaginatedData<T>` 定义为 `{list, total, page, page_size}` 嵌在 `data` 内——两端分页响应结构不一致~~ | ⚠️ 中 | ✅ 已修复 (2026-03-26) |
| 2 | ~~`RecordFormPage.tsx` / `VehicleFormPage.tsx` 中 Ant Design `addonAfter` 属性标记为 deprecated（Hint 级别）~~ | 💡 低 | ✅ 已修复 (2026-03-26) |
| 3 | ~~`shared` 包 TypeScript 编译有 `ImportMeta.env` 类型缺失警告~~ | 💡 低 | ✅ 已修复 (2026-03-26) |
| 4 | ~~右上角头像和语言选择"地球"图标不对齐~~ | 🐛 中 | ✅ 已修复 (2026-03-26) |
| 5 | ~~切换语言后浏览器标题仍然是默认中文~~ | 🐛 中 | ✅ 已修复 (2026-03-26) |
| 6 | ~~添加车辆页"车辆类型"字段 label 显示为"汽车"~~ | 🐛 中 | ✅ 已修复 (2026-03-26) |
| 7 | ~~燃油类型下拉菜单中"电动"显示为 `fuelTyp...`（缺少翻译）~~ | 🐛 中 | ✅ 已修复 (2026-03-26) |
| 8 | ~~统计页"里程趋势"标题显示为 `stats.distanceTrend`（缺少翻译）~~ | 🐛 中 | ✅ 已修复 (2026-03-26) |
| 9 | ~~加油记录表单中油/货币/里程单位无法选择~~ | 🐛 高 | ✅ 已修复 (2026-03-26) |
| 10 | ~~加油量/单价/总费用三个字段只能手动填写全部，不支持任意两项自动计算第三项~~ | 🐛 中 | ✅ 已修复 (2026-03-26) |
| 11 | ~~`FUEL_UNITS` / `DISTANCE_UNITS` 常量 label 硬编码中文，切换语言后单位选项不翻译~~ | 🐛 中 | ✅ 已修复 (2026-03-26) |
| 12 | ~~添加车辆页燃油类型下拉框宽度不足，"混合动力"等长文本被截断；油箱容量单位固定为 L，不支持加仑~~ | 🐛 中 | ✅ 已修复 (2026-03-26) |
| 13 | ~~右上角地球图标与用户头像垂直不对齐~~ | 🐛 中 | ✅ 已修复 (2026-03-26) |
| 14 | ~~排量(cc)字段仅摩托车可见，汽车无法填写排量~~ | 🐛 中 | ✅ 已修复 (2026-03-26) |
| 15 | ~~电动车选了"电动"燃油类型后，表单仍然使用油车逻辑（油箱容量/L/加油站/油耗）~~ | 🚀 高 | ✅ 已修复 (2026-03-26) |

> **当前无未修复的已知问题** 🎉

---

## 6. 变更日志

### 2026-03-26

- ✅ Docker Compose PostgreSQL 16 环境搭建
- ✅ 服务端口从 8080 改为 8098（前后端同步）
- ✅ 修复 CORS 错误（添加 `localhost:3000` 到允许列表）
- ✅ 修复 `formatNumber` TypeError（null 防护 + 后端补全缺失字段）
- ✅ 全面前后端 API 一致性审查并修复（10 项问题）
- ✅ 编写 API 详细文档、数据库文档、进度文档、开发调试文档
- ✅ **修复已知问题 #1**：前后端分页响应结构不一致
  - 废弃 `PaginatedData<T>`（嵌套在 `data` 内的 `{list, total, page, page_size}`）
  - 新增 `PageMeta` 接口 + 重写 `PaginatedResponse<T>` = `{ code, message, data: T[], meta: PageMeta }`
  - 更新 `RecordListPage.tsx` 解析逻辑：`data.data.list` → `resp.data`，`data.data.total` → `resp.meta.total`
  - 涉及文件：`shared/src/types/index.ts`、`shared/src/api/index.ts`、`web/src/pages/record/RecordListPage.tsx`
- ✅ **修复已知问题 #2**：InputNumber `addonAfter` deprecated hint
  - 将 `addonAfter` 替换为 `suffix`（fuel_amount、odometer、tank_capacity、engine_cc 共 4 处）
  - 涉及文件：`web/src/pages/record/RecordFormPage.tsx`、`web/src/pages/vehicle/VehicleFormPage.tsx`
- ✅ **修复已知问题 #3**：shared 包 `import.meta.env` 类型缺失
  - 新增 `shared/src/vite-env.d.ts`，声明 `ImportMetaEnv` 和 `ImportMeta` 接口
  - tsconfig.json 的 `include: ["src/**/*"]` 自动包含该声明文件
- ✅ **修复 UI 反馈 #4**：Header 头像和语言地球图标不对齐
  - `MainLayout.tsx` 中 `Space` 添加 `align="center"`
  - `GlobalOutlined` 包裹在 `inline-flex` 容器中确保垂直居中对齐
- ✅ **修复 UI 反馈 #5**：切换语言后浏览器标题仍为中文
  - 三个 locale JSON 新增 `app.title` 翻译
  - `MainLayout.tsx` 添加 `useEffect` 监听 `i18n.language` 变化，同步更新 `document.title` 和 `<html lang>`
  - 语言切换使用 `await i18n.changeLanguage()` 确保异步完成
- ✅ **修复 UI 反馈 #6**：车辆类型 label 显示为"汽车"
  - `VehicleFormPage.tsx` 中 `t('vehicleType.car')` → `t('vehicle.vehicleType')`
  - 三个 locale JSON 的 `vehicle` 节点新增 `vehicleType` 翻译
- ✅ **修复 UI 反馈 #7**：燃油类型"电动"显示为 `fuelTyp...`
  - 三个 locale JSON 的 `fuelType` 节点新增 `electric` 翻译（中:电动 / en:Electric / ja:電気）
- ✅ **修复 UI 反馈 #8**：统计页"里程趋势"未翻译
  - 三个 locale JSON 的 `stats` 节点新增 `distanceTrend` 翻译（中:里程趋势 / en:Distance Trend / ja:走行距離推移）
- ✅ **修复 UI 反馈 #9**：加油记录表单单位不可选
  - `RecordFormPage.tsx` 新增三个 `Select` 控件：燃油单位（L/gal）、货币（CNY/USD/EUR/JPY/GBP/KRW）、里程单位（km/mi）
  - `constants/index.ts` 新增 `FUEL_UNITS` 和 `DISTANCE_UNITS` 常量
  - 三个 locale JSON 新增 `fuelRecord.fuelUnit`、`fuelRecord.currency`、`fuelRecord.distanceUnit` 翻译
- ✅ **修复 UI 反馈 #10**：加油量/单价/总费用三值自动计算
  - 重写 `RecordFormPage.tsx` 的 `autoCalc` 逻辑：填写任意两个字段后自动计算第三个
  - 支持的计算方向：`加油量 × 单价 → 总费用`、`总费用 ÷ 加油量 → 单价`、`总费用 ÷ 单价 → 加油量`
- ✅ **修复 i18n 遗漏 #11**：`FUEL_UNITS` / `DISTANCE_UNITS` 常量 label 未国际化
  - `constants/index.ts` 中 label 从硬编码中文改为 i18n key（`unit.liter`、`unit.gallon`、`unit.km`、`unit.mile`）
  - 三个 locale JSON 新增 `unit` 节点（中: 升/加仑/公里/英里 / en: Liter/Gallon/Kilometer/Mile / ja: リットル/ガロン/キロメートル/マイル）
  - `RecordFormPage.tsx` 中 Select 选项使用 `t(u.label)` 翻译
- ✅ **修复 UI 反馈 #12**：燃油类型下拉框截断 + 油箱容量不支持加仑
  - `VehicleFormPage.tsx` 中燃油类型 `Select` 添加 `popupMatchSelectWidth={false}`，下拉宽度自适应内容
  - 油箱容量从固定 `suffix="L"` 改为 `InputNumber` + `Select` 单位选择器（L / gal），复用 `FUEL_UNITS` 常量
  - 三个 locale JSON 的 `vehicle.tankCapacity` 去掉硬编码 "(L)"，单位由选择器提供
- ✅ **修复 UI 反馈 #13**：Header 地球图标与头像垂直不对齐（第二次修复）
  - `MainLayout.tsx` 中右侧区域从 AntD `Space` 改为原生 `div` + `display: flex; align-items: center; gap: 16px` 布局
  - `GlobalOutlined` 直接作为 `Dropdown` 子元素，不再包裹额外容器
  - 移除未使用的 `Space` import
- ✅ **修复逻辑 #14**：排量(cc)字段从摩托车专属改为所有非电动车型可见
  - `VehicleFormPage.tsx` 排量显示条件从 `vehicleType === 'motorcycle'` 改为 `hasEngineCC(fuelType)`（即 `fuelType !== 'electric'`）
  - 排量上限从 3000 扩大到 10000（覆盖大排量汽车）
  - `constants/index.ts` 新增 `hasEngineCC()` / `isElectricVehicle()` 工具函数
  - 后端 `vehicle.go` model 注释从"摩托车常用"改为"燃油/混动车辆通用"
- ✅ **新增功能 #15**：电动车电耗统计全栈支持
  - **后端**：
    - `model/vehicle.go` 新增 `BatteryCapacity` 字段（decimal(6,2)，kWh）
    - `model/fuel_record.go` 的 `FuelUnit` 扩展支持 `kWh`，`FuelEfficiency` 兼容 `kWh/100km`
    - `dto/vehicle.go` 全三个 DTO（Create/Update/Response）同步新增 `battery_capacity`
    - `dto/fuel_record.go` 的 validate tag 从 `oneof=L gal` 扩展为 `oneof=L gal kWh`
    - `service/vehicle.go` 的 Create/Update/vehicleToResponse 同步映射 `BatteryCapacity`
  - **前端常量**：
    - `constants/index.ts` 新增 `ENERGY_UNITS`（kWh）、`EV_MEASUREMENT_SYSTEMS`（kWh/100km, km/kWh, mi/kWh）
  - **前端表单（VehicleFormPage）**：
    - 选择"电动"后：油箱容量 → 电池容量、单位自动切换 kWh、排量字段隐藏
    - 选择其他燃油类型后：自动恢复油箱容量 + L/gal 选项、排量字段显示
  - **前端表单（RecordFormPage）**：
    - 通过 `vehicleApi.getById` 获取车辆燃料类型，`isEv` 状态驱动表单适配
    - 电动车：加油日期 → 充电日期、加油站 → 充电站、加油量 → 充电量、燃油单位 → 能量单位(kWh)、是否加满 → 是否充满
  - **TS 类型**：`Vehicle` / `CreateVehicleRequest` 新增 `battery_capacity` 字段
  - **i18n（zh/en/ja）**：
    - 新增 `vehicle.batteryCapacity`、`fuelRecord.titleEv/chargingDate/chargingStation/chargingAmount/energyUnit/isFullCharge/energyConsumption`
    - 新增 `unit.kwh`、`measurement.kwh100km/kmKwh/miKwh`
    - 新增 `stats.totalEnergy/avgEnergyConsumption/energyConsumptionTrend`
