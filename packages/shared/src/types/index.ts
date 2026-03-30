// ============================================================
// GasTrack 共享类型定义
// 与后端 DTO JSON 字段完全对齐
// ============================================================

// ---------- 通用 ----------

export interface ApiResponse<T = unknown> {
  code: number;
  message: string;
  data: T;
}

/** 分页元信息，与后端 PageMeta 对齐 */
export interface PageMeta {
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
}

/**
 * 后端分页响应格式：
 * { code: 0, message: "success", data: T[], meta: { page, page_size, total, total_pages } }
 */
export interface PaginatedResponse<T> {
  code: number;
  message: string;
  data: T[];
  meta: PageMeta;
}

// ---------- Auth ----------

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  nickname: string;
  invite_code?: string;
}

/** 后端 AuthResponse: { access_token, refresh_token, expires_in, user } */
export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: User;
}

/** 兼容旧代码的 alias */
export type AuthTokens = AuthResponse;

export interface RefreshRequest {
  refresh_token: string;
}

// ---------- Registration Mode ----------

/** 后端 GET /api/v1/auth/registration-mode 返回 */
export interface RegistrationModeResponse {
  mode: 'open' | 'invite_only' | 'closed';
}

/** 邀请码验证响应 */
export interface ValidateInviteResponse {
  valid: boolean;
  remaining_uses?: number;
  expires_at?: string;
}

/** 邀请码详情 */
export interface InviteCode {
  id: string;
  code: string;
  created_by: string;
  creator_name?: string;
  max_uses: number;
  use_count: number;
  remaining_uses: number;
  expires_at?: string;
  note?: string;
  is_active: boolean;
  is_valid: boolean;
  created_at: string;
}

/** 创建邀请码请求 */
export interface CreateInviteRequest {
  max_uses?: number;
  expires_at?: string;
  note?: string;
}

/** 更新邀请码请求 */
export interface UpdateInviteRequest {
  is_active?: boolean;
  note?: string;
}

// ---------- User ----------

/** 后端 UserResponse 字段对齐 */
export interface User {
  id: string;
  email: string;
  nickname: string;
  avatar_url?: string;
  locale: string;
  timezone: string;
  country_code?: string;
  currency_code: string;
  unit_system: string;
  fuel_efficiency_unit: string;
  status: string;
  last_login_at?: string;
  created_at: string;
}

/** 后端 UpdateUserRequest 字段对齐 */
export interface UpdateUserRequest {
  nickname?: string;
  avatar_url?: string;
  locale?: string;
  timezone?: string;
  country_code?: string;
  currency_code?: string;
  unit_system?: string;
  fuel_efficiency_unit?: string;
}

export interface ChangePasswordRequest {
  old_password: string;
  new_password: string;
}

// ---------- Vehicle ----------

export type VehicleType = 'car' | 'motorcycle' | 'other';
export type FuelType = 'gasoline' | 'diesel' | 'hybrid' | 'electric';

/** 后端 VehicleResponse 字段对齐 */
export interface Vehicle {
  id: string;
  name: string;
  vehicle_type: VehicleType;
  brand: string;
  model: string;
  year: number;
  fuel_type: FuelType;
  fuel_grade?: string;
  tank_capacity: number;
  battery_capacity?: number;
  engine_cc?: number;
  license_plate?: string;
  photo_url?: string;
  is_default: boolean;
  is_archived: boolean;
  created_at: string;
  updated_at: string;
}

/** 后端 CreateVehicleRequest 字段对齐 */
export interface CreateVehicleRequest {
  name: string;
  vehicle_type: VehicleType;
  brand: string;
  model: string;
  year: number;
  fuel_type: FuelType;
  fuel_grade?: string;
  tank_capacity: number;
  battery_capacity?: number;
  engine_cc?: number;
  license_plate?: string;
  is_default?: boolean;
}

export type UpdateVehicleRequest = Partial<CreateVehicleRequest> & {
  is_archived?: boolean;
};

// ---------- Fuel Record ----------

/** 后端 FuelRecordResponse 字段对齐 */
export interface FuelRecord {
  id: string;
  vehicle_id: string;
  fuel_amount: number;
  fuel_unit: string;
  unit_price?: number;
  total_cost: number;
  currency_code: string;
  odometer: number;
  distance_unit: string;
  is_full_tank: boolean;
  fuel_grade?: string;
  station_name?: string;
  station_lat?: number;
  station_lng?: number;
  note?: string;
  receipt_url?: string;
  trip_distance?: number;
  fuel_efficiency?: number;
  refuel_date: string;
  created_at: string;
  updated_at: string;
}

/** 后端 CreateFuelRecordRequest 字段对齐 */
export interface CreateFuelRecordRequest {
  fuel_amount: number;
  fuel_unit?: string;
  unit_price?: number;
  total_cost: number;
  currency_code: string;
  odometer: number;
  distance_unit?: string;
  is_full_tank: boolean;
  fuel_grade?: string;
  station_name?: string;
  station_lat?: number;
  station_lng?: number;
  note?: string;
  refuel_date: string;
}

export type UpdateFuelRecordRequest = Partial<CreateFuelRecordRequest>;

// ---------- Stats ----------

/** 后端 VehicleStatsResponse 字段对齐 */
export interface VehicleStats {
  vehicle_id: string;
  vehicle_name: string;
  total_records: number;
  total_fuel: number;
  total_cost: number;
  total_distance: number;
  avg_efficiency: number;
  best_efficiency: number;
  worst_efficiency: number;
  avg_cost_per_km: number;
  avg_cost_per_fill: number;
  currency_code: string;
  fuel_efficiency_unit: string;
}

