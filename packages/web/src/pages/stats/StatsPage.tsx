import { useEffect, useState, useMemo } from 'react';
import { Row, Col, Card, Statistic, Select, Empty, Spin, Segmented, Space, theme } from 'antd';
import {
  DashboardOutlined,
  DollarOutlined,
  FileTextOutlined,
  ThunderboltOutlined,
  LeftOutlined,
  RightOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import ReactECharts from 'echarts-for-react';
import {
  useVehicleStore,
  statsApi,
  formatNumber,
  formatCurrency,
  useAuthStore,
  useThemeStore,
} from '@gastrack/shared';
import type { VehicleStats, PeriodStatsItem, PeriodStatsResponse } from '@gastrack/shared';
import { useIsMobile } from '../../hooks/useIsMobile';

export default function StatsPage() {
  const { t } = useTranslation();
  const { vehicles, fetchVehicles } = useVehicleStore();
  const user = useAuthStore((s) => s.user);
  const resolved = useThemeStore((s) => s.resolved);
  const { token } = theme.useToken();
  const isMobile = useIsMobile();

  const [selectedVehicleId, setSelectedVehicleId] = useState<string>('');
  const [stats, setStats] = useState<VehicleStats | null>(null);
  const [periodData, setPeriodData] = useState<PeriodStatsResponse | null>(null);
  const [period, setPeriod] = useState<'month' | 'year'>('month');
  const [selectedYear, setSelectedYear] = useState<number>(new Date().getFullYear());
  const [loading, setLoading] = useState(false);

  const currency = user?.currency_code || 'CNY';
  const isImperial = user?.unit_system === 'imperial';
  const fuelUnit = isImperial ? 'gal' : 'L';
  const distanceUnit = isImperial ? 'mi' : 'km';

  useEffect(() => {
    fetchVehicles();
  }, []);

  useEffect(() => {
    if (vehicles.length > 0 && !selectedVehicleId) {
      const defaultV = vehicles.find((v) => v.is_default) || vehicles[0];
      setSelectedVehicleId(defaultV.id);
    }
  }, [vehicles]);

  useEffect(() => {
    if (selectedVehicleId) {
      loadStats();
    }
  }, [selectedVehicleId, period, selectedYear]);

  const loadStats = async () => {
    setLoading(true);
    try {
      const [statsRes, periodRes] = await Promise.all([
        statsApi.vehicleStats(selectedVehicleId),
        statsApi.periodStats(selectedVehicleId, { period, year: selectedYear }),
      ]);
      setStats(statsRes.data.data);
      setPeriodData(periodRes.data.data);
    } catch {
      // 可能没有数据
    } finally {
      setLoading(false);
    }
  };

  const efficiencyUnit = stats?.fuel_efficiency_unit || periodData?.fuel_efficiency_unit || 'L/100km';

  // 月份标签
  const monthLabels = useMemo(() =>
    Array.from({ length: 12 }, (_, i) => t(`stats.month${i + 1}`)),
    [t],
  );

  /** 按月模式：把 items 填充为 12 个月的数组，缺失月份补0 */
  const fillMonthly = (items: PeriodStatsItem[]): PeriodStatsItem[] => {
    const map = new Map(items.map((item) => [item.period, item]));
    return Array.from({ length: 12 }, (_, i) => {
      const month = String(i + 1).padStart(2, '0');
      const key = `${selectedYear}-${month}`;
      return map.get(key) || {
        period: key,
        total_records: 0,
        total_fuel: 0,
        total_cost: 0,
        total_distance: 0,
        avg_efficiency: 0,
      };
    });
  };

  /** 按月模式的上一年同比也填充为 12 个月 */
  const fillMonthlyPrev = (items: PeriodStatsItem[]): PeriodStatsItem[] => {
    const map = new Map(items.map((item) => [item.period, item]));
    return Array.from({ length: 12 }, (_, i) => {
      const month = String(i + 1).padStart(2, '0');
      const key = `${selectedYear - 1}-${month}`;
      return map.get(key) || {
        period: key,
        total_records: 0,
        total_fuel: 0,
        total_cost: 0,
        total_distance: 0,
        avg_efficiency: 0,
      };
    });
  };

  // 准备图表数据
  const currentItems = useMemo(() => {
    if (!periodData) return [];
    return period === 'month' ? fillMonthly(periodData.items) : periodData.items;
  }, [periodData, period, selectedYear]);

  const prevItems = useMemo(() => {
    if (!periodData) return [];
    return period === 'month' ? fillMonthlyPrev(periodData.prev_items) : [];
  }, [periodData, period, selectedYear]);

  const xLabels = useMemo(() => {
    if (period === 'month') return monthLabels;
    return currentItems.map((item) => item.period);
  }, [period, currentItems, monthLabels]);

  const hasPrevData = prevItems.length > 0 && prevItems.some((item) => item.total_records > 0);

  // ECharts 暗色主题文字/坐标轴颜色
  const isDark = resolved === 'dark';
  const chartTextColor = isDark ? 'rgba(255,255,255,0.65)' : '#666';
  const chartAxisLineColor = isDark ? '#303030' : '#e0e0e0';
  const chartSplitLineColor = isDark ? '#303030' : '#f0f0f0';

  /** 构建含同比的图表 option */
  const buildChartOption = (
    field: keyof Pick<PeriodStatsItem, 'total_cost' | 'total_distance' | 'avg_efficiency' | 'total_records'>,
    yAxisName: string,
    type: 'bar' | 'line' = 'bar',
    color: string = '#1677ff',
  ) => {
    const currentSeries = currentItems.map((item) => item[field] || 0);
    const prevSeries = prevItems.map((item) => item[field] || 0);

    const series: object[] = [
      {
        name: period === 'month'
          ? `${selectedYear} ${t('stats.currentPeriod')}`
          : t('stats.currentPeriod'),
        type,
        data: currentSeries,
        smooth: type === 'line',
        areaStyle: type === 'line' ? { opacity: 0.1 } : undefined,
        itemStyle: {
          color,
          borderRadius: type === 'bar' ? [4, 4, 0, 0] : undefined,
        },
        barMaxWidth: 30,
      },
    ];

    if (hasPrevData && period === 'month') {
      series.push({
        name: `${selectedYear - 1} ${t('stats.previousPeriod')}`,
        type,
        data: prevSeries,
        smooth: type === 'line',
        areaStyle: type === 'line' ? { opacity: 0.05 } : undefined,
        itemStyle: {
          color: isDark ? '#555' : '#d9d9d9',
          borderRadius: type === 'bar' ? [4, 4, 0, 0] : undefined,
        },
        lineStyle: type === 'line' ? { type: 'dashed' as const } : undefined,
        barMaxWidth: 30,
      });
    }

    return {
      tooltip: { trigger: 'axis' as const },
      legend: hasPrevData && period === 'month'
        ? { top: 0, textStyle: { color: chartTextColor } }
        : undefined,
      xAxis: {
        type: 'category' as const,
        data: xLabels,
        axisLabel: { fontSize: 11, color: chartTextColor },
        axisLine: { lineStyle: { color: chartAxisLineColor } },
      },
      yAxis: {
        type: 'value' as const,
        name: yAxisName,
        nameTextStyle: { color: chartTextColor },
        axisLabel: { color: chartTextColor },
        splitLine: { lineStyle: { color: chartSplitLineColor } },
      },
      series,
      grid: { left: isMobile ? 45 : 55, right: 12, top: hasPrevData && period === 'month' ? 40 : 30, bottom: 30 },
    };
  };

  // 年份选择器可用年份范围
  const yearOptions = useMemo(() => {
    const currentYear = new Date().getFullYear();
    const years: number[] = [];
    for (let y = currentYear; y >= currentYear - 10; y--) {
      years.push(y);
    }
    return years.map((y) => ({ value: y, label: `${y}` }));
  }, []);

  if (vehicles.length === 0) {
    return (
      <div className="page-container">
        <div className="page-header">
          <h2>{t('stats.title')}</h2>
        </div>
        <Card>
          <Empty description={t('vehicle.noVehicle')} />
        </Card>
      </div>
    );
  }

  return (
    <div className="page-container">
      <div className="page-header">
        <h2 style={{ margin: 0 }}>{t('stats.title')}</h2>
        {!isMobile && (
          <Space wrap>
            <Select
              value={selectedVehicleId}
              onChange={setSelectedVehicleId}
              style={{ width: 160 }}
              options={vehicles.map((v) => ({
                value: v.id,
                label: v.name,
              }))}
            />
            <Segmented
              value={period}
              onChange={(val) => setPeriod(val as 'month' | 'year')}
              options={[
                { label: t('stats.byMonth'), value: 'month' },
                { label: t('stats.byYear'), value: 'year' },
              ]}
            />
            {period === 'month' && (
              <Space size={4}>
                <LeftOutlined
                  style={{ cursor: 'pointer', fontSize: 14, color: token.colorTextSecondary }}
                  onClick={() => setSelectedYear((y) => y - 1)}
                />
                <Select
                  value={selectedYear}
                  onChange={setSelectedYear}
                  style={{ width: 90 }}
                  options={yearOptions}
                />
                <RightOutlined
                  style={{
                    cursor: selectedYear >= new Date().getFullYear() ? 'not-allowed' : 'pointer',
                    fontSize: 14,
                    color: selectedYear >= new Date().getFullYear()
                      ? token.colorTextDisabled
                      : token.colorTextSecondary,
                  }}
                  onClick={() => {
                    if (selectedYear < new Date().getFullYear()) {
                      setSelectedYear((y) => y + 1);
                    }
                  }}
                />
              </Space>
            )}
          </Space>
        )}
      </div>

      {/* 移动端：筛选条件独立一行 */}
      {isMobile && (
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8, alignItems: 'center', marginBottom: 16 }}>
          <Select
            value={selectedVehicleId}
            onChange={setSelectedVehicleId}
            style={{ minWidth: 120, flex: 1 }}
            options={vehicles.map((v) => ({
              value: v.id,
              label: v.name,
            }))}
          />
          <Segmented
            value={period}
            onChange={(val) => setPeriod(val as 'month' | 'year')}
            options={[
              { label: t('stats.byMonth'), value: 'month' },
              { label: t('stats.byYear'), value: 'year' },
            ]}
          />
          {period === 'month' && (
            <Space size={4}>
              <LeftOutlined
                style={{ cursor: 'pointer', fontSize: 14, color: token.colorTextSecondary }}
                onClick={() => setSelectedYear((y) => y - 1)}
              />
              <Select
                value={selectedYear}
                onChange={setSelectedYear}
                style={{ width: 90 }}
                options={yearOptions}
              />
              <RightOutlined
                style={{
                  cursor: selectedYear >= new Date().getFullYear() ? 'not-allowed' : 'pointer',
                  fontSize: 14,
                  color: selectedYear >= new Date().getFullYear()
                    ? token.colorTextDisabled
                    : token.colorTextSecondary,
                }}
                onClick={() => {
                  if (selectedYear < new Date().getFullYear()) {
                    setSelectedYear((y) => y + 1);
                  }
                }}
              />
            </Space>
          )}
        </div>
      )}

      <Spin spinning={loading}>
        {/* 总览统计卡片 */}
        <Row gutter={isMobile ? [8, 8] : [16, 16]} style={{ marginBottom: 24 }}>
          <Col xs={12} sm={6}>
            <Card>
              <Statistic
                title={t('stats.totalRecords')}
                value={stats?.total_records || 0}
                prefix={<FileTextOutlined />}
              />
            </Card>
          </Col>
          <Col xs={12} sm={6}>
            <Card>
              <Statistic
                title={t('stats.totalCost')}
                value={
                  stats
                    ? formatCurrency(stats.total_cost, currency)
                    : '-'
                }
                prefix={<DollarOutlined />}
              />
            </Card>
          </Col>
          <Col xs={12} sm={6}>
            <Card>
              <Statistic
                title={t('stats.avgConsumption')}
                value={
                  stats ? `${formatNumber(stats.avg_efficiency)} ${efficiencyUnit}` : '-'
                }
                prefix={<DashboardOutlined />}
              />
            </Card>
          </Col>
          <Col xs={12} sm={6}>
            <Card>
              <Statistic
                title={t('stats.totalFuel')}
                value={stats ? `${formatNumber(stats.total_fuel)} ${fuelUnit}` : '-'}
                prefix={<ThunderboltOutlined />}
              />
            </Card>
          </Col>
        </Row>

        {/* 时段统计图表 */}
        {currentItems.length > 0 ? (
          <Row gutter={[16, 16]}>
            <Col xs={24} lg={12}>
              <Card title={period === 'month' ? t('stats.monthlyCost') : t('stats.yearlyCost')}>
                <ReactECharts
                  option={buildChartOption(
                    'total_cost',
                    currency,
                    'bar',
                    '#1677ff',
                  )}
                  style={{ height: isMobile ? 240 : 300 }}
                  theme={isDark ? 'dark' : undefined}
                />
              </Card>
            </Col>
            <Col xs={24} lg={12}>
              <Card title={period === 'month' ? t('stats.monthlyEfficiency') : t('stats.yearlyEfficiency')}>
                <ReactECharts
                  option={buildChartOption(
                    'avg_efficiency',
                    efficiencyUnit,
                    'line',
                    '#faad14',
                  )}
                  style={{ height: isMobile ? 240 : 300 }}
                  theme={isDark ? 'dark' : undefined}
                />
              </Card>
            </Col>
            <Col xs={24} lg={12}>
              <Card title={period === 'month' ? t('stats.monthlyDistance') : t('stats.yearlyDistance')}>
                <ReactECharts
                  option={buildChartOption(
                    'total_distance',
                    distanceUnit,
                    'bar',
                    '#52c41a',
                  )}
                  style={{ height: isMobile ? 240 : 300 }}
                  theme={isDark ? 'dark' : undefined}
                />
              </Card>
            </Col>
            <Col xs={24} lg={12}>
              <Card title={period === 'month' ? t('stats.monthlyRecords') : t('stats.yearlyRecords')}>
                <ReactECharts
                  option={buildChartOption(
                    'total_records',
                    '',
                    'bar',
                    '#722ed1',
                  )}
                  style={{ height: isMobile ? 240 : 300 }}
                  theme={isDark ? 'dark' : undefined}
                />
              </Card>
            </Col>
          </Row>
        ) : (
          <Card>
            <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />
          </Card>
        )}
      </Spin>
    </div>
  );
}
