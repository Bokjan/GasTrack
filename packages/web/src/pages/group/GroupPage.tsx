import { useState, useEffect, useCallback, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Card,
  Button,
  Space,
  Tag,
  Modal,
  Form,
  Input,
  InputNumber,
  message,
  Popconfirm,
  Typography,
  Tooltip,
  Empty,
  Tabs,
  Table,
  Select,
  Statistic,
  Row,
  Col,
  Descriptions,
  Switch,
  List,
  Badge,
  Segmented,
  Avatar,
} from 'antd';
import {
  PlusOutlined,
  CopyOutlined,
  DeleteOutlined,
  TeamOutlined,
  UserAddOutlined,
  EditOutlined,
  LogoutOutlined,
  ReloadOutlined,
  CrownOutlined,
  UserOutlined,
  TrophyOutlined,
  DollarOutlined,
  CaretUpOutlined,
  CaretDownOutlined,
  EnvironmentOutlined,
  ArrowUpOutlined,
  ArrowDownOutlined,
  MinusOutlined,
  InfoCircleOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { groupApi } from '@gastrack/shared/src/api';
import type {
  Group,
  CreateGroupRequest,
  UpdateGroupRequest,
  GroupMemberDetail,
  GroupOverviewResponse,
  SharedVehicleResponse,
  LeaderboardResponse,
  LeaderboardMetric,
  LeaderboardEnergy,
  GroupVehicleSummary,
  LeaderboardPeriod,
  GroupExpenseStatsResponse,
  GroupStationStatsResponse,
} from '@gastrack/shared';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import { useIsMobile } from '../../hooks/useIsMobile';
import { useLatestRequest } from '../../hooks/useLatestRequest';
import { useAuthStore, FUEL_GRADES, formatCurrency, formatNumber, convertEfficiencyForDisplay, litersToGallons, kmToMiles, isElectricVehicle, getApiErrorMessage } from '@gastrack/shared';
import { useCurrency } from '../../hooks/useCurrency';
import UnconvertedNotice from '../../components/UnconvertedNotice';

const { Text } = Typography;

/** 金额展示；外币金额（如加油站单价）附带换算为用户币种的参考值 */
function Money({ amount, currency, reference }: { amount: number; currency: string; reference?: string }) {
  return (
    <span>
      {formatCurrency(amount, currency)}
      {reference && (
        <Tooltip title={reference}>
          <InfoCircleOutlined style={{ marginLeft: 4, fontSize: 12, color: 'var(--gt-text-tertiary)', cursor: 'help' }} />
        </Tooltip>
      )}
    </span>
  );
}

export default function GroupPage() {
  const { t } = useTranslation();
  const { user } = useAuthStore();
  const isMobile = useIsMobile();
  const navigate = useNavigate();

  // --- 用户偏好派生 ---
  const currency = user?.currency_code || 'CNY';
  const isImperial = user?.unit_system === 'imperial';
  const efficiencyUnit = user?.fuel_efficiency_unit || 'L/100km';
  const fuelUnit = isImperial ? 'gal' : 'L';
  const distanceUnit = isImperial ? 'mi' : 'km';

  // --- 汇率：群组金额已由后端换算为查看者币种，仅加油站单价保留原币种 ---
  const { referenceText } = useCurrency();

  /** 将后端的加注量（燃油为 L，电动车为 kWh）按用户偏好转换 */
  const fuelDisplay = (amount: number, isEv = false): { value: number; unit: string } =>
    isEv ? { value: amount, unit: 'kWh' } : { value: isImperial ? litersToGallons(amount) : amount, unit: fuelUnit };

  /** 将后端的油量 (L) 按用户偏好转换 */
  const convertFuel = (liters: number): number => fuelDisplay(liters).value;

  /** 将后端的距离 (km) 按用户偏好转换 */
  const convertDistance = (km: number): number => {
    return isImperial ? kmToMiles(km) : km;
  };

  /** 将后端基准油耗（L/100km）或电耗（kWh/100km）按用户偏好转换 */
  const effDisplay = (base: number, isEv = false): { value: number; unit: string } =>
    convertEfficiencyForDisplay(base, efficiencyUnit, isEv);

  /** 将后端的油耗 (L/100km) 按用户偏好转换 */
  const convertEfficiency = (l100km: number): number => effDisplay(l100km).value;

  /** 加油站单价：后端统一为每升（或每 kWh），英制显示为每加仑 */
  const stationPrice = (price: number, priceUnit?: string): { value: number; unit: string } => {
    if (priceUnit === 'kWh') return { value: price, unit: 'kWh' };
    return isImperial ? { value: price * 3.78541, unit: 'gal' } : { value: price, unit: 'L' };
  };

  const [loading, setLoading] = useState(false);
  const [groups, setGroups] = useState<Group[]>([]);
  const [selectedGroup, setSelectedGroup] = useState<Group | null>(null);
  const [overview, setOverview] = useState<GroupOverviewResponse | null>(null);
  const [overviewLoading, setOverviewLoading] = useState(false);

  // Shared vehicles
  const [sharedVehicles, setSharedVehicles] = useState<SharedVehicleResponse[]>([]);
  const [sharingVehicleId, setSharingVehicleId] = useState<string | null>(null);

  // Leaderboard
  const [leaderboard, setLeaderboard] = useState<LeaderboardResponse | null>(null);
  const [leaderboardLoading, setLeaderboardLoading] = useState(false);
  const [lbMetric, setLbMetric] = useState<LeaderboardMetric>('efficiency');
  // 油耗排行只在同一能源类别内比较
  const [lbEnergy, setLbEnergy] = useState<LeaderboardEnergy>('fuel');
  const [lbPeriod, setLbPeriod] = useState<LeaderboardPeriod>('current_month');

  // Expense Stats
  const [expenseStats, setExpenseStats] = useState<GroupExpenseStatsResponse | null>(null);
  const [expenseLoading, setExpenseLoading] = useState(false);
  const [expensePeriod, setExpensePeriod] = useState<'month' | 'year'>('month');
  const [expenseYear, setExpenseYear] = useState(new Date().getFullYear());

  // Station Stats
  const [stationStats, setStationStats] = useState<GroupStationStatsResponse | null>(null);
  const [stationLoading, setStationLoading] = useState(false);
  const [stationFuelGrade, setStationFuelGrade] = useState('');
  const [stationMonths, setStationMonths] = useState(6);
  const [stationSortBy, setStationSortBy] = useState('frequency');

  // Modal states
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [joinModalOpen, setJoinModalOpen] = useState(false);
  const [editModalOpen, setEditModalOpen] = useState(false);
  const [creating, setCreating] = useState(false);
  const [joining, setJoining] = useState(false);
  const [updating, setUpdating] = useState(false);

  const [createForm] = Form.useForm();
  const [joinForm] = Form.useForm();
  const [editForm] = Form.useForm();

  const fetchGroups = useCallback(async () => {
    setLoading(true);
    try {
      const { data } = await groupApi.list();
      setGroups(data.data || []);
    } catch (err) {
      message.error(getApiErrorMessage(err));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchGroups();
  }, [fetchGroups]);

  // 选中群组时并行获取详情、数据汇总与共享车辆；快速切换群组时丢弃旧群组的响应
  const activeGroupIdRef = useRef<string | null>(null);
  const selectGroup = useCallback(async (group: Group) => {
    activeGroupIdRef.current = group.id;
    const isStale = () => activeGroupIdRef.current !== group.id;
    setOverviewLoading(true);

    const [detailRes, overviewRes, sharedRes] = await Promise.allSettled([
      groupApi.getById(group.id),
      groupApi.getOverview(group.id),
      groupApi.listSharedVehicles(group.id),
    ]);
    if (isStale()) return;

    if (detailRes.status === 'fulfilled') {
      setSelectedGroup(detailRes.value.data.data);
    } else {
      message.error(getApiErrorMessage(detailRes.reason));
    }
    setOverview(overviewRes.status === 'fulfilled' ? overviewRes.value.data.data : null);
    setSharedVehicles(sharedRes.status === 'fulfilled' ? sharedRes.value.data.data || [] : []);
    setOverviewLoading(false);
  }, []);

  const beginLeaderboard = useLatestRequest();
  const beginExpense = useLatestRequest();
  const beginStation = useLatestRequest();

  // 获取排行榜数据
  const fetchLeaderboard = useCallback(async (groupId: string, metric: LeaderboardMetric, period: LeaderboardPeriod, energy: LeaderboardEnergy) => {
    const isLatest = beginLeaderboard();
    setLeaderboardLoading(true);
    try {
      const { data } = await groupApi.getLeaderboard(groupId, { metric, period, energy });
      if (isLatest()) setLeaderboard(data.data);
    } catch {
      if (isLatest()) setLeaderboard(null);
    } finally {
      if (isLatest()) setLeaderboardLoading(false);
    }
  }, [beginLeaderboard]);

  // 获取费用统计数据
  const fetchExpenseStats = useCallback(async (groupId: string, period: 'month' | 'year', year: number) => {
    const isLatest = beginExpense();
    setExpenseLoading(true);
    try {
      const { data } = await groupApi.getExpenseStats(groupId, { period, year: period === 'month' ? year : undefined });
      if (isLatest()) setExpenseStats(data.data);
    } catch {
      if (isLatest()) setExpenseStats(null);
    } finally {
      if (isLatest()) setExpenseLoading(false);
    }
  }, [beginExpense]);

  // 获取加油站数据
  const fetchStationStats = useCallback(async (groupId: string, fuelGrade: string, months: number, sortBy: string) => {
    const isLatest = beginStation();
    setStationLoading(true);
    try {
      const { data } = await groupApi.getStationStats(groupId, {
        fuel_grade: fuelGrade || undefined,
        months,
        sort_by: sortBy,
      });
      if (isLatest()) setStationStats(data.data);
    } catch {
      if (isLatest()) setStationStats(null);
    } finally {
      if (isLatest()) setStationLoading(false);
    }
  }, [beginStation]);

  // 共享/取消共享车辆
  const handleShareToggle = async (vehicleId: string, isCurrentlyShared: boolean) => {
    if (!selectedGroup) return;
    setSharingVehicleId(vehicleId);
    try {
      if (isCurrentlyShared) {
        await groupApi.unshareVehicle(selectedGroup.id, vehicleId);
        message.success(t('group.unshareSuccess'));
        setSharedVehicles(prev => prev.filter(sv => sv.vehicle_id !== vehicleId));
      } else {
        const { data } = await groupApi.shareVehicle(selectedGroup.id, { vehicle_id: vehicleId });
        message.success(t('group.shareSuccess'));
        setSharedVehicles(prev => [...prev, data.data]);
      }
    } catch (err) {
      message.error(getApiErrorMessage(err));
    } finally {
      setSharingVehicleId(null);
    }
  };

  const handleCreate = async (values: CreateGroupRequest) => {
    setCreating(true);
    try {
      await groupApi.create(values);
      message.success(t('common.success'));
      setCreateModalOpen(false);
      createForm.resetFields();
      fetchGroups();
    } catch (err) {
      message.error(getApiErrorMessage(err));
    } finally {
      setCreating(false);
    }
  };

  const handleJoin = async (values: { invite_code: string }) => {
    setJoining(true);
    try {
      const { data } = await groupApi.join(values.invite_code);
      message.success(t('group.joinSuccess', { name: data.data.group_name }));
      setJoinModalOpen(false);
      joinForm.resetFields();
      fetchGroups();
    } catch (err) {
      message.error(getApiErrorMessage(err));
    } finally {
      setJoining(false);
    }
  };

  const handleUpdate = async (values: UpdateGroupRequest) => {
    if (!selectedGroup) return;
    setUpdating(true);
    try {
      const { data } = await groupApi.update(selectedGroup.id, values);
      message.success(t('group.updateSuccess'));
      setEditModalOpen(false);
      editForm.resetFields();
      setSelectedGroup(data.data);
      fetchGroups();
    } catch (err) {
      message.error(getApiErrorMessage(err));
    } finally {
      setUpdating(false);
    }
  };

  const handleDelete = async () => {
    if (!selectedGroup) return;
    try {
      await groupApi.delete(selectedGroup.id);
      message.success(t('group.deleteSuccess'));
      setSelectedGroup(null);
      setOverview(null);
      fetchGroups();
    } catch (err) {
      message.error(getApiErrorMessage(err));
    }
  };

  const handleLeave = async () => {
    if (!selectedGroup) return;
    try {
      await groupApi.leave(selectedGroup.id);
      message.success(t('group.leaveSuccess'));
      setSelectedGroup(null);
      setOverview(null);
      fetchGroups();
    } catch (err) {
      message.error(getApiErrorMessage(err));
    }
  };

  const handleCopyInviteCode = async (code: string) => {
    try {
      await navigator.clipboard.writeText(code);
      message.success(t('group.copied'));
    } catch (err) {
      message.error(getApiErrorMessage(err));
    }
  };

  const handleRegenerateCode = async () => {
    if (!selectedGroup) return;
    try {
      const { data } = await groupApi.regenerateInviteCode(selectedGroup.id);
      message.success(t('group.regenerateSuccess'));
      setSelectedGroup(data.data);
      fetchGroups();
    } catch (err) {
      message.error(getApiErrorMessage(err));
    }
  };

  const handleChangeRole = async (userId: string, role: 'admin' | 'member') => {
    if (!selectedGroup) return;
    try {
      await groupApi.updateMemberRole(selectedGroup.id, userId, { role });
      message.success(t('group.changeRoleSuccess'));
      selectGroup(selectedGroup);
    } catch (err) {
      message.error(getApiErrorMessage(err));
    }
  };

  const handleRemoveMember = async (userId: string) => {
    if (!selectedGroup) return;
    try {
      await groupApi.removeMember(selectedGroup.id, userId);
      message.success(t('group.removeMemberSuccess'));
      selectGroup(selectedGroup);
      fetchGroups();
    } catch (err) {
      message.error(getApiErrorMessage(err));
    }
  };

  const getRoleTag = (role: string) => {
    switch (role) {
      case 'owner':
        return <Tag color="gold" icon={<CrownOutlined />}>{t('group.roleOwner')}</Tag>;
      case 'admin':
        return <Tag color="blue">{t('group.roleAdmin')}</Tag>;
      default:
        return <Tag>{t('group.roleMember')}</Tag>;
    }
  };

  const isOwnerOrAdmin = selectedGroup?.my_role === 'owner' || selectedGroup?.my_role === 'admin';
  const isOwner = selectedGroup?.my_role === 'owner';

  // 成员列表 columns
  const memberColumns: ColumnsType<GroupMemberDetail> = [
    {
      title: t('group.members'),
      key: 'user',
      render: (_, record) => (
        <Space>
          <UserOutlined />
          <Text strong>{record.nickname || record.email}</Text>
          {record.email && record.nickname && (
            <Text type="secondary" style={{ fontSize: 12 }}>({record.email})</Text>
          )}
        </Space>
      ),
    },
    {
      title: t('group.role'),
      key: 'role',
      width: 120,
      render: (_, record) => getRoleTag(record.role),
    },
    {
      title: t('group.joinedAt'),
      dataIndex: 'joined_at',
      key: 'joined_at',
      width: 160,
      render: (val: string) => dayjs(val).format('YYYY-MM-DD HH:mm'),
    },
  ];

  // Owner 可以管理角色和移除成员
  if (isOwner) {
    memberColumns.push({
      title: '',
      key: 'actions',
      width: 180,
      render: (_, record) => {
        if (record.user_id === user?.id) return null;
        return (
          <Space size="small">
            <Select
              size="small"
              value={record.role === 'owner' ? undefined : record.role}
              disabled={record.role === 'owner'}
              style={{ width: 100 }}
              onChange={(val) => handleChangeRole(record.user_id, val)}
              options={[
                { label: t('group.roleAdmin'), value: 'admin' },
                { label: t('group.roleMember'), value: 'member' },
              ]}
            />
            <Popconfirm
              title={t('group.removeMemberConfirm')}
              onConfirm={() => handleRemoveMember(record.user_id)}
              okText={t('common.confirm')}
              cancelText={t('common.cancel')}
            >
              <Button type="text" size="small" danger icon={<DeleteOutlined />} />
            </Popconfirm>
          </Space>
        );
      },
    });
  } else if (selectedGroup?.my_role === 'admin') {
    // Admin 可以移除普通成员
    memberColumns.push({
      title: '',
      key: 'actions',
      width: 80,
      render: (_, record) => {
        if (record.user_id === user?.id || record.role === 'owner' || record.role === 'admin') return null;
        return (
          <Popconfirm
            title={t('group.removeMemberConfirm')}
            onConfirm={() => handleRemoveMember(record.user_id)}
            okText={t('common.confirm')}
            cancelText={t('common.cancel')}
          >
            <Button type="text" size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        );
      },
    });
  }

  // --- 群组列表卡片 ---
  const renderGroupCards = () => {
    if (groups.length === 0) {
      return (
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description={t('group.noGroups')}
        />
      );
    }

    return (
      <Row gutter={[16, 16]}>
        {groups.map((group) => (
          <Col key={group.id} xs={24} sm={12} lg={8}>
            <Card
              hoverable
              onClick={() => selectGroup(group)}
              style={{
                borderColor: selectedGroup?.id === group.id ? '#1677ff' : undefined,
              }}
            >
              <Card.Meta
                avatar={<TeamOutlined style={{ fontSize: 24, color: '#1677ff' }} />}
                title={group.name}
                description={
                  <Space direction="vertical" size={4}>
                    <Text type="secondary">
                      {t('group.memberCount', { count: group.member_count })}
                    </Text>
                    {group.my_role && getRoleTag(group.my_role)}
                  </Space>
                }
              />
            </Card>
          </Col>
        ))}
      </Row>
    );
  };

  // --- 群组详情面板 ---
  const renderGroupDetail = () => {
    if (!selectedGroup) return null;

    return (
      <Card style={{ marginTop: 16 }}>
        <Tabs
          defaultActiveKey="info"
          items={[
            {
              key: 'info',
              label: t('group.groupInfo'),
              children: (
                <div>
                  <Descriptions
                    column={isMobile ? 1 : 2}
                    bordered
                    size="small"
                  >
                    <Descriptions.Item label={t('group.name')}>
                      {selectedGroup.name}
                    </Descriptions.Item>
                    <Descriptions.Item label={t('group.owner')}>
                      {selectedGroup.owner_name || selectedGroup.owner_id}
                    </Descriptions.Item>
                    <Descriptions.Item label={t('group.inviteCode')}>
                      <Space>
                        <Text code>{selectedGroup.invite_code}</Text>
                        <Tooltip title={t('group.copyInviteCode')}>
                          <Button
                            type="text"
                            size="small"
                            icon={<CopyOutlined />}
                            onClick={() => handleCopyInviteCode(selectedGroup.invite_code)}
                          />
                        </Tooltip>
                        {isOwnerOrAdmin && (
                          <Popconfirm
                            title={t('group.regenerateConfirm')}
                            onConfirm={handleRegenerateCode}
                            okText={t('common.confirm')}
                            cancelText={t('common.cancel')}
                          >
                            <Tooltip title={t('group.regenerateCode')}>
                              <Button type="text" size="small" icon={<ReloadOutlined />} />
                            </Tooltip>
                          </Popconfirm>
                        )}
                      </Space>
                    </Descriptions.Item>
                    <Descriptions.Item label={t('group.maxMembers')}>
                      {selectedGroup.max_members}
                    </Descriptions.Item>
                    <Descriptions.Item label={t('group.members')}>
                      {selectedGroup.member_count}
                    </Descriptions.Item>
                    <Descriptions.Item label={t('group.createdAt')}>
                      {dayjs(selectedGroup.created_at).format('YYYY-MM-DD HH:mm')}
                    </Descriptions.Item>
                    {selectedGroup.description && (
                      <Descriptions.Item label={t('group.description')} span={2}>
                        {selectedGroup.description}
                      </Descriptions.Item>
                    )}
                  </Descriptions>

                  <Space style={{ marginTop: 16 }}>
                    {isOwnerOrAdmin && (
                      <Button
                        icon={<EditOutlined />}
                        onClick={() => {
                          editForm.setFieldsValue({
                            name: selectedGroup.name,
                            max_members: selectedGroup.max_members,
                            description: selectedGroup.description || '',
                          });
                          setEditModalOpen(true);
                        }}
                      >
                        {t('group.editGroup')}
                      </Button>
                    )}
                    {isOwner ? (
                      <Popconfirm
                        title={t('group.deleteConfirm')}
                        onConfirm={handleDelete}
                        okText={t('common.confirm')}
                        cancelText={t('common.cancel')}
                      >
                        <Button danger icon={<DeleteOutlined />}>
                          {t('group.deleteGroup')}
                        </Button>
                      </Popconfirm>
                    ) : (
                      <Popconfirm
                        title={t('group.leaveConfirm')}
                        onConfirm={handleLeave}
                        okText={t('common.confirm')}
                        cancelText={t('common.cancel')}
                      >
                        <Button danger icon={<LogoutOutlined />}>
                          {t('group.leaveGroup')}
                        </Button>
                      </Popconfirm>
                    )}
                  </Space>
                </div>
              ),
            },
            {
              key: 'members',
              label: `${t('group.memberManagement')} (${selectedGroup.member_count})`,
              children: (
                <Table
                  rowKey="user_id"
                  columns={memberColumns}
                  dataSource={selectedGroup.members || []}
                  pagination={false}
                  size="small"
                  scroll={isMobile ? { x: 500 } : undefined}
                />
              ),
            },
            {
              key: 'overview',
              label: t('group.overview'),
              children: (
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
              ),
            },
            // --- 排行榜 Tab ---
            {
              key: 'leaderboard',
              label: (
                <span>
                  <TrophyOutlined style={{ marginRight: 4 }} />
                  {t('group.leaderboard')}
                </span>
              ),
              children: (
                <div>
                  <Space wrap style={{ marginBottom: 16 }}>
                    <Segmented
                      value={lbMetric}
                      onChange={(val) => {
                        const m = val as LeaderboardMetric;
                        setLbMetric(m);
                        fetchLeaderboard(selectedGroup.id, m, lbPeriod, lbEnergy);
                      }}
                      options={[
                        { label: t('group.metricEfficiency'), value: 'efficiency' },
                        { label: t('group.metricCost'), value: 'cost' },
                        { label: t('group.metricDistance'), value: 'distance' },
                        { label: t('group.metricFrequency'), value: 'frequency' },
                      ]}
                    />
                    {lbMetric === 'efficiency' && (
                      <Segmented
                        value={lbEnergy}
                        onChange={(val) => {
                          const e = val as LeaderboardEnergy;
                          setLbEnergy(e);
                          fetchLeaderboard(selectedGroup.id, lbMetric, lbPeriod, e);
                        }}
                        options={[
                          { label: t('group.energyFuel'), value: 'fuel' },
                          { label: t('group.energyElectric'), value: 'electric' },
                        ]}
                      />
                    )}
                    <Select
                      value={lbPeriod}
                      onChange={(val) => {
                        setLbPeriod(val);
                        fetchLeaderboard(selectedGroup.id, lbMetric, val, lbEnergy);
                      }}
                      style={{ width: 120 }}
                      options={[
                        { label: t('group.periodCurrentMonth'), value: 'current_month' },
                        { label: t('group.periodLastMonth'), value: 'last_month' },
                        { label: t('group.periodLast3Months'), value: 'last_3_months' },
                        { label: t('group.periodCurrentYear'), value: 'current_year' },
                      ]}
                    />
                  </Space>

                  {leaderboard && leaderboard.rankings.length > 0 && (
                    <div style={{ marginBottom: 12 }}>
                      <Text type="secondary">
                        {t('group.groupAvg')}:{' '}
                        {lbMetric === 'cost' ? (
                          <Text strong><Money amount={leaderboard.group_avg} currency={currency} /></Text>
                        ) : (
                          <>
                            <Text strong>
                              {lbMetric === 'efficiency'
                                ? formatNumber(effDisplay(leaderboard.group_avg, leaderboard.unit === 'kWh/100km').value, 2)
                                : lbMetric === 'distance'
                                ? formatNumber(convertDistance(leaderboard.group_avg), 0)
                                : formatNumber(leaderboard.group_avg, 2)}
                            </Text>{' '}
                            {lbMetric === 'efficiency'
                              ? effDisplay(0, leaderboard.unit === 'kWh/100km').unit
                              : lbMetric === 'distance'
                              ? distanceUnit
                              : leaderboard.unit}
                          </>
                        )}
                      </Text>
                    </div>
                  )}

                  {leaderboardLoading ? (
                    <Card loading />
                  ) : leaderboard && leaderboard.rankings.length > 0 ? (
                    <List
                      dataSource={leaderboard.rankings}
                      renderItem={(item) => {
                        const medals = ['🥇', '🥈', '🥉'];
                        const medal = item.rank <= 3 ? medals[item.rank - 1] : undefined;
                        return (
                          <List.Item
                            style={{
                              background: item.is_self ? 'rgba(22, 119, 255, 0.06)' : undefined,
                              borderRadius: 8,
                              padding: '12px 16px',
                              marginBottom: 4,
                            }}
                          >
                            <List.Item.Meta
                              avatar={
                                medal ? (
                                  <span style={{ fontSize: 24 }}>{medal}</span>
                                ) : (
                                  <Avatar size="small" style={{ backgroundColor: '#f0f0f0', color: '#999' }}>
                                    {item.rank}
                                  </Avatar>
                                )
                              }
                              title={
                                <Space>
                                  <Text strong={item.is_self}>
                                    {item.nickname}
                                  </Text>
                                  <Text type="secondary" style={{ fontSize: 12 }}>
                                    {item.vehicle_name}
                                  </Text>
                                  {item.is_self && <Badge count="✦" style={{ backgroundColor: '#1677ff' }} />}
                                </Space>
                              }
                              description={
                                <Space size="large">
                                  <Text>
                                    {lbMetric === 'cost' ? (
                                      <Money amount={item.value} currency={currency} />
                                    ) : (
                                      <>
                                        {lbMetric === 'efficiency'
                                          ? formatNumber(effDisplay(item.value, leaderboard.unit === 'kWh/100km').value, 2)
                                          : lbMetric === 'distance'
                                          ? formatNumber(convertDistance(item.value), 0)
                                          : formatNumber(item.value, 2)}{' '}
                                        {lbMetric === 'efficiency'
                                          ? effDisplay(0, leaderboard.unit === 'kWh/100km').unit
                                          : lbMetric === 'distance'
                                          ? distanceUnit
                                          : leaderboard.unit}
                                      </>
                                    )}
                                  </Text>
                                  <Text type="secondary" style={{ fontSize: 12 }}>
                                    {t('group.recordCount')}: {item.record_count}
                                  </Text>
                                  <Text
                                    type={item.diff_from_avg < 0 ? 'success' : item.diff_from_avg > 0 ? 'danger' : 'secondary'}
                                    style={{ fontSize: 12 }}
                                  >
                                    {t('group.diffFromAvg')}: {item.diff_from_avg > 0 ? '+' : ''}{formatNumber(item.diff_from_avg, 1)}%
                                  </Text>
                                </Space>
                              }
                            />
                          </List.Item>
                        );
                      }}
                    />
                  ) : (
                    <Empty
                      image={Empty.PRESENTED_IMAGE_SIMPLE}
                      description={t('group.noLeaderboardData')}
                    />
                  )}
                </div>
              ),
            },
            // --- 费用看板 Tab ---
            {
              key: 'expense',
              label: (
                <span>
                  <DollarOutlined style={{ marginRight: 4 }} />
                  {t('group.expenseStats')}
                </span>
              ),
              children: (
                <div>
                  <Space wrap style={{ marginBottom: 16 }}>
                    <Segmented
                      value={expensePeriod}
                      onChange={(val) => {
                        const p = val as 'month' | 'year';
                        setExpensePeriod(p);
                        fetchExpenseStats(selectedGroup.id, p, expenseYear);
                      }}
                      options={[
                        { label: t('group.byMonth'), value: 'month' },
                        { label: t('group.byYear'), value: 'year' },
                      ]}
                    />
                    {expensePeriod === 'month' && (
                      <Select
                        value={expenseYear}
                        onChange={(val) => {
                          setExpenseYear(val);
                          fetchExpenseStats(selectedGroup.id, expensePeriod, val);
                        }}
                        style={{ width: 100 }}
                        options={Array.from({ length: 5 }, (_, i) => {
                          const y = new Date().getFullYear() - i;
                          return { label: `${y}`, value: y };
                        })}
                      />
                    )}
                  </Space>

                  {expenseLoading ? (
                    <Card loading />
                  ) : expenseStats ? (
                    <>
                      <UnconvertedNotice currencies={[expenseStats.unconverted_currencies]} />
                      {/* 统计卡片 */}
                      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
                        {[
                          {
                            title: t('group.totalExpense'),
                            value: { raw: 0, displayText: formatCurrency(expenseStats.summary.total_cost, expenseStats.currency_code || currency) },
                            change: expenseStats.summary.cost_change_pct,
                            precision: 2,
                            isCost: true,
                          },
                          {
                            title: t('group.totalFuelAmount'),
                            value: { raw: convertFuel(expenseStats.summary.total_fuel), displayText: '' },
                            change: expenseStats.summary.fuel_change_pct,
                            precision: 2,
                            suffix: fuelUnit,
                            isCost: false,
                          },
                          {
                            title: t('group.totalMileage'),
                            value: { raw: convertDistance(expenseStats.summary.total_distance), displayText: '' },
                            change: expenseStats.summary.distance_change_pct,
                            precision: 0,
                            suffix: distanceUnit,
                            isCost: false,
                          },
                          {
                            title: t('group.avgFuelEfficiency'),
                            value: { raw: convertEfficiency(expenseStats.summary.avg_efficiency), displayText: '' },
                            change: expenseStats.summary.efficiency_change_pct,
                            precision: 2,
                            suffix: efficiencyUnit,
                            isCost: false,
                          },
                          // 有电动车数据时单独展示充电量与电耗（与燃油分开统计）
                          ...((expenseStats.summary.total_energy ?? 0) > 0
                            ? [
                                {
                                  title: t('group.totalEnergyAmount'),
                                  value: { raw: expenseStats.summary.total_energy ?? 0, displayText: '' },
                                  change: 0,
                                  precision: 2,
                                  suffix: 'kWh',
                                  isCost: false,
                                },
                                {
                                  title: t('group.avgEnergyEfficiency'),
                                  value: { raw: effDisplay(expenseStats.summary.avg_energy_efficiency ?? 0, true).value, displayText: '' },
                                  change: 0,
                                  precision: 2,
                                  suffix: effDisplay(0, true).unit,
                                  isCost: false,
                                },
                              ]
                            : []),
                        ].map((card, idx) => (
                          <Col xs={12} sm={6} key={idx}>
                            <Card size="small">
                              {card.isCost ? (
                                <Statistic
                                  title={card.title}
                                  formatter={() => (
                                    <span>
                                      {card.value.displayText}
                                      {' '}
                                      {card.change !== 0 && (
                                        <Text
                                          style={{ fontSize: 12 }}
                                          type={card.change > 0 ? 'danger' : 'success'}
                                        >
                                          {card.change > 0 ? <CaretUpOutlined /> : <CaretDownOutlined />}
                                          {formatNumber(Math.abs(card.change), 1)}%
                                        </Text>
                                      )}
                                    </span>
                                  )}
                                />
                              ) : (
                                <Statistic
                                  title={card.title}
                                  value={card.value.raw}
                                  precision={card.precision}
                                  suffix={
                                    <span>
                                      {card.suffix}{' '}
                                      {card.change !== 0 && (
                                        <Text
                                          style={{ fontSize: 12 }}
                                          type={card.change > 0 ? 'danger' : 'success'}
                                        >
                                          {card.change > 0 ? <CaretUpOutlined /> : <CaretDownOutlined />}
                                          {formatNumber(Math.abs(card.change), 1)}%
                                        </Text>
                                      )}
                                    </span>
                                  }
                                />
                              )}
                            </Card>
                          </Col>
                        ))}
                      </Row>

                      {/* 费用趋势表格 */}
                      {expenseStats.trend_items.length > 0 && (
                        <>
                          <Text strong style={{ display: 'block', marginBottom: 12 }}>
                            {t('group.costTrend')}
                          </Text>
                          <Table
                            rowKey="period_label"
                            size="small"
                            pagination={false}
                            scroll={isMobile ? { x: 500 } : undefined}
                            dataSource={expenseStats.trend_items}
                            columns={[
                              {
                                title: expensePeriod === 'month' ? t('stats.byMonth') : t('stats.byYear'),
                                dataIndex: 'period_label',
                                key: 'period_label',
                              },
                              {
                                title: t('group.totalExpense'),
                                dataIndex: 'total_cost',
                                key: 'total_cost',
                                render: (val: number) => formatCurrency(val, expenseStats.currency_code || currency),
                              },
                              {
                                title: t('group.totalFuelAmount'),
                                dataIndex: 'total_fuel',
                                key: 'total_fuel',
                                render: (val: number) => `${formatNumber(convertFuel(val), 2)} ${fuelUnit}`,
                              },
                              {
                                title: t('group.totalMileage'),
                                dataIndex: 'total_distance',
                                key: 'total_distance',
                                render: (val: number) => `${formatNumber(convertDistance(val), 0)} ${distanceUnit}`,
                              },
                              {
                                title: t('group.avgFuelEfficiency'),
                                dataIndex: 'avg_efficiency',
                                key: 'avg_efficiency',
                                render: (val: number) => `${formatNumber(convertEfficiency(val), 2)} ${efficiencyUnit}`,
                              },
                            ]}
                          />
                        </>
                      )}

                      {/* 成员费用占比 */}
                      {expenseStats.member_breakdown.length > 0 && (
                        <>
                          <Text strong style={{ display: 'block', marginTop: 24, marginBottom: 12 }}>
                            {t('group.costBreakdown')}
                          </Text>
                            <List
                            dataSource={expenseStats.member_breakdown}
                            renderItem={(item) => (
                              <List.Item>
                                <List.Item.Meta
                                  title={item.nickname}
                                  description={
                                    <span>
                                      <Money amount={item.total_cost} currency={item.currency_code || currency} /> · {formatNumber(convertFuel(item.total_fuel), 2)} {fuelUnit}
                                    </span>
                                  }
                                />
                                <div style={{ minWidth: 60, textAlign: 'right' }}>
                                  <Text strong>{formatNumber(item.percentage, 1)}%</Text>
                                </div>
                              </List.Item>
                            )}
                          />
                        </>
                      )}
                    </>
                  ) : (
                    <Empty
                      image={Empty.PRESENTED_IMAGE_SIMPLE}
                      description={t('group.noExpenseData')}
                    />
                  )}
                </div>
              ),
            },
            // --- 加油站推荐 Tab ---
            {
              key: 'stations',
              label: (
                <span>
                  <EnvironmentOutlined style={{ marginRight: 4 }} />
                  {t('group.stationRecommend')}
                </span>
              ),
              children: (
                <div>
                  <Space wrap style={{ marginBottom: 16 }}>
                    <Select
                      value={stationFuelGrade || undefined}
                      placeholder={t('group.fuelGradeFilter')}
                      allowClear
                      style={{ width: 140 }}
                      onChange={(val) => {
                        const v = val || '';
                        setStationFuelGrade(v);
                        fetchStationStats(selectedGroup.id, v, stationMonths, stationSortBy);
                      }}
                      options={FUEL_GRADES.map((item) => ({
                        label: t(item.label),
                        value: item.value,
                      }))}
                    />
                    <Select
                      value={stationMonths}
                      style={{ width: 130 }}
                      onChange={(val) => {
                        setStationMonths(val);
                        fetchStationStats(selectedGroup.id, stationFuelGrade, val, stationSortBy);
                      }}
                      options={[
                        { label: t('group.recentMonths', { count: 3 }), value: 3 },
                        { label: t('group.recentMonths', { count: 6 }), value: 6 },
                        { label: t('group.recentMonths', { count: 12 }), value: 12 },
                      ]}
                    />
                    <Select
                      value={stationSortBy}
                      style={{ width: 120 }}
                      onChange={(val) => {
                        setStationSortBy(val);
                        fetchStationStats(selectedGroup.id, stationFuelGrade, stationMonths, val);
                      }}
                      options={[
                        { label: t('group.sortByFrequency'), value: 'frequency' },
                        { label: t('group.sortByPrice'), value: 'price' },
                        { label: t('group.sortByDate'), value: 'date' },
                      ]}
                    />
                  </Space>

                  {stationLoading ? (
                    <Card loading />
                  ) : stationStats && stationStats.stations.length > 0 ? (
                    <Row gutter={[16, 16]}>
                      {stationStats.stations.map((station, idx) => (
                        <Col xs={24} sm={12} lg={8} key={idx}>
                          <Card size="small" hoverable>
                            <Space direction="vertical" style={{ width: '100%' }} size={8}>
                              <Space align="center">
                                <EnvironmentOutlined style={{ color: '#1677ff', fontSize: 16 }} />
                                <Text strong style={{ fontSize: 15 }}>{station.station_name}</Text>
                              </Space>

                              <Row gutter={16}>
                                <Col span={12}>
                                  <Text type="secondary" style={{ fontSize: 12 }}>{t('group.avgPrice')}</Text>
                                  <div>
                                    <Text strong><Money amount={stationPrice(station.avg_unit_price, station.price_unit).value} currency={station.currency_code || currency} reference={referenceText(stationPrice(station.avg_unit_price, station.price_unit).value, station.currency_code || currency, currency)} /></Text>
                                    <Text type="secondary" style={{ fontSize: 12 }}>/{stationPrice(0, station.price_unit).unit}</Text>
                                  </div>
                                </Col>
                                <Col span={12}>
                                  <Text type="secondary" style={{ fontSize: 12 }}>{t('group.latestPrice')}</Text>
                                  <div>
                                    <Text strong><Money amount={stationPrice(station.latest_unit_price, station.price_unit).value} currency={station.currency_code || currency} reference={referenceText(stationPrice(station.latest_unit_price, station.price_unit).value, station.currency_code || currency, currency)} /></Text>
                                    <Text type="secondary" style={{ fontSize: 12 }}>/{stationPrice(0, station.price_unit).unit} </Text>
                                    <Text
                                      type={station.price_trend === 'up' ? 'danger' : station.price_trend === 'down' ? 'success' : 'secondary'}
                                      style={{ fontSize: 12 }}
                                    >
                                      {station.price_trend === 'up' && <><ArrowUpOutlined /> {t('group.priceTrendUp')}</>}
                                      {station.price_trend === 'down' && <><ArrowDownOutlined /> {t('group.priceTrendDown')}</>}
                                      {station.price_trend === 'stable' && <><MinusOutlined /> {t('group.priceTrendStable')}</>}
                                    </Text>
                                  </div>
                                </Col>
                              </Row>

                              <Row gutter={16}>
                                <Col span={12}>
                                  <Text type="secondary" style={{ fontSize: 12 }}>{t('group.visitCount')}</Text>
                                  <div><Text>{station.visit_count} {t('group.metricFrequency')}</Text></div>
                                </Col>
                                <Col span={12}>
                                  <Text type="secondary" style={{ fontSize: 12 }}>{t('group.latestVisit')}</Text>
                                  <div><Text>{dayjs(station.latest_visit).format('YYYY-MM-DD')}</Text></div>
                                </Col>
                              </Row>

                              {station.visitors.length > 0 && (
                                <div>
                                  <Text type="secondary" style={{ fontSize: 12 }}>{t('group.regulars')}</Text>
                                  <div style={{ marginTop: 4 }}>
                                    {station.visitors.map((v, vi) => (
                                      <Tag key={vi} style={{ marginBottom: 4 }}>
                                        {v.nickname} ({v.count})
                                      </Tag>
                                    ))}
                                  </div>
                                </div>
                              )}

                              {station.fuel_grades_seen.length > 0 && (
                                <div>
                                  <Text type="secondary" style={{ fontSize: 12 }}>{t('group.fuelGrades')}</Text>
                                  <div style={{ marginTop: 4 }}>
                                    {station.fuel_grades_seen.map((g, gi) => {
                                      const grade = FUEL_GRADES.find((fg) => fg.value === g);
                                      return (
                                        <Tag key={gi} color="blue" style={{ marginBottom: 4 }}>
                                          {grade ? t(grade.label) : g}
                                        </Tag>
                                      );
                                    })}
                                  </div>
                                </div>
                              )}
                            </Space>
                          </Card>
                        </Col>
                      ))}
                    </Row>
                  ) : (
                    <Empty
                      image={Empty.PRESENTED_IMAGE_SIMPLE}
                      description={t('group.noStationData')}
                    />
                  )}
                </div>
              ),
            },
          ]}
          onChange={(key) => {
            if (!selectedGroup) return;
            if (key === 'leaderboard' && !leaderboard) {
              fetchLeaderboard(selectedGroup.id, lbMetric, lbPeriod, lbEnergy);
            }
            if (key === 'expense' && !expenseStats) {
              fetchExpenseStats(selectedGroup.id, expensePeriod, expenseYear);
            }
            if (key === 'stations' && !stationStats) {
              fetchStationStats(selectedGroup.id, stationFuelGrade, stationMonths, stationSortBy);
            }
          }}
        />
      </Card>
    );
  };

  return (
    <div className="page-container">
      <div className="page-header">
        <h2>{t('group.title')}</h2>
        <Space>
          <Button
            icon={<UserAddOutlined />}
            onClick={() => setJoinModalOpen(true)}
          >
            {t('group.joinGroup')}
          </Button>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => setCreateModalOpen(true)}
          >
            {t('group.create')}
          </Button>
        </Space>
      </div>

      <Card loading={loading}>
        {renderGroupCards()}
      </Card>

      {renderGroupDetail()}

      {/* 创建群组弹窗 */}
      <Modal
        title={t('group.createGroup')}
        open={createModalOpen}
        onCancel={() => {
          setCreateModalOpen(false);
          createForm.resetFields();
        }}
        footer={null}
        destroyOnHidden
      >
        <Form
          form={createForm}
          layout="vertical"
          onFinish={handleCreate}
          initialValues={{ max_members: 10 }}
        >
          <Form.Item
            name="name"
            label={t('group.name')}
            rules={[{ required: true, message: t('common.required') }]}
          >
            <Input
              maxLength={100}
              placeholder={t('group.namePlaceholder')}
            />
          </Form.Item>

          <Form.Item
            name="max_members"
            label={t('group.maxMembers')}
          >
            <InputNumber min={2} max={50} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item
            name="description"
            label={t('group.description')}
          >
            <Input.TextArea
              rows={2}
              maxLength={500}
              placeholder={t('group.descriptionPlaceholder')}
            />
          </Form.Item>

          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => { setCreateModalOpen(false); createForm.resetFields(); }}>
                {t('common.cancel')}
              </Button>
              <Button type="primary" htmlType="submit" loading={creating}>
                {t('group.create')}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 加入群组弹窗 */}
      <Modal
        title={t('group.joinByCode')}
        open={joinModalOpen}
        onCancel={() => {
          setJoinModalOpen(false);
          joinForm.resetFields();
        }}
        footer={null}
        destroyOnHidden
      >
        <Form
          form={joinForm}
          layout="vertical"
          onFinish={handleJoin}
        >
          <Form.Item
            name="invite_code"
            label={t('group.inviteCode')}
            rules={[{ required: true, message: t('common.required') }]}
          >
            <Input
              maxLength={20}
              placeholder={t('group.inviteCodePlaceholder')}
              style={{ textTransform: 'uppercase' }}
            />
          </Form.Item>

          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => { setJoinModalOpen(false); joinForm.resetFields(); }}>
                {t('common.cancel')}
              </Button>
              <Button type="primary" htmlType="submit" loading={joining}>
                {t('group.joinGroup')}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 编辑群组弹窗 */}
      <Modal
        title={t('group.editGroup')}
        open={editModalOpen}
        onCancel={() => {
          setEditModalOpen(false);
          editForm.resetFields();
        }}
        footer={null}
        destroyOnHidden
      >
        <Form
          form={editForm}
          layout="vertical"
          onFinish={handleUpdate}
        >
          <Form.Item
            name="name"
            label={t('group.name')}
            rules={[{ required: true, message: t('common.required') }]}
          >
            <Input maxLength={100} />
          </Form.Item>

          <Form.Item
            name="max_members"
            label={t('group.maxMembers')}
          >
            <InputNumber min={2} max={50} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item
            name="description"
            label={t('group.description')}
          >
            <Input.TextArea rows={2} maxLength={500} />
          </Form.Item>

          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => { setEditModalOpen(false); editForm.resetFields(); }}>
                {t('common.cancel')}
              </Button>
              <Button type="primary" htmlType="submit" loading={updating}>
                {t('common.save')}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
