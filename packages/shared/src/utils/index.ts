import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc';
import timezone from 'dayjs/plugin/timezone';
import i18n from '../i18n';

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

/**
 * 日期选择器使用浏览器本地时区，而展示统一用资料中的时区。
 * 以下两个函数在两者之间转换"墙上时间"，保证录入与展示一致（例如出差时设备时区与资料时区不同）。
 */
const WALL_CLOCK_FMT = 'YYYY-MM-DDTHH:mm:ss';

/** 将 UTC 时间（缺省为当前时间）转为用户时区的墙上时间，供日期选择器显示 */
export function toPickerValue(dateStr?: string, tz?: string): dayjs.Dayjs {
  const d = dateStr ? dayjs.utc(dateStr) : dayjs.utc();
  return tz ? dayjs(d.tz(tz).format(WALL_CLOCK_FMT)) : d.local();
}

/** 将日期选择器的值按用户时区解释，返回带时区偏移的 ISO 8601 字符串 */
export function fromPickerValue(value: dayjs.Dayjs, tz?: string): string {
  return tz
    ? dayjs.tz(value.format(WALL_CLOCK_FMT), tz).format()
    : value.format('YYYY-MM-DDTHH:mm:ssZ');
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

/** 燃油单位对应的电耗单位 */
export const EV_EFFICIENCY_UNIT_OF: Record<string, string> = {
  'L/100km': 'kWh/100km',
  'km/L': 'km/kWh',
  'MPG': 'mi/kWh',
};

/**
 * 将后端基准油耗（L/100km）或电耗（kWh/100km）转为用户偏好单位。
 * 电动车按燃油单位的同类形式换算：kWh/100km、km/kWh、mi/kWh
 */
export function convertEfficiencyForDisplay(value: number, unit: string, isEv: boolean): { value: number; unit: string } {
  if (!isEv) return { value: convertFuelEfficiency(value, 'L/100km', unit), unit };
  const evUnit = EV_EFFICIENCY_UNIT_OF[unit] ?? 'kWh/100km';
  if (value <= 0 || evUnit === 'kWh/100km') return { value, unit: evUnit };
  const kmPerKwh = 100 / value;
  return { value: Math.round((evUnit === 'km/kWh' ? kmPerKwh : kmToMiles(kmPerKwh)) * 100) / 100, unit: evUnit };
}

/** 燃油车三种油耗单位 */
export const FUEL_EFFICIENCY_UNITS = ['L/100km', 'km/L', 'MPG'] as const;

/** 电动车三种能耗单位 */
export const EV_EFFICIENCY_UNITS = ['kWh/100km', 'km/kWh', 'mi/kWh'] as const;

/** 数字/金额格式化使用应用当前语言，而非浏览器语言 */
const appLocale = () => i18n.language || undefined;

/** 格式化数字：添加千分位 */
export function formatNumber(num: number | undefined | null, decimals = 2): string {
  const value = num == null || isNaN(num) ? 0 : num;
  return value.toLocaleString(appLocale(), {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  });
}

/**
 * 格式化金额。使用 Intl 按当前语言选择符号：本地币种显示短符号，
 * 外币带前缀区分（如中文界面 ¥ / JP¥，英文界面 CN¥ / ¥），避免 CNY 与 JPY 混淆。
 */
export function formatCurrency(amount: number | undefined | null, currency: string): string {
  const value = amount == null || isNaN(amount) ? 0 : amount;
  try {
    return new Intl.NumberFormat(appLocale(), { style: 'currency', currency }).format(value);
  } catch {
    // 非法币种代码时退化为 "XXX 1,234.00"
    return `${currency} ${formatNumber(value)}`;
  }
}

// ============================================================
// 汇率参考工具函数
// ============================================================

/**
 * 按汇率表（以 base 为基准，rates[X] 表示 "1 base = X"）在任意两种币种之间换算。
 * @returns 换算后的金额；缺少任一币种汇率时返回 null
 */
export function convertViaBase(
  amount: number,
  fromCurrency: string,
  toCurrency: string,
  base: string,
  rates: Record<string, number>,
): number | null {
  if (fromCurrency === toCurrency) return amount;
  const fromRate = fromCurrency === base ? 1 : rates[fromCurrency];
  const toRate = toCurrency === base ? 1 : rates[toCurrency];
  if (!fromRate || fromRate <= 0 || !toRate || toRate <= 0) return null;
  return (amount / fromRate) * toRate;
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

export { getApiErrorMessage } from './apiError';
