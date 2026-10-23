import { AxiosError, AxiosHeaders, type InternalAxiosRequestConfig } from 'axios';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import apiClient, { setUnauthorizedHandler } from '../api/client';
import { getApiErrorMessage } from '../utils';
import i18n from '../i18n';

/** 让 apiClient 的所有请求都以给定状态码失败 */
function failWith(status: number, data: unknown = {}) {
  apiClient.defaults.adapter = async (config: InternalAxiosRequestConfig) => {
    throw new AxiosError('failed', String(status), config, null, {
      status,
      statusText: '',
      data,
      headers: {},
      config: { ...config, headers: new AxiosHeaders() },
    });
  };
}

describe('401 handling', () => {
  const onUnauthorized = vi.fn();

  beforeEach(() => {
    localStorage.clear();
    localStorage.setItem('theme_mode', 'dark');
    localStorage.setItem('access_token', 'expired');
    onUnauthorized.mockReset();
    setUnauthorizedHandler(onUnauthorized);
  });

  afterEach(() => {
    delete apiClient.defaults.adapter;
  });

  it('treats a 401 from login as a normal error', async () => {
    failWith(401, { message_key: 'auth.invalid_credentials' });
    await expect(apiClient.post('/auth/login', {})).rejects.toBeInstanceOf(AxiosError);
    expect(onUnauthorized).not.toHaveBeenCalled();
    expect(localStorage.getItem('access_token')).toBe('expired');
  });

  it('expires the session without wiping other preferences', async () => {
    failWith(401);
    await expect(apiClient.get('/vehicles')).rejects.toBeInstanceOf(AxiosError);
    expect(onUnauthorized).toHaveBeenCalledOnce();
    expect(localStorage.getItem('access_token')).toBeNull();
    expect(localStorage.getItem('theme_mode')).toBe('dark');
  });
});

describe('getApiErrorMessage', () => {
  const errorWith = (data: unknown) =>
    new AxiosError('failed', '400', undefined, null, {
      status: 400,
      statusText: '',
      data,
      headers: {},
      config: { headers: new AxiosHeaders() },
    });

  it('localizes known message keys', async () => {
    await i18n.changeLanguage('zh-CN');
    expect(getApiErrorMessage(errorWith({ message_key: 'user.wrong_password', message: 'current password is incorrect' })))
      .toBe('当前密码不正确');
  });

  it('falls back to the English message only in English', async () => {
    const err = errorWith({ message_key: 'unknown.key', message: 'raw message' });
    await i18n.changeLanguage('en-US');
    expect(getApiErrorMessage(err)).toBe('raw message');
    await i18n.changeLanguage('ja-JP');
    expect(getApiErrorMessage(err)).toBe(i18n.t('common.error'));
  });

  it('reports network errors', async () => {
    await i18n.changeLanguage('en-US');
    expect(getApiErrorMessage(new AxiosError('Network Error'))).toBe(i18n.t('apiError.network'));
  });
});
