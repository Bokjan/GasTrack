import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Row, Col, Card, Statistic, Button, Empty, Space, List, Tag } from 'antd';
import {
  CarOutlined,
  FileTextOutlined,
  DashboardOutlined,
  DollarOutlined,
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
} from '@gastrack/shared';
import type { OverviewStats } from '@gastrack/shared';

export default function DashboardPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { vehicles, fetchVehicles, isLoading: vehiclesLoading } = useVehicleStore();
  const user = useAuthStore((s) => s.user);
  const [overview, setOverview] = useState<OverviewStats | null>(null);
  const [statsLoading, setStatsLoading] = useState(false);

  useEffect(() => {
    fetchVehicles();
    loadOverview();
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

  return (
    <div className="page-container">
      <div className="page-header">
        <h2>{t('nav.dashboard')}</h2>
      </div>

      {/* 统计卡片 */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={12} sm={6}>
          <Card loading={statsLoading}>
            <Statistic
              title={t('stats.totalRecords')}
              value={overview?.total_records || 0}
              prefix={<FileTextOutlined />}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card loading={statsLoading}>
            <Statistic
              title={t('stats.totalCost')}
              value={overview ? formatCurrency(overview.total_cost, currency) : '-'}
              prefix={<DollarOutlined />}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card loading={statsLoading}>
            <Statistic
              title={t('stats.totalDistance')}
              value={overview ? `${formatNumber(overview.total_distance, 0)} km` : '-'}
              prefix={<DashboardOutlined />}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card loading={statsLoading}>
            <Statistic
              title={t('stats.avgConsumption')}
              value={overview ? `${formatNumber(overview.avg_consumption)} L/100km` : '-'}
              prefix={<DashboardOutlined />}
            />
          </Card>
        </Col>
      </Row>

      {/* 我的车辆 */}
      <Card
        title={
          <Space>
            <CarOutlined />
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
                      {vehicle.is_default && <Tag color="blue">默认</Tag>}
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
