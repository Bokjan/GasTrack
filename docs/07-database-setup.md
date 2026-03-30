# GasTrack 数据库文档

> **数据库**: PostgreSQL 16
>
> **ORM**: GORM (Go)
>
> **迁移方式**: GORM AutoMigrate（开发阶段）
>
> **更新日期**: 2026-03-26

---

## 1. 环境搭建

### 1.1 Docker Compose 启动 PostgreSQL

```bash
# 在项目根目录执行
docker compose up -d
```

`docker-compose.yaml` 内容：

```yaml
services:
  postgres:
    image: postgres:16-alpine
    container_name: gastrack-postgres
    restart: unless-stopped
    ports:
      - "5432:5432"
    environment:
      POSTGRES_USER: gastrack
      POSTGRES_PASSWORD: gastrack
      POSTGRES_DB: gastrack
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U gastrack -d gastrack"]
      interval: 5s
      timeout: 5s
      retries: 5

volumes:
  pgdata:
    driver: local
```

### 1.2 连接信息

| 配置项 | 值 |
|--------|-----|
| Host | `localhost` |
| Port | `5432` |
| User | `gastrack` |
| Password | `gastrack` |
| Database | `gastrack` |
| SSL Mode | `disable` |

### 1.3 连接字符串

```
host=localhost port=5432 user=gastrack password=gastrack dbname=gastrack sslmode=disable
```

---

## 2. 初始化建库 SQL

> **注意**: 项目使用 GORM AutoMigrate 自动建表，正常启动后端服务即可自动创建表结构。以下 SQL 供手动初始化或参考使用。

