import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Row, Col, Card, Statistic, Button, Empty, Space, List, Tag, Divider, Tooltip } from 'antd';
import {
  PlusOutlined,
  RightOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import {
  useVehicleStore,
  statsApi,
  formatCurrency,
  formatNumber,
  useAuthStore,
  useExchangeRateStore,
  isElectricVehicle,
  convertAmount,
  getReferenceCurrency,
  sumConvertedCostsByCurrency,
} from '@gastrack/shared';
import type { OverviewStats, VehicleStats, Vehicle } from '@gastrack/shared';
import { useIsMobile } from '../../hooks/useIsMobile';

export default function DashboardPage() {
  const navigate = useNavigate();
  const { vehicles, fetchVehicles, isLoading: vehiclesLoading } = useVehicleStore();
  const user = useAuthStore((s) => s.user);
  const isMobile = useIsMobile();
  const [overview, setOverview] = useState<OverviewStats | null>(null);
  const [statsLoading, setStatsLoading] = useState(false);
  const { data: ratesData, fetchRates } = useExchangeRateStore();
  const { t } = useTranslation();

  useEffect(() => {
    fetchVehicles();
    loadOverview();
    if (user?.currency_code) {
      fetchRates(user.currency_code);
    }
  }, []);

  const loadOverview = async () => {
    setStatsLoading(true);
    try {
      const { data } = await statsApi.overview();
      setOverview(data.data);
    } catch {
      // 忽略错误，可能没有数据
    } finally {
      setStatsLoading(false);
    }
  };

  const currency = user?.currency_code || 'CNY';
  const isImperial = user?.unit_system === 'imperial';
  const distanceUnit = isImperial ? 'mi' : 'km';
  const efficiencyUnit = user?.fuel_efficiency_unit || 'L/100km';

  /** 使用 costs_by_currency 按汇率换算后的总费用（用户偏好币种） */
  const getConvertedCost = (costsByCurrency?: Record<string, number>, fallbackCost?: number): number => {
    const converted = sumConvertedCostsByCurrency(costsByCurrency, currency, ratesData?.rates);
    if (converted != null) return converted;
    // 如果无法换算（缺汇率或无分组数据），回退到原始值
    return fallbackCost ?? 0;
  };

  /** 全局概览的换算后总费用 */
  const overviewConvertedCost = overview
    ? getConvertedCost(overview.costs_by_currency, overview.total_cost)
    : 0;

  /** 全局概览的换算后开销总额 */
  const overviewExpenseCost = overview
    ? getConvertedCost(overview.expense_costs_by_currency, overview.total_expense_cost)
    : 0;

  const overviewCombinedCost = overviewConvertedCost + overviewExpenseCost;

  /** 汇率参考换算：综合总费用 → 另一种参考币种 */
  const refCurrency = getReferenceCurrency(currency, user?.reference_currency);
  const refCombinedCost = overviewCombinedCost && ratesData?.rates
    ? convertAmount(overviewCombinedCost, currency, refCurrency, ratesData.rates)
    : null;

  /** 根据 vehicle_id 找到对应车辆信息 */
  const findVehicle = (vehicleId: string): Vehicle | undefined =>
    vehicles.find((v) => v.id === vehicleId);

  /** 渲染单辆车的统计卡片 */
  const renderVehicleStats = (vs: VehicleStats, vehicle?: Vehicle) => {
    const isEv = vehicle ? isElectricVehicle(vehicle.fuel_type) : false;
    const vehicleConvertedCost = getConvertedCost(vs.costs_by_currency, vs.total_cost);
    const vehicleExpenseCost = getConvertedCost(vs.expense_costs_by_currency, vs.total_expense_cost);
    const vehicleCombinedCost = vehicleConvertedCost + vehicleExpenseCost;
    return (
      <Row gutter={isMobile ? [8, 8] : [16, 16]}>
        <Col xs={12} sm={6} lg={4}>
          <Card loading={statsLoading} size="small">
            <Statistic
              title={t('stats.totalRecords')}
              value={vs.total_records || 0}
              prefix={<span>📋</span>}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6} lg={4}>
          <Card loading={statsLoading} size="small">
            <Statistic
              title={t('stats.fuelCost')}
              value={formatCurrency(vehicleConvertedCost, currency)}
              prefix={<span>⛽</span>}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6} lg={4}>
          <Card loading={statsLoading} size="small">
            <Statistic
              title={t('stats.expenseCost')}
              value={formatCurrency(vehicleExpenseCost, currency)}
              prefix={<span>💸</span>}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6} lg={4}>
          <Card loading={statsLoading} size="small">
            <Statistic
              title={t('stats.combinedCost')}
              value={formatCurrency(vehicleCombinedCost, currency)}
              prefix={<span>💰</span>}
              valueStyle={{ color: 'var(--ant-color-primary)' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6} lg={4}>
          <Card loading={statsLoading} size="small">
            <Statistic
              title={t('stats.totalDistance')}
              value={`${formatNumber(vs.total_distance, 0)} ${distanceUnit}`}
              prefix={<span>🛣️</span>}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6} lg={4}>
          <Card loading={statsLoading} size="small">
            <Statistic
              title={isEv ? t('stats.avgEnergyConsumption') : t('stats.avgConsumption')}
              value={vs.avg_efficiency ? `${formatNumber(vs.avg_efficiency)} ${vs.fuel_efficiency_unit || efficiencyUnit}` : '-'}
              prefix={isEv ? <span>⚡</span> : <span>📊</span>}
            />
          </Card>
        </Col>
      </Row>
    );
  };

  const hasMultipleVehicles = vehicles.length > 1;
  const vehicleStatsList = overview?.vehicles || [];

  return (
    <div className="page-container">
      <div className="page-header">
        <h2>{t('nav.dashboard')}</h2>
      </div>

      {/* 统计卡片 — 按车辆维度展示 */}
      {vehicleStatsList.length > 0 && (
        <div style={{ marginBottom: 24 }}>
          {hasMultipleVehicles ? (
            // 多辆车：按车辆分组展示独立统计
            <>
              {/* 全局概览 */}
              <Row gutter={isMobile ? [8, 8] : [16, 16]} style={{ marginBottom: 16 }}>
                <Col xs={12} sm={6}>
                  <Card loading={statsLoading} size="small">
                    <Statistic
                      title={t('stats.totalRecords')}
                      value={overview?.total_records || 0}
                      prefix={<span>📋</span>}
                    />
                  </Card>
                </Col>
                <Col xs={12} sm={6}>
                  <Card loading={statsLoading} size="small">
                    <Statistic
                      title={t('stats.fuelCost')}
                      value={overview ? formatCurrency(overviewConvertedCost, currency) : '-'}
                      prefix={<span>⛽</span>}
                    />
                  </Card>
                </Col>
                <Col xs={12} sm={6}>
                  <Card loading={statsLoading} size="small">
                    <Statistic
                      title={t('stats.expenseCost')}
                      value={overview ? formatCurrency(overviewExpenseCost, currency) : '-'}
                      prefix={<span>💸</span>}
                    />
                  </Card>
                </Col>
                <Col xs={12} sm={6}>
                  <Tooltip
                    title={refCombinedCost != null ? t('exchangeRate.referenceAmount', { amount: formatCurrency(refCombinedCost, refCurrency) }) : undefined}
                  >
                    <Card loading={statsLoading} size="small">
                      <Statistic
                        title={t('stats.combinedCost')}
                        value={overview ? formatCurrency(overviewCombinedCost, currency) : '-'}
                        prefix={<span>💰</span>}
                        valueStyle={{ color: 'var(--ant-color-primary)' }}
                      />
                    </Card>
                  </Tooltip>
                </Col>
              </Row>

              <Divider orientation="left" style={{ margin: '8px 0 16px' }}>
                {t('stats.perVehicle')}
              </Divider>

              {vehicleStatsList.map((vs) => {
                const vehicle = findVehicle(vs.vehicle_id);
                return (
                  <div key={vs.vehicle_id} style={{ marginBottom: 16 }}>
                    <div
                      style={{
                        marginBottom: 8,
                        fontWeight: 500,
                        fontSize: 14,
                        display: 'flex',
                        alignItems: 'center',
                        gap: 8,
                        cursor: 'pointer',
                      }}
                      onClick={() => navigate(`/vehicles/${vs.vehicle_id}/records`)}
                    >
                      <span>
                        {vehicle?.fuel_type === 'electric' ? '⚡' :
                         vehicle?.vehicle_type === 'motorcycle' ? '🏍️' : '🚗'}
                      </span>
                      <span>{vs.vehicle_name}</span>
                      {vehicle && (
                        <Tag color={vehicle.fuel_type === 'electric' ? 'green' : 'blue'} style={{ marginLeft: 4 }}>
                          {t(`fuelType.${vehicle.fuel_type}`)}
                        </Tag>
                      )}
                      <RightOutlined style={{ fontSize: 12, color: 'var(--gt-text-tertiary)' }} />
                    </div>
                    {renderVehicleStats(vs, vehicle)}
                  </div>
                );
              })}
            </>
          ) : (
            // 单辆车：直接展示完整统计（不需要标题）
            renderVehicleStats(vehicleStatsList[0], findVehicle(vehicleStatsList[0].vehicle_id))
          )}
        </div>
      )}

      {/* 我的车辆 */}
      <Card
        title={
          <Space>
            <span>🚗</span>
            <span>{t('nav.vehicles')}</span>
          </Space>
        }
        extra={
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => navigate('/vehicles/new')}
          >
            {t('vehicle.addVehicle')}
          </Button>
        }
        loading={vehiclesLoading}
      >
        {vehicles.length === 0 ? (
          <Empty
            description={t('vehicle.noVehicle')}
            image={Empty.PRESENTED_IMAGE_SIMPLE}
          >
            <Button type="primary" onClick={() => navigate('/vehicles/new')}>
              {t('vehicle.addFirst')}
            </Button>
          </Empty>
        ) : (
          <List
            dataSource={vehicles}
            renderItem={(vehicle) => (
              <List.Item
                style={{ cursor: 'pointer' }}
                onClick={() => navigate(`/vehicles/${vehicle.id}/records`)}
                actions={[<RightOutlined key="go" />]}
              >
                <List.Item.Meta
                  avatar={
                    <div style={{ fontSize: 32 }}>
                      {vehicle.vehicle_type === 'motorcycle' ? '🏍️' : '🚗'}
                    </div>
                  }
                  title={
                    <Space>
                      <span>{vehicle.name}</span>
                      {vehicle.is_default && <Tag color="blue">{t('vehicle.default')}</Tag>}
                    </Space>
                  }
                  description={`${vehicle.brand} ${vehicle.model} · ${vehicle.year} · ${t(`fuelType.${vehicle.fuel_type}`)}`}
                />
              </List.Item>
            )}
          />
        )}
      </Card>
    </div>
  );
}
