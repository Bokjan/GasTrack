// ============================================================
// 常量定义：燃油类型、车辆类型、计量单位、货币、支持的语言
// ============================================================

export const FUEL_TYPES = [
  { value: 'gasoline', label: 'fuelType.gasoline' },
  { value: 'diesel', label: 'fuelType.diesel' },
  { value: 'hybrid', label: 'fuelType.hybrid' },
] as const;

export const VEHICLE_TYPES = [
  { value: 'car', label: 'vehicleType.car' },
  { value: 'motorcycle', label: 'vehicleType.motorcycle' },
  { value: 'other', label: 'vehicleType.other' },
] as const;

export const MEASUREMENT_SYSTEMS = [
  { value: 'metric_eu', label: 'measurement.metricEu', unit: 'L/100km' },
  { value: 'metric_jp', label: 'measurement.metricJp', unit: 'km/L' },
  { value: 'imperial', label: 'measurement.imperial', unit: 'MPG' },
] as const;

export const CURRENCIES = [
  { value: 'CNY', label: '¥ CNY', symbol: '¥' },
  { value: 'USD', label: '$ USD', symbol: '$' },
  { value: 'EUR', label: '€ EUR', symbol: '€' },
  { value: 'JPY', label: '¥ JPY', symbol: '¥' },
  { value: 'GBP', label: '£ GBP', symbol: '£' },
  { value: 'KRW', label: '₩ KRW', symbol: '₩' },
] as const;

export const SUPPORTED_LOCALES = [
  { value: 'zh-CN', label: '简体中文' },
  { value: 'en-US', label: 'English' },
  { value: 'ja-JP', label: '日本語' },
] as const;

export const DEFAULT_PAGE_SIZE = 20;
