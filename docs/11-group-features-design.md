# GasTrack 群组功能扩展设计

> **创建日期**: 2026-03-31
>
> **基于**: 3.10 家庭/群组管理（已完成）的扩展
>
> **状态**: ✅ 四个扩展功能已全部实现（后端 API + 前端 UI + i18n 三语言）
>
> **目标**: 让群组功能从"管理工具"升级为"家庭用车协作平台"

---

## 概览

在现有群组基础功能（CRUD + 成员管理 + 简单数据汇总表格）之上，新增四个功能模块：

| # | 功能 | 优先级 | 核心价值 | 状态 |
|---|------|--------|----------|------|
| 1 | 车辆共享标记 | P1 | 多人共用一辆车时，各自都能记录加油 | ✅ 已完成 |
| 2 | 群组油耗排行榜 / 驾驶 PK | P1 | 趣味性对比，激励家庭成员省油 | ✅ 已完成 |
| 3 | 群组费用统计看板 | P1 | "全家这个月加油花了多少钱"一目了然 | ✅ 已完成 |
| 4 | 加油站推荐共享 | P1 | 家庭成员之间共享低价加油站 | ✅ 已完成 |

---

## 1. 车辆共享标记

### 1.1 需求背景

现有模型：一辆车只属于一个用户（`Vehicle.UserID`），只有车辆拥有者才能为其记录加油。但在家庭场景中，一辆车可能全家人都在开——妈妈上班开、爸爸周末开、孩子偶尔也开。

### 1.2 功能设计

**核心概念**：在群组内，车辆拥有者可以将自己的车标记为"群组共享"，群组内其他成员就可以为该车辆记录加油。

**用户流程**：
1. 车主在群组详情页的"数据汇总" Tab 中，看到自己的车辆旁有一个"共享"开关
2. 开启共享后，群组内所有成员在新增加油记录时，可以选择该共享车辆
3. 加油记录的 `user_id` 记录的是"谁加的油"（操作人），`vehicle_id` 记录的是"哪辆车"
4. 车主可以随时关闭共享（不影响已有记录）

**权限规则**：
- 只有**车辆拥有者**可以开启/关闭共享
- 共享范围限定在**该群组内**（不是全局共享）
- 共享车辆对群组外用户**不可见**
- 群组成员对共享车辆的操作权限：**可新增记录，不可编辑/删除他人记录，不可编辑车辆信息**

### 1.3 数据模型

新增 `shared_vehicles` 关联表（多对多：群组 ↔ 车辆）：

```sql
CREATE TABLE shared_vehicles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id        UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    vehicle_id      UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    shared_by       UUID NOT NULL REFERENCES users(id),    -- 共享发起人（车主）
    shared_at       TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (group_id, vehicle_id)
);
CREATE INDEX idx_shared_vehicles_group ON shared_vehicles(group_id);
CREATE INDEX idx_shared_vehicles_vehicle ON shared_vehicles(vehicle_id);
```

**设计要点**：
- 不修改现有 `vehicles` 表结构，通过关联表实现，保持向后兼容
- `shared_by` 必须等于 `vehicles.user_id`（由 Service 层校验）
- 联合唯一索引 `(group_id, vehicle_id)` 防止重复共享
- 级联删除：群组删除时自动清理共享关系

### 1.4 API 设计

| HTTP 方法 | 路径 | 说明 |
|-----------|------|------|
| `POST` | `/api/v1/groups/{id}/shared-vehicles` | 共享车辆到群组 |
| `DELETE` | `/api/v1/groups/{id}/shared-vehicles/{vid}` | 取消车辆共享 |
| `GET` | `/api/v1/groups/{id}/shared-vehicles` | 获取群组内共享车辆列表 |

#### 1.4.1 共享车辆

```
POST /api/v1/groups/{id}/shared-vehicles
```

**请求体**
```json
{
  "vehicle_id": "uuid-of-vehicle"
}
```

**校验规则**：
- 请求者必须是该群组的成员
- 请求者必须是该车辆的拥有者（`vehicle.user_id == caller`）
- 车辆未被归档
- 车辆尚未在该群组中共享

