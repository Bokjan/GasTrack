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

/** 格式化数字：添加千分位 */
export function formatNumber(num: number, decimals = 2): string {
  return num.toLocaleString(undefined, {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  });
}

/** 格式化金额 */
export function formatCurrency(amount: number, currency: string): string {
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
  return `${symbol}${formatNumber(amount, decimals)}`;
}
