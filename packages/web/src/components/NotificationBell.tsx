import { useState, useEffect, useCallback } from 'react';
import { Badge, Popover, Drawer, List, Button, Typography, Space, Tag, Empty, Spin, message, theme } from 'antd';
import {
  BellOutlined,
  WarningOutlined,
  ToolOutlined,
  CheckOutlined,
  UserAddOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import {
  notificationApi,
  formatNumber,
  convertEfficiencyForDisplay,
  kmToMiles,
  isElectricVehicle,
  useAuthStore,
  useVehicleStore,
  getApiErrorMessage,
} from '@gastrack/shared';
import type { Notification } from '@gastrack/shared';
import { useIsMobile } from '../hooks/useIsMobile';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';

dayjs.extend(relativeTime);

const { Text } = Typography;

// 长标签（如日文）允许换行，避免撑出横向滚动
const tagStyle = { marginInlineEnd: 0, whiteSpace: 'normal', maxWidth: '100%' } as const;

/**
 * 解析后端结构化通知参数：新版为 JSON 对象，旧版为 "k1=v1;k2=v2"
 * 旧版中不含 '=' 的片段视为上一个值中的 ';'（如提醒标题本身含分号）
 */
function parseParams(raw: string): Record<string, string> {
  if (raw.startsWith('{')) {
    try {
      const obj = JSON.parse(raw) as Record<string, unknown>;
      return Object.fromEntries(Object.entries(obj).map(([k, v]) => [k, String(v)]));
    } catch {
      // 非法 JSON 时按旧格式解析
    }
  }
  const params: Record<string, string> = {};
  let lastKey: string | null = null;
  for (const seg of raw.split(';')) {
    const idx = seg.indexOf('=');
    if (idx > 0) {
      lastKey = seg.slice(0, idx);
      params[lastKey] = seg.slice(idx + 1);
    } else if (lastKey) {
      params[lastKey] += `;${seg}`;
    }
  }
  return params;
}

export default function NotificationBell() {
  const { t, i18n } = useTranslation();
  const isMobile = useIsMobile();
  const user = useAuthStore((s) => s.user);
  const vehicles = useVehicleStore((s) => s.vehicles);
  const fetchVehicles = useVehicleStore((s) => s.fetchVehicles);
  const efficiencyUnit = user?.fuel_efficiency_unit || 'L/100km';
  const isImperial = user?.unit_system === 'imperial';
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const { token } = theme.useToken();

  const fetchUnreadCount = useCallback(async () => {
    try {
      const { data } = await notificationApi.unreadCount();
      setUnreadCount(data.data.count);
    } catch {
      // silently ignore
    }
  }, []);

  const fetchNotifications = useCallback(async () => {
    setLoading(true);
    try {
      const { data } = await notificationApi.list();
      setNotifications(data.data || []);
    } catch {
      // silently ignore
    } finally {
      setLoading(false);
    }
  }, []);

  // 轮询未读数（每 60 秒）；后台标签页暂停，回到前台立即刷新
  useEffect(() => {
    fetchUnreadCount();
    const interval = setInterval(() => {
      if (document.visibilityState === 'visible') fetchUnreadCount();
    }, 60000);
    const onVisible = () => {
      if (document.visibilityState === 'visible') fetchUnreadCount();
    };
    document.addEventListener('visibilitychange', onVisible);
    return () => {
      clearInterval(interval);
      document.removeEventListener('visibilitychange', onVisible);
    };
  }, [fetchUnreadCount]);

  // 打开 Popover 时加载列表
  useEffect(() => {
    if (open) {
      fetchNotifications();
      // 油耗异常通知需按车辆类型（燃油/电动）选择单位
      if (vehicles.length === 0) fetchVehicles().catch(() => {});
    }
  }, [open, fetchNotifications, fetchVehicles, vehicles.length]);

  const handleMarkAsRead = async (id: string) => {
    try {
      await notificationApi.markAsRead(id);
      setNotifications((prev) =>
        prev.map((n) => (n.id === id ? { ...n, is_read: true } : n))
      );
      setUnreadCount((c) => Math.max(0, c - 1));
    } catch (err) {
      message.error(getApiErrorMessage(err));
    }
  };

  const handleMarkAllAsRead = async () => {
    try {
      await notificationApi.markAllAsRead();
      setNotifications((prev) => prev.map((n) => ({ ...n, is_read: true })));
      setUnreadCount(0);
    } catch (err) {
      message.error(getApiErrorMessage(err));
    }
  };

  const getNotificationIcon = (type: string) => {
    switch (type) {
      case 'anomaly_fuel':
        return <WarningOutlined style={{ color: '#faad14' }} />;
      case 'maintenance_due':
        return <ToolOutlined style={{ color: '#1890ff' }} />;
      case 'invite_used':
        return <UserAddOutlined style={{ color: '#52c41a' }} />;
      default:
        return <BellOutlined />;
    }
  };

  /**
   * 将结构化通知渲染为本地化标题与正文；旧版纯文本通知原样显示
   * showTag：标题与类型标签相同时不重复显示标签
   */
  const renderNotification = (item: Notification): { title: string; message: string; showTag: boolean } => {
    if (!item.title.startsWith('notification.') || !i18n.exists(item.title)) {
      return { title: item.title, message: item.message, showTag: true };
    }
    const params = parseParams(item.message);
    switch (item.type) {
      case 'anomaly_fuel': {
        // 后端参数为基准单位，换算后重新计算方向与偏离比例（km/L、MPG 下高低关系相反）
        const vehicle = vehicles.find((v) => v.id === item.vehicle_id);
        const isEv = vehicle ? isElectricVehicle(vehicle.fuel_type) : false;
        const current = convertEfficiencyForDisplay(Number(params.current), efficiencyUnit, isEv);
        const avg = convertEfficiencyForDisplay(Number(params.avg), efficiencyUnit, isEv);
        const pct = avg.value > 0 ? Math.round((Math.abs(current.value - avg.value) / avg.value) * 100) : Number(params.pct);
        return {
          title: t(item.title),
          message: t('notification.anomalyFuelMsg', {
            current: formatNumber(current.value, 1),
            avg: formatNumber(avg.value, 1),
            unit: current.unit,
            pct,
            direction: t(current.value < avg.value ? 'notification.directionLower' : 'notification.directionHigher'),
          }),
          showTag: false,
        };
      }
      case 'maintenance_due':
        // 按时间到期的提醒只有目标日期
        if (params.targetDate) {
          return {
            title: params.reminderTitle || t(item.title),
            message: t('notification.maintenanceDueDateMsg', { targetDate: params.targetDate }),
            showTag: !!params.reminderTitle,
          };
        }
        return {
          title: params.reminderTitle || t(item.title),
          // 后端里程统一为 km
          message: t('notification.maintenanceDueMsg', {
            currentOdo: formatNumber(isImperial ? kmToMiles(Number(params.currentOdo)) : Number(params.currentOdo), 0),
            targetOdo: formatNumber(isImperial ? kmToMiles(Number(params.targetOdo)) : Number(params.targetOdo), 0),
            unit: isImperial ? 'mi' : 'km',
          }),
          showTag: !!params.reminderTitle,
        };
      case 'invite_used':
        return { title: t(item.title), message: t('notification.inviteUsedMsg', params), showTag: false };
      default:
        return { title: t(item.title), message: item.message, showTag: false };
    }
  };

  const getNotificationTag = (type: string) => {
    switch (type) {
      case 'anomaly_fuel':
        return <Tag color="warning" style={tagStyle}>{t('notification.anomalyFuel')}</Tag>;
      case 'maintenance_due':
        return <Tag color="processing" style={tagStyle}>{t('notification.maintenanceDue')}</Tag>;
      case 'invite_used':
        return <Tag color="success" style={tagStyle}>{t('notification.inviteUsed')}</Tag>;
      default:
        return null;
    }
  };

  const content = (
    // 桌面端固定宽度：Popover 默认按内容收缩，会被标题栏宽度挤窄
    <div
      style={{
        width: isMobile ? '100%' : 360,
        maxWidth: 'calc(100vw - 32px)',
        maxHeight: isMobile ? 'calc(100vh - 120px)' : 440,
        overflowX: 'hidden',
        overflowY: 'auto',
      }}
    >
      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        padding: '0 0 8px',
        borderBottom: `1px solid ${token.colorBorderSecondary}`,
      }}>
        <Text strong style={{ flexShrink: 0 }}>{t('notification.title')}</Text>
        {unreadCount > 0 && (
          <Button type="link" size="small" onClick={handleMarkAllAsRead} style={{ flexShrink: 0 }}>
            <CheckOutlined /> {t('notification.markAllRead')}
          </Button>
        )}
      </div>

      {loading && notifications.length === 0 ? (
        <div style={{ textAlign: 'center', padding: '24px 0' }}><Spin /></div>
      ) : notifications.length === 0 ? (
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description={t('notification.noNotifications')}
          style={{ padding: '24px 0' }}
        />
      ) : (
        <List
          loading={loading}
          dataSource={notifications}
          renderItem={(item) => {
            const { title, message: body, showTag } = renderNotification(item);
            return (
              <List.Item
                key={item.id}
                style={{
                  cursor: item.is_read ? 'default' : 'pointer',
                  background: item.is_read ? 'transparent' : token.colorPrimaryBg,
                  padding: '8px 4px',
                }}
                onClick={() => {
                  if (!item.is_read) handleMarkAsRead(item.id);
                }}
              >
                <List.Item.Meta
                  avatar={getNotificationIcon(item.type)}
                  title={
                    <Space size={4} style={{ flexWrap: 'wrap' }}>
                      <Text
                        style={{
                          fontWeight: item.is_read ? 'normal' : 'bold',
                          fontSize: 13,
                          wordBreak: 'break-word',
                        }}
                      >
                        {title}
                      </Text>
                      {showTag && getNotificationTag(item.type)}
                    </Space>
                  }
                  description={
                    <Space direction="vertical" size={0} style={{ width: '100%' }}>
                      <Text type="secondary" style={{ fontSize: 12, wordBreak: 'break-word' }}>{body}</Text>
                      <Text type="secondary" style={{ fontSize: 11 }}>
                        {dayjs(item.created_at).fromNow()}
                      </Text>
                    </Space>
                  }
                />
                {!item.is_read && (
                  <div style={{
                    width: 8, height: 8, borderRadius: '50%',
                    background: token.colorPrimary, flexShrink: 0,
                  }} />
                )}
              </List.Item>
            );
          }}
        />
      )}
    </div>
  );

  const bellIcon = (
    <Badge count={unreadCount} size="small" offset={[-2, 4]}>
      <BellOutlined
        style={{ fontSize: 18, cursor: 'pointer' }}
        onClick={() => setOpen(true)}
      />
    </Badge>
  );

  // 手机端：使用 Drawer 从顶部滑出，避免 Popover 溢出
  if (isMobile) {
    return (
      <>
        {bellIcon}
        <Drawer
          title={null}
          placement="top"
          height="auto"
          open={open}
          onClose={() => setOpen(false)}
          closable
          styles={{ body: { padding: '12px 16px' } }}
        >
          {content}
        </Drawer>
      </>
    );
  }

  // 桌面端：保持 Popover
  return (
    <Popover
      content={content}
      trigger="click"
      open={open}
      onOpenChange={setOpen}
      placement="bottomRight"
      arrow={false}
    >
      <Badge count={unreadCount} size="small" offset={[-2, 4]}>
        <BellOutlined style={{ fontSize: 18, cursor: 'pointer' }} />
      </Badge>
    </Popover>
  );
}