**成功响应** `201 Created`
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "shared-vehicle-uuid",
    "group_id": "group-uuid",
    "vehicle_id": "vehicle-uuid",
    "vehicle_name": "家用车",
    "owner_name": "张三",
    "shared_at": "2026-03-31T10:00:00Z"
  }
}
```

#### 1.4.2 取消共享

```
DELETE /api/v1/groups/{id}/shared-vehicles/{vid}
```

**权限**：仅车辆拥有者或群组 Owner 可取消共享。

#### 1.4.3 获取共享车辆列表

```
GET /api/v1/groups/{id}/shared-vehicles
```

**响应**：返回共享车辆数组，每项包含车辆基本信息 + 车主昵称。

### 1.5 对现有功能的影响

**加油记录创建**：前端在选择车辆时，除了用户自己的车辆，还需要加载用户所在群组的共享车辆作为可选项。

改造 `GET /api/v1/vehicles` 接口，新增查询参数：

```
GET /api/v1/vehicles?include_shared=true
```

- `include_shared=false`（默认）：仅返回自己的车辆（现有行为）
- `include_shared=true`：返回自己的车辆 + 所有群组内的共享车辆（附带 `shared_from_group` 标识）

**加油记录权限**：为共享车辆创建的记录，`user_id` 为操作人，`vehicle_id` 为共享车辆。编辑/删除权限仍限于记录创建者本人和车辆拥有者。

### 1.6 前端设计

- **群组详情 → 数据汇总 Tab**：车辆列表中，车主自己的车辆行显示"共享"Switch 开关
- **加油记录表单 → 车辆选择**：下拉框分组显示——"我的车辆"和"共享车辆（来自XX群组）"
- **加油记录列表**：共享车辆的记录显示一个共享图标标识

### 1.7 实现层次

> ⚠️ **当前实现状态**：后端权限控制已全面完成（`verifyVehicleAccess` 统一鉴权模式已在 FuelRecordService / StatsService / ReminderService / VehicleService 四个 Service 中实现），`shared_vehicles` 数据模型已就绪。剩余工作：3 条 API（Share/Unshare/List）+ 前端 UI。

```
Model:    SharedVehicle (✅ 已实现)
DTO:      ShareVehicleRequest, SharedVehicleResponse (待实现)
Repo:     GroupRepository.IsVehicleSharedToUser (✅ 已实现)
          SharedVehicleRepository — Create/Delete/ListByGroup/ListByUser/Exists (待实现)
Service:  FuelRecordService.verifyVehicleAccess (✅ 已实现)
          StatsService.verifyVehicleAccess (✅ 已实现)
          ReminderService.verifyVehicleAccess (✅ 已实现)
          VehicleService — GetByID 共享回退 (✅ 已实现), List include_shared (待实现)
          SharedVehicleService — Share/Unshare/List + 权限校验 (待实现)
Handler:  GroupHandler — 新增 ShareVehicle/UnshareVehicle/ListSharedVehicles (待实现)
Router:   新增 3 条路由 (待实现)
```

---

## 2. 群组油耗排行榜 / 驾驶 PK

### 2.1 需求背景

现有的数据汇总只是一个干巴巴的表格，缺乏趣味性和互动感。排行榜让家庭成员之间产生良性竞争——"谁开车最省油？"

### 2.2 功能设计

**排行维度**（按实用性排序）：

| 维度 | 指标 | 说明 | 排序 |
|------|------|------|------|
| 🏆 油耗排行 | 平均油耗 (L/100km) | 核心排行，越低越好 | ASC |
| 💰 费用排行 | 当月总加油费用 | 谁花钱最多 | DESC |
| 📏 里程排行 | 当月总行驶里程 | 谁开得最多 | DESC |
| ⛽ 加油频次 | 当月加油次数 | 谁加油最勤 | DESC |

**排行粒度**：
- 按**成员 × 车辆**排行（一个成员多辆车分开排）
- 时间范围：默认**当月**，可切换为"上月"、"近 3 个月"、"今年"

**显示效果**：
- 前三名用 🥇🥈🥉 标识
- 每行显示：排名 + 成员昵称 + 车辆名 + 指标值 + 相比群组平均的差异百分比
- 自己的排名高亮显示

**数据安全规则**：
- 仅群组成员可查看排行榜
- 排行数据只基于成员的**非归档**车辆
- 至少有 2 条加油记录的车辆才参与排行（避免数据不足导致排名失真）

### 2.3 API 设计

```
GET /api/v1/groups/{id}/leaderboard
```

**🔒 需要认证**

**查询参数**

| 参数 | 类型 | 默认值 | 可选值 | 说明 |
|------|------|--------|--------|------|
| metric | string | `efficiency` | `efficiency` / `cost` / `distance` / `frequency` | 排行维度 |
| period | string | `current_month` | `current_month` / `last_month` / `last_3_months` / `current_year` | 时间范围 |

**成功响应** `200 OK`
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "group_id": "group-uuid",
    "group_name": "我的家庭群",
    "metric": "efficiency",
    "period": "current_month",
    "period_label": "2026年3月",
    "group_avg": 8.5,
    "unit": "L/100km",
    "rankings": [
      {
        "rank": 1,
        "user_id": "user-uuid-1",
        "nickname": "张三",
        "vehicle_id": "vehicle-uuid-1",
        "vehicle_name": "家用卡罗拉",
        "value": 6.8,
        "diff_from_avg": -20.0,
        "record_count": 5,
        "is_self": false
      },
      {
        "rank": 2,
        "user_id": "user-uuid-2",
        "nickname": "李四",
        "vehicle_id": "vehicle-uuid-2",
        "vehicle_name": "通勤小飞度",
        "value": 7.5,
        "diff_from_avg": -11.8,
        "record_count": 4,
        "is_self": true
      }
    ],
    "total_participants": 4
  }
}
```

