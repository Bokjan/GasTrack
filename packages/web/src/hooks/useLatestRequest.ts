import { useCallback, useRef } from 'react';

/**
 * 只采用最近一次请求的结果：begin() 返回本次请求的检查函数，
 * 之后若又发起了新请求，检查函数返回 false，调用方应丢弃该响应。
 */
export function useLatestRequest() {
  const counter = useRef(0);
  return useCallback(() => {
    const id = ++counter.current;
    return () => id === counter.current;
  }, []);
}
