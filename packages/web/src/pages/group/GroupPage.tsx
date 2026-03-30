import { useState, useEffect, useCallback } from 'react';
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
  LeaderboardPeriod,
  GroupExpenseStatsResponse,
  GroupStationStatsResponse,
} from '@gastrack/shared';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import { useIsMobile } from '../../hooks/useIsMobile';
import { useAuthStore, FUEL_GRADES } from '@gastrack/shared';

const { Text, Paragraph } = Typography;

export default function GroupPage() {
  const { t } = useTranslation();
  const { user } = useAuthStore();
  const isMobile = useIsMobile();
  const navigate = useNavigate();

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
    } catch {
      message.error(t('common.error'));
    } finally {
      setLoading(false);
    }
  }, [t]);

  useEffect(() => {
    fetchGroups();
  }, [fetchGroups]);

  // 选中群组时获取详情 + 数据汇总
  const selectGroup = useCallback(async (group: Group) => {
    try {
      const { data } = await groupApi.getById(group.id);
      setSelectedGroup(data.data);
    } catch {
      message.error(t('common.error'));
    }

    setOverviewLoading(true);
    try {
      const { data } = await groupApi.getOverview(group.id);
      setOverview(data.data);
    } catch {
      setOverview(null);
    } finally {
      setOverviewLoading(false);
    }

    // 获取共享车辆
    try {
      const { data } = await groupApi.listSharedVehicles(group.id);
      setSharedVehicles(data.data || []);
    } catch {
      setSharedVehicles([]);
    }
  }, [t]);

  // 获取排行榜数据
  const fetchLeaderboard = useCallback(async (groupId: string, metric: LeaderboardMetric, period: LeaderboardPeriod) => {
    setLeaderboardLoading(true);
    try {
      const { data } = await groupApi.getLeaderboard(groupId, { metric, period });
      setLeaderboard(data.data);
    } catch {
      setLeaderboard(null);
    } finally {
      setLeaderboardLoading(false);
    }
  }, []);

  // 获取费用统计数据
  const fetchExpenseStats = useCallback(async (groupId: string, period: 'month' | 'year', year: number) => {
    setExpenseLoading(true);
    try {
      const { data } = await groupApi.getExpenseStats(groupId, { period, year: period === 'month' ? year : undefined });
      setExpenseStats(data.data);
    } catch {
      setExpenseStats(null);
    } finally {
      setExpenseLoading(false);
    }
  }, []);

  // 获取加油站数据
  const fetchStationStats = useCallback(async (groupId: string, fuelGrade: string, months: number, sortBy: string) => {
    setStationLoading(true);
    try {
      const { data } = await groupApi.getStationStats(groupId, {
        fuel_grade: fuelGrade || undefined,
        months,
        sort_by: sortBy,
      });
      setStationStats(data.data);
    } catch {
      setStationStats(null);
    } finally {
      setStationLoading(false);
    }
  }, []);

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
    } catch {
      message.error(t('common.error'));
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
    } catch {
      message.error(t('common.error'));
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
    } catch {
      message.error(t('common.error'));
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
    } catch {
      message.error(t('common.error'));
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
    } catch {
      message.error(t('common.error'));
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
    } catch {
      message.error(t('common.error'));
    }
  };

  const handleCopyInviteCode = async (code: string) => {
    try {
      await navigator.clipboard.writeText(code);
      message.success(t('group.copied'));
    } catch {
      message.error(t('common.error'));
    }
  };

  const handleRegenerateCode = async () => {
    if (!selectedGroup) return;
    try {
      const { data } = await groupApi.regenerateInviteCode(selectedGroup.id);
      message.success(t('group.regenerateSuccess'));
      setSelectedGroup(data.data);
      fetchGroups();
    } catch {
      message.error(t('common.error'));
    }
  };

  const handleChangeRole = async (userId: string, role: 'admin' | 'member') => {
    if (!selectedGroup) return;
    try {
      await groupApi.updateMemberRole(selectedGroup.id, userId, { role });
      message.success(t('group.changeRoleSuccess'));
      selectGroup(selectedGroup);
    } catch {
      message.error(t('common.error'));
    }
  };

  const handleRemoveMember = async (userId: string) => {
    if (!selectedGroup) return;
    try {
      await groupApi.removeMember(selectedGroup.id, userId);
      message.success(t('group.removeMemberSuccess'));
      selectGroup(selectedGroup);
      fetchGroups();
    } catch {
      message.error(t('common.error'));
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
                      </Row>
                      <Table
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
                            render: (val: number) => val.toFixed(2),
                          },
                          {
                            title: t('group.totalFuel'),
                            dataIndex: 'total_fuel',
                            key: 'total_fuel',
                            width: 120,
                            render: (val: number) => `${val.toFixed(2)} L`,
                          },
                          {
                            title: t('group.avgEfficiency'),
                            dataIndex: 'avg_efficiency',
                            key: 'avg_efficiency',
                            width: 130,
                            render: (val: number) => `${val.toFixed(2)} L/100km`,
                          },
                          {
                            title: t('group.shareVehicle'),
                            key: 'shared',
                            width: 100,
                            render: (_: unknown, record: { vehicle_id: string; owner_id: string }) => {
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
                            render: (_: unknown, record: { vehicle_id: string; owner_id: string }) => {
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
                        fetchLeaderboard(selectedGroup.id, m, lbPeriod);
                      }}
                      options={[
                        { label: t('group.metricEfficiency'), value: 'efficiency' },
                        { label: t('group.metricCost'), value: 'cost' },
                        { label: t('group.metricDistance'), value: 'distance' },
                        { label: t('group.metricFrequency'), value: 'frequency' },
                      ]}
                    />
                    <Select
                      value={lbPeriod}
                      onChange={(val) => {
                        setLbPeriod(val);
                        fetchLeaderboard(selectedGroup.id, lbMetric, val);
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
                        <Text strong>{leaderboard.group_avg.toFixed(2)}</Text> {leaderboard.unit}
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
                                  <Text>{item.value.toFixed(2)} {leaderboard.unit}</Text>
                                  <Text type="secondary" style={{ fontSize: 12 }}>
                                    {t('group.recordCount')}: {item.record_count}
                                  </Text>
                                  <Text
                                    type={item.diff_from_avg < 0 ? 'success' : item.diff_from_avg > 0 ? 'danger' : 'secondary'}
                                    style={{ fontSize: 12 }}
                                  >
                                    {t('group.diffFromAvg')}: {item.diff_from_avg > 0 ? '+' : ''}{item.diff_from_avg.toFixed(1)}%
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
                      {/* 统计卡片 */}
                      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
                        {[
                          {
                            title: t('group.totalExpense'),
                            value: expenseStats.summary.total_cost,
                            change: expenseStats.summary.cost_change_pct,
                            precision: 2,
                            prefix: '¥',
                          },
                          {
                            title: t('group.totalFuelAmount'),
                            value: expenseStats.summary.total_fuel,
                            change: expenseStats.summary.fuel_change_pct,
                            precision: 2,
                            suffix: 'L',
                          },
                          {
                            title: t('group.totalMileage'),
                            value: expenseStats.summary.total_distance,
                            change: expenseStats.summary.distance_change_pct,
                            precision: 0,
                            suffix: 'km',
                          },
                          {
                            title: t('group.avgFuelEfficiency'),
                            value: expenseStats.summary.avg_efficiency,
                            change: expenseStats.summary.efficiency_change_pct,
                            precision: 2,
                            suffix: 'L/100km',
                          },
                        ].map((card, idx) => (
                          <Col xs={12} sm={6} key={idx}>
                            <Card size="small">
                              <Statistic
                                title={card.title}
                                value={card.value}
                                precision={card.precision}
                                prefix={card.prefix}
                                suffix={
                                  <span>
                                    {card.suffix}{' '}
                                    {card.change !== 0 && (
                                      <Text
                                        style={{ fontSize: 12 }}
                                        type={card.change > 0 ? 'danger' : 'success'}
                                      >
                                        {card.change > 0 ? <CaretUpOutlined /> : <CaretDownOutlined />}
                                        {Math.abs(card.change).toFixed(1)}%
                                      </Text>
                                    )}
                                  </span>
                                }
                              />
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
                                render: (val: number) => `¥${val.toFixed(2)}`,
                              },
                              {
                                title: t('group.totalFuelAmount'),
                                dataIndex: 'total_fuel',
                                key: 'total_fuel',
                                render: (val: number) => `${val.toFixed(2)} L`,
                              },
                              {
                                title: t('group.totalMileage'),
                                dataIndex: 'total_distance',
                                key: 'total_distance',
                                render: (val: number) => `${val.toFixed(0)} km`,
                              },
                              {
                                title: t('group.avgFuelEfficiency'),
                                dataIndex: 'avg_efficiency',
                                key: 'avg_efficiency',
                                render: (val: number) => `${val.toFixed(2)} L/100km`,
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
                                  description={`¥${item.total_cost.toFixed(2)} · ${item.total_fuel.toFixed(2)} L`}
                                />
                                <div style={{ minWidth: 60, textAlign: 'right' }}>
                                  <Text strong>{item.percentage.toFixed(1)}%</Text>
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
                                    <Text strong>¥{station.avg_unit_price.toFixed(2)}</Text>
                                    <Text type="secondary" style={{ fontSize: 12 }}>/L</Text>
                                  </div>
                                </Col>
                                <Col span={12}>
                                  <Text type="secondary" style={{ fontSize: 12 }}>{t('group.latestPrice')}</Text>
                                  <div>
                                    <Text strong>¥{station.latest_unit_price.toFixed(2)}</Text>
                                    <Text type="secondary" style={{ fontSize: 12 }}>/L </Text>
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
              fetchLeaderboard(selectedGroup.id, lbMetric, lbPeriod);
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
        destroyOnClose
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
        destroyOnClose
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
        destroyOnClose
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
