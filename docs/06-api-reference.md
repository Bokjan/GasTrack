# GasTrack API 接口文档 (V1)

> **Base URL**: `http://localhost:8098/api/v1`
>
> **认证方式**: Bearer Token（JWT）
>
> **内容类型**: `application/json; charset=utf-8`
>
> **更新日期**: 2026-03-26

---

## 1. 通用约定

### 1.1 统一响应格式

**成功响应**
```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

**分页响应**
```json
{
  "code": 0,
  "message": "success",
  "data": [ ... ],
  "meta": {
    "page": 1,
    "page_size": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

**错误响应**
```json
{
  "code": 4000,
  "message": "error description"
}
```

**校验错误响应 (422)**
```json
{
  "code": 4220,
  "message": "validation_error",
  "errors": { ... }
}
```

### 1.2 业务错误码

| HTTP 状态码 | 业务错误码 | 说明 |
|------------|-----------|------|
| 400 | 4000 | 错误请求（参数不合法） |
| 401 | 4010 | 未认证（Token 缺失/无效/过期） |
| 403 | 4030 | 无权限 |
| 404 | 4040 | 资源不存在 |
| 422 | 4220 | 校验错误 |
| 429 | 4290 | 请求频率超限 |
| 500 | 5000 | 服务器内部错误 |

### 1.3 认证说明

- 需认证的接口须在 Header 中携带：`Authorization: Bearer <access_token>`
- Access Token 有效期 15 分钟
- Refresh Token 有效期 7 天
- Token 过期后使用 Refresh Token 续期

### 1.4 限流

- 全局限流：100 请求/秒/IP，突发上限 200
- 超限返回 `429 Too Many Requests`

---

## 2. 认证接口 (Auth)

### 2.1 用户注册

```
POST /api/v1/auth/register
```

**无需认证**

**请求体**
| 字段 | 类型 | 必填 | 校验 | 说明 |
|------|------|------|------|------|
| email | string | ✅ | 合法邮箱 | 用户邮箱 |
| password | string | ✅ | 8-72 字符 | 密码 |
| nickname | string | ✅ | 1-100 字符 | 昵称 |
| locale | string | - | `en-US` / `zh-CN` / `ja-JP` | 偏好语言 |

**请求示例**
```json
{
  "email": "user@example.com",
  "password": "mypassword123",
  "nickname": "张三",
  "locale": "zh-CN"
}
```

**成功响应** `201 Created`
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2g...",
    "expires_in": 900,
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "user@example.com",
      "nickname": "张三",
      "locale": "zh-CN",
      "timezone": "UTC",
      "currency_code": "USD",
      "unit_system": "metric",
      "fuel_efficiency_unit": "L/100km",
      "status": "active",
      "created_at": "2026-03-26T10:00:00Z"
    }
  }
}
```

---

### 2.2 用户登录

```
POST /api/v1/auth/login
```

**无需认证**

**请求体**
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | ✅ | 用户邮箱 |
| password | string | ✅ | 密码 |

**成功响应** `200 OK`

响应格式同「2.1 注册」。

---

### 2.3 刷新 Token

```
POST /api/v1/auth/refresh
```

**无需认证**

**请求体**
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| refresh_token | string | ✅ | 刷新令牌 |

**成功响应** `200 OK`

响应格式同「2.1 注册」（返回新的 access_token 和 refresh_token）。

---

### 2.4 用户登出

```
POST /api/v1/auth/logout
```

**🔒 需要认证**

**成功响应** `204 No Content`

（无响应体）

---

## 3. 用户接口 (User)

### 3.1 获取当前用户资料

```
GET /api/v1/users/me
```

**🔒 需要认证**

**成功响应** `200 OK`
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "nickname": "张三",
    "avatar_url": "https://example.com/avatar.jpg",
    "locale": "zh-CN",
    "timezone": "Asia/Shanghai",
    "country_code": "CN",
    "currency_code": "CNY",
    "unit_system": "metric",
    "fuel_efficiency_unit": "L/100km",
    "status": "active",
    "last_login_at": "2026-03-26T10:00:00Z",
    "created_at": "2026-03-01T08:00:00Z"
  }
}
```