### 2.4 后端实现

**SQL 查询核心**（以 efficiency 排行为例）：

```sql
SELECT 
    v.id AS vehicle_id,
    v.name AS vehicle_name,
    v.user_id,
    COUNT(fr.id) AS record_count,
    AVG(fr.fuel_efficiency) AS avg_efficiency,
    SUM(fr.total_cost) AS total_cost,
    SUM(fr.trip_distance) AS total_distance
FROM vehicles v
JOIN group_members gm ON gm.user_id = v.user_id AND gm.group_id = $1
JOIN fuel_records fr ON fr.vehicle_id = v.id
WHERE v.deleted_at IS NULL 
  AND v.is_archived = false
  AND fr.refuel_date >= $2          -- 时间范围起始
  AND fr.refuel_date < $3           -- 时间范围结束
  AND fr.fuel_efficiency > 0        -- 排除无效油耗记录
GROUP BY v.id, v.name, v.user_id
HAVING COUNT(fr.id) >= 2            -- 至少 2 条记录
ORDER BY avg_efficiency ASC         -- 油耗越低越好
```

**共享车辆的处理**：共享车辆的加油记录按"操作人"归入其名下排行（因为是谁开的影响油耗），但也在车辆名后附注"(共享)"标识。

### 2.5 实现层次

```
DTO:      LeaderboardRequest (query params), LeaderboardResponse, LeaderboardEntry (新增)
Repo:     GroupRepository — 新增 GetLeaderboard(ctx, groupID, metric, startDate, endDate)
Service:  GroupService — 新增 GetLeaderboard(ctx, groupID, userID, metric, period)
Handler:  GroupHandler — 新增 GetLeaderboard
Router:   新增 1 条路由
```

### 2.6 前端设计

- **群组详情页新增 Tab**："排行榜" 🏆（与现有的 群组信息 / 成员管理 / 数据汇总 并列）
- **排行维度**：顶部 Segmented 切换（油耗 / 费用 / 里程 / 频次）
- **时间范围**：右侧 Select 下拉（本月 / 上月 / 近3月 / 今年）
- **排行列表**：Ant Design List 组件，前三名 Badge + 自己行高亮
- **空状态**：群组内数据不足时显示友好提示（"群组数据积累中，至少需要2条加油记录才能参与排行"）

---

## 3. 群组费用统计看板

### 3.1 需求背景

当前的数据汇总只是一个车辆维度的表格，缺乏时间维度的趋势分析和直观的图表可视化。家庭管理者最关心的问题是："全家这个月/这个季度/今年加油花了多少钱？趋势是升还是降？"

### 3.2 功能设计

**看板内容**：

#### 3.2.1 顶部统计卡片（4张）

| 卡片 | 指标 | 计算方式 |
|------|------|----------|
| 💰 总费用 | 选定时段内群组所有成员的加油总费用 | `SUM(total_cost)` |
| ⛽ 总加油量 | 选定时段内总加油升数 | `SUM(fuel_amount)` |
| 📏 总里程 | 选定时段内总行驶里程 | `SUM(trip_distance)` |
| 📊 平均油耗 | 选定时段内群组整体平均油耗 | `AVG(fuel_efficiency)` |