```sql
-- ============================================================
-- GasTrack 数据库初始化脚本
-- PostgreSQL 16+
-- 执行方式: psql -h localhost -U gastrack -d gastrack -f init.sql
-- ============================================================

-- 启用 UUID 生成函数（PostgreSQL 13+ 内置 gen_random_uuid()）
-- CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================
-- 1. users 用户表
-- ============================================================
CREATE TABLE IF NOT EXISTS users (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email                VARCHAR(255) UNIQUE NOT NULL,
    password_hash        VARCHAR(255) NOT NULL,
    nickname             VARCHAR(100) NOT NULL,
    avatar_url           VARCHAR(500) DEFAULT '',
    locale               VARCHAR(10) DEFAULT 'en-US',
    timezone             VARCHAR(50) DEFAULT 'UTC',
    country_code         VARCHAR(5) DEFAULT '',
    currency_code        VARCHAR(3) DEFAULT 'USD',
    unit_system          VARCHAR(10) DEFAULT 'metric',
    fuel_efficiency_unit VARCHAR(10) DEFAULT 'L/100km',
    status               VARCHAR(20) DEFAULT 'active',
    last_login_at        TIMESTAMPTZ,
    created_at           TIMESTAMPTZ DEFAULT NOW(),
    updated_at           TIMESTAMPTZ DEFAULT NOW(),
    deleted_at           TIMESTAMPTZ                     -- GORM 软删除
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

COMMENT ON TABLE  users IS '用户表';
COMMENT ON COLUMN users.locale IS '偏好语言: en-US / zh-CN / ja-JP';
COMMENT ON COLUMN users.unit_system IS '计量体系: metric / imperial';
COMMENT ON COLUMN users.fuel_efficiency_unit IS '油耗单位: L/100km / km/L / MPG';
COMMENT ON COLUMN users.status IS '状态: active / suspended / deleted';

-- ============================================================
-- 2. vehicles 车辆表
-- ============================================================
CREATE TABLE IF NOT EXISTS vehicles (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name           VARCHAR(100) NOT NULL,
    vehicle_type   VARCHAR(20) NOT NULL DEFAULT 'car',
    brand          VARCHAR(100) DEFAULT '',
    model          VARCHAR(100) DEFAULT '',
    year           INT DEFAULT 0,
    fuel_type      VARCHAR(20) NOT NULL DEFAULT 'gasoline',
    tank_capacity  DECIMAL(6,2) DEFAULT 0,
    battery_capacity DECIMAL(6,2) DEFAULT 0,
    engine_cc      INT DEFAULT 0,
    license_plate  VARCHAR(20) DEFAULT '',
    photo_url      VARCHAR(500) DEFAULT '',
    is_default     BOOLEAN DEFAULT false,
    is_archived    BOOLEAN DEFAULT false,
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    updated_at     TIMESTAMPTZ DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_vehicles_user_id ON vehicles(user_id);
CREATE INDEX IF NOT EXISTS idx_vehicles_deleted_at ON vehicles(deleted_at);

COMMENT ON TABLE  vehicles IS '车辆表';
COMMENT ON COLUMN vehicles.vehicle_type IS '车辆类型: car / motorcycle / other';
COMMENT ON COLUMN vehicles.fuel_type IS '燃油/能源类型: gasoline / diesel / hybrid / electric';
COMMENT ON COLUMN vehicles.tank_capacity IS '油箱容量（升），燃油车使用';
COMMENT ON COLUMN vehicles.battery_capacity IS '电池容量（kWh），电动车使用';
COMMENT ON COLUMN vehicles.engine_cc IS '排量（cc），燃油/混动车辆通用';

-- ============================================================
-- 3. fuel_records 加油记录表
-- ============================================================
CREATE TABLE IF NOT EXISTS fuel_records (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id      UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id),

    -- 加油/充电数据
    fuel_amount     DECIMAL(8,3) NOT NULL,
    fuel_unit       VARCHAR(5) DEFAULT 'L',
    unit_price      DECIMAL(10,4) DEFAULT 0,
    total_cost      DECIMAL(10,2) NOT NULL,
    currency_code   VARCHAR(3) NOT NULL,

    -- 里程数据
    odometer        DECIMAL(10,1) NOT NULL,
    distance_unit   VARCHAR(5) DEFAULT 'km',

    -- 加油详情
    is_full_tank    BOOLEAN DEFAULT true,
    fuel_grade      VARCHAR(20) DEFAULT '',
    station_name    VARCHAR(200) DEFAULT '',
    station_lat     DECIMAL(10,7) DEFAULT 0,
    station_lng     DECIMAL(10,7) DEFAULT 0,
    note            TEXT DEFAULT '',
    receipt_url     VARCHAR(500) DEFAULT '',

    -- 计算字段（冗余存储）
    trip_distance   DECIMAL(10,1) DEFAULT 0,
    fuel_efficiency DECIMAL(6,2) DEFAULT 0,

    refuel_date     TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_fuel_records_vehicle ON fuel_records(vehicle_id, refuel_date DESC);
CREATE INDEX IF NOT EXISTS idx_fuel_records_user ON fuel_records(user_id, refuel_date DESC);
CREATE INDEX IF NOT EXISTS idx_fuel_records_deleted_at ON fuel_records(deleted_at);

COMMENT ON TABLE  fuel_records IS '加油/充电记录表';
COMMENT ON COLUMN fuel_records.fuel_unit IS '燃油/能量单位: L / gal / kWh';
COMMENT ON COLUMN fuel_records.distance_unit IS '距离单位: km / mi';
COMMENT ON COLUMN fuel_records.fuel_efficiency IS '油耗/电耗值（L/100km 或 kWh/100km 存储基准）';
COMMENT ON COLUMN fuel_records.trip_distance IS '本次行驶距离（根据里程表差值计算）';

-- ============================================================
-- 4. refresh_tokens 刷新令牌表
-- ============================================================
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  VARCHAR(255) NOT NULL,
    device_info VARCHAR(255) DEFAULT '',
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_deleted_at ON refresh_tokens(deleted_at);

COMMENT ON TABLE  refresh_tokens IS '刷新令牌表';
COMMENT ON COLUMN refresh_tokens.token_hash IS 'Refresh Token 的哈希值';
COMMENT ON COLUMN refresh_tokens.device_info IS '设备信息';

-- ============================================================
-- 5. 预留：群组表（P1 阶段实现）
-- ============================================================
-- CREATE TABLE IF NOT EXISTS groups ( ... );
-- CREATE TABLE IF NOT EXISTS group_members ( ... );
```

