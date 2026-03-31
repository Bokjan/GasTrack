import { useEffect } from 'react';
import { useRegisterSW } from 'virtual:pwa-register/react';
import { Button, notification } from 'antd';
import { ReloadOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';

/**
 * PWA 更新提示：当 Service Worker 检测到新版本时，
 * 弹出通知引导用户刷新页面以使用最新版。
 */
export default function PWAUpdatePrompt() {
  const { t } = useTranslation();
  const [api, contextHolder] = notification.useNotification();

  const {
    needRefresh: [needRefresh, setNeedRefresh],
    updateServiceWorker,
  } = useRegisterSW({
    onRegisteredSW(_swUrl, registration) {
      // 每小时检查一次更新
      if (registration) {
        setInterval(() => {
          registration.update();
        }, 60 * 60 * 1000);
      }
    },
    onRegisterError(error) {
      console.error('SW registration error:', error);
    },
  });

  useEffect(() => {
    if (needRefresh) {
      api.info({
        key: 'pwa-update',
        message: t('pwa.updateAvailable', 'New version available'),
        description: t('pwa.updateDescription', 'A new version is available. Refresh to update.'),
        btn: (
          <Button
            type="primary"
            size="small"
            icon={<ReloadOutlined />}
            onClick={() => updateServiceWorker(true)}
          >
            {t('pwa.refresh', 'Refresh')}
          </Button>
        ),
        duration: 0,
        onClose: () => setNeedRefresh(false),
      });
    }
  }, [needRefresh, api, t, updateServiceWorker, setNeedRefresh]);

  return <>{contextHolder}</>;
}