每张卡片下方显示**环比变化**（与上一时段相比的百分比增减）。

#### 3.2.2 费用趋势图

- **图表类型**：堆叠柱状图（ECharts）
- **X 轴**：时间（月维度 → 各月份，年维度 → 各年份）
- **Y 轴**：费用金额
- **堆叠维度**：按成员分色堆叠（每个成员一种颜色），可直观看出每个人的费用占比
- **与现有 Stats 的区别**：Stats 页是单车维度，这里是群组全员汇总

#### 3.2.3 费用占比饼图

- **图表类型**：环形饼图（ECharts）
- **切片**：按成员维度切分费用占比
- **Tooltip**：显示成员昵称、费用金额、百分比

#### 3.2.4 时间维度控制

- **维度**：按月 / 按年（Segmented 切换）
- **年份选择**：按月模式下可选择年份（Select 下拉）

### 3.3 API 设计

```
GET /api/v1/groups/{id}/expense-stats
```

**🔒 需要认证**

**查询参数**

| 参数 | 类型 | 默认值 | 可选值 | 说明 |
|------|------|--------|--------|------|
| period | string | `month` | `month` / `year` | 聚合维度 |
| year | int | 当前年 | 如 `2026` | 查询年份（仅 period=month 时有效） |

**成功响应** `200 OK`
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "group_id": "group-uuid",
    "group_name": "我的家庭群",
    "period": "month",
    "year": 2026,
    "summary": {
      "total_cost": 28500.00,
      "total_fuel": 3200.5,
      "total_distance": 38000.0,
      "avg_efficiency": 8.42,
      "cost_change_pct": -5.2,
      "fuel_change_pct": -3.1,
      "distance_change_pct": 2.0,
      "efficiency_change_pct": -1.8
    },
    "member_breakdown": [
      {
        "user_id": "user-uuid-1",
        "nickname": "张三",
        "total_cost": 15000.00,
        "total_fuel": 1700.0,
        "percentage": 52.6
      },
      {
        "user_id": "user-uuid-2",
        "nickname": "李四",
        "total_cost": 13500.00,
        "total_fuel": 1500.5,
        "percentage": 47.4
      }
    ],
    "trend_items": [
      {
        "period_label": "2026-01",
        "total_cost": 4200.00,
        "total_fuel": 480.0,
        "total_distance": 5600.0,
        "avg_efficiency": 8.57,
        "by_member": [
          {
            "user_id": "user-uuid-1",
            "nickname": "张三",
            "cost": 2300.00
          },
          {
            "user_id": "user-uuid-2",
            "nickname": "李四",
            "cost": 1900.00
          }
        ]
      }
    ],
    "prev_trend_items": [
      {
        "period_label": "2025-01",
        "total_cost": 3800.00,
        "total_fuel": 440.0,
        "total_distance": 5200.0,
        "avg_efficiency": 8.46
      }
    ]
  }
}
```

### 3.4 后端实现

**主查询**（趋势数据，以月维度为例）：

```sql
-- 按月按成员聚合
SELECT 
    TO_CHAR(fr.refuel_date, 'YYYY-MM') AS period_label,
    gm.user_id,
    SUM(fr.total_cost) AS cost,
    SUM(fr.fuel_amount) AS fuel,
    SUM(fr.trip_distance) AS distance,
    AVG(fr.fuel_efficiency) AS avg_eff,
    COUNT(fr.id) AS records
FROM fuel_records fr
JOIN vehicles v ON v.id = fr.vehicle_id AND v.deleted_at IS NULL AND v.is_archived = false
JOIN group_members gm ON gm.user_id = v.user_id AND gm.group_id = $1
WHERE EXTRACT(YEAR FROM fr.refuel_date) = $2
GROUP BY period_label, gm.user_id
ORDER BY period_label ASC, gm.user_id ASC
```

**环比计算**：Service 层查询当前时段和上一时段的汇总数据，计算百分比变化。

**同比数据**：按月模式下，额外查询上一年同期数据（复用逻辑模式同现有 Stats 的 `prev_items`）。

### 3.5 实现层次

```
DTO:      GroupExpenseStatsRequest, GroupExpenseStatsResponse, 
          GroupExpenseSummary, MemberCostBreakdown, GroupTrendItem, MemberCostItem (新增)
