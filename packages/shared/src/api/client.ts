import axios from 'axios';
import type { ApiResponse, AuthResponse } from '../types';

const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8098/api/v1',
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' },
});

// ---------- 请求拦截器：自动添加 Token ----------
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// ---------- 会话失效处理 ----------
/** 登录/注册/刷新接口自身的 401 是业务错误（如密码错误），不应触发刷新或登出 */
const AUTH_ENDPOINTS = ['/auth/login', '/auth/register', '/auth/refresh'];

let unauthorizedHandler: (() => void) | null = null;

/** 注册会话失效回调（由 authStore 设置，用于重置登录状态并交由路由跳转登录页） */
export function setUnauthorizedHandler(handler: () => void) {
  unauthorizedHandler = handler;
}

const handleSessionExpired = () => {
  localStorage.removeItem('access_token');
  localStorage.removeItem('refresh_token');
  if (unauthorizedHandler) {
    unauthorizedHandler();
  } else {
    window.location.href = '/login';
  }
};

// ---------- 响应拦截器：统一错误处理 + Token 刷新 ----------
let isRefreshing = false;
let failedQueue: Array<{
  resolve: (token: string) => void;
  reject: (err: unknown) => void;
}> = [];

const processQueue = (error: unknown, token: string | null) => {
  failedQueue.forEach((p) => {
    if (error) {
      p.reject(error);
    } else {
      p.resolve(token!);
    }
  });
  failedQueue = [];
};

apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    const isAuthEndpoint = AUTH_ENDPOINTS.some((p) => originalRequest?.url?.startsWith(p));

    if (error.response?.status === 401 && originalRequest && !originalRequest._retry && !isAuthEndpoint) {
      const refreshToken = localStorage.getItem('refresh_token');
      if (!refreshToken) {
        handleSessionExpired();
        return Promise.reject(error);
      }

      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject });
        }).then((token) => {
          originalRequest.headers.Authorization = `Bearer ${token}`;
          return apiClient(originalRequest);
        });
      }

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        const { data } = await axios.post<ApiResponse<AuthResponse>>(
          `${apiClient.defaults.baseURL}/auth/refresh`,
          { refresh_token: refreshToken },
        );
        const tokens = data.data;
        localStorage.setItem('access_token', tokens.access_token);
        localStorage.setItem('refresh_token', tokens.refresh_token);

        processQueue(null, tokens.access_token);
        originalRequest.headers.Authorization = `Bearer ${tokens.access_token}`;
        return apiClient(originalRequest);
      } catch (refreshError) {
        // refresh token 只能使用一次：若其他标签页已抢先完成轮换，直接改用其写入的新 token
        const latestRefresh = localStorage.getItem('refresh_token');
        const latestAccess = localStorage.getItem('access_token');
        if (latestRefresh && latestRefresh !== refreshToken && latestAccess) {
          processQueue(null, latestAccess);
          originalRequest.headers.Authorization = `Bearer ${latestAccess}`;
          return apiClient(originalRequest);
        }
        processQueue(refreshError, null);
        // 仅在服务端明确拒绝 refresh token 时登出；网络错误或 5xx 保留登录状态，待恢复后重试
        const status = axios.isAxiosError(refreshError) ? refreshError.response?.status : undefined;
        if (status !== undefined && status < 500) {
          handleSessionExpired();
        }
        return Promise.reject(refreshError);
      } finally {
        isRefreshing = false;
      }
    }

    return Promise.reject(error);
  },
);

export default apiClient;
