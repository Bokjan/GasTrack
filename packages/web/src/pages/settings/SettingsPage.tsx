import { useState } from 'react';
import { Card, Form, Input, Select, Button, Space, message, Divider, Popconfirm, Typography, Segmented } from 'antd';
import { BulbOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import {
  useAuthStore,
  useThemeStore,
  userApi,
  CURRENCIES,
  MEASUREMENT_SYSTEMS,
  EV_MEASUREMENT_SYSTEMS,
  SUPPORTED_LOCALES,
  TIMEZONES,
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

  if (!user) return null;

  // 所有油耗/电耗单位合并
  const efficiencyOptions = [...MEASUREMENT_SYSTEMS, ...EV_MEASUREMENT_SYSTEMS];

  const handleSavePreferences = async (values: {
    nickname: string;
    locale: string;
    timezone: string;
    unit_system: string;
    currency_code: string;
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
