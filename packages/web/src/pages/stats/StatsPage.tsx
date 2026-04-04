import { useEffect, useState, useMemo } from 'react';
import { Row, Col, Card, Statistic, Select, Empty, Spin, Segmented, Space, theme, Tooltip } from 'antd';
import {
  DashboardOutlined,
  DollarOutlined,
  FileTextOutlined,
  ThunderboltOutlined,
  LeftOutlined,
  RightOutlined,
  WalletOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import ReactECharts from 'echarts-for-react';
import {
  useVehicleStore,
  statsApi,
  expenseApi,
  formatNumber,
  formatCurrency,
  useAuthStore,
  useThemeStore,
  useExchangeRateStore,
  convertAmount,
  getReferenceCurrency,
  sumConvertedCostsByCurrency,
} from '@gastrack/shared';
import type {
  VehicleStats,
  PeriodStatsItem,
  PeriodStatsResponse,
  ExpensePeriodStatsResponse,
  ExpensePeriodStatsItem,
  ExpenseStatsResponse,
} from '@gastrack/shared';
import { useIsMobile } from '../../hooks/useIsMobile';

type StatsTab = 'fuel' | 'expense' | 'combined';

export default function StatsPage() {
  const { t } = useTranslation();
  const { vehicles, fetchVehicles } = useVehicleStore();
  const user = useAuthStore((s) => s.user);
  const resolved = useThemeStore((s) => s.resolved);
  const { token } = theme.useToken();
  const isMobile = useIsMobile();

  const [activeTab, setActiveTab] = useState<StatsTab>('fuel');
  const [selectedVehicleId, setSelectedVehicleId] = useState<string>('');
  const [stats, setStats] = useState<VehicleStats | null>(null);
  const [periodData, setPeriodData] = useState<PeriodStatsResponse | null>(null);
  const [expensePeriodData, setExpensePeriodData] = useState<ExpensePeriodStatsResponse | null>(null);
  const [expenseStats, setExpenseStats] = useState<ExpenseStatsResponse | null>(null);
  const [period, setPeriod] = useState<'month' | 'year'>('month');
  const [selectedYear, setSelectedYear] = useState<number>(new Date().getFullYear());
  const [loading, setLoading] = useState(false);
  const { data: ratesData, fetchRates } = useExchangeRateStore();

  const currency = user?.currency_code || 'CNY';
  const refCurrency = getReferenceCurrency(currency, user?.reference_currency);
  const isImperial = user?.unit_system === 'imperial';
  const fuelUnit = isImperial ? 'gal' : 'L';
  const distanceUnit = isImperial ? 'mi' : 'km';

  useEffect(() => {
    fetchVehicles();
    if (currency) fetchRates(currency);
  }, []);

  useEffect(() => {
    if (vehicles.length > 0 && !selectedVehicleId) {
      const defaultV = vehicles.find((v) => v.is_default) || vehicles[0];
      setSelectedVehicleId(defaultV.id);
    }
  }, [vehicles]);

  useEffect(() => {
    if (selectedVehicleId) loadStats();
  }, [selectedVehicleId, period, selectedYear]);

  const loadStats = async () => {
    setLoading(true);
    try {
      const [statsRes, periodRes, expPeriodRes, expStatsRes] = await Promise.all([
        statsApi.vehicleStats(selectedVehicleId),
        statsApi.periodStats(selectedVehicleId, { period, year: selectedYear }),
        statsApi.expensePeriodStats(selectedVehicleId, { period, year: selectedYear }),
        expenseApi.getStats(selectedVehicleId),
      ]);
      setStats(statsRes.data.data);
      setPeriodData(periodRes.data.data);
      setExpensePeriodData(expPeriodRes.data.data);
      setExpenseStats(expStatsRes.data.data);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  };

  const efficiencyUnit = stats?.fuel_efficiency_unit || periodData?.fuel_efficiency_unit || 'L/100km';

  const monthLabels = useMemo(() =>
    Array.from({ length: 12 }, (_, i) => t(`stats.month${i + 1}`)), [t]);

  // --- Fill helpers ---
  const fillMonthlyFuel = (items: PeriodStatsItem[], yr: number): PeriodStatsItem[] => {
    const map = new Map(items.map((item) => [item.period, item]));
    return Array.from({ length: 12 }, (_, i) => {
      const key = `${yr}-${String(i + 1).padStart(2, '0')}`;
      return map.get(key) || { period: key, total_records: 0, total_fuel: 0, total_cost: 0, total_distance: 0, avg_efficiency: 0 };
    });
  };

  const fillMonthlyExpense = (items: ExpensePeriodStatsItem[], yr: number): ExpensePeriodStatsItem[] => {
    const map = new Map(items.map((item) => [item.period, item]));
    return Array.from({ length: 12 }, (_, i) => {
      const key = `${yr}-${String(i + 1).padStart(2, '0')}`;
      return map.get(key) || { period: key, total_records: 0, total_amount: 0 };
    });
  };

  // --- Fuel period data ---
  const fuelCurrent = useMemo(() => {
    if (!periodData) return [];
    return period === 'month' ? fillMonthlyFuel(periodData.items, selectedYear) : periodData.items;
  }, [periodData, period, selectedYear]);

  const fuelPrev = useMemo(() => {
    if (!periodData) return [];
    return period === 'month' ? fillMonthlyFuel(periodData.prev_items, selectedYear - 1) : [];
  }, [periodData, period, selectedYear]);

  // --- Expense period data ---
  const expCurrent = useMemo(() => {
    if (!expensePeriodData) return [];
    return period === 'month' ? fillMonthlyExpense(expensePeriodData.items, selectedYear) : expensePeriodData.items;
  }, [expensePeriodData, period, selectedYear]);

  const expPrev = useMemo(() => {
    if (!expensePeriodData) return [];
    return period === 'month' ? fillMonthlyExpense(expensePeriodData.prev_items, selectedYear - 1) : [];
  }, [expensePeriodData, period, selectedYear]);

  const fuelXLabels = useMemo(() => period === 'month' ? monthLabels : fuelCurrent.map((i) => i.period), [period, fuelCurrent, monthLabels]);
  const expXLabels = useMemo(() => period === 'month' ? monthLabels : expCurrent.map((i) => i.period), [period, expCurrent, monthLabels]);

  const hasFuelPrev = fuelPrev.length > 0 && fuelPrev.some((i) => i.total_records > 0);
  const hasExpPrev = expPrev.length > 0 && expPrev.some((i) => i.total_records > 0);

  // --- Theme ---
  const isDark = resolved === 'dark';
  const chartTextColor = isDark ? 'rgba(255,255,255,0.65)' : '#666';
  const chartAxisLineColor = isDark ? '#303030' : '#e0e0e0';
  const chartSplitLineColor = isDark ? '#303030' : '#f0f0f0';
  const prevColor = isDark ? '#555' : '#d9d9d9';

  // --- Chart builders ---
  const buildFuelChart = (
    field: keyof Pick<PeriodStatsItem, 'total_cost' | 'total_distance' | 'avg_efficiency' | 'total_records'>,
    yAxisName: string, type: 'bar' | 'line' = 'bar', color = '#1677ff',
  ) => {
    const series: object[] = [{
      name: period === 'month' ? `${selectedYear} ${t('stats.currentPeriod')}` : t('stats.currentPeriod'),
      type, data: fuelCurrent.map((i) => i[field] || 0), smooth: type === 'line',
      areaStyle: type === 'line' ? { opacity: 0.1 } : undefined,
      itemStyle: { color, borderRadius: type === 'bar' ? [4, 4, 0, 0] : undefined }, barMaxWidth: 30,
    }];
    if (hasFuelPrev && period === 'month') {
      series.push({
        name: `${selectedYear - 1} ${t('stats.previousPeriod')}`, type,
        data: fuelPrev.map((i) => i[field] || 0), smooth: type === 'line',
        areaStyle: type === 'line' ? { opacity: 0.05 } : undefined,
        itemStyle: { color: prevColor, borderRadius: type === 'bar' ? [4, 4, 0, 0] : undefined },
        lineStyle: type === 'line' ? { type: 'dashed' as const } : undefined, barMaxWidth: 30,
      });
    }
    return {
      tooltip: { trigger: 'axis' as const },
      legend: hasFuelPrev && period === 'month' ? { top: 0, textStyle: { color: chartTextColor } } : undefined,
      xAxis: { type: 'category' as const, data: fuelXLabels, axisLabel: { fontSize: 11, color: chartTextColor }, axisLine: { lineStyle: { color: chartAxisLineColor } } },
      yAxis: { type: 'value' as const, name: yAxisName, nameTextStyle: { color: chartTextColor }, axisLabel: { color: chartTextColor }, splitLine: { lineStyle: { color: chartSplitLineColor } } },
      series, grid: { left: isMobile ? 45 : 55, right: 12, top: hasFuelPrev && period === 'month' ? 40 : 30, bottom: 30 },
    };
  };

  const buildExpenseChart = (
    field: keyof Pick<ExpensePeriodStatsItem, 'total_amount' | 'total_records'>,
    yAxisName: string, type: 'bar' | 'line' = 'bar', color = '#ff7a45',
  ) => {
    const series: object[] = [{
      name: period === 'month' ? `${selectedYear} ${t('stats.currentPeriod')}` : t('stats.currentPeriod'),
      type, data: expCurrent.map((i) => i[field] || 0), smooth: type === 'line',
      itemStyle: { color, borderRadius: type === 'bar' ? [4, 4, 0, 0] : undefined }, barMaxWidth: 30,
    }];
    if (hasExpPrev && period === 'month') {
      series.push({
        name: `${selectedYear - 1} ${t('stats.previousPeriod')}`, type,
        data: expPrev.map((i) => i[field] || 0), smooth: type === 'line',
        itemStyle: { color: prevColor, borderRadius: type === 'bar' ? [4, 4, 0, 0] : undefined }, barMaxWidth: 30,
      });
    }
    return {
      tooltip: { trigger: 'axis' as const },
      legend: hasExpPrev && period === 'month' ? { top: 0, textStyle: { color: chartTextColor } } : undefined,
      xAxis: { type: 'category' as const, data: expXLabels, axisLabel: { fontSize: 11, color: chartTextColor }, axisLine: { lineStyle: { color: chartAxisLineColor } } },
      yAxis: { type: 'value' as const, name: yAxisName, nameTextStyle: { color: chartTextColor }, axisLabel: { color: chartTextColor }, splitLine: { lineStyle: { color: chartSplitLineColor } } },
      series, grid: { left: isMobile ? 45 : 55, right: 12, top: hasExpPrev && period === 'month' ? 40 : 30, bottom: 30 },
    };
  };

  const buildCombinedCostChart = () => {
    const fuelSeries = fuelCurrent.map((i) => i.total_cost || 0);
    const expSeries = expCurrent.map((i) => i.total_amount || 0);
    const labels = period === 'month' ? monthLabels : fuelCurrent.map((i) => i.period);
    return {
      tooltip: { trigger: 'axis' as const },
      legend: { top: 0, textStyle: { color: chartTextColor } },
      xAxis: { type: 'category' as const, data: labels, axisLabel: { fontSize: 11, color: chartTextColor }, axisLine: { lineStyle: { color: chartAxisLineColor } } },
      yAxis: { type: 'value' as const, name: currency, nameTextStyle: { color: chartTextColor }, axisLabel: { color: chartTextColor }, splitLine: { lineStyle: { color: chartSplitLineColor } } },
      series: [
        { name: t('stats.fuelPortion'), type: 'bar', stack: 'total', data: fuelSeries, itemStyle: { color: '#1677ff', borderRadius: [0, 0, 0, 0] }, barMaxWidth: 30 },
        { name: t('stats.expensePortion'), type: 'bar', stack: 'total', data: expSeries, itemStyle: { color: '#ff7a45', borderRadius: [4, 4, 0, 0] }, barMaxWidth: 30 },
      ],
      grid: { left: isMobile ? 45 : 55, right: 12, top: 40, bottom: 30 },
    };
  };

  const buildCategoryPieChart = () => {
    const breakdown = expenseStats?.category_breakdown || [];
    return {
      tooltip: { trigger: 'item' as const, formatter: '{b}: {c} ({d}%)' },
      legend: { orient: 'vertical' as const, right: 10, top: 'center', textStyle: { color: chartTextColor } },
      series: [{
        type: 'pie', radius: ['40%', '70%'], avoidLabelOverlap: false,
        itemStyle: { borderRadius: 6, borderColor: isDark ? '#1f1f1f' : '#fff', borderWidth: 2 },
        label: { show: false }, emphasis: { label: { show: true, fontSize: 14, fontWeight: 'bold' } },
        data: breakdown.map((c) => ({ value: c.total_amount, name: t(`expense.category.${c.category}`) })),
      }],
    };
  };

  // --- Year selector ---
  const yearOptions = useMemo(() => {
    const cur = new Date().getFullYear();
    return Array.from({ length: 11 }, (_, i) => ({ value: cur - i, label: `${cur - i}` }));
  }, []);

  const yearSelector = (
    <Space size={4}>
      <LeftOutlined style={{ cursor: 'pointer', fontSize: 14, color: token.colorTextSecondary }} onClick={() => setSelectedYear((y) => y - 1)} />
      <Select value={selectedYear} onChange={setSelectedYear} style={{ width: 90 }} options={yearOptions} />
      <RightOutlined
        style={{ cursor: selectedYear >= new Date().getFullYear() ? 'not-allowed' : 'pointer', fontSize: 14, color: selectedYear >= new Date().getFullYear() ? token.colorTextDisabled : token.colorTextSecondary }}
        onClick={() => { if (selectedYear < new Date().getFullYear()) setSelectedYear((y) => y + 1); }}
      />
    </Space>
  );

  // --- Computed totals ---
  const fuelTotalCost = stats ? (sumConvertedCostsByCurrency(stats.costs_by_currency, currency, ratesData?.rates) ?? stats.total_cost) : 0;
  const expenseTotalCost = expensePeriodData ? (sumConvertedCostsByCurrency(expensePeriodData.costs_by_currency, currency, ratesData?.rates) ?? (expenseStats?.totals_by_currency?.[0]?.total_amount || 0)) : 0;
  const combinedCost = fuelTotalCost + expenseTotalCost;

  if (vehicles.length === 0) {
    return (
      <div className="page-container">
        <div className="page-header"><h2>{t('stats.title')}</h2></div>
        <Card><Empty description={t('vehicle.noVehicle')} /></Card>
      </div>
    );
  }

  // --- Filters bar ---
  const filtersBar = (
    <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8, alignItems: 'center', marginBottom: 16 }}>
      <Select value={selectedVehicleId} onChange={setSelectedVehicleId} style={{ minWidth: 120, flex: isMobile ? 1 : undefined, width: isMobile ? undefined : 160 }}
        options={vehicles.map((v) => ({ value: v.id, label: v.name }))} />
      <Segmented value={period} onChange={(val) => setPeriod(val as 'month' | 'year')}
        options={[{ label: t('stats.byMonth'), value: 'month' }, { label: t('stats.byYear'), value: 'year' }]} />
      {period === 'month' && yearSelector}
    </div>
  );

  const chartH = isMobile ? 240 : 300;

  return (
    <div className="page-container">
      <div className="page-header">
        <h2 style={{ margin: 0 }}>{t('stats.title')}</h2>
      </div>

      {filtersBar}

      <Segmented
        value={activeTab}
        onChange={(val) => setActiveTab(val as StatsTab)}
        options={[
          { label: t('stats.tabFuel'), value: 'fuel' },
          { label: t('stats.tabExpense'), value: 'expense' },
          { label: t('stats.tabCombined'), value: 'combined' },
        ]}
        block={isMobile}
        style={{ marginBottom: 20 }}
      />

      <Spin spinning={loading}>
        {/* ========== Tab: Fuel ========== */}
        {activeTab === 'fuel' && (
          <>
            <Row gutter={isMobile ? [8, 8] : [16, 16]} style={{ marginBottom: 24 }}>
              <Col xs={12} sm={6}>
                <Card><Statistic title={t('stats.totalRecords')} value={stats?.total_records || 0} prefix={<FileTextOutlined />} /></Card>
              </Col>
              <Col xs={12} sm={6}>
                <Tooltip title={stats && ratesData?.rates ? (() => {
                  const ref = convertAmount(fuelTotalCost, currency, refCurrency, ratesData.rates);
                  return ref != null ? t('exchangeRate.referenceAmount', { amount: formatCurrency(ref, refCurrency) }) : undefined;
                })() : undefined}>
                  <Card><Statistic title={t('stats.fuelCost')} value={stats ? formatCurrency(fuelTotalCost, currency) : '-'} prefix={<DollarOutlined />} /></Card>
                </Tooltip>
              </Col>
              <Col xs={12} sm={6}>
                <Card><Statistic title={t('stats.avgConsumption')} value={stats ? `${formatNumber(stats.avg_efficiency)} ${efficiencyUnit}` : '-'} prefix={<DashboardOutlined />} /></Card>
              </Col>
              <Col xs={12} sm={6}>
                <Card><Statistic title={t('stats.totalFuel')} value={stats ? `${formatNumber(stats.total_fuel)} ${fuelUnit}` : '-'} prefix={<ThunderboltOutlined />} /></Card>
              </Col>
            </Row>
            {fuelCurrent.length > 0 ? (
              <Row gutter={[16, 16]}>
                <Col xs={24} lg={12}>
                  <Card title={period === 'month' ? t('stats.monthlyCost') : t('stats.yearlyCost')}>
                    <ReactECharts option={buildFuelChart('total_cost', currency, 'bar', '#1677ff')} style={{ height: chartH }} theme={isDark ? 'dark' : undefined} />
                  </Card>
                </Col>
                <Col xs={24} lg={12}>
                  <Card title={period === 'month' ? t('stats.monthlyEfficiency') : t('stats.yearlyEfficiency')}>
                    <ReactECharts option={buildFuelChart('avg_efficiency', efficiencyUnit, 'line', '#faad14')} style={{ height: chartH }} theme={isDark ? 'dark' : undefined} />
                  </Card>
                </Col>
                <Col xs={24} lg={12}>
                  <Card title={period === 'month' ? t('stats.monthlyDistance') : t('stats.yearlyDistance')}>
                    <ReactECharts option={buildFuelChart('total_distance', distanceUnit, 'bar', '#52c41a')} style={{ height: chartH }} theme={isDark ? 'dark' : undefined} />
                  </Card>
                </Col>
                <Col xs={24} lg={12}>
                  <Card title={period === 'month' ? t('stats.monthlyRecords') : t('stats.yearlyRecords')}>
                    <ReactECharts option={buildFuelChart('total_records', '', 'bar', '#722ed1')} style={{ height: chartH }} theme={isDark ? 'dark' : undefined} />
                  </Card>
                </Col>
              </Row>
            ) : (<Card><Empty image={Empty.PRESENTED_IMAGE_SIMPLE} /></Card>)}
          </>
        )}

        {/* ========== Tab: Expense ========== */}
        {activeTab === 'expense' && (
          <>
            <Row gutter={isMobile ? [8, 8] : [16, 16]} style={{ marginBottom: 24 }}>
              <Col xs={12} sm={6}>
                <Card><Statistic title={t('stats.totalExpenseRecords')} value={expenseStats?.total_records || 0} prefix={<FileTextOutlined />} /></Card>
              </Col>
              <Col xs={12} sm={6}>
                <Card><Statistic title={t('stats.totalExpenseAmount')} value={expenseStats ? formatCurrency(expenseTotalCost, currency) : '-'} prefix={<WalletOutlined />} /></Card>
              </Col>
              <Col xs={12} sm={6}>
                <Card><Statistic title={t('expense.stats.maintenanceCost')} value={expenseStats?.category_breakdown?.find((c) => c.category === 'maintenance')?.total_amount || 0} prefix={<DollarOutlined />} precision={2} valueStyle={{ fontSize: isMobile ? 16 : 20 }} /></Card>
              </Col>
              <Col xs={12} sm={6}>
                <Card><Statistic title={t('expense.stats.last30Days')} value={expenseStats?.last_30_days_amount || 0} prefix={<DollarOutlined />} precision={2} suffix={expenseStats?.last_30_days_currency || currency} valueStyle={{ fontSize: isMobile ? 16 : 20 }} /></Card>
              </Col>
            </Row>
            <Row gutter={[16, 16]}>
              <Col xs={24} lg={12}>
                <Card title={period === 'month' ? t('stats.monthlyExpense') : t('stats.yearlyExpense')}>
                  {expCurrent.length > 0 ? (
                    <ReactECharts option={buildExpenseChart('total_amount', currency, 'bar', '#ff7a45')} style={{ height: chartH }} theme={isDark ? 'dark' : undefined} />
                  ) : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />}
                </Card>
              </Col>
              <Col xs={24} lg={12}>
                <Card title={period === 'month' ? t('stats.monthlyExpenseRecords') : t('stats.yearlyExpenseRecords')}>
                  {expCurrent.length > 0 ? (
                    <ReactECharts option={buildExpenseChart('total_records', '', 'bar', '#eb2f96')} style={{ height: chartH }} theme={isDark ? 'dark' : undefined} />
                  ) : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />}
                </Card>
              </Col>
              <Col xs={24} lg={12}>
                <Card title={t('stats.categoryBreakdown')}>
                  {(expenseStats?.category_breakdown?.length || 0) > 0 ? (
                    <ReactECharts option={buildCategoryPieChart()} style={{ height: chartH }} theme={isDark ? 'dark' : undefined} />
                  ) : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />}
                </Card>
              </Col>
            </Row>
          </>
        )}

        {/* ========== Tab: Combined ========== */}
        {activeTab === 'combined' && (
          <>
            <Row gutter={isMobile ? [8, 8] : [16, 16]} style={{ marginBottom: 24 }}>
              <Col xs={24} sm={8}>
                <Card><Statistic title={t('stats.fuelCost')} value={stats ? formatCurrency(fuelTotalCost, currency) : '-'} prefix={<span>⛽</span>} /></Card>
              </Col>
              <Col xs={24} sm={8}>
                <Card><Statistic title={t('stats.expenseCost')} value={expenseStats ? formatCurrency(expenseTotalCost, currency) : '-'} prefix={<span>💸</span>} /></Card>
              </Col>
              <Col xs={24} sm={8}>
                <Tooltip title={ratesData?.rates ? (() => {
                  const ref = convertAmount(combinedCost, currency, refCurrency, ratesData.rates);
                  return ref != null ? t('exchangeRate.referenceAmount', { amount: formatCurrency(ref, refCurrency) }) : undefined;
                })() : undefined}>
                  <Card><Statistic title={t('stats.combinedCost')} value={formatCurrency(combinedCost, currency)} prefix={<DollarOutlined />} valueStyle={{ color: token.colorPrimary, fontWeight: 600 }} /></Card>
                </Tooltip>
              </Col>
            </Row>
            <Row gutter={[16, 16]}>
              <Col xs={24}>
                <Card title={period === 'month' ? t('stats.monthlyCombined') : t('stats.yearlyCombined')}>
                  {fuelCurrent.length > 0 ? (
                    <ReactECharts option={buildCombinedCostChart()} style={{ height: chartH }} theme={isDark ? 'dark' : undefined} />
                  ) : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />}
                </Card>
              </Col>
            </Row>
          </>
        )}
      </Spin>
    </div>
  );
}
