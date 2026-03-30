import { Grid } from 'antd';

const { useBreakpoint } = Grid;

/** 返回 true 当屏幕宽度 < md (768px) */
export function useIsMobile() {
  const screens = useBreakpoint();
  return !screens.md;
}
