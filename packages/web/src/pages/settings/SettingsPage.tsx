import { useState, useEffect } from 'react';
import { Card, Form, Input, Select, Button, Space, message, Divider, Popconfirm, Typography, Segmented, Table, Spin, Radio } from 'antd';
import { BulbOutlined, DownloadOutlined, SwapOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import {
  useAuthStore,
  useThemeStore,
  useExchangeRateStore,
  userApi,
  CURRENCIES,
  MEASUREMENT_SYSTEMS,
  EV_MEASUREMENT_SYSTEMS,
  SUPPORTED_LOCALES,
  TIMEZONES,
  formatNumber,
} from '@gastrack/shared';
import type { ChangePasswordRequest } from '@gastrack/shared';
import type { ThemeMode } from '@gastrack/shared';

const { Title, Text } = Typography;

export default function SettingsPage() {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const { user, updateProfile, logout } = useAuthStore();
  const { mode: themeMode, setMode: setThemeMode } = useThemeStore();
  const [profileForm] = Form.useForm();
  const [passwordForm] = Form.useForm();
  const [saving, setSaving] = useState(false);
  const [changingPwd, setChangingPwd] = useState(false);
  const [exporting, setExporting] = useState(false);
  const [exportScope, setExportScope] = useState<'basic' | 'full'>('basic');
  const [exportFormat, setExportFormat] = useState<'csv' | 'zip' | 'json'>('csv');

  const { data: exchangeRateData, isLoading: ratesLoading, fetchRates } = useExchangeRateStore();

  useEffect(() => {
    if (user?.currency_code) {
      fetchRates(user.currency_code);
    }
  }, [user?.currency_code]);

  if (!user) return null;

  // 所有油耗/电耗单位合并
  const efficiencyOptions = [...MEASUREMENT_SYSTEMS, ...EV_MEASUREMENT_SYSTEMS];

  const handleSavePreferences = async (values: {
    nickname: string;
    locale: string;
    timezone: string;
    unit_system: string;
    currency_code: string;
    reference_currency: string;
    fuel_efficiency_unit: string;
  }) => {
    setSaving(true);
    try {
      await updateProfile({
        nickname: values.nickname,
        locale: values.locale,
        timezone: values.timezone,
        unit_system: values.unit_system,
        currency_code: values.currency_code,
        reference_currency: values.reference_currency ?? '',
        fuel_efficiency_unit: values.fuel_efficiency_unit,
      });
      // 同步前端语言
      if (values.locale !== i18n.language) {
        await i18n.changeLanguage(values.locale);
        localStorage.setItem('locale', values.locale);
      }
      message.success(t('settings.profileUpdated'));
    } catch {
      message.error(t('common.error'));
    } finally {
      setSaving(false);
    }
  };

  const handleChangePassword = async (values: ChangePasswordRequest & { confirm_password: string }) => {
    if (values.new_password !== values.confirm_password) {
      message.error(t('settings.passwordMismatch'));
      return;
    }
    setChangingPwd(true);
    try {
      await userApi.changePassword({
        old_password: values.old_password,
        new_password: values.new_password,
      });
      message.success(t('settings.passwordChanged'));
      passwordForm.resetFields();
    } catch {
      message.error(t('common.error'));
    } finally {
      setChangingPwd(false);
    }
  };

  const handleDeleteAccount = async () => {
    try {
      await userApi.deleteAccount();
      await logout();
      navigate('/login');
    } catch {
      message.error(t('common.error'));
    }
  };

  const handleExportData = async () => {
    setExporting(true);
    try {
      // scope=full 时自动升级为 zip（后端会自动处理，但前端也同步）
      const actualFormat = exportScope === 'full' && exportFormat === 'csv' ? 'zip' : exportFormat;
      const response = await userApi.exportData({ format: actualFormat, scope: exportScope });

      // 根据格式确定 MIME type
      const mimeMap: Record<string, string> = {
        csv: 'text/csv; charset=utf-8',
        zip: 'application/zip',
        json: 'application/json; charset=utf-8',
      };
      const blob = new Blob([response.data], { type: mimeMap[actualFormat] || 'application/octet-stream' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      // 从响应头提取文件名，或使用默认名
      const contentDisposition = response.headers['content-disposition'];
      const filenameMatch = contentDisposition?.match(/filename="?([^"]+)"?/);
      const extMap: Record<string, string> = { csv: '.csv', zip: '.zip', json: '.json' };
      link.download = filenameMatch ? filenameMatch[1] : `gastrack-export${extMap[actualFormat] || '.csv'}`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
      message.success(t('settings.exportSuccess'));
    } catch {
      message.error(t('common.error'));
    } finally {
      setExporting(false);
    }
  };

  const themeOptions: { label: string; value: ThemeMode }[] = [
    { label: t('settings.themeLight'), value: 'light' },
    { label: t('settings.themeDark'), value: 'dark' },
    { label: t('settings.themeSystem'), value: 'system' },
  ];

  return (
    <div className="page-container">
      <div className="page-header">
        <h2>{t('settings.title')}</h2>
      </div>

      <Space direction="vertical" size="large" style={{ width: '100%', maxWidth: 600 }}>
        {/* 外观主题 */}
        <Card
          title={
            <Space>
              <BulbOutlined />
              <span>{t('settings.theme')}</span>
            </Space>
          }
        >
          <Segmented
            value={themeMode}
            onChange={(val) => setThemeMode(val as ThemeMode)}
            options={themeOptions}
            block
          />
        </Card>

        {/* 个人资料 & 偏好设置 */}
        <Card title={t('settings.preferences')}>
          <Form
            form={profileForm}
            layout="vertical"
            onFinish={handleSavePreferences}
            initialValues={{
              nickname: user.nickname,
              locale: user.locale || i18n.language,
              timezone: user.timezone || Intl.DateTimeFormat().resolvedOptions().timeZone,
              unit_system: user.unit_system || 'metric',
              currency_code: user.currency_code || 'CNY',
              reference_currency: user.reference_currency || '',
              fuel_efficiency_unit: user.fuel_efficiency_unit || 'L/100km',
            }}
          >
            <Form.Item name="nickname" label={t('settings.nickname')} rules={[{ required: true, message: t('common.required') }]}>
              <Input />
            </Form.Item>

            <Form.Item label={t('settings.email')}>
              <Input value={user.email} disabled />
            </Form.Item>

            <Form.Item name="locale" label={t('settings.language')}>
              <Select
                options={SUPPORTED_LOCALES.map((l) => ({
                  value: l.value,
                  label: l.label,
                }))}
              />
            </Form.Item>

            <Form.Item name="timezone" label={t('settings.timezone')}>
              <Select
                showSearch
                placeholder={t('settings.timezonePlaceholder')}
                optionFilterProp="label"
                options={TIMEZONES.map((tz) => ({
                  value: tz.value,
                  label: t(tz.label),
                }))}
              />
            </Form.Item>

            <Form.Item name="unit_system" label={t('settings.unitSystem')}>
              <Select
                options={[
                  { value: 'metric', label: t('settings.metric') },
                  { value: 'imperial', label: t('settings.imperial') },
                ]}
              />
            </Form.Item>

            <Form.Item name="currency_code" label={t('settings.currency')}>
              <Select
                options={CURRENCIES.map((c) => ({
                  value: c.value,
                  label: c.label,
                }))}
              />
            </Form.Item>

            <Form.Item
              name="reference_currency"
              label={t('exchangeRate.referenceCurrency')}
              tooltip={t('exchangeRate.referenceCurrencyHint')}
            >
              <Select
                allowClear
                placeholder={t('exchangeRate.referenceCurrencyAuto')}
                options={[
                  { value: '', label: t('exchangeRate.referenceCurrencyAuto') },
                  ...CURRENCIES
                    .filter((c) => c.value !== (profileForm.getFieldValue('currency_code') || user.currency_code))
                    .map((c) => ({
                      value: c.value,
                      label: `${c.symbol} ${t(`exchangeRate.currencyName.${c.value}`)} (${c.value})`,
                    })),
                ]}
              />
            </Form.Item>

            <Form.Item name="fuel_efficiency_unit" label={t('settings.fuelEfficiencyUnit')}>
              <Select
                options={efficiencyOptions.map((m) => ({
                  value: m.unit,
                  label: `${t(m.label)}`,
                }))}
              />
            </Form.Item>

            <Form.Item>
              <Button type="primary" htmlType="submit" loading={saving}>
                {t('common.save')}
              </Button>
            </Form.Item>
          </Form>
        </Card>

        {/* 修改密码 */}
        <Card title={t('settings.changePassword')}>
          <Form
            form={passwordForm}
            layout="vertical"
            onFinish={handleChangePassword}
          >
            <Form.Item
              name="old_password"
              label={t('settings.oldPassword')}
              rules={[{ required: true, message: t('common.required') }]}
            >
              <Input.Password />
            </Form.Item>

            <Form.Item
              name="new_password"
              label={t('settings.newPassword')}
              rules={[{ required: true, message: t('common.required') }, { min: 6 }]}
            >
              <Input.Password />
            </Form.Item>

            <Form.Item
              name="confirm_password"
              label={t('settings.confirmNewPassword')}
              rules={[{ required: true, message: t('common.required') }]}
            >
              <Input.Password />
            </Form.Item>

            <Form.Item>
              <Button type="primary" htmlType="submit" loading={changingPwd}>
                {t('settings.changePassword')}
              </Button>
            </Form.Item>
          </Form>
        </Card>

        {/* 汇率参考 */}
        <Card
          title={
            <Space>
              <SwapOutlined />
              <span>{t('exchangeRate.title')}</span>
            </Space>
          }
        >
          {ratesLoading ? (
            <div style={{ textAlign: 'center', padding: 24 }}>
              <Spin />
              <div style={{ marginTop: 8 }}>
                <Text type="secondary">{t('exchangeRate.loading')}</Text>
              </div>
            </div>
          ) : exchangeRateData?.rates ? (
            <Space direction="vertical" size="small" style={{ width: '100%' }}>
              <Table
                dataSource={Object.entries(exchangeRateData.rates).map(([code, rate]) => ({
                  key: code,
                  code,
                  name: t(`exchangeRate.currencyName.${code}`),
                  symbol: CURRENCIES.find((c) => c.value === code)?.symbol || code,
                  rate: formatNumber(rate, code === 'JPY' || code === 'KRW' ? 2 : 4),
                }))}
                columns={[
                  {
                    title: t('exchangeRate.baseCurrency'),
                    dataIndex: 'code',
                    key: 'code',
                    render: (_: unknown, row: { code: string; symbol: string; name: string }) => `${row.symbol} ${row.code}`,
                  },
                  {
                    title: t('exchangeRate.currencyName.CNY').includes('人民币') ? '名称' : 'Name',
                    dataIndex: 'name',
                    key: 'name',
                  },
                  {
                    title: `1 ${exchangeRateData.base} =`,
                    dataIndex: 'rate',
                    key: 'rate',
                  },
                ]}
                pagination={false}
                size="small"
              />
              <Text type="secondary" style={{ fontSize: 12 }}>
                {t('exchangeRate.lastUpdated', { date: exchangeRateData.date })}
              </Text>
              <Text type="secondary" style={{ fontSize: 12 }}>
                {t('exchangeRate.disclaimer')}
              </Text>
            </Space>
          ) : (
            <Text type="secondary">{t('exchangeRate.noData')}</Text>
          )}
        </Card>

        {/* 数据与隐私 */}
        <Card title={t('settings.dataPrivacy')}>
          <Space direction="vertical" size="middle" style={{ width: '100%' }}>
            <div>
              <Text>{t('settings.exportDescription')}</Text>
            </div>

            <div>
              <Text strong style={{ display: 'block', marginBottom: 8 }}>{t('settings.exportScope')}</Text>
              <Radio.Group
                value={exportScope}
                onChange={(e) => setExportScope(e.target.value)}
                optionType="button"
                buttonStyle="solid"
                options={[
                  { label: t('settings.exportScopeBasic'), value: 'basic' },
                  { label: t('settings.exportScopeFull'), value: 'full' },
                ]}
              />
              <div style={{ marginTop: 4 }}>
                <Text type="secondary" style={{ fontSize: 12 }}>
                  {exportScope === 'basic'
                    ? t('settings.exportScopeBasicHint')
                    : t('settings.exportScopeFullHint')}
                </Text>
              </div>
            </div>

            <div>
              <Text strong style={{ display: 'block', marginBottom: 8 }}>{t('settings.exportFormat')}</Text>
              <Radio.Group
                value={exportScope === 'full' && exportFormat === 'csv' ? 'zip' : exportFormat}
                onChange={(e) => setExportFormat(e.target.value)}
                optionType="button"
                buttonStyle="solid"
                options={[
                  { label: 'CSV', value: 'csv', disabled: exportScope === 'full' },
                  { label: 'ZIP', value: 'zip' },
                  { label: 'JSON', value: 'json' },
                ]}
              />
              <div style={{ marginTop: 4 }}>
                <Text type="secondary" style={{ fontSize: 12 }}>
                  {t('settings.exportFormatHint')}
                </Text>
              </div>
            </div>

            <Button
              type="primary"
              icon={<DownloadOutlined />}
              loading={exporting}
              onClick={handleExportData}
            >
              {t('settings.exportData')}
            </Button>
            <Divider style={{ margin: '8px 0' }} />
            <Space split={<Divider type="vertical" />}>
              <a onClick={() => navigate('/privacy')} style={{ cursor: 'pointer' }}>
                {t('legal.privacyPolicy')}
              </a>
              <a onClick={() => navigate('/terms')} style={{ cursor: 'pointer' }}>
                {t('legal.termsOfService')}
              </a>
            </Space>
          </Space>
        </Card>

        {/* 注销账号 */}
        <Card>
          <Title level={5} type="danger">{t('settings.deleteAccountWarning')}</Title>
          <Text type="secondary">{t('settings.deleteAccountConfirm')}</Text>
          <Divider />
          <Popconfirm
            title={t('settings.deleteAccountConfirm')}
            onConfirm={handleDeleteAccount}
            okText={t('common.confirm')}
            cancelText={t('common.cancel')}
            okButtonProps={{ danger: true }}
          >
            <Button danger>{t('settings.deleteAccount')}</Button>
          </Popconfirm>
        </Card>
      </Space>
    </div>
  );
}
