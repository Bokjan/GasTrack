import { create } from 'zustand';
import type { ExchangeRateResponse } from '../types';
import { exchangeRateApi } from '../api';

interface ExchangeRateState {
  /** 汇率数据（按 base 币种缓存） */
  data: ExchangeRateResponse | null;
  /** 是否正在加载 */
  isLoading: boolean;
  /** 加载错误 */
  error: string | null;
  /** 上次拉取时间戳 */
  lastFetched: number | null;

  /** 拉取汇率数据 */
  fetchRates: (base?: string) => Promise<void>;
  /** 重置 store */
  reset: () => void;
}

/** 缓存有效期：30 分钟（毫秒） */
const CACHE_TTL = 30 * 60 * 1000;

export const useExchangeRateStore = create<ExchangeRateState>((set, get) => ({
  data: null,
  isLoading: false,
  error: null,
  lastFetched: null,

  fetchRates: async (base?: string) => {
    const state = get();

    // 如果缓存有效且 base 匹配，跳过请求
    if (
      state.data &&
      state.lastFetched &&
      Date.now() - state.lastFetched < CACHE_TTL &&
      (!base || state.data.base === base)
    ) {
      return;
    }

    set({ isLoading: true, error: null });
    try {
      const { data } = await exchangeRateApi.getRates(base);
      set({
        data: data.data,
        lastFetched: Date.now(),
      });
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch exchange rates';
      set({ error: message });
    } finally {
      set({ isLoading: false });
    }
  },

  reset: () => {
    set({ data: null, isLoading: false, error: null, lastFetched: null });
  },
}));
