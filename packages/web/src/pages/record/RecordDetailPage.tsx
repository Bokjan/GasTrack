import { useEffect, useState, useMemo } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Card,
  Button,
  Space,
  Spin,
  message,
  Popconfirm,
  Tag,
  Row,
  Col,
  Divider,
  Descriptions,
  Typography,
  Rate,
  Tooltip,
} from 'antd';
import {
  ArrowLeftOutlined,
  EditOutlined,
  DeleteOutlined,
  DollarOutlined,
  ThunderboltOutlined,
  FireOutlined,
  TrophyOutlined,
  WalletOutlined,
  DashboardOutlined,
  CarOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import {
  fuelRecordApi,
  vehicleApi,
  statsApi,
  formatCurrency,
  formatNumber,
  formatDateTime,
  convertFuelEfficiency,
  FUEL_EFFICIENCY_UNITS,
  FUEL_GRADES,
  isElectricVehicle,
  useAuthStore,
} from '@gastrack/shared';
import type { FuelRecord, Vehicle, VehicleStats } from '@gastrack/shared';
import { useIsMobile } from '../../hooks/useIsMobile';

const { Title, Text } = Typography;

/** 油耗评级：返回 1~5 星及对应 key */
function getEfficiencyRating(
  current: number,
  avg: number,
  isLowerBetter: boolean,
): { stars: number; key: string } {
  if (!avg || !current) return { stars: 3, key: 'normal' };
  const diff = isLowerBetter
    ? (avg - current) / avg          // L/100km: 低于均值为正
    : (current - avg) / avg;          // km/L, MPG: 高于均值为正
  if (diff > 0.1) return { stars: 5, key: 'excellent' };
  if (diff > 0.05) return { stars: 4, key: 'good' };
  if (diff > -0.05) return { stars: 3, key: 'normal' };
  if (diff > -0.15) return { stars: 2, key: 'belowAvg' };
  return { stars: 1, key: 'poor' };
}

export default function RecordDetailPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { vehicleId, recordId } = useParams<{
    vehicleId: string;
    recordId: string;
  }>();
  const user = useAuthStore((s) => s.user);
  const isMobile = useIsMobile();

  const [record, setRecord] = useState<FuelRecord | null>(null);
  const [vehicle, setVehicle] = useState<Vehicle | null>(null);
  const [stats, setStats] = useState<VehicleStats | null>(null);
  const [loading, setLoading] = useState(true);

  const currency = user?.currency_code || 'CNY';
  const isImperial = user?.unit_system === 'imperial';
  const fuelUnit = isImperial ? 'gal' : 'L';
  const distanceUnit = isImperial ? 'mi' : 'km';
  const efficiencyUnit = user?.fuel_efficiency_unit || 'L/100km';
  const userTimezone = user?.timezone;

  useEffect(() => {
    if (vehicleId && recordId) {
      loadData();
    }
  }, [vehicleId, recordId]);

  const loadData = async () => {
    setLoading(true);
    try {
      const [recordRes, vehicleRes, statsRes] = await Promise.all([
        fuelRecordApi.getById(vehicleId!, recordId!),
        vehicleApi.getById(vehicleId!),
        statsApi.vehicleStats(vehicleId!),
      ]);
      setRecord(recordRes.data.data);
      setVehicle(vehicleRes.data.data);
      setStats(statsRes.data.data);
    } catch {
      message.error(t('common.error'));
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async () => {
    try {
      await fuelRecordApi.delete(vehicleId!, recordId!);
      message.success(t('common.success'));
      navigate(`/vehicles/${vehicleId}/records`);
    } catch {
      message.error(t('common.error'));
    }
  };

  const isEv = vehicle ? isElectricVehicle(vehicle.fuel_type) : false;

  /** 判断当前油耗单位是否为"越低越好" */
  const isLowerBetter = efficiencyUnit === 'L/100km' || efficiencyUnit === 'kWh/100km';

  // ── 智能分析数据 ────────────────────────
  const insights = useMemo(() => {
    if (!record || !stats) return null;

    const efficiency = record.fuel_efficiency;
    const avgEfficiency = stats.avg_efficiency;
    const tripDistance = record.trip_distance;
    const avgDistance = stats.total_distance && stats.total_records
      ? stats.total_distance / stats.total_records : 0;
    const totalCost = record.total_cost;
    const avgCostPerFill = stats.avg_cost_per_fill;
    const unitPrice = record.unit_price;
    const avgCostPerKm = stats.avg_cost_per_km;

    // 油耗评级
    const rating = efficiency && avgEfficiency
      ? getEfficiencyRating(efficiency, avgEfficiency, isLowerBetter)
      : null;

    // 单价对比（通过 total_cost / fuel_amount 粗略计算平均单价）
    const avgUnitPrice = stats.total_fuel && stats.total_cost
      ? stats.total_cost / stats.total_fuel : 0;

    // 单次每公里成本
    const costPerKm = tripDistance && totalCost ? totalCost / tripDistance : 0;

    // 油箱利用率
    const tankCapacity = vehicle?.tank_capacity || 0;
    const tankUsage = tankCapacity > 0 ? record.fuel_amount / tankCapacity : 0;

    return {
      efficiency,
      avgEfficiency,
      rating,
      unitPrice,
      avgUnitPrice,
      totalCost,
      avgCostPerFill,
      costPerKm,
      avgCostPerKm,
      tripDistance,
      avgDistance,
      tankUsage,
      tankCapacity,
    };
  }, [record, stats, vehicle, isLowerBetter]);

  /** 偏差百分比文字 */
  const diffText = (current: number, avg: number) => {
    if (!avg) return '';
    const pct = ((current - avg) / avg) * 100;
    const abs = Math.abs(pct).toFixed(1);
    if (pct > 0) return `+${abs}%`;
    if (pct < 0) return `-${abs}%`;
    return '0%';
  };

  const diffColor = (current: number, avg: number, lowerIsBetter = false) => {
    if (!avg) return undefined;
    const pct = (current - avg) / avg;
    if (lowerIsBetter) return pct < -0.02 ? '#52c41a' : pct > 0.02 ? '#ff4d4f' : undefined;
    return pct > 0.02 ? '#52c41a' : pct < -0.02 ? '#ff4d4f' : undefined;
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
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate(`/vehicles/${vehicleId}/records`)}>
          {t('common.back')}
        </Button>
      </div>
    );
  }

  return (
    <div className="page-container">
      {/* ── 页头 ── */}
      <div className="page-header">
        <Space>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate(`/vehicles/${vehicleId}/records`)}
          />
          <h2>{isMobile ? t('recordDetail.title') : `${isEv ? t('fuelRecord.titleEv') : t('fuelRecord.title')} — ${t('recordDetail.title')}`}</h2>
        </Space>
        <Space>
          <Button
            icon={<EditOutlined />}
            onClick={() => navigate(`/vehicles/${vehicleId}/records/${recordId}/edit`)}
          >
            {isMobile ? '' : t('common.edit')}
          </Button>
          <Popconfirm
            title={t('fuelRecord.deleteConfirm')}
            onConfirm={handleDelete}
          >
            <Button danger icon={<DeleteOutlined />}>
              {isMobile ? '' : t('common.delete')}
            </Button>
          </Popconfirm>
        </Space>
      </div>

      <Row gutter={[16, 16]}>
        {/* ── ① 基本信息卡片 ── */}
        <Col xs={24} lg={12}>
          <Card
            title={
              <Space>
                {isEv ? <ThunderboltOutlined /> : <FireOutlined />}
                <span>{t('recordDetail.basicInfo')}</span>
              </Space>
            }
          >
            <Descriptions column={1} labelStyle={{ fontWeight: 500, width: isMobile ? 90 : 140 }} colon={false} size={isMobile ? 'small' : 'middle'}>
              <Descriptions.Item label={isEv ? t('fuelRecord.chargingDate') : t('fuelRecord.fuelDate')}>
                {formatDateTime(record.refuel_date, userTimezone, 'YYYY-MM-DD HH:mm')}
              </Descriptions.Item>

              {record.station_name && (
                <Descriptions.Item label={isEv ? t('fuelRecord.chargingStation') : t('fuelRecord.station')}>
                  {record.station_name}
                </Descriptions.Item>
              )}

              <Descriptions.Item label={isEv ? t('fuelRecord.chargingAmount') : t('fuelRecord.fuelAmount')}>
                <Text strong>{formatNumber(record.fuel_amount)} {record.fuel_unit || fuelUnit}</Text>
              </Descriptions.Item>

              {record.unit_price != null && (
                <Descriptions.Item label={t('fuelRecord.pricePerUnit')}>
                  {formatCurrency(record.unit_price, record.currency_code || currency)}/{record.fuel_unit || fuelUnit}
                </Descriptions.Item>
              )}

              <Descriptions.Item label={t('fuelRecord.totalCost')}>
                <Text strong style={{ fontSize: 18, color: 'var(--ant-color-primary)' }}>
                  {formatCurrency(record.total_cost, record.currency_code || currency)}
                </Text>
              </Descriptions.Item>

              <Descriptions.Item label={t('fuelRecord.odometer')}>
                {formatNumber(record.odometer, 0)} {record.distance_unit || distanceUnit}
              </Descriptions.Item>

              {record.trip_distance != null && record.trip_distance > 0 && (
                <Descriptions.Item label={t('fuelRecord.tripDistance')}>
                  {formatNumber(record.trip_distance, 1)} {record.distance_unit || distanceUnit}
                </Descriptions.Item>
              )}

              {record.fuel_efficiency != null && record.fuel_efficiency > 0 && (
                <Descriptions.Item label={isEv ? t('fuelRecord.energyConsumption') : t('fuelRecord.consumption')}>
                  <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}>
                    <Tag color="blue" style={{ margin: 0 }}>
                      {formatNumber(record.fuel_efficiency)} {efficiencyUnit}
                    </Tag>
                    {/* 其他单位换算 */}
                    {!isEv && FUEL_EFFICIENCY_UNITS
                      .filter((u) => u !== efficiencyUnit)
                      .map((u) => (
                        <Tag key={u} style={{ margin: 0 }}>
                          {formatNumber(convertFuelEfficiency(record.fuel_efficiency!, efficiencyUnit, u))} {u}
                        </Tag>
                      ))}
                  </div>
                </Descriptions.Item>
              )}

              <Descriptions.Item label={isEv ? t('fuelRecord.isFullCharge') : t('fuelRecord.isFullTank')}>
                {record.is_full_tank
                  ? <Tag color="green">✓ {t('common.yes')}</Tag>
                  : <Tag>✗ {t('common.no')}</Tag>}
              </Descriptions.Item>

              {record.fuel_grade && (
                <Descriptions.Item label={t('fuelRecord.fuelGrade')}>
                  <Tag>{t(FUEL_GRADES.find((g) => g.value === record.fuel_grade)?.label || record.fuel_grade)}</Tag>
                </Descriptions.Item>
              )}

              {record.note && (
                <Descriptions.Item label={t('fuelRecord.notes')}>
                  {record.note}
                </Descriptions.Item>
              )}
            </Descriptions>
          </Card>
        </Col>

        {/* ── ② 智能分析卡片 ── */}
        <Col xs={24} lg={12}>
          <Card
            title={
              <Space>
                <TrophyOutlined />
                <span>{t('recordDetail.insights')}</span>
              </Space>
            }
          >
            {insights && stats && stats.total_records > 1 ? (
              <Space direction="vertical" size="middle" style={{ width: '100%' }}>
                {/* 油耗/电耗评级 */}
                {insights.rating && (
                  <div>
                    <Title level={5} style={{ marginBottom: 4 }}>
                      {isEv ? t('recordDetail.energyRating') : t('recordDetail.efficiencyRating')}
                    </Title>
                    <Space align="center" wrap>
                      <Rate disabled value={insights.rating.stars} />
                      <Tag color={insights.rating.stars >= 4 ? 'green' : insights.rating.stars >= 3 ? 'blue' : 'orange'}>
                        {t(`recordDetail.rating.${insights.rating.key}`)}
                      </Tag>
                    </Space>
                    <div style={{ marginTop: 4, wordBreak: 'break-word' }}>
                      <Text type="secondary" style={{ fontSize: isMobile ? 12 : 14 }}>
                        {t('recordDetail.currentValue')}: {formatNumber(insights.efficiency!)} {efficiencyUnit}
                        {isMobile ? <br /> : ' · '}
                        {t('recordDetail.avgValue')}: {formatNumber(insights.avgEfficiency)} {efficiencyUnit}
                        {' '}
                        <Text style={{ color: diffColor(insights.efficiency!, insights.avgEfficiency, isLowerBetter) }}>
                          ({diffText(insights.efficiency!, insights.avgEfficiency)})
                        </Text>
                      </Text>
                    </div>
                  </div>
                )}

                <Divider style={{ margin: '4px 0' }} />

                {/* 单价对比 */}
                {insights.unitPrice != null && insights.avgUnitPrice > 0 && (
                  <div>
                    <Space>
                      <DollarOutlined />
                      <Text strong>{t('recordDetail.priceComparison')}</Text>
                    </Space>
                    <div style={{ marginTop: 4, wordBreak: 'break-word' }}>
                      <Text style={{ fontSize: isMobile ? 12 : 14 }}>
                        {t('recordDetail.thisTime')}: {formatCurrency(insights.unitPrice, currency)}/{fuelUnit}
                        {isMobile ? <br /> : ' · '}
                        {t('recordDetail.historical')}: {formatCurrency(insights.avgUnitPrice, currency)}/{fuelUnit}
                        {' '}
                        <Text style={{ color: diffColor(insights.unitPrice, insights.avgUnitPrice, true) }}>
                          ({diffText(insights.unitPrice, insights.avgUnitPrice)})
                        </Text>
                      </Text>
                    </div>
                  </div>
                )}

                {/* 单次花费对比 */}
                {insights.avgCostPerFill > 0 && (
                  <div>
                    <Space>
                      <WalletOutlined />
                      <Text strong>{t('recordDetail.costComparison')}</Text>
                    </Space>
                    <div style={{ marginTop: 4, wordBreak: 'break-word' }}>
                      <Text style={{ fontSize: isMobile ? 12 : 14 }}>
                        {t('recordDetail.thisTime')}: {formatCurrency(insights.totalCost, currency)}
                        {isMobile ? <br /> : ' · '}
                        {t('recordDetail.avgPerFill')}: {formatCurrency(insights.avgCostPerFill, currency)}
                        {' '}
                        <Text style={{ color: diffColor(insights.totalCost, insights.avgCostPerFill, true) }}>
                          ({diffText(insights.totalCost, insights.avgCostPerFill)})
                        </Text>
                      </Text>
                    </div>
                  </div>
                )}

                {/* 每公里/英里成本 */}
                {insights.costPerKm > 0 && insights.avgCostPerKm > 0 && (
                  <div>
                    <Space>
                      <DashboardOutlined />
                      <Text strong>{t('recordDetail.costPerDistance', { unit: distanceUnit })}</Text>
                    </Space>
                    <div style={{ marginTop: 4, wordBreak: 'break-word' }}>
                      <Text style={{ fontSize: isMobile ? 12 : 14 }}>
                        {t('recordDetail.thisTime')}: {formatCurrency(insights.costPerKm, currency)}/{distanceUnit}
                        {isMobile ? <br /> : ' · '}
                        {t('recordDetail.historical')}: {formatCurrency(insights.avgCostPerKm, currency)}/{distanceUnit}
                        {' '}
                        <Text style={{ color: diffColor(insights.costPerKm, insights.avgCostPerKm, true) }}>
                          ({diffText(insights.costPerKm, insights.avgCostPerKm)})
                        </Text>
                      </Text>
                    </div>
                  </div>
                )}

                {/* 行驶里程对比 */}
                {insights.tripDistance != null && insights.tripDistance > 0 && insights.avgDistance > 0 && (
                  <div>
                    <Space>
                      <CarOutlined />
                      <Text strong>{t('recordDetail.distanceComparison')}</Text>
                    </Space>
                    <div style={{ marginTop: 4, wordBreak: 'break-word' }}>
                      <Text style={{ fontSize: isMobile ? 12 : 14 }}>
                        {t('recordDetail.thisTrip')}: {formatNumber(insights.tripDistance, 1)} {distanceUnit}
                        {isMobile ? <br /> : ' · '}
                        {t('recordDetail.avgTrip')}: {formatNumber(insights.avgDistance, 1)} {distanceUnit}
                        {' '}
                        <Text style={{ color: diffColor(insights.tripDistance, insights.avgDistance) }}>
                          ({diffText(insights.tripDistance, insights.avgDistance)})
                        </Text>
                      </Text>
                    </div>
                  </div>
                )}

                {/* 油箱/电池利用率 */}
                {insights.tankUsage > 0 && insights.tankCapacity > 0 && (
                  <div>
                    <Space>
                      {isEv ? <ThunderboltOutlined /> : <FireOutlined />}
                      <Text strong>{isEv ? t('recordDetail.batteryUsage') : t('recordDetail.tankUsage')}</Text>
                    </Space>
                    <div style={{ marginTop: 4, wordBreak: 'break-word' }}>
                      <Tooltip title={`${formatNumber(record.fuel_amount)} / ${formatNumber(insights.tankCapacity)} ${fuelUnit}`}>
                        <Text style={{ fontSize: isMobile ? 12 : 14 }}>
                          {(insights.tankUsage * 100).toFixed(0)}%
                          {' '}
                          <Text type="secondary" style={{ fontSize: isMobile ? 12 : 14 }}>
                            ({formatNumber(record.fuel_amount)} / {formatNumber(insights.tankCapacity)} {record.fuel_unit || fuelUnit})
                          </Text>
                        </Text>
                      </Tooltip>
                    </div>
                  </div>
                )}
              </Space>
            ) : (
              <Text type="secondary">{t('recordDetail.noInsights')}</Text>
            )}
          </Card>
        </Col>
      </Row>
    </div>
  );
}