---

## 3. 表结构说明

### 3.1 ER 关系图

```
users ──1:N──► vehicles ──1:N──► fuel_records
  │               │
  │──1:N──► refresh_tokens
  │──1:N──► invite_codes
  │──1:N──► reminders (via vehicles)
  └──1:N──► notifications
```

### 3.2 公共字段

所有表继承自 GORM `BaseModel`，包含以下公共字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键，自动生成 `gen_random_uuid()` |
| created_at | TIMESTAMPTZ | 创建时间（自动填充） |
| updated_at | TIMESTAMPTZ | 更新时间（自动更新） |
| deleted_at | TIMESTAMPTZ | 软删除时间（GORM 软删除） |

### 3.3 users 用户表

| 字段 | 类型 | 约束 | 默认值 | 说明 |
|------|------|------|--------|------|
| id | UUID | PK | `gen_random_uuid()` | 用户 ID |
| email | VARCHAR(255) | UNIQUE, NOT NULL | - | 邮箱 |
| password_hash | VARCHAR(255) | NOT NULL | - | 密码哈希（bcrypt） |
| nickname | VARCHAR(100) | NOT NULL | - | 昵称 |
| avatar_url | VARCHAR(500) | - | `''` | 头像 URL |
| locale | VARCHAR(10) | - | `'en-US'` | 偏好语言 |
| timezone | VARCHAR(50) | - | `'UTC'` | 时区 |
| country_code | VARCHAR(5) | - | `''` | 国家代码 |
| currency_code | VARCHAR(3) | - | `'USD'` | 货币代码 |
| unit_system | VARCHAR(10) | - | `'metric'` | 计量体系 |
| fuel_efficiency_unit | VARCHAR(10) | - | `'L/100km'` | 油耗单位 |
| status | VARCHAR(20) | - | `'active'` | 状态 |
| last_login_at | TIMESTAMPTZ | - | NULL | 最后登录时间 |

**索引**: `idx_users_email`(email), `idx_users_deleted_at`(deleted_at)

### 3.4 vehicles 车辆表

| 字段 | 类型 | 约束 | 默认值 | 说明 |
|------|------|------|--------|------|
| id | UUID | PK | `gen_random_uuid()` | 车辆 ID |
| user_id | UUID | FK → users(id), NOT NULL | - | 所属用户 |
| name | VARCHAR(100) | NOT NULL | - | 车辆名称 |
| vehicle_type | VARCHAR(20) | NOT NULL | `'car'` | 车辆类型 |
| brand | VARCHAR(100) | - | `''` | 品牌 |
| model | VARCHAR(100) | - | `''` | 型号 |
| year | INT | - | 0 | 年份 |
| fuel_type | VARCHAR(20) | NOT NULL | `'gasoline'` | 燃油/能源类型 |
| tank_capacity | DECIMAL(6,2) | - | 0 | 油箱容量（燃油车） |
| battery_capacity | DECIMAL(6,2) | - | 0 | 电池容量 kWh（电动车） |
| engine_cc | INT | - | 0 | 排量 cc（燃油/混动） |
| license_plate | VARCHAR(20) | - | `''` | 车牌号 |
| photo_url | VARCHAR(500) | - | `''` | 照片 |
| is_default | BOOLEAN | - | `false` | 是否默认 |
| is_archived | BOOLEAN | - | `false` | 是否归档 |

**索引**: `idx_vehicles_user_id`(user_id), `idx_vehicles_deleted_at`(deleted_at)

### 3.5 fuel_records 加油记录表

