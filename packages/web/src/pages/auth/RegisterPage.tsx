import { useState, useEffect, useRef, useCallback } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Form, Input, Button, message, Typography } from 'antd';
import {
  MailOutlined,
  LockOutlined,
  UserOutlined,
  GiftOutlined,
  CheckCircleFilled,
  CloseCircleFilled,
  LoadingOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useAuthStore } from '@gastrack/shared';
import { authApi, inviteApi } from '@gastrack/shared/src/api';

const { Text } = Typography;

type InviteStatus = 'idle' | 'validating' | 'valid' | 'invalid';

export default function RegisterPage() {
  const [loading, setLoading] = useState(false);
  const [registrationMode, setRegistrationMode] = useState<string>('invite_only');
  const [inviteStatus, setInviteStatus] = useState<InviteStatus>('idle');
  const [inviteHint, setInviteHint] = useState('');
  const debounceTimer = useRef<ReturnType<typeof setTimeout>>();
  const navigate = useNavigate();
  const { t } = useTranslation();
  const register = useAuthStore((s) => s.register);

  // 查询注册模式
  useEffect(() => {
    authApi.getRegistrationMode().then(({ data }) => {
      setRegistrationMode(data.data.mode);
    }).catch(() => {
      // 查询失败默认 invite_only
    });
  }, []);

  // 邀请码实时校验（debounce 500ms）
  const validateInviteCode = useCallback((code: string) => {
    if (debounceTimer.current) clearTimeout(debounceTimer.current);

    if (!code || code.trim().length === 0) {
      setInviteStatus('idle');
      setInviteHint('');
      return;
    }

    setInviteStatus('validating');
    debounceTimer.current = setTimeout(async () => {
      try {
        const { data } = await inviteApi.validate(code.trim());
        const result = data.data;
        if (result.valid) {
          setInviteStatus('valid');
          if (result.remaining_uses && result.remaining_uses > 0) {
            setInviteHint(t('invite.remainingUses', { count: result.remaining_uses }));
          } else {
            setInviteHint(t('invite.valid'));
          }
        } else {
          setInviteStatus('invalid');
          setInviteHint(t('invite.invalid'));
        }
      } catch {
        setInviteStatus('invalid');
        setInviteHint(t('invite.invalid'));
      }
    }, 500);
  }, [t]);

  const onFinish = async (values: {
    email: string;
    password: string;
    nickname: string;
    inviteCode?: string;
  }) => {
    setLoading(true);
    try {
      await register(values.email, values.password, values.nickname, values.inviteCode);
      message.success(t('auth.registerSuccess'));
      navigate('/');
    } catch (err: unknown) {
      const error = err as { response?: { data?: { message?: string } } };
      message.error(error.response?.data?.message || t('common.error'));
    } finally {
      setLoading(false);
    }
  };

  // 注册完全关闭
  if (registrationMode === 'closed') {
    return (
      <div className="auth-container">
        <div className="auth-card">
          <div className="logo">
            <div style={{ fontSize: 48 }}>⛽</div>
            <h1>GasTrack</h1>
            <p>{t('auth.registrationClosed')}</p>
          </div>
          <div style={{ textAlign: 'center', marginTop: 24 }}>
            <span>{t('auth.hasAccount')} </span>
            <Link to="/login">{t('auth.goLogin')}</Link>
          </div>
        </div>
      </div>
    );
  }

  const showInviteField = registrationMode === 'invite_only';

  // 邀请码输入框后缀图标
  const inviteSuffix = (() => {
    switch (inviteStatus) {
      case 'validating':
        return <LoadingOutlined style={{ color: '#1677ff' }} />;
      case 'valid':
        return <CheckCircleFilled style={{ color: '#52c41a' }} />;
      case 'invalid':
        return <CloseCircleFilled style={{ color: '#ff4d4f' }} />;
      default:
        return null;
    }
  })();

  return (
    <div className="auth-container">
      <div className="auth-card">
        <div className="logo">
          <div style={{ fontSize: 48 }}>⛽</div>
          <h1>GasTrack</h1>
          <p>{t('auth.register')}</p>
        </div>

        <Form layout="vertical" onFinish={onFinish} size="large">
          {showInviteField && (
            <Form.Item
              name="inviteCode"
              rules={[{ required: true, message: t('invite.required') }]}
              help={
                inviteHint && (
                  <Text
                    type={inviteStatus === 'valid' ? 'success' : inviteStatus === 'invalid' ? 'danger' : undefined}
                    style={{ fontSize: 12 }}
                  >
                    {inviteHint}
                  </Text>
                )
              }
              validateStatus={
                inviteStatus === 'valid' ? 'success' :
                inviteStatus === 'invalid' ? 'error' :
                inviteStatus === 'validating' ? 'validating' : undefined
              }
            >
              <Input
                prefix={<GiftOutlined />}
                placeholder={t('invite.placeholder')}
                suffix={inviteSuffix}
                onChange={(e) => validateInviteCode(e.target.value)}
                style={{
                  borderColor:
                    inviteStatus === 'valid' ? '#52c41a' :
                    inviteStatus === 'invalid' ? '#ff4d4f' : undefined,
                }}
              />
            </Form.Item>
          )}

          <Form.Item
            name="nickname"
            rules={[{ required: true, message: t('common.required') }]}
          >
            <Input prefix={<UserOutlined />} placeholder={t('auth.nickname')} />
          </Form.Item>

          <Form.Item
            name="email"
            rules={[
              { required: true, message: t('common.required') },
              { type: 'email', message: t('auth.invalidEmail') },
            ]}
          >
            <Input prefix={<MailOutlined />} placeholder={t('auth.email')} />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[
              { required: true, message: t('common.required') },
              { min: 8, message: t('auth.passwordMinLength') },
            ]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder={t('auth.password')}
            />
          </Form.Item>

          <Form.Item
            name="confirmPassword"
            dependencies={['password']}
            rules={[
              { required: true, message: t('common.required') },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('password') === value) {
                    return Promise.resolve();
                  }
                  return Promise.reject(new Error(t('auth.passwordMismatch')));
                },
              }),
            ]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder={t('auth.confirmPassword')}
            />
          </Form.Item>

          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              loading={loading}
              block
              disabled={showInviteField && inviteStatus !== 'valid'}
            >
              {t('auth.register')}
            </Button>
          </Form.Item>
        </Form>

        <div style={{ textAlign: 'center' }}>
          <span>{t('auth.hasAccount')} </span>
          <Link to="/login">{t('auth.goLogin')}</Link>
        </div>
      </div>
    </div>
  );
}