Repo:     GroupRepository — 新增 GetGroupExpenseByPeriod(ctx, groupID, year, period)
Service:  GroupService — 新增 GetExpenseStats(ctx, groupID, userID, period, year)
Handler:  GroupHandler — 新增 GetExpenseStats
Router:   新增 1 条路由
```

### 3.6 前端设计

- **群组详情页新增 Tab**："费用看板" 💰（与排行榜并列）
- **顶部**：4 张 Statistic 卡片（Row + Col 4列），环比变化用绿色/红色箭头
- **中部左侧**：堆叠柱状图（ECharts，按成员分色，与现有 StatsPage 复用图表风格）
- **中部右侧**：环形饼图（ECharts，成员费用占比）
- **维度控制**：Segmented（月/年）+ Select（年份选择），同现有 StatsPage 交互方式
- **暗色模式**：复用现有 ECharts 暗色 token 体系

---

## 4. 加油站推荐共享

### 4.1 需求背景

家庭成员经常去不同的加油站，哪个站便宜、哪个站油品好，这些信息值得在群组内共享。"爸爸常去的XX加油站比较便宜"——现在可以通过数据说话。

### 4.2 功能设计

**核心逻辑**：基于群组所有成员的加油记录中的 `station_name` + `unit_price` 数据，自动聚合出群组成员常去的加油站列表。

**展示内容**：

| 字段 | 说明 | 来源 |
|------|------|------|
| 加油站名 | 站点名称 | `fuel_records.station_name` |
| 平均油价 | 该站所有记录的平均单价 | `AVG(unit_price)` |
| 最新油价 | 该站最近一次加油的单价 | 最新一条记录的 `unit_price` |
| 加油次数 | 群组成员在该站的总加油次数 | `COUNT(*)` |
| 常客 | 哪些成员去过该站 | 成员昵称列表 |
| 最近加油 | 最近一次加油日期 | `MAX(refuel_date)` |
| 价格趋势 | 相比上次涨了还是跌了 | 最近两次记录对比 |

**数据安全**：
- 仅群组成员可查看
- 只聚合群组成员的数据（不跨群组）
- `station_name` 为空的记录自动排除

**筛选**：
- 按燃油标号筛选（92/95/98 等）
- 按时间范围筛选（默认近 6 个月）

### 4.3 API 设计

```
GET /api/v1/groups/{id}/stations
```

**🔒 需要认证**

**查询参数**

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| fuel_grade | string | - | 可选，按燃油标号筛选（如 `92`/`95`/`98`） |
| months | int | `6` | 时间范围，最近 N 个月的数据（1-24） |
| sort_by | string | `frequency` | 排序：`frequency`（加油次数）/ `avg_price`（平均油价）/ `latest_date`（最近加油） |

**成功响应** `200 OK`
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "group_id": "group-uuid",
    "group_name": "我的家庭群",
    "total_stations": 8,
    "stations": [
      {
        "station_name": "中石化望京站",
        "avg_unit_price": 7.85,
        "latest_unit_price": 7.92,
        "price_trend": "up",
        "currency_code": "CNY",
        "visit_count": 12,
        "visitors": [
          {"user_id": "uuid-1", "nickname": "张三", "count": 8},
          {"user_id": "uuid-2", "nickname": "李四", "count": 4}
        ],
        "latest_visit": "2026-03-28T14:30:00Z",
        "fuel_grades_seen": ["92", "95"]
      },
      {
        "station_name": "壳牌北辰西路站",
        "avg_unit_price": 8.12,
        "latest_unit_price": 8.05,
        "price_trend": "down",
        "currency_code": "CNY",
        "visit_count": 6,
        "visitors": [
          {"user_id": "uuid-2", "nickname": "李四", "count": 6}
        ],
        "latest_visit": "2026-03-25T10:00:00Z",
        "fuel_grades_seen": ["95"]
      }
    ]
  }
}
```

### 4.4 后端实现

**SQL 查询核心**：

