import { create } from 'zustand';
import type { User, AuthResponse, UpdateUserRequest } from '../types';
import { authApi, userApi } from '../api';

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
      set({ isAuthenticated: true, user: authData.user });
    } finally {
      set({ isLoading: false });
    }
  },

  register: async (email, password, nickname, inviteCode) => {
    set({ isLoading: true });
    try {
      const { data } = await authApi.register({ email, password, nickname, invite_code: inviteCode });
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

  updateProfile: async (req) => {
    const { data } = await userApi.updateProfile(req);
    set({ user: data.data });
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
