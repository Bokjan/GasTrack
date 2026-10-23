import { Spin } from 'antd';

/** 页面代码块加载中的占位 */
export default function PageFallback() {
  return (
    <div style={{ display: 'flex', justifyContent: 'center', padding: 80 }}>
      <Spin size="large" />
    </div>
  );
}
