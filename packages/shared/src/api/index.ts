import apiClient from './client';
import type {
  ApiResponse,
  AuthResponse,
  LoginRequest,
  RegisterRequest,
  RefreshRequest,
  User,
  UpdateUserRequest,
  ChangePasswordRequest,
  Vehicle,
  CreateVehicleRequest,
  UpdateVehicleRequest,
  FuelRecord,
  CreateFuelRecordRequest,
  UpdateFuelRecordRequest,
  VehicleStats,
  OverviewStats,
  FuelEfficiencyTrendResponse,
  PaginatedResponse,
} from '../types';

// ==================== Auth ====================

export const authApi = {
  login: (data: LoginRequest) =>
    apiClient.post<ApiResponse<AuthResponse>>('/auth/login', data),

  register: (data: RegisterRequest) =>
    apiClient.post<ApiResponse<AuthResponse>>('/auth/register', data),

  refresh: (data: RefreshRequest) =>
    apiClient.post<ApiResponse<AuthResponse>>('/auth/refresh', data),

  logout: () => apiClient.post<ApiResponse<null>>('/auth/logout'),
};

// ==================== User ====================

export const userApi = {
  getProfile: () =>
    apiClient.get<ApiResponse<User>>('/users/me'),

  updateProfile: (data: UpdateUserRequest) =>
    apiClient.patch<ApiResponse<User>>('/users/me', data),

  changePassword: (data: ChangePasswordRequest) =>
    apiClient.put<ApiResponse<null>>('/users/me/password', data),

  deleteAccount: () =>
    apiClient.delete<ApiResponse<null>>('/users/me'),
};

// ==================== Vehicle ====================

export const vehicleApi = {
  list: () =>
    apiClient.get<ApiResponse<Vehicle[]>>('/vehicles'),

  create: (data: CreateVehicleRequest) =>
    apiClient.post<ApiResponse<Vehicle>>('/vehicles', data),

  getById: (id: string) =>
    apiClient.get<ApiResponse<Vehicle>>(`/vehicles/${id}`),

  update: (id: string, data: UpdateVehicleRequest) =>
    apiClient.patch<ApiResponse<Vehicle>>(`/vehicles/${id}`, data),

  delete: (id: string) =>
    apiClient.delete<ApiResponse<null>>(`/vehicles/${id}`),
};

// ==================== Fuel Record ====================

export const fuelRecordApi = {
  list: (vehicleId: string, params?: { page?: number; page_size?: number }) =>
    apiClient.get<PaginatedResponse<FuelRecord>>(
      `/vehicles/${vehicleId}/records`,
      { params },
    ),

  create: (vehicleId: string, data: CreateFuelRecordRequest) =>
    apiClient.post<ApiResponse<FuelRecord>>(
      `/vehicles/${vehicleId}/records`,
      data,
    ),

  getById: (vehicleId: string, recordId: string) =>
    apiClient.get<ApiResponse<FuelRecord>>(
      `/vehicles/${vehicleId}/records/${recordId}`,
    ),

  update: (vehicleId: string, recordId: string, data: UpdateFuelRecordRequest) =>
    apiClient.patch<ApiResponse<FuelRecord>>(
      `/vehicles/${vehicleId}/records/${recordId}`,
      data,
    ),

  delete: (vehicleId: string, recordId: string) =>
    apiClient.delete<ApiResponse<null>>(
      `/vehicles/${vehicleId}/records/${recordId}`,
    ),
};

// ==================== Stats ====================

export const statsApi = {
  vehicleStats: (vehicleId: string) =>
    apiClient.get<ApiResponse<VehicleStats>>(`/vehicles/${vehicleId}/stats`),

  overview: () =>
    apiClient.get<ApiResponse<OverviewStats>>('/stats/overview'),

  /** 后端路由: GET /vehicles/:id/efficiency-trend */
  efficiencyTrend: (vehicleId: string, params?: { limit?: number }) =>
    apiClient.get<ApiResponse<FuelEfficiencyTrendResponse>>(
      `/vehicles/${vehicleId}/efficiency-trend`,
      { params },
    ),
};
