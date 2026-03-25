// ============================================================
// GasTrack 共享类型定义
// ============================================================

// ---------- 通用 ----------

export interface ApiResponse<T = unknown> {
  code: number;
  message: string;
  data: T;
}

export interface PaginatedData<T> {
  list: T[];
  total: number;
  page: number;
  page_size: number;
}

export type PaginatedResponse<T> = ApiResponse<PaginatedData<T>>;

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

export interface AuthTokens {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export interface RefreshRequest {
  refresh_token: string;
}

// ---------- User ----------

export type MeasurementSystem = 'metric_eu' | 'metric_jp' | 'imperial';

export interface User {
  id: string;
  email: string;
  nickname: string;
  avatar_url: string;
  timezone: string;
  locale: string;
  currency: string;
  measurement_system: MeasurementSystem;
  created_at: string;
  updated_at: string;
}

export interface UpdateUserRequest {
  nickname?: string;
  avatar_url?: string;
  timezone?: string;
  locale?: string;
  currency?: string;
  measurement_system?: MeasurementSystem;
}

export interface ChangePasswordRequest {
  old_password: string;
  new_password: string;
}

// ---------- Vehicle ----------

export type VehicleType = 'car' | 'motorcycle' | 'other';
export type FuelType = 'gasoline' | 'diesel' | 'hybrid';

export interface Vehicle {
  id: string;
  user_id: string;
  name: string;
  vehicle_type: VehicleType;
  brand: string;
  model: string;
  year: number;
  fuel_type: FuelType;
  tank_capacity: number;
  engine_cc?: number;
  is_default: boolean;
  photo_url: string;
  created_at: string;
  updated_at: string;
}

export interface CreateVehicleRequest {
  name: string;
  vehicle_type: VehicleType;
  brand: string;
  model: string;
  year: number;
  fuel_type: FuelType;
  tank_capacity: number;
  engine_cc?: number;
  is_default?: boolean;
}

export type UpdateVehicleRequest = Partial<CreateVehicleRequest>;

// ---------- Fuel Record ----------

export interface FuelRecord {
  id: string;
  vehicle_id: string;
  fuel_date: string;
  station: string;
  fuel_amount: number;
  price_per_unit: number;
  total_cost: number;
  odometer: number;
  is_full_tank: boolean;
  fuel_consumption?: number;
  notes: string;
  created_at: string;
  updated_at: string;
}

export interface CreateFuelRecordRequest {
  fuel_date: string;
  station?: string;
  fuel_amount: number;
  price_per_unit: number;
  total_cost: number;
  odometer: number;
  is_full_tank: boolean;
  notes?: string;
}

export type UpdateFuelRecordRequest = Partial<CreateFuelRecordRequest>;

// ---------- Stats ----------

export interface VehicleStats {
  vehicle_id: string;
  total_records: number;
  total_fuel: number;
  total_cost: number;
  total_distance: number;
  avg_consumption: number;
  best_consumption: number;
  worst_consumption: number;
  avg_price_per_unit: number;
}

export interface OverviewStats {
  total_vehicles: number;
  total_records: number;
  total_fuel: number;
  total_cost: number;
  total_distance: number;
  avg_consumption: number;
}

export interface ConsumptionTrend {
  date: string;
  consumption: number;
  cost: number;
  distance: number;
}
