import { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Typography, Button, Card, Space, Tag, Modal, Form, Input, Select,
  InputNumber, DatePicker, Switch, Empty, Spin, message, Popconfirm,
  Row, Col, Grid, Tooltip, Divider,
} from 'antd';
import {
  PlusOutlined, DeleteOutlined, EditOutlined, WalletOutlined,
  CarOutlined, DashboardOutlined, CalendarOutlined, ToolOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import dayjs from 'dayjs';
import {
  reminderApi,
  vehicleApi,
  useAuthStore,
} from '@gastrack/shared';
import type { Reminder, Vehicle, CreateReminderRequest } from '@gastrack/shared';

const { Title, Text } = Typography;
const { useBreakpoint } = Grid;

/** 保养类别图标/颜色映射 */
const CATEGORY_CONFIG: Record<string, { color: string; icon: string }> = {
  oil_change:    { color: 'orange',  icon: '🛢️' },
  tire_rotation: { color: 'blue',    icon: '🔄' },
  brake_pads:    { color: 'red',     icon: '🛑' },
  air_filter:    { color: 'green',   icon: '🌬️' },
  transmission:  { color: 'purple',  icon: '⚙️' },
  coolant:       { color: 'cyan',    icon: '💧' },
  spark_plugs:   { color: 'gold',    icon: '⚡' },
  battery:       { color: 'lime',    icon: '🔋' },
  tire_replace:  { color: 'volcano', icon: '🛞' },
  inspection:    { color: 'magenta', icon: '📋' },
  custom:        { color: 'default', icon: '🔧' },
};

export default function ReminderPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const user = useAuthStore((s) => s.user);
  const screens = useBreakpoint();
  const isMobile = !screens.md;

  const [reminders, setReminders] = useState<Reminder[]>([]);
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [loading, setLoading] = useState(true);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingReminder, setEditingReminder] = useState<Reminder | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [form] = Form.useForm();

  const isImperial = user?.unit_system === 'imperial';
  const distUnit = isImperial ? 'mi' : 'km';

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const [remindersRes, vehiclesRes] = await Promise.all([
        reminderApi.list(),
        vehicleApi.list(),
      ]);
      setReminders(remindersRes.data.data || []);
      setVehicles(vehiclesRes.data.data || []);
    } catch {
      message.error(t('common.error'));
    } finally {
      setLoading(false);
    }
  }, [t]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleCreate = () => {
    setEditingReminder(null);
    form.resetFields();
    form.setFieldsValue({ trigger: 'both' });
    setModalOpen(true);
  };

  const handleEdit = (reminder: Reminder) => {
    setEditingReminder(reminder);
    form.setFieldsValue({
      vehicle_id: reminder.vehicle_id,
      category: reminder.category,
      title: reminder.title,
      description: reminder.description,
      trigger: reminder.trigger,
      mileage_interval: reminder.mileage_interval,
      time_interval_days: reminder.time_interval_days,
      last_mileage: reminder.last_mileage,
      last_date: reminder.last_date ? dayjs(reminder.last_date) : undefined,
    });
    setModalOpen(true);
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setSubmitting(true);

      const payload: CreateReminderRequest = {
        vehicle_id: values.vehicle_id,
        category: values.category,
        title: values.title,
        description: values.description || '',
        trigger: values.trigger,
        mileage_interval: values.mileage_interval || 0,
        time_interval_days: values.time_interval_days || 0,
        last_mileage: values.last_mileage || 0,
        last_date: values.last_date ? values.last_date.format('YYYY-MM-DD') : '',
      };

      if (editingReminder) {
        await reminderApi.update(editingReminder.id, payload);
      } else {
        await reminderApi.create(payload);
      }

      message.success(t('common.success'));
      setModalOpen(false);
      fetchData();
    } catch {
      // form validation error or API error
    } finally {
      setSubmitting(false);
    }
  };

  const handleToggle = async (reminder: Reminder, enabled: boolean) => {
    try {
      await reminderApi.update(reminder.id, { is_enabled: enabled });
      fetchData();
    } catch {
      message.error(t('common.error'));
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await reminderApi.delete(id);
      message.success(t('common.success'));
      fetchData();
    } catch {
      message.error(t('common.error'));
    }
  };

  const categoryOptions = [
    'oil_change', 'tire_rotation', 'brake_pads', 'air_filter',
    'transmission', 'coolant', 'spark_plugs', 'battery',
    'tire_replace', 'inspection', 'custom',
  ].map((key) => ({
    value: key,
    label: `${CATEGORY_CONFIG[key]?.icon || '🔧'} ${t(`reminder.category.${key}`)}`,
  }));

  const triggerValue = Form.useWatch('trigger', form);

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', padding: 80 }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div style={{ padding: isMobile ? 16 : 24, maxWidth: 900, margin: '0 auto' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Title level={3} style={{ margin: 0 }}>
          <ToolOutlined style={{ marginRight: 8 }} />
          {t('reminder.title')}
        </Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
          {!isMobile && t('reminder.add')}
        </Button>
      </div>

      {reminders.length === 0 ? (
        <Card>
          <Empty description={t('reminder.noReminders')} />
        </Card>
      ) : (
        <Row gutter={[16, 16]}>
          {reminders.map((reminder) => {
            const cfg = CATEGORY_CONFIG[reminder.category] || CATEGORY_CONFIG.custom;
            const isOverdue = reminder.is_overdue || (
              reminder.next_date && new Date(reminder.next_date) < new Date()
            );

            return (
              <Col xs={24} sm={24} md={12} key={reminder.id}>
                <Card
                  size="small"
                  style={{
                    opacity: reminder.is_enabled ? 1 : 0.6,
                    borderLeft: `3px solid ${isOverdue && reminder.is_enabled ? '#ff4d4f' : 'transparent'}`,
                  }}
                  actions={[
                    <Tooltip title={t('expense.completeAndLog')} key="expense">
                      <WalletOutlined onClick={() => navigate(
                        `/vehicles/${reminder.vehicle_id}/expenses/new?reminder_id=${reminder.id}&category=maintenance&maintenance_category=${reminder.category}&title=${encodeURIComponent(reminder.title)}`
                      )} />
                    </Tooltip>,
                    <Tooltip title={t('common.edit')} key="edit">
                      <EditOutlined onClick={() => handleEdit(reminder)} />
                    </Tooltip>,
                    <Popconfirm
                      key="delete"
                      title={t('reminder.deleteConfirm')}
                      onConfirm={() => handleDelete(reminder.id)}
                    >
                      <DeleteOutlined style={{ color: '#ff4d4f' }} />
                    </Popconfirm>,
                  ]}
                >
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                    <Space direction="vertical" size={4} style={{ flex: 1 }}>
                      <Space>
                        <span style={{ fontSize: 20 }}>{cfg.icon}</span>
                        <Text strong>{reminder.title}</Text>
                        {isOverdue && reminder.is_enabled && (
                          <Tag color="error">{t('reminder.overdue')}</Tag>
                        )}
                      </Space>
                      <Space size={4}>
                        <Tag icon={<CarOutlined />} color="processing">{reminder.vehicle_name}</Tag>
                        <Tag color={cfg.color}>{t(`reminder.category.${reminder.category}`)}</Tag>
                      </Space>
                    </Space>
                    <Switch
                      size="small"
                      checked={reminder.is_enabled}
                      onChange={(checked) => handleToggle(reminder, checked)}
                    />
                  </div>

                  <Divider style={{ margin: '8px 0' }} />

                  <Space direction="vertical" size={2} style={{ width: '100%' }}>
                    {(reminder.trigger === 'mileage' || reminder.trigger === 'both') && (reminder.next_mileage ?? 0) > 0 && (
                      <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                        <Text type="secondary">
                          <DashboardOutlined style={{ marginRight: 4 }} />
                          {t('reminder.nextMileage')}
                        </Text>
                        <Text>{(reminder.next_mileage ?? 0).toLocaleString()} {distUnit}</Text>
                      </div>
                    )}
                    {(reminder.trigger === 'time' || reminder.trigger === 'both') && reminder.next_date && (
                      <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                        <Text type="secondary">
                          <CalendarOutlined style={{ marginRight: 4 }} />
                          {t('reminder.nextDate')}
                        </Text>
                        <Text>{dayjs(reminder.next_date).format('YYYY-MM-DD')}</Text>
                      </div>
                    )}
                    {(reminder.mileage_interval ?? 0) > 0 && (
                      <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                        <Text type="secondary">{t('reminder.interval')}</Text>
                        <Text>{t('reminder.everyMileage', { value: (reminder.mileage_interval ?? 0).toLocaleString(), unit: distUnit })}</Text>
                      </div>
                    )}
                    {(reminder.time_interval_days ?? 0) > 0 && (
                      <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                        <Text type="secondary">{t('reminder.interval')}</Text>
                        <Text>{t('reminder.everyDays', { count: reminder.time_interval_days ?? 0 })}</Text>
                      </div>
                    )}
                  </Space>
                </Card>
              </Col>
            );
          })}
        </Row>
      )}

      {/* 创建/编辑弹窗 */}
      <Modal
        title={editingReminder ? t('reminder.editReminder') : t('reminder.addReminder')}
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        onOk={handleSubmit}
        confirmLoading={submitting}
        destroyOnHidden
        width={520}
      >
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item
            name="vehicle_id"
            label={t('reminder.vehicle')}
            rules={[{ required: true }]}
          >
            <Select
              placeholder={t('reminder.selectVehicle')}
              options={vehicles.map((v) => ({ value: v.id, label: v.name }))}
            />
          </Form.Item>

          <Form.Item
            name="category"
            label={t('reminder.categoryLabel')}
            rules={[{ required: true }]}
          >
            <Select options={categoryOptions} placeholder={t('reminder.selectCategory')} />
          </Form.Item>

          <Form.Item
            name="title"
            label={t('reminder.titleLabel')}
            rules={[{ required: true, max: 200 }]}
          >
            <Input placeholder={t('reminder.titlePlaceholder')} />
          </Form.Item>

          <Form.Item name="description" label={t('reminder.description')}>
            <Input.TextArea rows={2} />
          </Form.Item>

          <Form.Item
            name="trigger"
            label={t('reminder.triggerType')}
            rules={[{ required: true }]}
          >
            <Select
              options={[
                { value: 'mileage', label: t('reminder.triggerMileage') },
                { value: 'time', label: t('reminder.triggerTime') },
                { value: 'both', label: t('reminder.triggerBoth') },
              ]}
            />
          </Form.Item>

          {(triggerValue === 'mileage' || triggerValue === 'both') && (
            <Form.Item
              name="mileage_interval"
              label={t('reminder.mileageInterval', { unit: distUnit })}
              rules={[{ required: true, type: 'number', min: 1 }]}
            >
              <InputNumber style={{ width: '100%' }} min={1} addonAfter={distUnit} />
            </Form.Item>
          )}

          {(triggerValue === 'time' || triggerValue === 'both') && (
            <Form.Item
              name="time_interval_days"
              label={t('reminder.timeInterval')}
              rules={[{ required: true, type: 'number', min: 1 }]}
            >
              <InputNumber style={{ width: '100%' }} min={1} addonAfter={t('reminder.days')} />
            </Form.Item>
          )}

          <Form.Item name="last_mileage" label={t('reminder.lastMileage', { unit: distUnit })}>
            <InputNumber style={{ width: '100%' }} min={0} addonAfter={distUnit} />
          </Form.Item>

          <Form.Item name="last_date" label={t('reminder.lastDate')}>
            <DatePicker style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