---

### 3.2 更新用户资料

```
PATCH /api/v1/users/me
```

**🔒 需要认证**

**请求体**（所有字段均可选）
| 字段 | 类型 | 校验 | 说明 |
|------|------|------|------|
| nickname | string | 1-100 字符 | 昵称 |
| avatar_url | string | 合法 URL | 头像 URL |
| locale | string | `en-US` / `zh-CN` / `ja-JP` | 偏好语言 |
| timezone | string | 最长 50 字符 | 时区（如 `Asia/Shanghai`） |
| country_code | string | 2 字符 | ISO 3166-1 alpha-2 国家代码 |
| currency_code | string | 3 字符 | ISO 4217 货币代码 |
| unit_system | string | `metric` / `imperial` | 计量单位体系 |
| fuel_efficiency_unit | string | `L/100km` / `km/L` / `MPG` | 油耗单位 |

**成功响应** `200 OK`

返回更新后的完整用户对象（格式同「3.1」）。

---

### 3.3 修改密码

```
PUT /api/v1/users/me/password
```

**🔒 需要认证**

**请求体**
| 字段 | 类型 | 必填 | 校验 | 说明 |
|------|------|------|------|------|
| old_password | string | ✅ | - | 旧密码 |
| new_password | string | ✅ | 8-72 字符 | 新密码 |

**成功响应** `204 No Content`

---

### 3.4 注销账号

```
DELETE /api/v1/users/me
```

**🔒 需要认证**

**成功响应** `204 No Content`

---

## 4. 车辆接口 (Vehicle)

### 4.1 获取车辆列表

```
GET /api/v1/vehicles
```

**🔒 需要认证**

**查询参数**
| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| include_archived | string | `"false"` | 是否包含已归档车辆 |

**成功响应** `200 OK`
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "name": "家用车",
      "vehicle_type": "car",
      "brand": "Toyota",
      "model": "Camry",
      "year": 2023,
      "fuel_type": "gasoline",
      "tank_capacity": 60.0,
      "engine_cc": 2000,
      "license_plate": "京A12345",
      "photo_url": "",
      "is_default": true,
      "is_archived": false,
      "created_at": "2026-03-01T08:00:00Z",
      "updated_at": "2026-03-20T12:00:00Z"
    }
  ]
}
```

---

### 4.2 添加车辆

```
POST /api/v1/vehicles
```

**🔒 需要认证**

**请求体**
| 字段 | 类型 | 必填 | 校验 | 说明 |
|------|------|------|------|------|
| name | string | ✅ | 1-100 字符 | 车辆名称 |
| vehicle_type | string | ✅ | `car` / `motorcycle` / `other` | 车辆类型 |
| brand | string | - | 最长 100 字符 | 品牌 |
| model | string | - | 最长 100 字符 | 型号 |
| year | int | - | 1900-2100 | 年份 |
| fuel_type | string | ✅ | `gasoline` / `diesel` / `hybrid` / `electric` | 燃油类型 |
| tank_capacity | float | - | > 0 | 油箱容量（升） |
| engine_cc | int | - | > 0 | 排量（cc） |
| license_plate | string | - | 最长 20 字符 | 车牌号 |
| is_default | bool | - | - | 是否设为默认车辆 |

**成功响应** `201 Created`

返回创建的车辆对象（格式同「4.1」的列表项）。

---

### 4.3 获取车辆详情

```
GET /api/v1/vehicles/{id}
```

**🔒 需要认证**

**路径参数**
| 参数 | 类型 | 说明 |
|------|------|------|
| id | UUID | 车辆 ID |

**成功响应** `200 OK`

返回车辆对象（格式同「4.1」的列表项）。

---

### 4.4 编辑车辆

```
PATCH /api/v1/vehicles/{id}
```

**🔒 需要认证**

**请求体**（所有字段均可选，格式同「4.2」，另加）
| 字段 | 类型 | 说明 |
|------|------|------|
| is_archived | bool | 是否归档 |

**成功响应** `200 OK`

返回更新后的车辆对象。

---

### 4.5 删除车辆

```
DELETE /api/v1/vehicles/{id}
```

**🔒 需要认证**

**成功响应** `204 No Content`

---

## 5. 加油记录接口 (Fuel Record)

### 5.1 获取加油记录列表（分页）

```
GET /api/v1/vehicles/{id}/records
```

**🔒 需要认证**

**路径参数**
| 参数 | 类型 | 说明 |
|------|------|------|
| id | UUID | 车辆 ID |

**查询参数**
| 参数 | 类型 | 默认值 | 范围 | 说明 |
|------|------|--------|------|------|
| page | int | 1 | ≥ 1 | 页码 |
| page_size | int | 20 | 1-100 | 每页条数 |

**成功响应** `200 OK`
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": "770e8400-e29b-41d4-a716-446655440002",
      "vehicle_id": "660e8400-e29b-41d4-a716-446655440001",
      "fuel_amount": 45.5,
      "fuel_unit": "L",
      "unit_price": 7.89,
      "total_cost": 358.99,
      "currency_code": "CNY",
      "odometer": 15230.5,
      "distance_unit": "km",
      "is_full_tank": true,
      "fuel_grade": "95",
      "station_name": "中石化望京站",
      "station_lat": 39.9876543,
      "station_lng": 116.4712345,
      "note": "",
      "receipt_url": "",
      "trip_distance": 520.3,
      "fuel_efficiency": 8.75,
      "refuel_date": "2026-03-25T14:30:00Z",
      "created_at": "2026-03-25T14:35:00Z",
      "updated_at": "2026-03-25T14:35:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "page_size": 20,
    "total": 42,
    "total_pages": 3
  }
}
```

