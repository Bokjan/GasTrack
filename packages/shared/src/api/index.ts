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
  Group,
  CreateGroupRequest,
  UpdateGroupRequest,
  UpdateMemberRoleRequest,
  JoinGroupResponse,
  GroupOverviewResponse,
  SharedVehicleResponse,
  ShareVehicleRequest,
  LeaderboardResponse,
  GroupExpenseStatsResponse,
  GroupStationStatsResponse,
  ExchangeRateResponse,
  ExpenseRecord,
  CreateExpenseRequest,
  UpdateExpenseRequest,
  ExpenseListFilter,
  ExpenseStatsResponse,
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

  /** 导出用户数据（支持 format=csv|zip|json，scope=basic|full） */
  exportData: (params?: { format?: 'csv' | 'zip' | 'json'; scope?: 'basic' | 'full' }) =>
    apiClient.get<Blob>('/users/me/export', {
      responseType: 'blob',
      params,
    }),
};

// ==================== Vehicle ====================

export const vehicleApi = {
  list: (params?: { include_shared?: boolean }) =>
    apiClient.get<ApiResponse<Vehicle[]>>('/vehicles', { params }),

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

// ==================== Group ====================

export const groupApi = {
  /** 获取我所在的群组列表 */
  list: () =>
    apiClient.get<ApiResponse<Group[]>>('/groups'),

  /** 创建群组 */
  create: (data: CreateGroupRequest) =>
    apiClient.post<ApiResponse<Group>>('/groups', data),

  /** 获取群组详情 */
  getById: (id: string) =>
    apiClient.get<ApiResponse<Group>>(`/groups/${id}`),

  /** 更新群组信息 */
  update: (id: string, data: UpdateGroupRequest) =>
    apiClient.patch<ApiResponse<Group>>(`/groups/${id}`, data),

  /** 删除群组 */
  delete: (id: string) =>
    apiClient.delete<ApiResponse<null>>(`/groups/${id}`),

  /** 通过邀请码加入群组 */
  join: (inviteCode: string) =>
    apiClient.post<ApiResponse<JoinGroupResponse>>('/groups/join', { invite_code: inviteCode }),

  /** 重新生成邀请码 */
  regenerateInviteCode: (id: string) =>
    apiClient.post<ApiResponse<Group>>(`/groups/${id}/regenerate-invite`),

  /** 退出群组 */
  leave: (id: string) =>
    apiClient.post<ApiResponse<null>>(`/groups/${id}/leave`),

  /** 获取群组数据汇总 */
  getOverview: (id: string) =>
    apiClient.get<ApiResponse<GroupOverviewResponse>>(`/groups/${id}/overview`),

  /** 更新成员角色 */
  updateMemberRole: (groupId: string, userId: string, data: UpdateMemberRoleRequest) =>
    apiClient.patch<ApiResponse<null>>(`/groups/${groupId}/members/${userId}`, data),

  /** 移除成员 */
  removeMember: (groupId: string, userId: string) =>
    apiClient.delete<ApiResponse<null>>(`/groups/${groupId}/members/${userId}`),

  // --- 车辆共享 ---

  /** 共享车辆到群组 */
  shareVehicle: (groupId: string, data: ShareVehicleRequest) =>
    apiClient.post<ApiResponse<SharedVehicleResponse>>(`/groups/${groupId}/shared-vehicles`, data),

  /** 取消车辆共享 */
  unshareVehicle: (groupId: string, vehicleId: string) =>
    apiClient.delete<ApiResponse<null>>(`/groups/${groupId}/shared-vehicles/${vehicleId}`),

  /** 获取群组内共享车辆列表 */
  listSharedVehicles: (groupId: string) =>
    apiClient.get<ApiResponse<SharedVehicleResponse[]>>(`/groups/${groupId}/shared-vehicles`),

  // --- 排行榜 ---

  /** 获取群组排行榜 */
  getLeaderboard: (groupId: string, params?: { metric?: string; period?: string }) =>
    apiClient.get<ApiResponse<LeaderboardResponse>>(`/groups/${groupId}/leaderboard`, { params }),

  // --- 费用统计看板 ---

  /** 获取群组费用统计 */
  getExpenseStats: (groupId: string, params?: { period?: string; year?: number }) =>
    apiClient.get<ApiResponse<GroupExpenseStatsResponse>>(`/groups/${groupId}/expense-stats`, { params }),

  // --- 加油站推荐共享 ---

  /** 获取群组加油站推荐 */
  getStationStats: (groupId: string, params?: { fuel_grade?: string; months?: number; sort_by?: string }) =>
    apiClient.get<ApiResponse<GroupStationStatsResponse>>(`/groups/${groupId}/stations`, { params }),
};

// ==================== Exchange Rate ====================

export const exchangeRateApi = {
  /** 获取汇率参考数据 */
  getRates: (base?: string) =>
    apiClient.get<ApiResponse<ExchangeRateResponse>>('/exchange-rates', { params: base ? { base } : undefined }),
};

// ==================== Expense Record ====================

export const expenseApi = {
  /** 获取开销记录列表 */
  list: (vehicleId: string, params?: ExpenseListFilter) =>
    apiClient.get<PaginatedResponse<ExpenseRecord>>(
      `/vehicles/${vehicleId}/expenses`,
      { params },
    ),

  /** 创建开销记录 */
  create: (vehicleId: string, data: CreateExpenseRequest) =>
    apiClient.post<ApiResponse<ExpenseRecord>>(
      `/vehicles/${vehicleId}/expenses`,
      data,
    ),

  /** 获取开销记录详情 */
  getById: (vehicleId: string, expenseId: string) =>
    apiClient.get<ApiResponse<ExpenseRecord>>(
      `/vehicles/${vehicleId}/expenses/${expenseId}`,
    ),

  /** 更新开销记录 */
  update: (vehicleId: string, expenseId: string, data: UpdateExpenseRequest) =>
    apiClient.patch<ApiResponse<ExpenseRecord>>(
      `/vehicles/${vehicleId}/expenses/${expenseId}`,
      data,
    ),

  /** 删除开销记录 */
  delete: (vehicleId: string, expenseId: string) =>
    apiClient.delete<ApiResponse<null>>(
      `/vehicles/${vehicleId}/expenses/${expenseId}`,
    ),

  /** 获取开销统计 */
  getStats: (vehicleId: string) =>
    apiClient.get<ApiResponse<ExpenseStatsResponse>>(
      `/vehicles/${vehicleId}/expense-stats`,
    ),

  /** 获取商家名称建议列表 */
  getVendorSuggestions: (vehicleId: string) =>
    apiClient.get<ApiResponse<string[]>>(
      `/vehicles/${vehicleId}/expense-vendors`,
    ),
};