| 字段 | 类型 | 约束 | 默认值 | 说明 |
|------|------|------|--------|------|
| id | UUID | PK | `gen_random_uuid()` | 记录 ID |
| vehicle_id | UUID | FK → vehicles(id), NOT NULL | - | 所属车辆 |
| user_id | UUID | FK → users(id), NOT NULL | - | 所属用户 |
| fuel_amount | DECIMAL(8,3) | NOT NULL | - | 加油量/充电量 |
| fuel_unit | VARCHAR(5) | - | `'L'` | 燃油/能量单位（L/gal/kWh） |
| unit_price | DECIMAL(10,4) | - | 0 | 单价 |
| total_cost | DECIMAL(10,2) | NOT NULL | - | 总费用 |
| currency_code | VARCHAR(3) | NOT NULL | - | 货币代码 |
| odometer | DECIMAL(10,1) | NOT NULL | - | 里程表读数 |
| distance_unit | VARCHAR(5) | - | `'km'` | 距离单位 |
| is_full_tank | BOOLEAN | - | `true` | 是否加满 |
| fuel_grade | VARCHAR(20) | - | `''` | 燃油标号 |
| station_name | VARCHAR(200) | - | `''` | 加油站名称 |
| station_lat | DECIMAL(10,7) | - | 0 | 加油站纬度 |
| station_lng | DECIMAL(10,7) | - | 0 | 加油站经度 |
| note | TEXT | - | `''` | 备注 |
| receipt_url | VARCHAR(500) | - | `''` | 小票照片 |
| trip_distance | DECIMAL(10,1) | - | 0 | 本次行驶距离 |
| fuel_efficiency | DECIMAL(6,2) | - | 0 | 油耗/电耗 |
| refuel_date | TIMESTAMPTZ | NOT NULL | - | 加油日期 |

**索引**: `idx_fuel_records_vehicle`(vehicle_id, refuel_date DESC), `idx_fuel_records_user`(user_id, refuel_date DESC)

### 3.6 refresh_tokens 刷新令牌表

| 字段 | 类型 | 约束 | 默认值 | 说明 |
|------|------|------|--------|------|
| id | UUID | PK | `gen_random_uuid()` | 令牌 ID |
| user_id | UUID | FK → users(id), NOT NULL | - | 所属用户 |
| token_hash | VARCHAR(255) | NOT NULL | - | Token 哈希 |
| device_info | VARCHAR(255) | - | `''` | 设备信息 |
| expires_at | TIMESTAMPTZ | NOT NULL | - | 过期时间 |

**索引**: `idx_refresh_tokens_user_id`(user_id)

---

## 4. 数据库配置

### 4.1 配置文件 (`server/config.yaml`)

```yaml
database:
  host: localhost
  port: 5432
  user: gastrack
  password: gastrack
  dbname: gastrack
  sslmode: disable
  max_open_conns: 25
  max_idle_conns: 5
```

### 4.2 连接池参数

| 参数 | 值 | 说明 |
|------|-----|------|
| max_open_conns | 25 | 最大打开连接数 |
| max_idle_conns | 5 | 最大空闲连接数 |
| conn_max_lifetime | 30min | 连接最大生存时间 |
| conn_max_idle_time | 5min | 空闲连接最大保持时间 |

### 4.3 GORM AutoMigrate

启动后端服务时，GORM 会自动执行 `AutoMigrate`，创建/更新以下 7 张表：

```go
db.AutoMigrate(
    &model.User{},
    &model.Vehicle{},
    &model.FuelRecord{},
    &model.RefreshToken{},
    &model.InviteCode{},
    &model.Reminder{},
    &model.Notification{},
)
```

> AutoMigrate 只会创建缺失的表/列/索引，不会删除现有列或更改列类型。

---

## 5. 常见操作

### 5.1 手动连接数据库

```bash
# 通过 Docker 容器连接
docker exec -it gastrack-postgres psql -U gastrack -d gastrack

# 通过 psql 客户端连接
psql -h localhost -p 5432 -U gastrack -d gastrack
```

### 5.2 查看表结构

```sql
\dt                           -- 列出所有表
\d users                      -- 查看 users 表结构
\d fuel_records               -- 查看 fuel_records 表结构
```

### 5.3 重置数据库

```bash
# 停止并删除容器和数据卷
docker compose down -v

# 重新启动（数据库将被重新创建）
docker compose up -d
```

### 5.4 备份与恢复

```bash
# 备份
docker exec gastrack-postgres pg_dump -U gastrack gastrack > backup.sql

# 恢复
docker exec -i gastrack-postgres psql -U gastrack -d gastrack < backup.sql
```
