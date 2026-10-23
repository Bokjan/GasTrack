import axios from 'axios';
import i18n from '../i18n';

/**
 * 从 API 错误中取出可展示给用户的本地化消息。
 * 优先使用后端返回的 message_key 对应的翻译；没有翻译时英文界面显示后端原文，其余语言显示 fallback。
 */
export function getApiErrorMessage(err: unknown, fallback?: string): string {
  const generic = fallback ?? i18n.t('common.error');
  if (!axios.isAxiosError(err)) return generic;
  if (!err.response) return i18n.t('apiError.network');

  const data = err.response.data as { message?: string; message_key?: string } | undefined;
  const key = data?.message_key ? `apiError.${data.message_key}` : '';
  if (key && i18n.exists(key)) return i18n.t(key);
  if (data?.message && i18n.language.startsWith('en')) return data.message;
  return generic;
}
