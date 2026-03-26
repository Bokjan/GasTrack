import { create } from 'zustand';

export type ThemeMode = 'light' | 'dark' | 'system';

interface ThemeState {
  /** 用户选择的主题模式 */
  mode: ThemeMode;
  /** 实际应用的主题（resolved，system 模式下由系统决定） */
  resolved: 'light' | 'dark';

  setMode: (mode: ThemeMode) => void;
  /** 内部方法：根据系统偏好更新 resolved */
  _syncSystem: () => void;
}

/** 从 localStorage 读取持久化的主题模式 */
function loadMode(): ThemeMode {
  const stored = localStorage.getItem('theme_mode');
  if (stored === 'light' || stored === 'dark' || stored === 'system') return stored;
  return 'system'; // 默认跟随系统
}

/** 获取系统当前偏好 */
function getSystemTheme(): 'light' | 'dark' {
  if (typeof window === 'undefined') return 'light';
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

/** 根据 mode 计算 resolved */
function resolve(mode: ThemeMode): 'light' | 'dark' {
  if (mode === 'system') return getSystemTheme();
  return mode;
}

export const useThemeStore = create<ThemeState>((set, get) => {
  const initialMode = loadMode();

  return {
    mode: initialMode,
    resolved: resolve(initialMode),

    setMode: (mode) => {
      localStorage.setItem('theme_mode', mode);
      set({ mode, resolved: resolve(mode) });
    },

    _syncSystem: () => {
      const { mode } = get();
      if (mode === 'system') {
        set({ resolved: getSystemTheme() });
      }
    },
  };
});

// 监听系统主题变化
if (typeof window !== 'undefined') {
  window
    .matchMedia('(prefers-color-scheme: dark)')
    .addEventListener('change', () => {
      useThemeStore.getState()._syncSystem();
    });
}