---

### 5.2 添加加油记录

```
POST /api/v1/vehicles/{id}/records
```

**🔒 需要认证**

**请求体**
| 字段 | 类型 | 必填 | 校验 | 说明 |
|------|------|------|------|------|
| fuel_amount | float | ✅ | > 0 | 加油量 |
| fuel_unit | string | - | `L` / `gal` | 燃油单位（默认 `L`） |
| unit_price | float | - | ≥ 0 | 单价 |
| total_cost | float | ✅ | > 0 | 总费用 |
| currency_code | string | ✅ | 3 字符 | 货币代码（如 `CNY`） |
| odometer | float | ✅ | > 0 | 里程表读数 |
| distance_unit | string | - | `km` / `mi` | 距离单位（默认 `km`） |
| is_full_tank | bool | - | - | 是否加满 |
| fuel_grade | string | - | 最长 20 字符 | 燃油标号（如 `92`/`95`/`98`） |
| station_name | string | - | 最长 200 字符 | 加油站名称 |
| station_lat | float | - | - | 加油站纬度 |
| station_lng | float | - | - | 加油站经度 |
| note | string | - | 最长 1000 字符 | 备注 |
| refuel_date | string | ✅ | ISO 8601 | 加油日期 |

**成功响应** `201 Created`

返回创建的加油记录对象（格式同「5.1」的列表项）。

---

### 5.3 获取加油记录详情

```
GET /api/v1/vehicles/{id}/records/{rid}
```

**🔒 需要认证**

**成功响应** `200 OK`

返回加油记录对象（格式同「5.1」的列表项）。

---

### 5.4 编辑加油记录

```
PATCH /api/v1/vehicles/{id}/records/{rid}
```

**🔒 需要认证**

**请求体**：所有字段均可选，格式同「5.2」。

**成功响应** `200 OK`

返回更新后的加油记录对象。

---

### 5.5 删除加油记录

```
DELETE /api/v1/vehicles/{id}/records/{rid}
```

**🔒 需要认证**

**成功响应** `204 No Content`

---

## 6. 统计接口 (Stats)

### 6.1 车辆统计

```
GET /api/v1/vehicles/{id}/stats
```

**🔒 需要认证**

