import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Card,
  Table,
  Button,
  Space,
  Popconfirm,
  message,
  Tag,
  Typography,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ArrowLeftOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import {
  fuelRecordApi,
  vehicleApi,
  formatCurrency,
  formatNumber,
  useAuthStore,
} from '@gastrack/shared';
import type { FuelRecord, Vehicle } from '@gastrack/shared';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';

export default function RecordListPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { vehicleId } = useParams<{ vehicleId: string }>();
  const user = useAuthStore((s) => s.user);

  const [vehicle, setVehicle] = useState<Vehicle | null>(null);
  const [records, setRecords] = useState<FuelRecord[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);

  const currency = user?.currency_code || 'CNY';
  const isImperial = user?.unit_system === 'imperial';
  const fuelUnit = isImperial ? 'gal' : 'L';
  const distanceUnit = isImperial ? 'mi' : 'km';
  const efficiencyUnit = user?.fuel_efficiency_unit || 'L/100km';

  useEffect(() => {
    if (vehicleId) {
      loadVehicle();
      loadRecords(1);
    }
  }, [vehicleId]);

  const loadVehicle = async () => {
    try {
      const { data } = await vehicleApi.getById(vehicleId!);
      setVehicle(data.data);
    } catch {
      message.error(t('common.error'));
    }
  };

  const loadRecords = async (p: number) => {
    setLoading(true);
    try {
      const { data: resp } = await fuelRecordApi.list(vehicleId!, {
        page: p,
        page_size: 20,
      });
      // 后端分页格式: { code, message, data: FuelRecord[], meta: { page, page_size, total, total_pages } }
      setRecords(resp.data);
      setTotal(resp.meta.total);
      setPage(p);
    } catch {
      message.error(t('common.error'));
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (recordId: string) => {
    try {
      await fuelRecordApi.delete(vehicleId!, recordId);
      message.success(t('common.success'));
      loadRecords(page);
    } catch {
      message.error(t('common.error'));
    }
  };

  const columns: ColumnsType<FuelRecord> = [
    {
      title: t('fuelRecord.fuelDate'),
      dataIndex: 'refuel_date',
      width: 120,
      render: (v: string) => dayjs(v).format('YYYY-MM-DD'),
    },
    {
      title: t('fuelRecord.station'),
      dataIndex: 'station_name',
      ellipsis: true,
      render: (v: string) => v || '-',
    },
    {
      title: t('fuelRecord.fuelAmount'),
      dataIndex: 'fuel_amount',
      width: 100,
      render: (v: number) => `${formatNumber(v)} ${fuelUnit}`,
    },
    {
      title: t('fuelRecord.pricePerUnit'),
      dataIndex: 'unit_price',
      width: 100,
      render: (v: number) => formatCurrency(v, currency),
    },
    {
      title: t('fuelRecord.totalCost'),
      dataIndex: 'total_cost',
      width: 110,
      render: (v: number) => formatCurrency(v, currency),
    },
    {
      title: t('fuelRecord.odometer'),
      dataIndex: 'odometer',
      width: 110,
      render: (v: number) => `${formatNumber(v, 0)} ${distanceUnit}`,
    },
    {
      title: t('fuelRecord.consumption'),
      dataIndex: 'fuel_efficiency',
      width: 110,
      render: (v?: number) =>
        v ? (
          <Tag color="blue">{formatNumber(v)} {efficiencyUnit}</Tag>
        ) : (
          <Tag>-</Tag>
        ),
    },
    {
      title: t('fuelRecord.isFullTank'),
      dataIndex: 'is_full_tank',
      width: 80,
      render: (v: boolean) =>
        v ? <Tag color="green">✓</Tag> : <Tag>✗</Tag>,
    },
    {
      title: '',
      key: 'actions',
      width: 100,
      render: (_: unknown, record: FuelRecord) => (
        <Space>
          <Button
            type="text"
            size="small"
            icon={<EditOutlined />}
            onClick={() =>
              navigate(
                `/vehicles/${vehicleId}/records/${record.id}/edit`,
              )
            }
          />
          <Popconfirm
            title={t('fuelRecord.deleteConfirm')}
            onConfirm={() => handleDelete(record.id)}
          >
            <Button type="text" size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div className="page-container">
      <div className="page-header">
        <Space>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/vehicles')}
          />
          <h2>
            {vehicle
              ? `${vehicle.name} - ${t('fuelRecord.title')}`
              : t('fuelRecord.title')}
          </h2>
        </Space>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => navigate(`/vehicles/${vehicleId}/records/new`)}
        >
          {t('fuelRecord.addRecord')}
        </Button>
      </div>

      {vehicle && (
        <Card size="small" style={{ marginBottom: 16 }}>
          <Space split={<Typography.Text type="secondary">|</Typography.Text>}>
            <span>
              {vehicle.brand} {vehicle.model}
            </span>
            <span>{vehicle.year}</span>
            <Tag>{t(`fuelType.${vehicle.fuel_type}`)}</Tag>
            <span>{isImperial ? (vehicle.tank_capacity / 3.78541).toFixed(1) : vehicle.tank_capacity} {fuelUnit}</span>
          </Space>
        </Card>
      )}

      <Card>
        <Table
          columns={columns}
          dataSource={records}
          rowKey="id"
          loading={loading}
          pagination={{
            current: page,
            total,
            pageSize: 20,
            onChange: loadRecords,
            showTotal: (total) => t('common.totalItems', { total }),
          }}
          scroll={{ x: 900 }}
          size="middle"
        />
      </Card>
    </div>
  );
}
