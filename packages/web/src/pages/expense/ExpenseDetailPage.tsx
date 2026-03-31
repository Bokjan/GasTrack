import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Card,
  Button,
  Space,
  Spin,
  message,
  Popconfirm,
  Tag,
  Descriptions,
  Typography,
} from 'antd';
import {
  ArrowLeftOutlined,
  EditOutlined,
  DeleteOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import {
  expenseApi,
  vehicleApi,
  formatCurrency,
  formatNumber,
  formatDateTime,
  useAuthStore,
} from '@gastrack/shared';
import type { ExpenseRecord, Vehicle } from '@gastrack/shared';
import { useIsMobile } from '../../hooks/useIsMobile';

const { Text } = Typography;

const CATEGORY_COLORS: Record<string, string> = {
  maintenance: 'blue',
  repair: 'red',
  insurance: 'green',
  parking: 'orange',
  toll: 'purple',
  car_wash: 'cyan',
  inspection: 'gold',
  parts: 'geekblue',
  fine: 'magenta',
  other: 'default',
};

export default function ExpenseDetailPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { vehicleId, expenseId } = useParams<{
    vehicleId: string;
    expenseId: string;
  }>();
  const user = useAuthStore((s) => s.user);
  const isMobile = useIsMobile();

  const [record, setRecord] = useState<ExpenseRecord | null>(null);
  const [vehicle, setVehicle] = useState<Vehicle | null>(null);
  const [loading, setLoading] = useState(true);

  const currency = user?.currency_code || 'CNY';
  const userTimezone = user?.timezone;

  useEffect(() => {
    if (vehicleId && expenseId) {
      loadData();
    }
  }, [vehicleId, expenseId]);

  const loadData = async () => {
    setLoading(true);
    try {
      const [recordRes, vehicleRes] = await Promise.all([
        expenseApi.getById(vehicleId!, expenseId!),
        vehicleApi.getById(vehicleId!),
      ]);
      setRecord(recordRes.data.data);
      setVehicle(vehicleRes.data.data);
    } catch {
      message.error(t('common.error'));
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async () => {
    try {
      await expenseApi.delete(vehicleId!, expenseId!);
      message.success(t('expense.deleteSuccess'));
      navigate(`/vehicles/${vehicleId}/expenses`);
    } catch {
      message.error(t('common.error'));
    }
  };

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: 80 }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!record) {
    return (
      <div className="page-container">
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate(`/vehicles/${vehicleId}/expenses`)}>
          {t('common.back')}
        </Button>
      </div>
    );
  }

  return (
    <div className="page-container">
      {/* 页头 */}
      <div className="page-header">
        <Space>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate(`/vehicles/${vehicleId}/expenses`)}
          />
          <h2>
            {isMobile
              ? t('expense.detail')
              : `${vehicle?.name || ''} - ${t('expense.detail')}`}
          </h2>
        </Space>
        <Space>
          <Button
            icon={<EditOutlined />}
            onClick={() => navigate(`/vehicles/${vehicleId}/expenses/${expenseId}/edit`)}
          >
            {isMobile ? '' : t('common.edit')}
          </Button>
          <Popconfirm
            title={t('expense.deleteConfirm')}
            onConfirm={handleDelete}
          >
            <Button danger icon={<DeleteOutlined />}>
              {isMobile ? '' : t('common.delete')}
            </Button>
          </Popconfirm>
        </Space>
      </div>

      <Card>
        <Descriptions
          column={1}
          labelStyle={{ fontWeight: 500, width: isMobile ? 90 : 140 }}
          colon={false}
          size={isMobile ? 'small' : 'middle'}
        >
          <Descriptions.Item label={t('expense.category.label')}>
            <Tag color={CATEGORY_COLORS[record.category] || 'default'}>
              {t(`expense.category.${record.category}`)}
            </Tag>
          </Descriptions.Item>

          {record.maintenance_category && (
            <Descriptions.Item label={t('expense.maintenanceCategory.label')}>
              <Tag>{t(`expense.maintenanceCategory.${record.maintenance_category}`)}</Tag>
            </Descriptions.Item>
          )}

          <Descriptions.Item label={t('expense.field.title')}>
            <Text strong>{record.title}</Text>
          </Descriptions.Item>

          <Descriptions.Item label={t('expense.field.amount')}>
            <Text strong style={{ fontSize: 18, color: 'var(--ant-color-primary)' }}>
              {formatCurrency(record.amount, record.currency_code || currency)}
            </Text>
          </Descriptions.Item>

          <Descriptions.Item label={t('expense.field.date')}>
            {formatDateTime(record.expense_date, userTimezone, 'YYYY-MM-DD')}
          </Descriptions.Item>

          {record.vendor_name && (
            <Descriptions.Item label={t('expense.field.vendor')}>
              {record.vendor_name}
            </Descriptions.Item>
          )}

          {record.odometer != null && record.odometer > 0 && (
            <Descriptions.Item label={t('expense.field.odometer')}>
              {formatNumber(record.odometer, 0)} {record.distance_unit || 'km'}
            </Descriptions.Item>
          )}

          {record.reminder_id && (
            <Descriptions.Item label={t('expense.reminderLink.label')}>
              <Tag color="blue">{t('expense.reminderLink.linked')}</Tag>
            </Descriptions.Item>
          )}

          {record.note && (
            <Descriptions.Item label={t('expense.field.note')}>
              {record.note}
            </Descriptions.Item>
          )}

          {vehicle && (
            <Descriptions.Item label={t('vehicle.title')}>
              {vehicle.name} ({vehicle.brand} {vehicle.model} {vehicle.year})
            </Descriptions.Item>
          )}

          <Descriptions.Item label={t('common.createdAt')}>
            {formatDateTime(record.created_at, userTimezone, 'YYYY-MM-DD HH:mm')}
          </Descriptions.Item>

          <Descriptions.Item label={t('common.updatedAt')}>
            {formatDateTime(record.updated_at, userTimezone, 'YYYY-MM-DD HH:mm')}
          </Descriptions.Item>
        </Descriptions>
      </Card>
    </div>
  );
}
