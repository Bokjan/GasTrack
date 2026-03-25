import { create } from 'zustand';
import type { User, AuthTokens } from '../types';
import { authApi, userApi } from '../api';

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;

  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, nickname: string) => Promise<void>;
  logout: () => Promise<void>;
  fetchProfile: () => Promise<void>;
  setTokens: (tokens: AuthTokens) => void;
  reset: () => void;
}

const saveTokens = (tokens: AuthTokens) => {
  localStorage.setItem('access_token', tokens.access_token);
  localStorage.setItem('refresh_token', tokens.refresh_token);
};

const clearTokens = () => {
  localStorage.removeItem('access_token');
  localStorage.removeItem('refresh_token');
};

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isAuthenticated: !!localStorage.getItem('access_token'),
  isLoading: false,

  login: async (email, password) => {
    set({ isLoading: true });
    try {
      const { data } = await authApi.login({ email, password });
      saveTokens(data.data);
      set({ isAuthenticated: true });

      // 获取用户资料
      const profile = await userApi.getProfile();
      set({ user: profile.data.data });
    } finally {
      set({ isLoading: false });
    }
  },

  register: async (email, password, nickname) => {
    set({ isLoading: true });
    try {
      const { data } = await authApi.register({ email, password, nickname });
      saveTokens(data.data);
      set({ isAuthenticated: true });

      const profile = await userApi.getProfile();
      set({ user: profile.data.data });
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
      clearTokens();
      set({ user: null, isAuthenticated: false });
    }
  },

  fetchProfile: async () => {
    set({ isLoading: true });
    try {
      const { data } = await userApi.getProfile();
      set({ user: data.data, isAuthenticated: true });
    } catch {
      clearTokens();
      set({ user: null, isAuthenticated: false });
    } finally {
      set({ isLoading: false });
    }
  },

  setTokens: (tokens) => {
    saveTokens(tokens);
    set({ isAuthenticated: true });
  },

  reset: () => {
    clearTokens();
    set({ user: null, isAuthenticated: false, isLoading: false });
  },
}));