**成功响应** `200 OK`
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "vehicle_id": "660e8400-e29b-41d4-a716-446655440001",
    "vehicle_name": "家用车",
    "total_records": 42,
    "total_fuel": 1890.5,
    "total_cost": 14920.80,
    "total_distance": 21500.0,
    "avg_efficiency": 8.79,
    "best_efficiency": 6.50,
    "worst_efficiency": 12.30,
    "avg_cost_per_km": 0.69,
    "avg_cost_per_fill": 355.26,
    "currency_code": "CNY",
    "fuel_efficiency_unit": "L/100km"
  }
}
```

---

### 6.2 全局统计总览

```
GET /api/v1/stats/overview
```

**🔒 需要认证**

**成功响应** `200 OK`
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total_vehicles": 2,
    "total_records": 85,
    "total_fuel": 3780.0,
    "total_cost": 29800.50,
    "total_distance": 43000.0,
    "avg_consumption": 8.79,
    "currency_code": "CNY",
    "vehicles": [
      {
        "vehicle_id": "...",
        "vehicle_name": "家用车",
        "total_records": 42,
        "total_fuel": 1890.5,
        "total_cost": 14920.80,
        "total_distance": 21500.0,
        "avg_efficiency": 8.79,
        "best_efficiency": 6.50,
        "worst_efficiency": 12.30,
        "avg_cost_per_km": 0.69,
        "avg_cost_per_fill": 355.26,
        "currency_code": "CNY",
        "fuel_efficiency_unit": "L/100km"
      }
    ]
  }
}
```

---

### 6.3 油耗趋势

```
GET /api/v1/vehicles/{id}/efficiency-trend
```

**🔒 需要认证**

**查询参数**
| 参数 | 类型 | 默认值 | 范围 | 说明 |
|------|------|--------|------|------|
| limit | int | 30 | 1-100 | 返回最近 N 条趋势数据 |

**成功响应** `200 OK`
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "vehicle_id": "660e8400-e29b-41d4-a716-446655440001",
    "vehicle_name": "家用车",
    "efficiency_unit": "L/100km",
    "items": [
      {
        "date": "2026-03-25",
        "fuel_efficiency": 8.5,
        "trip_distance": 520.3
      },
      {
        "date": "2026-03-18",
        "fuel_efficiency": 9.2,
        "trip_distance": 480.0
      }
    ]
  }
}
```

---

### 6.4 健康检查

```
GET /api/v1/health
```

**无需认证**

**成功响应** `200 OK`
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "ok"
  }
}
```

---

## 7. API 路由汇总表

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| POST | `/api/v1/auth/register` | ❌ | 用户注册 |
| POST | `/api/v1/auth/login` | ❌ | 用户登录 |
| POST | `/api/v1/auth/refresh` | ❌ | 刷新 Token |
| POST | `/api/v1/auth/logout` | ✅ | 用户登出 |
| GET | `/api/v1/health` | ❌ | 健康检查 |
| GET | `/api/v1/users/me` | ✅ | 获取用户资料 |
| PATCH | `/api/v1/users/me` | ✅ | 更新用户资料 |
| PUT | `/api/v1/users/me/password` | ✅ | 修改密码 |
| DELETE | `/api/v1/users/me` | ✅ | 注销账号 |
| GET | `/api/v1/vehicles` | ✅ | 车辆列表 |
| POST | `/api/v1/vehicles` | ✅ | 添加车辆 |
| GET | `/api/v1/vehicles/{id}` | ✅ | 车辆详情 |
| PATCH | `/api/v1/vehicles/{id}` | ✅ | 编辑车辆 |
| DELETE | `/api/v1/vehicles/{id}` | ✅ | 删除车辆 |
| GET | `/api/v1/vehicles/{id}/records` | ✅ | 加油记录列表 |
| POST | `/api/v1/vehicles/{id}/records` | ✅ | 添加加油记录 |
| GET | `/api/v1/vehicles/{id}/records/{rid}` | ✅ | 加油记录详情 |
| PATCH | `/api/v1/vehicles/{id}/records/{rid}` | ✅ | 编辑加油记录 |
| DELETE | `/api/v1/vehicles/{id}/records/{rid}` | ✅ | 删除加油记录 |
| GET | `/api/v1/vehicles/{id}/stats` | ✅ | 车辆统计 |
| GET | `/api/v1/vehicles/{id}/efficiency-trend` | ✅ | 油耗趋势 |
| GET | `/api/v1/stats/overview` | ✅ | 全局统计总览 |
