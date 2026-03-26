import React, { useEffect } from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import { ConfigProvider, theme } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import { useThemeStore } from '@gastrack/shared';
import '@gastrack/shared/src/i18n';
import App from './App';
import './styles/global.css';

function ThemeRoot() {
  const resolved = useThemeStore((s) => s.resolved);

  // 同步 data-theme 到 <html>，让 CSS 变量生效
  useEffect(() => {
    document.documentElement.setAttribute('data-theme', resolved);
  }, [resolved]);

  const isDark = resolved === 'dark';

  return (
    <ConfigProvider
      locale={zhCN}
      theme={{
        algorithm: isDark ? theme.darkAlgorithm : theme.defaultAlgorithm,
        token: {
          colorPrimary: isDark ? '#4096ff' : '#1677ff',
          borderRadius: 8,
        },
        components: isDark
          ? {
              // 暗色模式组件级 token 微调
              Tag: {
                defaultBg: 'rgba(255,255,255,0.08)',
                defaultColor: 'rgba(255,255,255,0.75)',
              },
              Card: {
                actionsBg: 'rgba(255,255,255,0.04)',
              },
              Button: {
                defaultBg: 'transparent',
                defaultBorderColor: 'rgba(255,255,255,0.25)',
                defaultColor: 'rgba(255,255,255,0.85)',
                defaultHoverBg: 'rgba(255,255,255,0.08)',
                defaultHoverBorderColor: '#4096ff',
                defaultHoverColor: '#4096ff',
              },
              Menu: {
                darkItemSelectedBg: 'rgba(64,150,255,0.15)',
                darkItemSelectedColor: '#4096ff',
                darkItemHoverBg: 'rgba(255,255,255,0.06)',
              },
              Input: {
                activeBorderColor: '#4096ff',
                hoverBorderColor: '#4096ff',
              },
              Switch: {
                colorPrimary: '#4096ff',
                colorPrimaryHover: '#69b1ff',
              },
            }
          : {},
      }}
    >
      <BrowserRouter>
        <App />
      </BrowserRouter>
    </ConfigProvider>
  );
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <ThemeRoot />
  </React.StrictMode>,
);
