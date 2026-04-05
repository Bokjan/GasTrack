import { useEffect } from 'react';
import { useRegisterSW } from 'virtual:pwa-register/react';
import { Button, notification } from 'antd';
import { ReloadOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';

/**
 * PWA 更新提示：当 Service Worker 检测到新版本时，
 * 弹出不可关闭的通知，强制用户刷新以使用最新版。
 */
export default function PWAUpdatePrompt() {
  const { t } = useTranslation();
  const [api, contextHolder] = notification.useNotification();

  const {
    needRefresh: [needRefresh],
    updateServiceWorker,
  } = useRegisterSW({
    onRegisteredSW(_swUrl, registration) {
      // 每 10 分钟检查一次更新
      if (registration) {
        setInterval(() => {
          registration.update();
        }, 10 * 60 * 1000);
      }
    },
    onRegisterError(error) {
      console.error('SW registration error:', error);
    },
  });

  useEffect(() => {
    if (needRefresh) {
      api.warning({
        key: 'pwa-update',
        message: t('pwa.updateAvailable', 'New version available'),
        description: t('pwa.updateDescription', 'A new version is available. Please refresh to continue.'),
        btn: (
          <Button
            type="primary"
            icon={<ReloadOutlined />}
            onClick={() => updateServiceWorker(true)}
          >
            {t('pwa.refresh', 'Refresh')}
          </Button>
        ),
        duration: 0,
        closable: false,
      });
    }
  }, [needRefresh, api, t, updateServiceWorker]);

  return <>{contextHolder}</>;
}
