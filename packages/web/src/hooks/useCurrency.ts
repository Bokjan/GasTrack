import { useCallback, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import {
  useAuthStore,
  useExchangeRateStore,
  convertViaBase,
  formatCurrency,
  getReferenceCurrency,
} from '@gastrack/shared';

/**
 * 用户偏好币种与参考汇率。
 * 统计金额已由后端换算为偏好币种，这里只负责"≈ 参考币种"提示及外币记录的参考换算。
 */
export function useCurrency() {
  const { t } = useTranslation();
  const user = useAuthStore((s) => s.user);
  const { data: ratesData, fetchRates } = useExchangeRateStore();

  const userCurrency = user?.currency_code;
  const currency = userCurrency || 'CNY';
  const refCurrency = getReferenceCurrency(currency, user?.reference_currency);

  // 等用户资料加载后再拉取，避免以默认币种为基准
  useEffect(() => {
    if (userCurrency) fetchRates(userCurrency);
  }, [userCurrency, fetchRates]);

  /** 任意两币种换算（无汇率时返回 null） */
  const convert = useCallback(
    (amount: number, from: string, to: string) =>
      ratesData ? convertViaBase(amount, from, to, ratesData.base, ratesData.rates) : null,
    [ratesData],
  );

  /** "≈ $12.34" 形式的参考金额提示；同币种或无汇率时返回 undefined */
  const referenceText = useCallback(
    (amount: number, from: string = currency, to: string = refCurrency) => {
      if (from === to) return undefined;
      const value = convert(amount, from, to);
      return value != null ? t('exchangeRate.referenceAmount', { amount: formatCurrency(value, to) }) : undefined;
    },
    [convert, currency, refCurrency, t],
  );

  return { currency, refCurrency, rates: ratesData, convert, referenceText };
}
