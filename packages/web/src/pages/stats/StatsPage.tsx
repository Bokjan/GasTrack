import { useEffect, useState } from 'react';
import { Row, Col, Card, Statistic, Select, Empty, Spin } from 'antd';
import {
  DashboardOutlined,
  DollarOutlined,
  FileTextOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import ReactECharts from 'echarts-for-react';
import {
  useVehicleStore,
  statsApi,
  formatNumber,
  formatCurrency,
  useAuthStore,
} from '@gastrack/shared';
import type { VehicleStats, ConsumptionTrend } from '@gastrack/shared';

export default function StatsPage() {
  const { t } = useTranslation();
  const { vehicles, fetchVehicles } = useVehicleStore();
  const user = useAuthStore((s) => s.user);

  const [selectedVehicleId, setSelectedVehicleId] = useState<string>('');
  const [stats, setStats] = useState<VehicleStats | null>(null);
  const [trend, setTrend] = useState<ConsumptionTrend[]>([]);
  const [efficiencyUnit, setEfficiencyUnit] = useState<string>('L/100km');
  const [loading, setLoading] = useState(false);

  const currency = user?.currency_code || 'CNY';

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
  }, [selectedVehicleId]);

  const loadStats = async () => {
    setLoading(true);
    try {
      const [statsRes, trendRes] = await Promise.all([
        statsApi.vehicleStats(selectedVehicleId),
        statsApi.efficiencyTrend(selectedVehicleId, { limit: 30 }),
      ]);
      setStats(statsRes.data.data);
      const trendData = trendRes.data.data;
      setTrend(trendData.items || []);
      setEfficiencyUnit(trendData.efficiency_unit || 'L/100km');
    } catch {
      // 可能没有数据
    } finally {
      setLoading(false);
    }
  };

  // 油耗趋势图配置
  const consumptionChartOption = {
    tooltip: { trigger: 'axis' as const },
    xAxis: {
      type: 'category' as const,
      data: trend.map((item) => item.date),
    },
    yAxis: { type: 'value' as const, name: efficiencyUnit },
    series: [
      {
        name: t('stats.avgConsumption'),
        type: 'line',
        data: trend.map((item) => item.fuel_efficiency),
        smooth: true,
        areaStyle: { opacity: 0.15 },
        itemStyle: { color: '#1677ff' },
      },
    ],
    grid: { left: 50, right: 20, top: 40, bottom: 30 },
  };

  // 里程趋势图配置
  const distanceChartOption = {
    tooltip: { trigger: 'axis' as const },
    xAxis: {
      type: 'category' as const,
      data: trend.map((item) => item.date),
    },
    yAxis: { type: 'value' as const, name: 'km' },
    series: [
      {
        name: t('stats.totalDistance'),
        type: 'bar',
        data: trend.map((item) => item.trip_distance),
        itemStyle: { color: '#52c41a', borderRadius: [4, 4, 0, 0] },
      },
    ],
    grid: { left: 60, right: 20, top: 40, bottom: 30 },
  };

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
        <h2>{t('stats.title')}</h2>
        <Select
          value={selectedVehicleId}
          onChange={setSelectedVehicleId}
          style={{ width: 200 }}
          options={vehicles.map((v) => ({
            value: v.id,
            label: v.name,
          }))}
        />
      </div>

      <Spin spinning={loading}>
        {/* 统计卡片 */}
        <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
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
                  stats ? `${formatNumber(stats.avg_efficiency)} ${stats.fuel_efficiency_unit || 'L/100km'}` : '-'
                }
                prefix={<DashboardOutlined />}
              />
            </Card>
          </Col>
          <Col xs={12} sm={6}>
            <Card>
              <Statistic
                title={t('stats.totalFuel')}
                value={stats ? `${formatNumber(stats.total_fuel)} L` : '-'}
                prefix={<ThunderboltOutlined />}
              />
            </Card>
          </Col>
        </Row>

        {/* 油耗趋势 */}
        <Row gutter={[16, 16]}>
          <Col xs={24} lg={12}>
            <Card title={t('stats.consumptionTrend')}>
              {trend.length > 0 ? (
                <ReactECharts option={consumptionChartOption} style={{ height: 300 }} />
              ) : (
                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />
              )}
            </Card>
          </Col>
          <Col xs={24} lg={12}>
            <Card title={t('stats.distanceTrend')}>
              {trend.length > 0 ? (
                <ReactECharts option={distanceChartOption} style={{ height: 300 }} />
              ) : (
                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />
              )}
            </Card>
          </Col>
        </Row>
      </Spin>
    </div>
  );
}
