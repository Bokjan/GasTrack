import { useState, useEffect, useCallback } from 'react';
import {
  Card,
  Table,
  Button,
  Space,
  Tag,
  Modal,
  Form,
  InputNumber,
  DatePicker,
  Input,
  message,
  Popconfirm,
  Typography,
  Tooltip,
  Switch,
  Empty,
} from 'antd';
import {
  PlusOutlined,
  CopyOutlined,
  DeleteOutlined,
  GiftOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { inviteApi } from '@gastrack/shared/src/api';
import type { InviteCode, CreateInviteRequest } from '@gastrack/shared';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';

const { Text, Paragraph } = Typography;

export default function InviteManagePage() {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [invites, setInvites] = useState<InviteCode[]>([]);
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [creating, setCreating] = useState(false);
  const [form] = Form.useForm();

  const fetchInvites = useCallback(async () => {
    setLoading(true);
    try {
      const { data } = await inviteApi.list();
      setInvites(data.data || []);
    } catch {
      message.error(t('common.error'));
    } finally {
      setLoading(false);
    }
  }, [t]);

  useEffect(() => {
    fetchInvites();
  }, [fetchInvites]);

  const handleCreate = async (values: {
    max_uses?: number;
    expires_at?: dayjs.Dayjs;
    note?: string;
  }) => {
    setCreating(true);
    try {
      const req: CreateInviteRequest = {
        max_uses: values.max_uses,
        note: values.note,
      };
      if (values.expires_at) {
        req.expires_at = values.expires_at.toISOString();
      }
      await inviteApi.create(req);
      message.success(t('invite.createSuccess'));
      setCreateModalOpen(false);
      form.resetFields();
      fetchInvites();
    } catch {
      message.error(t('common.error'));
    } finally {
      setCreating(false);
    }
  };

  const handleCopy = async (code: string) => {
    try {
      await navigator.clipboard.writeText(code);
      message.success(t('invite.copied'));
    } catch {
      message.error(t('common.error'));
    }
  };

  const handleToggleActive = async (record: InviteCode) => {
    try {
      await inviteApi.update(record.id, { is_active: !record.is_active });
      message.success(t('common.success'));
      fetchInvites();
    } catch {
      message.error(t('common.error'));
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await inviteApi.delete(id);
      message.success(t('common.success'));
      fetchInvites();
    } catch {
      message.error(t('common.error'));
    }
  };

  /** 状态 Tag */
  const renderStatus = (record: InviteCode) => {
    if (!record.is_active) {
      return <Tag color="default">{t('invite.inactive')}</Tag>;
    }
    if (record.expires_at && dayjs(record.expires_at).isBefore(dayjs())) {
      return <Tag color="orange">{t('invite.expired')}</Tag>;
    }
    if (record.max_uses > 0 && record.use_count >= record.max_uses) {
      return <Tag color="red">{t('invite.used')}</Tag>;
    }
    return <Tag color="green">{t('invite.active')}</Tag>;
  };

  const columns: ColumnsType<InviteCode> = [
    {
      title: t('invite.code'),
      dataIndex: 'code',
      key: 'code',
      render: (code: string) => (
        <Space>
          <Text code copyable={false} style={{ fontSize: 14, fontWeight: 600 }}>
            {code}
          </Text>
          <Tooltip title={t('invite.copyToClipboard')}>
            <Button
              type="text"
              size="small"
              icon={<CopyOutlined />}
              onClick={() => handleCopy(code)}
            />
          </Tooltip>
        </Space>
      ),
    },
    {
      title: t('invite.statusLabel'),
      key: 'status',
      width: 100,
      render: (_, record) => renderStatus(record),
    },
    {
      title: t('invite.usage'),
      key: 'usage',
      width: 120,
      render: (_, record) => (
        <Text>
          {record.use_count} / {record.max_uses > 0 ? record.max_uses : t('invite.unlimited')}
        </Text>
      ),
    },
    {
      title: t('invite.expiresAt'),
      dataIndex: 'expires_at',
      key: 'expires_at',
      width: 180,
      render: (val: string | undefined) =>
        val ? dayjs(val).format('YYYY-MM-DD HH:mm') : <Text type="secondary">—</Text>,
    },
    {
      title: t('invite.note'),
      dataIndex: 'note',
      key: 'note',
      ellipsis: true,
      render: (val: string | undefined) =>
        val || <Text type="secondary">—</Text>,
    },
    {
      title: t('invite.createdAt'),
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (val: string) => dayjs(val).format('YYYY-MM-DD HH:mm'),
    },
    {
      title: t('invite.actions'),
      key: 'actions',
      width: 140,
      render: (_, record) => (
        <Space size="small">
          <Tooltip title={record.is_active ? t('invite.deactivate') : t('invite.activate')}>
            <Switch
              size="small"
              checked={record.is_active}
              onChange={() => handleToggleActive(record)}
            />
          </Tooltip>
          <Popconfirm
            title={t('invite.deleteConfirm')}
            onConfirm={() => handleDelete(record.id)}
            okText={t('common.confirm')}
            cancelText={t('common.cancel')}
          >
            <Button type="text" size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div className="page-container">
      <div className="page-header">
        <h2>{t('invite.manageTitle')}</h2>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => setCreateModalOpen(true)}
        >
          {t('invite.create')}
        </Button>
      </div>

      <Card>
        <Paragraph type="secondary" style={{ marginBottom: 16 }}>
          <GiftOutlined style={{ marginRight: 6 }} />
          {t('invite.generateForFriend')}
        </Paragraph>

        <Table
          rowKey="id"
          columns={columns}
          dataSource={invites}
          loading={loading}
          pagination={false}
          scroll={{ x: 900 }}
          locale={{
            emptyText: (
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description={t('invite.noInvites')}
              />
            ),
          }}
        />
      </Card>

      {/* 创建邀请码弹窗 */}
      <Modal
        title={t('invite.create')}
        open={createModalOpen}
        onCancel={() => {
          setCreateModalOpen(false);
          form.resetFields();
        }}
        footer={null}
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleCreate}
          initialValues={{ max_uses: 1 }}
        >
          <Form.Item
            name="max_uses"
            label={t('invite.maxUses')}
            extra={t('invite.maxUsesHint')}
          >
            <InputNumber min={0} style={{ width: '100%' }} placeholder="0 = unlimited" />
          </Form.Item>

          <Form.Item
            name="expires_at"
            label={t('invite.expiresAt')}
            extra={t('invite.expiresHint')}
          >
            <DatePicker
              showTime
              style={{ width: '100%' }}
              disabledDate={(current) => current && current < dayjs().startOf('day')}
              placeholder={t('invite.expiresPlaceholder')}
            />
          </Form.Item>

          <Form.Item name="note" label={t('invite.note')}>
            <Input.TextArea
              rows={2}
              maxLength={200}
              placeholder={t('invite.notePlaceholder')}
            />
          </Form.Item>

          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button
                onClick={() => {
                  setCreateModalOpen(false);
                  form.resetFields();
                }}
              >
                {t('common.cancel')}
              </Button>
              <Button type="primary" htmlType="submit" loading={creating}>
                {t('invite.create')}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