/** 后端 OverviewStatsResponse 字段对齐 */
export interface OverviewStats {
  total_vehicles: number;
  total_records: number;
  total_fuel: number;
  total_cost: number;
  total_distance: number;
  avg_consumption: number;
  currency_code: string;
  vehicles: VehicleStats[];
}

/** 后端 FuelEfficiencyTrendItem 字段对齐 */
export interface ConsumptionTrend {
  date: string;
  fuel_efficiency: number;
  trip_distance: number;
}

/** 后端 FuelEfficiencyTrendResponse 字段对齐 */
export interface FuelEfficiencyTrendResponse {
  vehicle_id: string;
  vehicle_name: string;
  efficiency_unit: string;
  items: ConsumptionTrend[];
}

// ---------- Period Stats ----------

/** 后端 PeriodStatsItem 字段对齐 */
export interface PeriodStatsItem {
  period: string;         // "2026-01" 或 "2026"
  total_records: number;
  total_fuel: number;
  total_cost: number;
  total_distance: number;
  avg_efficiency: number;
}

/** 后端 PeriodStatsResponse 字段对齐 */
export interface PeriodStatsResponse {
  vehicle_id: string;
  vehicle_name: string;
  period: string;         // "month" | "year"
  year?: number;
  currency_code: string;
  fuel_efficiency_unit: string;
  items: PeriodStatsItem[];
  prev_items: PeriodStatsItem[];
}

// ---------- Reminder ----------

/** 保养提醒类别 */
export type MaintenanceCategory =
  | 'oil_change' | 'tire_rotation' | 'brake_pads' | 'air_filter'
  | 'transmission' | 'coolant' | 'spark_plugs' | 'battery'
  | 'tire_replace' | 'inspection' | 'custom';

/** 触发方式 */
export type ReminderTrigger = 'mileage' | 'time' | 'both';

/** 后端 ReminderResponse 字段对齐 */
export interface Reminder {
  id: string;
  vehicle_id: string;
  vehicle_name: string;
  type: string;
  category: MaintenanceCategory;
  title: string;
  description?: string;
  trigger: ReminderTrigger;
  mileage_interval?: number;
  time_interval_days?: number;
  last_mileage?: number;
  last_date?: string;
  next_mileage?: number;
  next_date?: string;
  is_enabled: boolean;
  is_overdue: boolean;
  created_at: string;
  updated_at: string;
}

/** 创建提醒请求 */
export interface CreateReminderRequest {
  vehicle_id: string;
  category: MaintenanceCategory;
  title: string;
  description?: string;
  trigger: ReminderTrigger;
  mileage_interval?: number;
  time_interval_days?: number;
  last_mileage?: number;
  last_date?: string;
}

/** 更新提醒请求 */
export interface UpdateReminderRequest {
  category?: MaintenanceCategory;
  title?: string;
  description?: string;
  trigger?: ReminderTrigger;
  mileage_interval?: number;
  time_interval_days?: number;
  last_mileage?: number;
  last_date?: string;
  is_enabled?: boolean;
}

// ---------- Notification ----------

/** 通知类型 */
export type NotificationType = 'anomaly_fuel' | 'maintenance_due' | 'invite_used';

/** 后端 NotificationResponse 字段对齐 */
export interface Notification {
  id: string;
  vehicle_id?: string;
  type: NotificationType;
  title: string;
  message: string;
  reminder_id?: string;
  record_id?: string;
  is_read: boolean;
  created_at: string;
}

// ---------- Group ----------

/** 群组成员角色 */
export type GroupRole = 'owner' | 'admin' | 'member';

/** 群组成员详情 */
export interface GroupMemberDetail {
  user_id: string;
  nickname: string;
  email: string;
  role: GroupRole;
  joined_at: string;
}

/** 后端 GroupResponse 字段对齐 */
export interface Group {
  id: string;
  name: string;
  owner_id: string;
  owner_name?: string;
  invite_code: string;
  max_members: number;
  description?: string;
  member_count: number;
  my_role?: GroupRole;
  members?: GroupMemberDetail[];
  created_at: string;
}

/** 创建群组请求 */
export interface CreateGroupRequest {
  name: string;
  max_members?: number;
  description?: string;
}

/** 更新群组请求 */
export interface UpdateGroupRequest {
  name?: string;
  max_members?: number;
  description?: string;
}

/** 更新成员角色请求 */
export interface UpdateMemberRoleRequest {
  role: 'admin' | 'member';
}

/** 加入群组请求 */
export interface JoinGroupRequest {
  invite_code: string;
}

/** 加入群组响应 */
export interface JoinGroupResponse {
  group_id: string;
  group_name: string;
  role: string;
}

/** 群组车辆汇总 */
export interface GroupVehicleSummary {
  vehicle_id: string;
  vehicle_name: string;
  owner_name: string;
  vehicle_type: string;
  fuel_type: string;
  total_records: number;
  total_cost: number;
  total_fuel: number;
  avg_efficiency: number;
}

/** 群组数据汇总响应 */
export interface GroupOverviewResponse {
  group_id: string;
  group_name: string;
  member_count: number;
  vehicle_count: number;
  vehicles: GroupVehicleSummary[];
}
