import { useState, useEffect } from 'react';
import { Alert, Button, Space, Typography } from 'antd';
import { DownloadOutlined, AppleOutlined, ShareAltOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';

interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>;
}

/**
 * PWA 安装引导：
 * - Android/Desktop: 监听 beforeinstallprompt 事件，显示安装按钮
 * - iOS Safari: 提示用户手动通过"分享 → 添加到主屏幕"安装
 */
export default function InstallPrompt() {
  const { t } = useTranslation();
  const [deferredPrompt, setDeferredPrompt] = useState<BeforeInstallPromptEvent | null>(null);
  const [showIOSPrompt, setShowIOSPrompt] = useState(false);
  const [dismissed, setDismissed] = useState(false);

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

  const handleInstall = async () => {
    if (!deferredPrompt) return;
    deferredPrompt.prompt();
    const { outcome } = await deferredPrompt.userChoice;
    if (outcome === 'accepted') {
      setDeferredPrompt(null);
    }
  };

  const handleDismiss = () => {
    setDismissed(true);
    setDeferredPrompt(null);
    setShowIOSPrompt(false);
    sessionStorage.setItem('pwa-install-dismissed', 'true');
  };

  if (dismissed) return null;

  // Android / Desktop: 有安装事件
  if (deferredPrompt) {
    return (
      <Alert
        type="info"
        showIcon
        icon={<DownloadOutlined />}
        closable
        onClose={handleDismiss}
        message={t('pwa.installTitle', 'Install GasTrack')}
        description={
          <Space direction="vertical" size={8}>
            <Typography.Text type="secondary">
              {t('pwa.installDescription', 'Install GasTrack to your home screen for quick access and a better experience.')}
            </Typography.Text>
            <Button type="primary" size="small" icon={<DownloadOutlined />} onClick={handleInstall}>
              {t('pwa.installButton', 'Install')}
            </Button>
          </Space>
        }
        style={{ margin: '12px 16px 0' }}
      />
    );
  }

  // iOS Safari: 引导手动添加
  if (showIOSPrompt) {
    return (
      <Alert
        type="info"
        showIcon
        icon={<AppleOutlined />}
        closable
        onClose={handleDismiss}
        message={t('pwa.iosInstallTitle', 'Add to Home Screen')}
        description={
          <Typography.Text type="secondary">
            {t(
              'pwa.iosInstallDescription',
              'Tap the Share button {shareIcon} then select "Add to Home Screen" to install GasTrack.',
            ).replace('{shareIcon}', '')}
            {' '}
            <ShareAltOutlined style={{ color: '#1677ff' }} />
            {' → '}
            <strong>{t('pwa.addToHomeScreen', 'Add to Home Screen')}</strong>
          </Typography.Text>
        }
        style={{ margin: '12px 16px 0' }}
      />
    );
  }

  return null;
}
