import { create } from 'zustand';
import axios from 'axios';
import type { User, AuthResponse, UpdateUserRequest } from '../types';
import { authApi, userApi } from '../api';
import { setUnauthorizedHandler } from '../api/client';
import i18n from '../i18n';
import { useVehicleStore } from './vehicleStore';
import { useExchangeRateStore } from './exchangeRateStore';

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;

  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, nickname: string, inviteCode?: string) => Promise<void>;
  logout: () => Promise<void>;
  fetchProfile: () => Promise<void>;
  updateProfile: (data: UpdateUserRequest) => Promise<void>;
  setTokens: (tokens: AuthResponse) => void;
  reset: () => void;
}

const saveTokens = (tokens: AuthResponse) => {
  localStorage.setItem('access_token', tokens.access_token);
  localStorage.setItem('refresh_token', tokens.refresh_token);
};

const clearTokens = () => {
  localStorage.removeItem('access_token');
  localStorage.removeItem('refresh_token');
};

/** 清理上一个用户留在内存与 Service Worker 缓存中的数据（共享设备隐私） */
const clearSessionData = () => {
  clearTokens();
  useVehicleStore.getState().reset();
  useExchangeRateStore.getState().reset();
  if (typeof caches !== 'undefined') {
    caches.delete('api-cache').catch(() => {});
  }
};

/** 应用账号保存的语言偏好 */
const applyUserLocale = (user: User) => {
  if (user.locale && user.locale !== i18n.language && i18n.hasResourceBundle(user.locale, 'translation')) {
    i18n.changeLanguage(user.locale);
    localStorage.setItem('locale', user.locale);
  }
};

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isAuthenticated: !!localStorage.getItem('access_token'),
  isLoading: false,

  login: async (email, password) => {
    set({ isLoading: true });
    try {
      const { data } = await authApi.login({ email, password });
      const authData = data.data;
      saveTokens(authData);
      applyUserLocale(authData.user);
      set({ isAuthenticated: true, user: authData.user });
    } finally {
      set({ isLoading: false });
    }
  },

  register: async (email, password, nickname, inviteCode) => {
    set({ isLoading: true });
    try {
      const { data } = await authApi.register({
        email,
        password,
        nickname,
        locale: i18n.language,
        invite_code: inviteCode,
      });
      const authData = data.data;
      saveTokens(authData);
      set({ isAuthenticated: true, user: authData.user });
    } finally {
      set({ isLoading: false });
    }
  },

  logout: async () => {
    try {
      await authApi.logout();
    } catch {
      // 忽略登出 API 错误
    } finally {
      clearSessionData();
      set({ user: null, isAuthenticated: false });
    }
  },

  fetchProfile: async () => {
    set({ isLoading: true });
    try {
      const { data } = await userApi.getProfile();
      applyUserLocale(data.data);
      set({ user: data.data, isAuthenticated: true });
    } catch (err) {
      // 仅在明确未认证时登出；离线、超时、5xx 等保持登录状态（PWA 离线打开不应被登出）
      if (axios.isAxiosError(err) && err.response?.status === 401) {
        clearSessionData();
        set({ user: null, isAuthenticated: false });
      }
    } finally {
      set({ isLoading: false });
    }
  },

  updateProfile: async (req) => {
    const { data } = await userApi.updateProfile(req);
    set({ user: data.data });
  },

  setTokens: (tokens) => {
    saveTokens(tokens);
    set({ isAuthenticated: true });
  },

  reset: () => {
    clearSessionData();
    set({ user: null, isAuthenticated: false, isLoading: false });
  },
}));

// token 刷新失败等会话失效场景：重置状态，由路由守卫跳转到登录页（不整页刷新）
setUnauthorizedHandler(() => useAuthStore.getState().reset());
