import { useEffect, useState, useCallback } from 'react';
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
  Pagination,
  Select,
  Input,
  Row,
  Col,
  Statistic,
  DatePicker,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ArrowLeftOutlined,
  EyeOutlined,
  FilterOutlined,
  ClearOutlined,
  DollarOutlined,
  ToolOutlined,
  SettingOutlined,
  CalendarOutlined,
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
import type { ExpenseRecord, Vehicle, ExpenseStatsResponse, ExpenseCategory } from '@gastrack/shared';
import type { ColumnsType } from 'antd/es/table';
import { useIsMobile } from '../../hooks/useIsMobile';
import dayjs from 'dayjs';

const { RangePicker } = DatePicker;

const EXPENSE_CATEGORIES: ExpenseCategory[] = [
  'maintenance', 'repair', 'insurance', 'parking', 'toll',
  'car_wash', 'inspection', 'parts', 'fine', 'other',
];

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

export default function ExpenseListPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { vehicleId } = useParams<{ vehicleId: string }>();
  const user = useAuthStore((s) => s.user);
  const isMobile = useIsMobile();

  const [vehicle, setVehicle] = useState<Vehicle | null>(null);
  const [records, setRecords] = useState<ExpenseRecord[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [stats, setStats] = useState<ExpenseStatsResponse | null>(null);

  // 筛选状态
  const [filterCategory, setFilterCategory] = useState<string>('');
  const [filterKeyword, setFilterKeyword] = useState('');
  const [filterDateRange, setFilterDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(null);
  const [showFilters, setShowFilters] = useState(false);

  const currency = user?.currency_code || 'CNY';
  const userTimezone = user?.timezone;

  useEffect(() => {
    if (vehicleId) {
      loadVehicle();
      loadRecords(1);
      loadStats();
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

  const loadStats = async () => {
    try {
      const { data } = await expenseApi.getStats(vehicleId!);
      setStats(data.data);
    } catch {
      // silent
    }
  };

  const loadRecords = useCallback(async (p: number) => {
    setLoading(true);
    try {
      const params: Record<string, unknown> = { page: p, page_size: 20 };
      if (filterCategory) params.category = filterCategory;
      if (filterKeyword) params.keyword = filterKeyword;
      if (filterDateRange) {
        params.start_date = filterDateRange[0].format('YYYY-MM-DD');
        params.end_date = filterDateRange[1].format('YYYY-MM-DD');
      }
      const { data: resp } = await expenseApi.list(vehicleId!, params);
      setRecords(resp.data);
      setTotal(resp.meta.total);
      setPage(p);
    } catch {
      message.error(t('common.error'));
    } finally {
      setLoading(false);
    }
  }, [vehicleId, filterCategory, filterKeyword, filterDateRange, t]);

  // 筛选条件变更后刷新
  useEffect(() => {
    if (vehicleId) {
      loadRecords(1);
    }
  }, [filterCategory, filterKeyword, filterDateRange]);

  const handleDelete = async (expenseId: string) => {
    try {
      await expenseApi.delete(vehicleId!, expenseId);
      message.success(t('expense.deleteSuccess'));
      loadRecords(page);
      loadStats();
    } catch {
      message.error(t('common.error'));
    }
  };

  const clearFilters = () => {
    setFilterCategory('');
    setFilterKeyword('');
    setFilterDateRange(null);
  };

  const columns: ColumnsType<ExpenseRecord> = [
    {
      title: t('expense.field.date'),
      dataIndex: 'expense_date',
      width: 120,
      render: (v: string) => formatDateTime(v, userTimezone, 'YYYY-MM-DD'),
    },
    {
      title: t('expense.category.label'),
      dataIndex: 'category',
      width: 100,
      render: (v: string) => (
        <Tag color={CATEGORY_COLORS[v] || 'default'}>
          {t(`expense.category.${v}`)}
        </Tag>
      ),
    },
    {
      title: t('expense.field.title'),
      dataIndex: 'title',
      ellipsis: true,
    },
    {
      title: t('expense.field.amount'),
      dataIndex: 'amount',
      width: 130,
      render: (v: number, record: ExpenseRecord) =>
        formatCurrency(v, record.currency_code || currency),
    },
    {
      title: t('expense.field.vendor'),
      dataIndex: 'vendor_name',
      ellipsis: true,
      render: (v: string) => v || '-',
    },
    {
      title: t('expense.field.odometer'),
      dataIndex: 'odometer',
      width: 110,
      render: (v: number, record: ExpenseRecord) =>
        v ? `${formatNumber(v, 0)} ${record.distance_unit || 'km'}` : '-',
    },
    {
      title: '',
      key: 'actions',
      width: 100,
      render: (_: unknown, record: ExpenseRecord) => (
        <Space>
          <Button
            type="text"
            size="small"
            icon={<EyeOutlined />}
            onClick={(e) => {
              e.stopPropagation();
              navigate(`/vehicles/${vehicleId}/expenses/${record.id}`);
            }}
          />
          <Button
            type="text"
            size="small"
            icon={<EditOutlined />}
            onClick={(e) => {
              e.stopPropagation();
              navigate(`/vehicles/${vehicleId}/expenses/${record.id}/edit`);
            }}
          />
          <Popconfirm
            title={t('expense.deleteConfirm')}
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
      {records.length === 0 && !loading && (
        <div style={{ textAlign: 'center', padding: 40, color: '#999' }}>
          <p>{t('expense.empty')}</p>
          <p style={{ fontSize: 12 }}>{t('expense.emptyHint')}</p>
        </div>
      )}
      {records.map((record) => (
        <Card
          key={record.id}
          size="small"
          className="mobile-record-card"
          onClick={() => navigate(`/vehicles/${vehicleId}/expenses/${record.id}`)}
        >
          <div className="card-header">
            <Space>
              <Tag color={CATEGORY_COLORS[record.category] || 'default'} style={{ margin: 0 }}>
                {t(`expense.category.${record.category}`)}
              </Tag>
              <span className="date">
                {formatDateTime(record.expense_date, userTimezone, 'YYYY-MM-DD')}
              </span>
            </Space>
            <span className="cost">
              {formatCurrency(record.amount, record.currency_code || currency)}
            </span>
          </div>

          <div className="card-row">
            <span className="label">{t('expense.field.title')}</span>
            <span className="value">{record.title}</span>
          </div>

          {record.vendor_name && (
            <div className="card-row">
              <span className="label">{t('expense.field.vendor')}</span>
              <span className="value">{record.vendor_name}</span>
            </div>
          )}

          <div className="card-actions" onClick={(e) => e.stopPropagation()}>
            <Button
              type="text"
              size="small"
              icon={<EyeOutlined />}
              onClick={() => navigate(`/vehicles/${vehicleId}/expenses/${record.id}`)}
            />
            <Button
              type="text"
              size="small"
              icon={<EditOutlined />}
              onClick={() => navigate(`/vehicles/${vehicleId}/expenses/${record.id}/edit`)}
            />
            <Popconfirm
              title={t('expense.deleteConfirm')}
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
                ? t('expense.title')
                : `${vehicle.name} - ${t('expense.title')}`
              : t('expense.title')}
          </h2>
        </Space>
        <Space>
          <Button
            icon={<FilterOutlined />}
            onClick={() => setShowFilters(!showFilters)}
          />
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => navigate(`/vehicles/${vehicleId}/expenses/new`)}
          >
            {isMobile ? '' : t('expense.add')}
          </Button>
        </Space>
      </div>

      {/* 统计摘要 */}
      {stats && (
        <Row gutter={[12, 12]} style={{ marginBottom: 16 }}>
          <Col xs={12} sm={6}>
            <Card size="small">
              <Statistic
                title={t('expense.stats.totalExpense')}
                value={stats.totals_by_currency?.[0]?.total_amount || 0}
                prefix={<DollarOutlined />}
                precision={2}
                suffix={stats.totals_by_currency?.[0]?.currency_code || currency}
                valueStyle={{ fontSize: isMobile ? 16 : 20 }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={6}>
            <Card size="small">
              <Statistic
                title={t('expense.stats.maintenanceCost')}
                value={stats.category_breakdown?.find(c => c.category === 'maintenance')?.total_amount || 0}
                prefix={<ToolOutlined />}
                precision={2}
                valueStyle={{ fontSize: isMobile ? 16 : 20 }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={6}>
            <Card size="small">
              <Statistic
                title={t('expense.stats.repairCost')}
                value={stats.category_breakdown?.find(c => c.category === 'repair')?.total_amount || 0}
                prefix={<SettingOutlined />}
                precision={2}
                valueStyle={{ fontSize: isMobile ? 16 : 20 }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={6}>
            <Card size="small">
              <Statistic
                title={t('expense.stats.last30Days')}
                value={stats.last_30_days_amount || 0}
                prefix={<CalendarOutlined />}
                precision={2}
                suffix={stats.last_30_days_currency || currency}
                valueStyle={{ fontSize: isMobile ? 16 : 20 }}
              />
            </Card>
          </Col>
        </Row>
      )}

      {/* 筛选区 */}
      {showFilters && (
        <Card size="small" style={{ marginBottom: 16 }}>
          <Row gutter={[12, 12]} align="middle">
            <Col xs={24} sm={6}>
              <Select
                allowClear
                placeholder={t('expense.filter.category')}
                value={filterCategory || undefined}
                onChange={(v) => setFilterCategory(v || '')}
                style={{ width: '100%' }}
              >
                {EXPENSE_CATEGORIES.map((cat) => (
                  <Select.Option key={cat} value={cat}>
                    <Tag color={CATEGORY_COLORS[cat]} style={{ margin: 0 }}>{t(`expense.category.${cat}`)}</Tag>
                  </Select.Option>
                ))}
              </Select>
            </Col>
            <Col xs={24} sm={8}>
              <RangePicker
                value={filterDateRange}
                onChange={(dates) => setFilterDateRange(dates as [dayjs.Dayjs, dayjs.Dayjs] | null)}
                style={{ width: '100%' }}
                placeholder={[t('expense.filter.startDate'), t('expense.filter.endDate')]}
              />
            </Col>
            <Col xs={24} sm={8}>
              <Input.Search
                placeholder={t('expense.filter.keywordPlaceholder')}
                value={filterKeyword}
                onChange={(e) => setFilterKeyword(e.target.value)}
                allowClear
              />
            </Col>
            <Col xs={24} sm={2}>
              <Button icon={<ClearOutlined />} onClick={clearFilters} block>
                {isMobile ? t('expense.filter.clear') : ''}
              </Button>
            </Col>
          </Row>
        </Card>
      )}

      {/* 车辆信息摘要 */}
      {vehicle && (
        <Card size="small" style={{ marginBottom: 16 }}>
          <Space wrap split={<Typography.Text type="secondary">|</Typography.Text>}>
            <span>{vehicle.brand} {vehicle.model}</span>
            <span>{vehicle.year}</span>
            <Tag style={{ margin: 0 }}>{t(`fuelType.${vehicle.fuel_type}`)}</Tag>
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
            scroll={{ x: 900 }}
            size="middle"
            onRow={(record) => ({
              onClick: () => navigate(`/vehicles/${vehicleId}/expenses/${record.id}`),
              style: { cursor: 'pointer' },
            })}
            locale={{ emptyText: t('expense.empty') }}
          />
        )}
      </Card>
    </div>
  );
}
