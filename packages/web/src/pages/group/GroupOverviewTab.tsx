import { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button, Card, Col, Empty, Row, Space, Statistic, Switch, Table, Tag, Typography, message } from 'antd';
import { InfoCircleOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { formatNumber, getApiErrorMessage, groupApi, isElectricVehicle } from '@gastrack/shared';
import type { GroupOverviewResponse, GroupVehicleSummary, SharedVehicleResponse } from '@gastrack/shared';
import { useIsMobile } from '../../hooks/useIsMobile';
import { useLatestRequest } from '../../hooks/useLatestRequest';
import { useGroupUnits } from './useGroupUnits';
import Money from './Money';

const { Text } = Typography;

/** 群组数据汇总：成员车辆统计与共享开关 */
export default function GroupOverviewTab({ groupId }: { groupId: string }) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const isMobile = useIsMobile();
  const { user, currency, fuelDisplay, effDisplay } = useGroupUnits();

  const [overview, setOverview] = useState<GroupOverviewResponse | null>(null);
  const [overviewLoading, setOverviewLoading] = useState(true);
  const [sharedVehicles, setSharedVehicles] = useState<SharedVehicleResponse[]>([]);
  const [sharingVehicleId, setSharingVehicleId] = useState<string | null>(null);
  const begin = useLatestRequest();

  const load = useCallback(async () => {
    const isLatest = begin();
    setOverviewLoading(true);
    const [overviewRes, sharedRes] = await Promise.allSettled([
      groupApi.getOverview(groupId),
      groupApi.listSharedVehicles(groupId),
    ]);
    if (!isLatest()) return;
    setOverview(overviewRes.status === 'fulfilled' ? overviewRes.value.data.data : null);
    setSharedVehicles(sharedRes.status === 'fulfilled' ? sharedRes.value.data.data || [] : []);
    setOverviewLoading(false);
  }, [groupId, begin]);

  useEffect(() => {
    load();
  }, [load]);

  // 共享/取消共享车辆
  const handleShareToggle = async (vehicleId: string, isCurrentlyShared: boolean) => {
    setSharingVehicleId(vehicleId);
    try {
      if (isCurrentlyShared) {
        await groupApi.unshareVehicle(groupId, vehicleId);
        message.success(t('group.unshareSuccess'));
        setSharedVehicles(prev => prev.filter(sv => sv.vehicle_id !== vehicleId));
      } else {
        const { data } = await groupApi.shareVehicle(groupId, { vehicle_id: vehicleId });
        message.success(t('group.shareSuccess'));
        setSharedVehicles(prev => [...prev, data.data]);
      }
    } catch (err) {
      message.error(getApiErrorMessage(err));
    } finally {
      setSharingVehicleId(null);
    }
  };

  return (
    <div>
      {overviewLoading ? (
        <Card loading />
      ) : overview && overview.vehicles.length > 0 ? (
        <>
          <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
            <Col xs={12} sm={6}>
              <Statistic
                title={t('group.members')}
                value={overview.member_count}
              />
            </Col>
            <Col xs={12} sm={6}>
              <Statistic
                title={t('nav.vehicles')}
                value={overview.vehicle_count}
              />
            </Col>
            <Col xs={24}>
              <Text type="secondary" style={{ fontSize: 12 }}>
                <InfoCircleOutlined style={{ marginRight: 4 }} />
                {t('group.statsSharedOnlyHint')}
              </Text>
            </Col>
          </Row>
          <Table<GroupVehicleSummary>
            rowKey="vehicle_id"
            size="small"
            scroll={isMobile ? { x: 700 } : undefined}
            pagination={false}
            dataSource={overview.vehicles}
            columns={[
              {
                title: t('nav.vehicles'),
                dataIndex: 'vehicle_name',
                key: 'vehicle_name',
              },
              {
                title: t('group.vehicleOwner'),
                dataIndex: 'owner_name',
                key: 'owner_name',
              },
              {
                title: t('group.totalRecords'),
                dataIndex: 'total_records',
                key: 'total_records',
                width: 100,
              },
              {
                title: t('group.totalCost'),
                dataIndex: 'total_cost',
                key: 'total_cost',
                width: 120,
                render: (val: number, record: { vehicle_id: string; owner_id: string; currency_code?: string }) => <Money amount={val} currency={record.currency_code || currency} />,
              },
              {
                title: t('group.totalFuel'),
                dataIndex: 'total_fuel',
                key: 'total_fuel',
                width: 120,
                render: (val: number, record: GroupVehicleSummary) => {
                  const d = fuelDisplay(val, isElectricVehicle(record.fuel_type));
                  return `${formatNumber(d.value, 2)} ${d.unit}`;
                },
              },
              {
                title: t('group.avgEfficiency'),
                dataIndex: 'avg_efficiency',
                key: 'avg_efficiency',
                width: 130,
                render: (val: number, record: GroupVehicleSummary) => {
                  const d = effDisplay(val, isElectricVehicle(record.fuel_type));
                  return `${formatNumber(d.value, 2)} ${d.unit}`;
                },
              },
              {
                title: t('group.shareVehicle'),
                key: 'shared',
                width: 100,
                render: (_: unknown, record: GroupVehicleSummary) => {
                  const isShared = sharedVehicles.some(sv => sv.vehicle_id === record.vehicle_id);
                  const isMyVehicle = record.owner_id === user?.id;
                  if (!isMyVehicle) {
                    return isShared ? (
                      <Tag color="green">{t('group.shared')}</Tag>
                    ) : (
                      <Tag>{t('group.notShared')}</Tag>
                    );
                  }
                  return (
                    <Switch
                      size="small"
                      checked={isShared}
                      loading={sharingVehicleId === record.vehicle_id}
                      onChange={() => handleShareToggle(record.vehicle_id, isShared)}
                    />
                  );
                },
              },
              {
                title: '',
                key: 'actions',
                width: 200,
                render: (_: unknown, record: GroupVehicleSummary) => {
                  const isShared = sharedVehicles.some(sv => sv.vehicle_id === record.vehicle_id);
                  const isMyVehicle = record.owner_id === user?.id;
                  // 只有自己的车辆或已共享的车辆才显示操作按钮
                  if (!isMyVehicle && !isShared) return null;
                  return (
                    <Space size="small">
                      <Button
                        type="link"
                        size="small"
                        onClick={() => navigate(`/vehicles/${record.vehicle_id}/records`)}
                      >
                        {t('group.viewRecords')}
                      </Button>
                      {isShared && !isMyVehicle && (
                        <Button
                          type="link"
                          size="small"
                          onClick={() => navigate(`/records/new?vehicleId=${record.vehicle_id}`)}
                        >
                          {t('group.addRecordForShared')}
                        </Button>
                      )}
                    </Space>
                  );
                },
              },
            ]}
          />
        </>
      ) : (
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description={t('group.noVehicles')}
        />
      )}
    </div>
  );
}
