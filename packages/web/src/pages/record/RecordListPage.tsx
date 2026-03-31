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
  Tooltip,
  Pagination,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ArrowLeftOutlined,
  EyeOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import {
  fuelRecordApi,
  vehicleApi,
  formatCurrency,
  formatNumber,
  formatDateTime,
  convertFuelEfficiency,
  litersToGallons,
  FUEL_EFFICIENCY_UNITS,
  useAuthStore,
  useExchangeRateStore,
  convertAmount,
  getReferenceCurrencies,
} from '@gastrack/shared';
import type { FuelRecord, Vehicle } from '@gastrack/shared';
import type { ColumnsType } from 'antd/es/table';
import { useIsMobile } from '../../hooks/useIsMobile';

export default function RecordListPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { vehicleId } = useParams<{ vehicleId: string }>();
  const user = useAuthStore((s) => s.user);
  const isMobile = useIsMobile();

  const { data: ratesData, fetchRates } = useExchangeRateStore();

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
    if (user?.currency_code) {
      fetchRates(user.currency_code);
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

  const userTimezone = user?.timezone;

  /** 生成汇率换算 Tooltip 文本 */
  const getRateTooltip = (amount: number, suffix?: string) => {
    if (!ratesData?.rates) return undefined;
    const base = ratesData.base;
    if (base !== currency) return undefined;
    const lines = getReferenceCurrencies(currency, user?.reference_currency)
      .map((refCode) => {
        const refAmount = convertAmount(amount, currency, refCode, ratesData.rates);
        return refAmount != null ? `≈ ${formatCurrency(refAmount, refCode)}${suffix || ''}` : null;
      })
      .filter(Boolean);
    return lines.length > 0 ? lines.join('  ·  ') : undefined;
  };

  const columns: ColumnsType<FuelRecord> = [
    {
      title: t('fuelRecord.fuelDate'),
      dataIndex: 'refuel_date',
      width: 120,
      render: (v: string) => (
        <Tooltip title={formatDateTime(v, userTimezone, 'YYYY-MM-DD HH:mm')}>
          <span style={{ cursor: 'pointer' }}>
            {formatDateTime(v, userTimezone, 'YYYY-MM-DD')}
          </span>
        </Tooltip>
      ),
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
      render: (v: number, record: FuelRecord) => `${formatNumber(v)} ${record.fuel_unit || fuelUnit}`,
    },
    {
      title: t('fuelRecord.pricePerUnit'),
      dataIndex: 'unit_price',
      width: 100,
      render: (v: number, record: FuelRecord) => {
        const unit = record.fuel_unit || fuelUnit;
        const tip = v != null ? getRateTooltip(v, `/${unit}`) : undefined;
        const text = formatCurrency(v, record.currency_code || currency);
        return tip ? (
          <Tooltip title={tip}>
            <span style={{ cursor: 'help' }}>{text}</span>
          </Tooltip>
        ) : text;
      },
    },
    {
      title: t('fuelRecord.totalCost'),
      dataIndex: 'total_cost',
      width: 110,
      render: (v: number, record: FuelRecord) => {
        const tip = v != null ? getRateTooltip(v) : undefined;
        const text = formatCurrency(v, record.currency_code || currency);
        return tip ? (
          <Tooltip title={tip}>
            <span style={{ cursor: 'help' }}>{text}</span>
          </Tooltip>
        ) : text;
      },
    },
    {
      title: t('fuelRecord.odometer'),
      dataIndex: 'odometer',
      width: 110,
      render: (v: number, record: FuelRecord) => `${formatNumber(v, 0)} ${record.distance_unit || distanceUnit}`,
    },
    {
      title: t('fuelRecord.tripDistance'),
      dataIndex: 'trip_distance',
      width: 100,
      render: (v: number | undefined, record: FuelRecord) =>
        v ? `${formatNumber(v, 1)} ${record.distance_unit || distanceUnit}` : '-',
    },
    {
      title: t('fuelRecord.consumption'),
      dataIndex: 'fuel_efficiency',
      width: 130,
      render: (v?: number) => {
        if (!v) return <Tag>-</Tag>;

        // 判断是否为电动车单位（不参与油耗互转）
        const isEvUnit = ['kWh/100km', 'km/kWh', 'mi/kWh'].includes(efficiencyUnit);

        if (isEvUnit) {
          return <Tag color="blue">{formatNumber(v)} {efficiencyUnit}</Tag>;
        }

        // 构建其他两种单位的换算值
        const otherUnits = FUEL_EFFICIENCY_UNITS
          .filter((u) => u !== efficiencyUnit)
          .map((u) => {
            const converted = convertFuelEfficiency(v, efficiencyUnit, u);
            return `${formatNumber(converted)} ${u}`;
          });

        return (
          <Tooltip
            title={otherUnits.join('\n')}
            overlayInnerStyle={{ whiteSpace: 'pre-line' }}
          >
            <Tag color="blue" style={{ cursor: 'pointer' }}>
              {formatNumber(v)} {efficiencyUnit}
            </Tag>
          </Tooltip>
        );
      },
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
          <Tooltip title={t('recordDetail.title')}>
            <Button
              type="text"
              size="small"
              icon={<EyeOutlined />}
              onClick={(e) => {
                e.stopPropagation();
                navigate(`/vehicles/${vehicleId}/records/${record.id}`);
              }}
            />
          </Tooltip>
          <Button
            type="text"
            size="small"
            icon={<EditOutlined />}
            onClick={(e) => {
              e.stopPropagation();
              navigate(
                `/vehicles/${vehicleId}/records/${record.id}/edit`,
              );
            }}
          />
          <Popconfirm
            title={t('fuelRecord.deleteConfirm')}
            onConfirm={() => handleDelete(record.id)}
          >
            <Button
              type="text"
              size="small"
              danger
              icon={<DeleteOutlined />}
              onClick={(e) => e.stopPropagation()}
            />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  /** 移动端卡片列表 */
  const renderMobileCards = () => (
    <div className="mobile-card-list">
      {records.map((record) => (
        <Card
          key={record.id}
          size="small"
          className="mobile-record-card"
          onClick={() => navigate(`/vehicles/${vehicleId}/records/${record.id}`)}
        >
          <div className="card-header">
            <span className="date">
              {formatDateTime(record.refuel_date, userTimezone, 'YYYY-MM-DD')}
            </span>
            <Tooltip title={getRateTooltip(record.total_cost)}>
              <span className="cost" style={getRateTooltip(record.total_cost) ? { cursor: 'help' } : undefined}>
                {formatCurrency(record.total_cost, currency)}
              </span>
            </Tooltip>
          </div>

          {record.station_name && (
            <div className="card-row">
              <span className="label">{t('fuelRecord.station')}</span>
              <span className="value">{record.station_name}</span>
            </div>
          )}

          <div className="card-row">
            <span className="label">{t('fuelRecord.fuelAmount')}</span>
            <span className="value">{formatNumber(record.fuel_amount)} {record.fuel_unit || fuelUnit}</span>
          </div>

          <div className="card-row">
            <span className="label">{t('fuelRecord.odometer')}</span>
            <span className="value">{formatNumber(record.odometer, 0)} {record.distance_unit || distanceUnit}</span>
          </div>

          {record.fuel_efficiency != null && record.fuel_efficiency > 0 && (
            <div className="card-row">
              <span className="label">{t('fuelRecord.consumption')}</span>
              <span className="value">
                <Tag color="blue" style={{ margin: 0 }}>
                  {formatNumber(record.fuel_efficiency)} {efficiencyUnit}
                </Tag>
              </span>
            </div>
          )}

          <div className="card-row">
            <span className="label">{t('fuelRecord.isFullTank')}</span>
            <span className="value">
              {record.is_full_tank ? <Tag color="green" style={{ margin: 0 }}>✓</Tag> : <Tag style={{ margin: 0 }}>✗</Tag>}
            </span>
          </div>

          <div className="card-actions" onClick={(e) => e.stopPropagation()}>
            <Button
              type="text"
              size="small"
              icon={<EyeOutlined />}
              onClick={() => navigate(`/vehicles/${vehicleId}/records/${record.id}`)}
            />
            <Button
              type="text"
              size="small"
              icon={<EditOutlined />}
              onClick={() => navigate(`/vehicles/${vehicleId}/records/${record.id}/edit`)}
            />
            <Popconfirm
              title={t('fuelRecord.deleteConfirm')}
              onConfirm={() => handleDelete(record.id)}
            >
              <Button type="text" size="small" danger icon={<DeleteOutlined />} />
            </Popconfirm>
          </div>
        </Card>
      ))}

      {total > 20 && (
        <div style={{ textAlign: 'center', padding: '12px 0' }}>
          <Pagination
            current={page}
            total={total}
            pageSize={20}
            onChange={loadRecords}
            size="small"
            simple
          />
        </div>
      )}
    </div>
  );

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
              ? isMobile
                ? t('fuelRecord.title')
                : `${vehicle.name} - ${t('fuelRecord.title')}`
              : t('fuelRecord.title')}
          </h2>
        </Space>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => navigate(`/vehicles/${vehicleId}/records/new`)}
        >
          {isMobile ? '' : t('fuelRecord.addRecord')}
        </Button>
      </div>

      {vehicle && (
        <Card size="small" style={{ marginBottom: 16 }}>
          <Space wrap split={<Typography.Text type="secondary">|</Typography.Text>}>
            <span>
              {vehicle.brand} {vehicle.model}
            </span>
            <span>{vehicle.year}</span>
            <Tag style={{ margin: 0 }}>{t(`fuelType.${vehicle.fuel_type}`)}</Tag>
            <span>{isImperial ? formatNumber(litersToGallons(vehicle.tank_capacity), 1) : vehicle.tank_capacity} {fuelUnit}</span>
            {vehicle.license_plate && <span>{vehicle.license_plate}</span>}
          </Space>
        </Card>
      )}

      <Card loading={loading}>
        {isMobile ? (
          renderMobileCards()
        ) : (
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
            scroll={{ x: 1050 }}
            size="middle"
            onRow={(record) => ({
              onClick: () => navigate(`/vehicles/${vehicleId}/records/${record.id}`),
              style: { cursor: 'pointer' },
            })}
          />
        )}
      </Card>
    </div>
  );
}
