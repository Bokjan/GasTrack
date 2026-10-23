import { useEffect, useRef } from 'react';
import { useRegisterSW } from 'virtual:pwa-register/react';
import { Button, notification } from 'antd';
import { ReloadOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';

/** 检查更新的间隔：10 分钟 */
const UPDATE_CHECK_INTERVAL = 10 * 60 * 1000;

/**
 * PWA 更新提示：当 Service Worker 检测到新版本时弹出通知，
 * 由用户选择刷新时机（可先关闭通知，保存完表单后再刷新）。
 */
export default function PWAUpdatePrompt() {
  const { t } = useTranslation();
  const [api, contextHolder] = notification.useNotification();
  const intervalRef = useRef<number>();

  const {
    needRefresh: [needRefresh],
    updateServiceWorker,
  } = useRegisterSW({
    onRegisteredSW(_swUrl, registration) {
      if (!registration) return;
      window.clearInterval(intervalRef.current);
      intervalRef.current = window.setInterval(() => {
        // 后台标签页不检查，回到前台时再说
        if (document.visibilityState === 'visible') {
          registration.update();
        }
      }, UPDATE_CHECK_INTERVAL);
    },
    onRegisterError(error) {
      console.error('SW registration error:', error);
    },
  });

  useEffect(() => () => window.clearInterval(intervalRef.current), []);

  useEffect(() => {
    if (needRefresh) {
      api.info({
        key: 'pwa-update',
        message: t('pwa.updateAvailable'),
        description: t('pwa.updateDescription'),
        btn: (
          <Button
            type="primary"
            icon={<ReloadOutlined />}
            onClick={() => updateServiceWorker(true)}
          >
            {t('pwa.refresh')}
          </Button>
        ),
        duration: 0,
      });
    }
  }, [needRefresh, api, t, updateServiceWorker]);

  return <>{contextHolder}</>;
}
