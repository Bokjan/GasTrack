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
  PeriodStatsResponse,
  PaginatedResponse,
  RegistrationModeResponse,
  ValidateInviteResponse,
  InviteCode,
  CreateInviteRequest,
  UpdateInviteRequest,
  Reminder,
  CreateReminderRequest,
  UpdateReminderRequest,
  Notification,
} from '../types';

// 注意: PaginatedResponse<T> 的 Axios 响应为 AxiosResponse<PaginatedResponse<T>>
// 即 response.data = { code, message, data: T[], meta: { page, page_size, total, total_pages } }

// ==================== Auth ====================

export const authApi = {
  login: (data: LoginRequest) =>
    apiClient.post<ApiResponse<AuthResponse>>('/auth/login', data),

  register: (data: RegisterRequest) =>
    apiClient.post<ApiResponse<AuthResponse>>('/auth/register', data),

  refresh: (data: RefreshRequest) =>
    apiClient.post<ApiResponse<AuthResponse>>('/auth/refresh', data),

  logout: () => apiClient.post<ApiResponse<null>>('/auth/logout'),

  /** 查询当前注册模式（open / invite_only / closed） */
  getRegistrationMode: () =>
    apiClient.get<ApiResponse<RegistrationModeResponse>>('/auth/registration-mode'),
};

// ==================== Invite ====================

export const inviteApi = {
  /** 验证邀请码是否有效（公开接口） */
  validate: (code: string) =>
    apiClient.get<ApiResponse<ValidateInviteResponse>>(`/invites/${code}`),

  /** 创建邀请码 */
  create: (data: CreateInviteRequest) =>
    apiClient.post<ApiResponse<InviteCode>>('/invites', data),

  /** 查询我创建的邀请码列表 */
  list: () =>
    apiClient.get<ApiResponse<InviteCode[]>>('/invites'),

  /** 更新邀请码 */
  update: (id: string, data: UpdateInviteRequest) =>
    apiClient.patch<ApiResponse<InviteCode>>(`/invites/${id}`, data),

  /** 删除邀请码 */
  delete: (id: string) =>
    apiClient.delete<ApiResponse<null>>(`/invites/${id}`),
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

  /** 导出用户全部数据（CSV 文件下载） */
  exportData: () =>
    apiClient.get<Blob>('/users/me/export', {
      responseType: 'blob',
    }),
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

  /** 获取加油站/充电站名称建议列表 */
  getStationSuggestions: (vehicleId: string) =>
    apiClient.get<ApiResponse<string[]>>(
      `/vehicles/${vehicleId}/stations`,
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

  /** 后端路由: GET /vehicles/:id/period-stats?period=month&year=2026 */
  periodStats: (vehicleId: string, params: { period: 'month' | 'year'; year?: number }) =>
    apiClient.get<ApiResponse<PeriodStatsResponse>>(
      `/vehicles/${vehicleId}/period-stats`,
      { params },
    ),
};

// ==================== Reminder ====================

export const reminderApi = {
  /** 获取提醒列表 */
  list: () =>
    apiClient.get<ApiResponse<Reminder[]>>('/reminders'),

  /** 创建提醒 */
  create: (data: CreateReminderRequest) =>
    apiClient.post<ApiResponse<Reminder>>('/reminders', data),

  /** 获取提醒详情 */
  getById: (id: string) =>
    apiClient.get<ApiResponse<Reminder>>(`/reminders/${id}`),

  /** 更新提醒 */
  update: (id: string, data: UpdateReminderRequest) =>
    apiClient.patch<ApiResponse<Reminder>>(`/reminders/${id}`, data),

  /** 删除提醒 */
  delete: (id: string) =>
    apiClient.delete<ApiResponse<null>>(`/reminders/${id}`),
};

// ==================== Notification ====================

export const notificationApi = {
  /** 获取通知列表 */
  list: () =>
    apiClient.get<ApiResponse<Notification[]>>('/notifications'),

  /** 获取未读通知数 */
  unreadCount: () =>
    apiClient.get<ApiResponse<{ count: number }>>('/notifications/unread-count'),

  /** 标记通知为已读 */
  markAsRead: (id: string) =>
    apiClient.patch<ApiResponse<null>>(`/notifications/${id}/read`),

  /** 标记所有通知为已读 */
  markAllAsRead: () =>
    apiClient.post<ApiResponse<null>>('/notifications/read-all'),

  /** 删除通知 */
  delete: (id: string) =>
    apiClient.delete<ApiResponse<null>>(`/notifications/${id}`),
};