```sql
-- 群组成员常去的加油站聚合
SELECT 
    fr.station_name,
    AVG(fr.unit_price) AS avg_unit_price,
    COUNT(*) AS visit_count,
    MAX(fr.refuel_date) AS latest_visit,
    ARRAY_AGG(DISTINCT fr.fuel_grade) FILTER (WHERE fr.fuel_grade IS NOT NULL AND fr.fuel_grade != '') AS fuel_grades,
    fr.currency_code
FROM fuel_records fr
JOIN vehicles v ON v.id = fr.vehicle_id AND v.deleted_at IS NULL
JOIN group_members gm ON gm.user_id = v.user_id AND gm.group_id = $1
WHERE fr.station_name IS NOT NULL 
  AND fr.station_name != ''
  AND fr.unit_price > 0
  AND fr.refuel_date >= NOW() - INTERVAL '$2 months'
GROUP BY fr.station_name, fr.currency_code
ORDER BY visit_count DESC
```

**最新油价 & 趋势**：额外子查询取每个站点最新两条记录计算趋势方向。

**visitors 列表**：额外查询每个站点被哪些成员光顾过，附带各自次数。

### 4.5 实现层次

```
DTO:      GroupStationStatsRequest, GroupStationStatsResponse, 
          StationInfo, StationVisitor (新增)
Repo:     GroupRepository — 新增 GetGroupStationStats(ctx, groupID, months, fuelGrade, sortBy)
                           新增 GetStationVisitors(ctx, groupID, stationNames, months)
                           新增 GetStationLatestPrices(ctx, groupID, stationNames, months)
Service:  GroupService — 新增 GetStationStats(ctx, groupID, userID, months, fuelGrade, sortBy)
Handler:  GroupHandler — 新增 GetStationStats
Router:   新增 1 条路由
```

### 4.6 前端设计

- **群组详情页新增 Tab**："加油站" ⛽（与排行榜/费用看板并列）
- **顶部筛选栏**：燃油标号 Select + 时间范围 Select + 排序方式 Select
- **列表**：Ant Design Table / Card List
  - 站名（大号字体）
  - 平均油价（显眼数字 + 趋势箭头 ↑↓）
  - 加油次数 Badge
  - 常客 Avatar.Group
  - 最近加油日期
- **空状态**：群组内尚无加油站数据时，提示"记录加油时别忘了填写加油站名哦！"
- **移动端**：Table → 卡片列表适配

---

## 5. 数据库变更汇总

### 5.1 新增表

```sql
-- 车辆共享（功能 1）
CREATE TABLE shared_vehicles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id        UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    vehicle_id      UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    shared_by       UUID NOT NULL REFERENCES users(id),
    shared_at       TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (group_id, vehicle_id)
);
CREATE INDEX idx_shared_vehicles_group ON shared_vehicles(group_id);
CREATE INDEX idx_shared_vehicles_vehicle ON shared_vehicles(vehicle_id);
```

### 5.2 现有表无变更

功能 2（排行榜）、3（费用看板）、4（加油站推荐）纯粹基于现有数据做 SQL 聚合查询，无需新增或修改任何现有表。

---

## 6. API 路由汇总（新增 6 条）

| HTTP 方法 | 路径 | 说明 |
|-----------|------|------|
| `POST` | `/api/v1/groups/{id}/shared-vehicles` | 共享车辆到群组 |
| `DELETE` | `/api/v1/groups/{id}/shared-vehicles/{vid}` | 取消车辆共享 |
| `GET` | `/api/v1/groups/{id}/shared-vehicles` | 获取群组内共享车辆列表 |
| `GET` | `/api/v1/groups/{id}/leaderboard` | 群组油耗排行榜 |
| `GET` | `/api/v1/groups/{id}/expense-stats` | 群组费用统计看板 |
| `GET` | `/api/v1/groups/{id}/stations` | 加油站推荐共享 |

---

## 7. 前端 Tab 结构（群组详情页）

改造后的群组详情页 Tabs 布局：

```
┌──────────────────────────────────────────────────────┐
│  群组信息  │  成员管理  │  数据汇总  │  排行榜 🏆  │  费用看板 💰  │  加油站 ⛽  │
└──────────────────────────────────────────────────────┘
```

**现有 3 个 Tab**（保持不变）：
- 群组信息（名称/描述/邀请码）
- 成员管理（成员列表/角色管理）
- 数据汇总（车辆维度的简单表格 + 共享开关）

**新增 3 个 Tab**：
- 排行榜 🏆（功能 2）
- 费用看板 💰（功能 3）
- 加油站 ⛽（功能 4）

---

## 8. i18n 新增翻译键（预估）

