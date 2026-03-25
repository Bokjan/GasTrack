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
