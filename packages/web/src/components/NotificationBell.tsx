import { useState, useEffect, useCallback } from 'react';
import { Badge, Popover, Drawer, List, Button, Typography, Space, Tag, Empty, message } from 'antd';
import {
  BellOutlined,
  WarningOutlined,
  ToolOutlined,
  CheckOutlined,
  UserAddOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { notificationApi } from '@gastrack/shared';
import type { Notification } from '@gastrack/shared';
import { useIsMobile } from '../hooks/useIsMobile';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import 'dayjs/locale/zh-cn';
import 'dayjs/locale/ja';

dayjs.extend(relativeTime);

const { Text } = Typography;

export default function NotificationBell() {
  const { t, i18n } = useTranslation();
  const isMobile = useIsMobile();
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);

  // 设置 dayjs locale
  useEffect(() => {
    const lang = i18n.language;
    if (lang.startsWith('zh')) dayjs.locale('zh-cn');
    else if (lang.startsWith('ja')) dayjs.locale('ja');
    else dayjs.locale('en');
  }, [i18n.language]);

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

  // 轮询未读数（每 60 秒）
  useEffect(() => {
    fetchUnreadCount();
    const interval = setInterval(fetchUnreadCount, 60000);
    return () => clearInterval(interval);
  }, [fetchUnreadCount]);

  // 打开 Popover 时加载列表
  useEffect(() => {
    if (open) {
      fetchNotifications();
    }
  }, [open, fetchNotifications]);

  const handleMarkAsRead = async (id: string) => {
    try {
      await notificationApi.markAsRead(id);
      setNotifications((prev) =>
        prev.map((n) => (n.id === id ? { ...n, is_read: true } : n))
      );
      setUnreadCount((c) => Math.max(0, c - 1));
    } catch {
      message.error(t('common.error'));
    }
  };

  const handleMarkAllAsRead = async () => {
    try {
      await notificationApi.markAllAsRead();
      setNotifications((prev) => prev.map((n) => ({ ...n, is_read: true })));
      setUnreadCount(0);
    } catch {
      message.error(t('common.error'));
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

  const getNotificationTag = (type: string) => {
    switch (type) {
      case 'anomaly_fuel':
        return <Tag color="warning">{t('notification.anomalyFuel')}</Tag>;
      case 'maintenance_due':
        return <Tag color="processing">{t('notification.maintenanceDue')}</Tag>;
      case 'invite_used':
        return <Tag color="success">{t('notification.inviteUsed')}</Tag>;
      default:
        return null;
    }
  };

  const content = (
    <div style={{ maxHeight: isMobile ? 'calc(100vh - 120px)' : 440, overflow: 'auto' }}>
      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        padding: '0 0 8px',
        borderBottom: '1px solid #f0f0f0',
      }}>
        <Text strong style={{ flexShrink: 0 }}>{t('notification.title')}</Text>
        {unreadCount > 0 && (
          <Button type="link" size="small" onClick={handleMarkAllAsRead} style={{ flexShrink: 0 }}>
            <CheckOutlined /> {t('notification.markAllRead')}
          </Button>
        )}
      </div>

      {notifications.length === 0 ? (
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description={t('notification.noNotifications')}
          style={{ padding: '24px 0' }}
        />
      ) : (
        <List
          loading={loading}
          dataSource={notifications}
          renderItem={(item) => (
            <List.Item
              key={item.id}
              style={{
                cursor: item.is_read ? 'default' : 'pointer',
                background: item.is_read ? 'transparent' : 'rgba(24, 144, 255, 0.04)',
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
                      {item.title}
                    </Text>
                    {getNotificationTag(item.type)}
                  </Space>
                }
                description={
                  <Space direction="vertical" size={0} style={{ width: '100%' }}>
                    <Text type="secondary" style={{ fontSize: 12, wordBreak: 'break-word' }}>{item.message}</Text>
                    <Text type="secondary" style={{ fontSize: 11 }}>
                      {dayjs(item.created_at).fromNow()}
                    </Text>
                  </Space>
                }
              />
              {!item.is_read && (
                <div style={{
                  width: 8, height: 8, borderRadius: '50%',
                  background: '#1890ff', flexShrink: 0,
                }} />
              )}
            </List.Item>
          )}
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