| 功能 | 预估新增键数 |
|------|-------------|
| 车辆共享 | ~15 键 |
| 排行榜 | ~20 键 |
| 费用看板 | ~20 键 |
| 加油站推荐 | ~15 键 |
| **合计** | **~70 键 × 3语 = ~210 条翻译** |

---

## 9. 建议实施顺序

| 步骤 | 功能 | 依赖 | 预估工作量 |
|------|------|------|-----------|
| ① | 车辆共享标记 | 无 | 后端 1.5 天 + 前端 1 天 |
| ② | 群组费用统计看板 | 无（纯聚合查询）| 后端 1 天 + 前端 1.5 天（ECharts 图表）|
| ③ | 群组油耗排行榜 | 无（可与 ② 并行）| 后端 0.5 天 + 前端 1 天 |
| ④ | 加油站推荐共享 | 无 | 后端 1 天 + 前端 1 天 |

**总计预估**：后端 4 天 + 前端 4.5 天 ≈ **1.5 周**

> 建议先做 ① 车辆共享，因为它涉及数据模型变更和权限改造，是后续功能的基础（共享车辆的记录也会进入排行榜和费用看板统计）。② 和 ③ 可并行开发，④ 最后。

---

## 10. 前端单位/货币国际化改造 ✅

> **完成日期**: 2026-03-31
>
> **问题**: GroupPage.tsx 中 15+ 处硬编码单位（`L/100km`、`L`、`km`）和货币符号（`¥`、`prefix: '¥'`），不尊重用户的计量系统偏好（metric/imperial）和币种设置

### 10.1 改造方案

**策略**：纯前端改造，后端不变。后端始终返回 metric 基准数据（L/100km、km、L），前端根据用户偏好转换。

**新增基础设施**（在 GroupPage 组件内）：

| 组件/函数 | 用途 |
|-----------|------|
| `useExchangeRateStore` | 获取汇率缓存（30min TTL） |
| `formatConvertedCost(amount)` | 智能货币换算：CNY→用户偏好币种，返回 `{ text, converted }` |
| `convertFuel(liters)` | L→gal（imperial）或原值（metric） |
| `convertDistance(km)` | km→mi（imperial）或原值（metric） |
| `convertEfficiency(l100km)` | L/100km→km/L / MPG（按 `efficiencyUnit`） |
| `<ConvertedCost amount={n}>` | 带"经换算"Tooltip 提示的金额展示组件 |

### 10.2 修复清单

| Tab | 位置 | 原硬编码 | 修复后 |
|-----|------|---------|--------|
| Overview | total_cost 列 | `val.toFixed(2)` | `<ConvertedCost>` |
| Overview | total_fuel 列 | `${val} L` | `${convertFuel(val)} ${fuelUnit}` |
| Overview | avg_efficiency 列 | `${val} L/100km` | `${convertEfficiency(val)} ${efficiencyUnit}` |
| Leaderboard | group_avg | `leaderboard.unit` | 按 metric 动态转换 |
| Leaderboard | item.value | `leaderboard.unit` | 按 metric 动态转换 |
| Expense Stats | 总费用卡片 | `prefix: '¥'` | `<ConvertedCost>` + 换算提示 |
| Expense Stats | 总油量卡片 | `suffix: 'L'` | `fuelUnit` |
| Expense Stats | 总里程卡片 | `suffix: 'km'` | `distanceUnit` |
| Expense Stats | 平均油耗卡片 | `suffix: 'L/100km'` | `efficiencyUnit` |
| Expense Stats | 趋势表费用列 | `` `¥${val}` `` | `<ConvertedCost>` |
| Expense Stats | 趋势表油量列 | `${val} L` | `${convertFuel(val)} ${fuelUnit}` |
| Expense Stats | 趋势表里程列 | `${val} km` | `${convertDistance(val)} ${distanceUnit}` |
| Expense Stats | 趋势表效率列 | `${val} L/100km` | `${convertEfficiency(val)} ${efficiencyUnit}` |
| Expense Stats | 成员占比 | `¥...L` | `<ConvertedCost>` + `fuelUnit` |
| Stations | 均价/最新价 | `user?.unit_system === 'imperial' ? 'gal' : 'L'` | `fuelUnit` |
| Stations | 货币 fallback | `user?.currency_code \|\| 'CNY'` | `currency` |

### 10.3 新增 i18n 键

| 键名 | zh-CN | en-US | ja-JP |
|------|-------|-------|-------|
| `group.converted` | 经换算 | Converted | 換算済み |
