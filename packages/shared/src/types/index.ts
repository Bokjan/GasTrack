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
  reference_currency?: string;
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
  reference_currency?: string;
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
  /** 共享车辆来源群组 ID（仅 include_shared=true 时有值） */
  shared_from_group_id?: string;
  /** 共享车辆来源群组名称（仅 include_shared=true 时有值） */
  shared_from_group_name?: string;
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
  /** 按原始入账币种分组的费用明细 */
  costs_by_currency?: Record<string, number>;
  /** 其他开销总额 */
  total_expense_cost?: number;
  /** 按原始入账币种分组的开销明细 */
  expense_costs_by_currency?: Record<string, number>;
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
  /** 按原始入账币种分组的总费用明细 */
  costs_by_currency?: Record<string, number>;
  vehicles: VehicleStats[];
  /** 全局其他开销总额 */
  total_expense_cost?: number;
  /** 全局按币种分组的开销明细 */
  expense_costs_by_currency?: Record<string, number>;
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

// ---------- Expense Period Stats ----------

/** 后端 ExpensePeriodStatsItem 字段对齐 */
export interface ExpensePeriodStatsItem {
  period: string;
  total_records: number;
  total_amount: number;
}

/** 后端 ExpensePeriodStatsResponse 字段对齐 */
export interface ExpensePeriodStatsResponse {
  vehicle_id: string;
  vehicle_name: string;
  period: string;
  year?: number;
  currency_code: string;
  costs_by_currency?: Record<string, number>;
  items: ExpensePeriodStatsItem[];
  prev_items: ExpensePeriodStatsItem[];
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
  owner_id: string;
  owner_name: string;
  vehicle_type: string;
  fuel_type: string;
  currency_code: string;
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

// ---------- Shared Vehicle ----------

/** 共享车辆请求 */
export interface ShareVehicleRequest {
  vehicle_id: string;
}

/** 共享车辆响应 */
export interface SharedVehicleResponse {
  id: string;
  group_id: string;
  vehicle_id: string;
  vehicle_name: string;
  owner_name: string;
  shared_at: string;
}

// ---------- Leaderboard ----------

/** 排行榜条目 */
export interface LeaderboardEntry {
  rank: number;
  user_id: string;
  nickname: string;
  vehicle_id: string;
  vehicle_name: string;
  value: number;
  diff_from_avg: number;
  record_count: number;
  is_self: boolean;
}

/** 排行榜响应 */
export interface LeaderboardResponse {
  group_id: string;
  group_name: string;
  metric: string;
  period: string;
  period_label: string;
  group_avg: number;
  unit: string;
  rankings: LeaderboardEntry[];
  total_participants: number;
}

/** 排行指标类型 */
export type LeaderboardMetric = 'efficiency' | 'cost' | 'distance' | 'frequency';

/** 排行时间范围 */
export type LeaderboardPeriod = 'current_month' | 'last_month' | 'last_3_months' | 'current_year';

// ---------- Group Expense Stats ----------

/** 群组费用统计摘要 */
export interface GroupExpenseSummary {
  total_cost: number;
  total_fuel: number;
  total_distance: number;
  avg_efficiency: number;
  cost_change_pct: number;
  fuel_change_pct: number;
  distance_change_pct: number;
  efficiency_change_pct: number;
}

/** 成员费用分解 */
export interface MemberCostBreakdown {
  user_id: string;
  nickname: string;
  total_cost: number;
  currency_code: string;
  total_fuel: number;
  percentage: number;
}

/** 成员费用项（趋势图） */
export interface MemberCostItem {
  user_id: string;
  nickname: string;
  cost: number;
  currency_code: string;
}

/** 群组趋势项 */
export interface GroupTrendItem {
  period_label: string;
  total_cost: number;
  total_fuel: number;
  total_distance: number;
  avg_efficiency: number;
  by_member?: MemberCostItem[];
}

/** 群组费用统计看板响应 */
export interface GroupExpenseStatsResponse {
  group_id: string;
  group_name: string;
  period: string;
  year?: number;
  summary: GroupExpenseSummary;
  member_breakdown: MemberCostBreakdown[];
  trend_items: GroupTrendItem[];
  prev_trend_items?: GroupTrendItem[];
}

// ---------- Group Station Stats ----------

/** 加油站常客 */
export interface StationVisitor {
  user_id: string;
  nickname: string;
  count: number;
}

/** 加油站信息 */
export interface StationInfo {
  station_name: string;
  avg_unit_price: number;
  latest_unit_price: number;
  price_trend: 'up' | 'down' | 'stable';
  currency_code: string;
  visit_count: number;
  visitors: StationVisitor[];
  latest_visit: string;
  fuel_grades_seen: string[];
}

/** 加油站推荐共享响应 */
export interface GroupStationStatsResponse {
  group_id: string;
  group_name: string;
  total_stations: number;
  stations: StationInfo[];
}

// ---------- Exchange Rate ----------

/** 汇率参考响应（只读展示） */
export interface ExchangeRateResponse {
  base: string;
  date: string;
  rates: Record<string, number>;
}

// ---------- Expense Record ----------

/** 开销类别 */
export type ExpenseCategory =
  | 'maintenance' | 'repair' | 'insurance' | 'parking'
  | 'toll' | 'car_wash' | 'inspection' | 'parts' | 'fine' | 'tax' | 'other';

/** 后端 ExpenseResponse 字段对齐 */
export interface ExpenseRecord {
  id: string;
  vehicle_id: string;
  user_id: string;
  category: ExpenseCategory;
  maintenance_category?: MaintenanceCategory;
  title: string;
  amount: number;
  currency_code: string;
  vendor_name?: string;
  odometer?: number;
  distance_unit?: string;
  note?: string;
  receipt_url?: string;
  expense_date: string;
  reminder_id?: string;
  created_at: string;
  updated_at: string;
}

/** 后端 CreateExpenseRequest 字段对齐 */
export interface CreateExpenseRequest {
  category: ExpenseCategory;
  maintenance_category?: MaintenanceCategory;
  title: string;
  amount: number;
  currency_code: string;
  vendor_name?: string;
  odometer?: number;
  distance_unit?: string;
  note?: string;
  expense_date: string;
  reminder_id?: string;
}

/** 更新开销请求 */
export type UpdateExpenseRequest = Partial<CreateExpenseRequest>;

/** 开销列表筛选参数 */
export interface ExpenseListFilter {
  page?: number;
  page_size?: number;
  category?: ExpenseCategory;
  start_date?: string;
  end_date?: string;
  keyword?: string;
  min_amount?: number;
  max_amount?: number;
}

/** 按币种汇总 */
export interface ExpenseCurrencyTotal {
  currency_code: string;
  total_amount: number;
  record_count: number;
}

/** 按分类汇总 */
export interface ExpenseCategoryBreakdown {
  category: string;
  total_amount: number;
  record_count: number;
  percentage: number;
}

/** 月度趋势 */
export interface ExpenseMonthlyTrend {
  period: string;
  total_amount: number;
  record_count: number;
}

/** 开销统计响应 */
export interface ExpenseStatsResponse {
  vehicle_id: string;
  total_records: number;
  totals_by_currency: ExpenseCurrencyTotal[];
  category_breakdown: ExpenseCategoryBreakdown[];
  monthly_trend: ExpenseMonthlyTrend[];
  last_30_days_amount: number;
  last_30_days_currency: string;
}
