import { useState, useEffect } from 'react';
import { Button, theme } from 'antd';
import { DownloadOutlined, CloseOutlined, ShareAltOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';

interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>;
}

/**
 * PWA 安装引导（底部浮动卡片）：
 * - Android/Desktop: 监听 beforeinstallprompt 事件，显示安装按钮
 * - iOS Safari: 提示用户手动通过"分享 → 添加到主屏幕"安装
 */
export default function InstallPrompt() {
  const { t } = useTranslation();
  const { token } = theme.useToken();
  const [deferredPrompt, setDeferredPrompt] = useState<BeforeInstallPromptEvent | null>(null);
  const [showIOSPrompt, setShowIOSPrompt] = useState(false);
  const [dismissed, setDismissed] = useState(false);
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    // 检查是否已安装（standalone 模式下不再显示）
    const isStandalone = window.matchMedia('(display-mode: standalone)').matches
      || (navigator as any).standalone === true;
    if (isStandalone) return;

    // 检查是否之前已经关闭过提示
    const wasDismissed = sessionStorage.getItem('pwa-install-dismissed');
    if (wasDismissed) {
      setDismissed(true);
      return;
    }

    // Android / Desktop Chrome: 监听安装事件
    const handler = (e: Event) => {
      e.preventDefault();
      setDeferredPrompt(e as BeforeInstallPromptEvent);
    };
    window.addEventListener('beforeinstallprompt', handler);

    // iOS Safari 检测
    const isIOS = /iPhone|iPad|iPod/.test(navigator.userAgent);
    const isSafari = /Safari/.test(navigator.userAgent) && !/Chrome/.test(navigator.userAgent);
    if (isIOS && isSafari) {
      setShowIOSPrompt(true);
    }

    return () => window.removeEventListener('beforeinstallprompt', handler);
  }, []);

  // 延迟入场动画
  useEffect(() => {
    if ((deferredPrompt || showIOSPrompt) && !dismissed) {
      const timer = setTimeout(() => setVisible(true), 500);
      return () => clearTimeout(timer);
    }
  }, [deferredPrompt, showIOSPrompt, dismissed]);

  const handleInstall = async () => {
    if (!deferredPrompt) return;
    deferredPrompt.prompt();
    const { outcome } = await deferredPrompt.userChoice;
    if (outcome === 'accepted') {
      setDeferredPrompt(null);
    }
  };

  const handleDismiss = () => {
    setVisible(false);
    // 等动画结束再真正卸载
    setTimeout(() => {
      setDismissed(true);
      setDeferredPrompt(null);
      setShowIOSPrompt(false);
      sessionStorage.setItem('pwa-install-dismissed', 'true');
    }, 300);
  };

  if (dismissed) return null;
  if (!deferredPrompt && !showIOSPrompt) return null;

  const isIOS = showIOSPrompt && !deferredPrompt;

  return (
    <div
      style={{
        position: 'fixed',
        bottom: 0,
        left: 0,
        right: 0,
        zIndex: 1050,
        padding: '0 16px 16px',
        pointerEvents: 'none',
        transform: visible ? 'translateY(0)' : 'translateY(100%)',
        opacity: visible ? 1 : 0,
        transition: 'transform 0.3s ease, opacity 0.3s ease',
      }}
    >
      <div
        style={{
          maxWidth: 420,
          margin: '0 auto',
          pointerEvents: 'auto',
          background: token.colorBgElevated,
          borderRadius: token.borderRadiusLG,
          boxShadow: '0 -2px 16px rgba(0,0,0,0.15)',
          border: `1px solid ${token.colorBorderSecondary}`,
          padding: '16px 20px',
          display: 'flex',
          alignItems: 'center',
          gap: 14,
        }}
      >
        {/* App 图标 */}
        <img
          src="/favicon.svg"
          alt="GasTrack"
          style={{
            width: 44,
            height: 44,
            borderRadius: 10,
            flexShrink: 0,
          }}
        />

        {/* 文案 */}
        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{ fontWeight: 600, fontSize: 15, color: token.colorText, lineHeight: 1.3 }}>
            {isIOS
              ? t('pwa.iosInstallTitle', 'Add to Home Screen')
              : t('pwa.installTitle', 'Install GasTrack')
            }
          </div>
          <div style={{ fontSize: 12, color: token.colorTextSecondary, marginTop: 2, lineHeight: 1.4 }}>
            {isIOS ? (
              <>
                {t('pwa.iosInstallHint', 'Tap')} <ShareAltOutlined style={{ color: token.colorPrimary, fontSize: 13 }} /> {t('pwa.iosInstallHintSuffix', 'then "Add to Home Screen"')}
              </>
            ) : (
              t('pwa.installDescription', 'Add to home screen for quick access.')
            )}
          </div>
        </div>

        {/* 操作按钮 */}
        {!isIOS && (
          <Button
            type="primary"
            size="small"
            icon={<DownloadOutlined />}
            onClick={handleInstall}
            style={{ borderRadius: 6, flexShrink: 0 }}
          >
            {t('pwa.installButton', 'Install')}
          </Button>
        )}

        {/* 关闭 */}
        <CloseOutlined
          onClick={handleDismiss}
          style={{
            fontSize: 14,
            color: token.colorTextQuaternary,
            cursor: 'pointer',
            flexShrink: 0,
            padding: 4,
          }}
        />
      </div>
    </div>
  );
}
