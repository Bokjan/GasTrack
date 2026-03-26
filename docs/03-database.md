# GasTrack 数据库设计

## 1. 数据库选型：PostgreSQL

**选择理由：**
- 优秀的 JSON 支持（用户偏好等灵活数据）
- 强大的地理空间扩展 PostGIS（加油站位置）
- 完善的多语言/Unicode 支持
- 高性能聚合查询（统计报表）
- 成熟的生态与丰富的 ORM 支持

## 2. ER 关系概览

```
users ──1:N──► vehicles ──1:N──► fuel_records
  │                                    │
  │──1:N──► user_preferences           │
  │                                    │
  │──N:M──► groups (via group_members)  │
  │                                    │
  └────────────────────────────────────┘
              stats (computed)
```

## 3. 核心表结构

### 3.1 users - 用户表
```sql
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           VARCHAR(255) UNIQUE NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    nickname        VARCHAR(100) NOT NULL,
    avatar_url      VARCHAR(500),
    locale          VARCHAR(10) DEFAULT 'en-US',  -- 偏好语言: zh-CN/en-US/ja-JP
    timezone        VARCHAR(50) DEFAULT 'UTC',
    country_code    VARCHAR(5),            -- ISO 3166-1 alpha-2
    currency_code   VARCHAR(3) DEFAULT 'USD', -- ISO 4217
    unit_system     VARCHAR(10) DEFAULT 'metric', -- metric / imperial
    fuel_efficiency_unit VARCHAR(10) DEFAULT 'L/100km', -- L/100km / km/L / MPG
    status          VARCHAR(20) DEFAULT 'active', -- active/suspended/deleted
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_users_email ON users(email);
```

### 3.2 vehicles - 车辆表
```sql
CREATE TABLE vehicles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            VARCHAR(100) NOT NULL,  -- 用户自定义名称，如"家用车"
    vehicle_type    VARCHAR(20) NOT NULL DEFAULT 'car', -- car/motorcycle/other
    brand           VARCHAR(100),
    model           VARCHAR(100),
    year            INT,
    fuel_type       VARCHAR(20) NOT NULL,   -- gasoline/diesel/hybrid/electric
    tank_capacity   DECIMAL(6,2),           -- 油箱容量（升），燃油车使用
    battery_capacity DECIMAL(6,2),          -- 电池容量（kWh），电动车使用
    engine_cc       INT,                    -- 排量(cc)，燃油/混动车辆通用
    license_plate   VARCHAR(20),
    photo_url       VARCHAR(500),
    is_default      BOOLEAN DEFAULT false,
    is_archived     BOOLEAN DEFAULT false,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_vehicles_user ON vehicles(user_id);
```

### 3.3 fuel_records - 加油记录表
```sql
CREATE TABLE fuel_records (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id      UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id),

    -- 加油/充电数据（存原始值）
    fuel_amount     DECIMAL(8,3) NOT NULL,  -- 加油量/充电量
    fuel_unit       VARCHAR(5) DEFAULT 'L', -- L / gal / kWh
    unit_price      DECIMAL(10,4),          -- 单价
    total_cost      DECIMAL(10,2) NOT NULL, -- 总费用
    currency_code   VARCHAR(3) NOT NULL,    -- 币种

    -- 里程数据
    odometer        DECIMAL(10,1) NOT NULL, -- 当前里程表读数
    distance_unit   VARCHAR(5) DEFAULT 'km',-- km / mi

    -- 加油详情
    is_full_tank    BOOLEAN DEFAULT true,   -- 是否加满
    fuel_grade      VARCHAR(20),            -- 92/95/98/diesel 等
    station_name    VARCHAR(200),
    station_lat     DECIMAL(10,7),
    station_lng     DECIMAL(10,7),
    note            TEXT,
    receipt_url     VARCHAR(500),           -- 小票照片

    -- 计算字段（冗余存储提高查询性能）
    trip_distance   DECIMAL(10,1),          -- 本次行驶距离
    fuel_efficiency DECIMAL(6,2),           -- 油耗/电耗（L/100km 或 kWh/100km 存储基准）

    refuel_date     TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_fuel_records_vehicle ON fuel_records(vehicle_id, refuel_date DESC);
CREATE INDEX idx_fuel_records_user ON fuel_records(user_id, refuel_date DESC);
```

### 3.4 groups / group_members - 群组表
```sql
CREATE TABLE groups (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(100) NOT NULL,
    owner_id        UUID NOT NULL REFERENCES users(id),
    invite_code     VARCHAR(20) UNIQUE,
    max_members     INT DEFAULT 10,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE group_members (
    group_id        UUID REFERENCES groups(id) ON DELETE CASCADE,
    user_id         UUID REFERENCES users(id) ON DELETE CASCADE,
    role            VARCHAR(20) DEFAULT 'member', -- owner/admin/member
    joined_at       TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (group_id, user_id)
);
```

### 3.5 refresh_tokens - 刷新令牌表
```sql
CREATE TABLE refresh_tokens (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      VARCHAR(255) NOT NULL,
    device_info     VARCHAR(255),
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id);
```

## 4. 缓存策略（无 Redis，进程内缓存）

本项目初期不引入 Redis，使用 Go 进程内缓存（`go-cache` 或 `sync.Map`）替代：

**可行性分析：**
- 油耗记录系统属于**读多写少、用户数据隔离**的场景
- 初期用户量不大，单实例部署，进程内缓存完全够用
- 省去 Redis 运维成本，降低部署复杂度
- 当未来需要多实例部署或需要分布式缓存时，再引入 Redis

**缓存项：**

| 缓存内容 | 实现方式 | TTL |
|----------|---------|-----|
| 用户资料 | go-cache (内存) | 30 min |
| 车辆统计 | go-cache (内存) | 10 min |
| 翻译资源 | 启动时全量加载到内存 | 不过期 |
| API 限流 | Nginx `limit_req` 模块 | - |
| JWT 黑名单 | PostgreSQL 表 + 内存缓存 | Token 剩余有效期 |

**限流方案（替代 Redis）：**
- 全局限流：Nginx `limit_req_zone` 实现 IP 级别限流
- 业务限流：Go 内置 `golang.org/x/time/rate` 令牌桶算法
- 登录防暴力破解：PostgreSQL 记录失败次数 + 内存缓存
