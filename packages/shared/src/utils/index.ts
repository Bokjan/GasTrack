import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc';
import timezone from 'dayjs/plugin/timezone';

dayjs.extend(utc);
dayjs.extend(timezone);

// ============================================================
// 日期时间格式化（时区感知）
// ============================================================

/**
 * 将 UTC 时间字符串按用户时区格式化。
 * @param dateStr  后端返回的 ISO 8601 时间字符串（通常为 UTC）
 * @param tz       用户时区，如 'Asia/Shanghai'、'America/New_York'
 * @param fmt      dayjs 格式化模式，默认 'YYYY-MM-DD'
 */
export function formatDateTime(
  dateStr: string,
  tz?: string,
  fmt = 'YYYY-MM-DD',
): string {
  const d = dayjs.utc(dateStr);
  return tz ? d.tz(tz).format(fmt) : d.local().format(fmt);
}

// ============================================================
// 单位换算工具函数
// ============================================================

/** 升/百公里 → 公里/升 */
export function lper100kmToKmpl(lper100km: number): number {
  if (lper100km <= 0) return 0;
  return Math.round((100 / lper100km) * 100) / 100;
}

/** 公里/升 → 升/百公里 */
export function kmplToLper100km(kmpl: number): number {
  if (kmpl <= 0) return 0;
  return Math.round((100 / kmpl) * 100) / 100;
}

/** 公里/升 → MPG */
export function kmplToMpg(kmpl: number): number {
  return Math.round(kmpl * 2.352 * 100) / 100;
}

/** MPG → 公里/升 */
export function mpgToKmpl(mpg: number): number {
  return Math.round((mpg / 2.352) * 100) / 100;
}

/** 升 → 加仑 */
export function litersToGallons(liters: number): number {
  return Math.round(liters * 0.264172 * 1000) / 1000;
}

/** 加仑 → 升 */
export function gallonsToLiters(gallons: number): number {
  return Math.round(gallons * 3.78541 * 1000) / 1000;
}

/** 公里 → 英里 */
export function kmToMiles(km: number): number {
  return Math.round(km * 0.621371 * 1000) / 1000;
}

/** 英里 → 公里 */
export function milesToKm(miles: number): number {
  return Math.round(miles * 1.60934 * 1000) / 1000;
}

// ============================================================
// 油耗单位通用转换（以 L/100km 为基准）
// ============================================================

const MPG_FACTOR = 235.215;

/** 将油耗值从 from 单位转为 to 单位 */
export function convertFuelEfficiency(
  value: number,
  from: string,
  to: string,
): number {
  if (from === to || value <= 0) return value;

  // 先转为 L/100km
  let l100km: number;
  switch (from) {
    case 'L/100km':
      l100km = value;
      break;
    case 'km/L':
      l100km = value > 0 ? 100 / value : 0;
      break;
    case 'MPG':
      l100km = value > 0 ? MPG_FACTOR / value : 0;
      break;
    // 电动车单位保持原样
    default:
      return value;
  }

  // 再从 L/100km 转为目标
  switch (to) {
    case 'L/100km':
      return Math.round(l100km * 100) / 100;
    case 'km/L':
      return l100km > 0 ? Math.round((100 / l100km) * 100) / 100 : 0;
    case 'MPG':
      return l100km > 0 ? Math.round((MPG_FACTOR / l100km) * 100) / 100 : 0;
    default:
      return value;
  }
}

/** 燃油车三种油耗单位 */
export const FUEL_EFFICIENCY_UNITS = ['L/100km', 'km/L', 'MPG'] as const;

/** 电动车三种能耗单位 */
export const EV_EFFICIENCY_UNITS = ['kWh/100km', 'km/kWh', 'mi/kWh'] as const;

/** 格式化数字：添加千分位 */
export function formatNumber(num: number | undefined | null, decimals = 2): string {
  if (num == null || isNaN(num)) return (0).toLocaleString(undefined, {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  });
  return num.toLocaleString(undefined, {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  });
}

/** 格式化金额 */
export function formatCurrency(amount: number | undefined | null, currency: string): string {
  const symbols: Record<string, string> = {
    CNY: '¥',
    USD: '$',
    EUR: '€',
    JPY: '¥',
    GBP: '£',
    KRW: '₩',
  };
  const symbol = symbols[currency] || currency;
  const decimals = currency === 'JPY' || currency === 'KRW' ? 0 : 2;
  return `${symbol}${formatNumber(amount ?? 0, decimals)}`;
}

// ============================================================
// 汇率参考工具函数
// ============================================================

/**
 * 使用汇率数据将金额从一种货币换算为另一种货币。
 * 仅供参考展示，不影响原始数据。
 * @param amount       原始金额
 * @param fromCurrency 原始币种（如 "CNY"）
 * @param toCurrency   目标币种（如 "USD"）
 * @param rates        汇率表（以 fromCurrency 为基准）
 * @returns 换算后的金额，如果无法换算则返回 null
 */
export function convertAmount(
  amount: number,
  fromCurrency: string,
  toCurrency: string,
  rates: Record<string, number>,
): number | null {
  if (fromCurrency === toCurrency) return amount;
  const rate = rates[toCurrency];
  if (rate == null || rate <= 0) return null;
  return amount * rate;
}

/**
 * 获取参考币种：
 * - 如果用户已设置 reference_currency，优先使用
 * - 否则自动推导：USD 用户默认显示 EUR，其他用户默认显示 USD
 */
export function getReferenceCurrency(userCurrency: string, referenceCurrency?: string): string {
  if (referenceCurrency) return referenceCurrency;
  return userCurrency === 'USD' ? 'EUR' : 'USD';
}

/**
 * 获取记录详情页应展示的参考币种列表：
 * - 如果用户已设置 reference_currency，只返回该币种
 * - 否则返回最多 3 个其他币种
 */
export function getReferenceCurrencies(currentCurrency: string, referenceCurrency?: string): string[] {
  if (referenceCurrency && referenceCurrency !== currentCurrency) {
    return [referenceCurrency];
  }
  const all = ['USD', 'EUR', 'CNY', 'JPY', 'GBP', 'KRW'];
  return all.filter((c) => c !== currentCurrency).slice(0, 3);
}
