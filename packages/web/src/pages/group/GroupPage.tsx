import { useState, useEffect, useCallback } from 'react';
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
  Row,
  Col,
  Descriptions,
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
  EnvironmentOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import type {
  Group,
  CreateGroupRequest,
  UpdateGroupRequest,
  GroupMemberDetail,
} from '@gastrack/shared';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import { useIsMobile } from '../../hooks/useIsMobile';
import { useLatestRequest } from '../../hooks/useLatestRequest';
import { useAuthStore, getApiErrorMessage, groupApi } from '@gastrack/shared';
import GroupOverviewTab from './GroupOverviewTab';
import GroupLeaderboardTab from './GroupLeaderboardTab';
import GroupExpenseTab from './GroupExpenseTab';
import GroupStationTab from './GroupStationTab';

const { Text } = Typography;

export default function GroupPage() {
  const { t } = useTranslation();
  const { user } = useAuthStore();
  const isMobile = useIsMobile();

  const [loading, setLoading] = useState(false);
  const [groups, setGroups] = useState<Group[]>([]);
  const [selectedGroup, setSelectedGroup] = useState<Group | null>(null);
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

  // 选中群组时获取详情；快速切换群组时丢弃旧群组的响应
  const beginSelect = useLatestRequest();
  const selectGroup = useCallback(async (group: Group) => {
    const isLatest = beginSelect();
    try {
      const { data } = await groupApi.getById(group.id);
      if (isLatest()) setSelectedGroup(data.data);
    } catch (err) {
      if (isLatest()) message.error(getApiErrorMessage(err));
    }
  }, [beginSelect]);

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
        {/* 以群组 id 作为 key：切换群组时各标签页重新加载，避免显示上一个群组的数据 */}
        <Tabs
          key={selectedGroup.id}
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
              children: <GroupOverviewTab groupId={selectedGroup.id} />,
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
              children: <GroupLeaderboardTab groupId={selectedGroup.id} />,
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
              children: <GroupExpenseTab groupId={selectedGroup.id} />,
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
              children: <GroupStationTab groupId={selectedGroup.id} />,
            },
          ]}
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
