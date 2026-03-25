// ============================================================
// 常量定义：燃油类型、车辆类型、计量单位、货币、支持的语言
// ============================================================

export const FUEL_TYPES = [
  { value: 'gasoline', label: 'fuelType.gasoline' },
  { value: 'diesel', label: 'fuelType.diesel' },
  { value: 'hybrid', label: 'fuelType.hybrid' },
  { value: 'electric', label: 'fuelType.electric' },
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

/** 电动车能效计量体系 */
export const EV_MEASUREMENT_SYSTEMS = [
  { value: 'kwh_100km', label: 'measurement.kwh100km', unit: 'kWh/100km' },
  { value: 'km_kwh', label: 'measurement.kmKwh', unit: 'km/kWh' },
  { value: 'mi_kwh', label: 'measurement.miKwh', unit: 'mi/kWh' },
] as const;

export const CURRENCIES = [
  { value: 'CNY', label: '¥ CNY', symbol: '¥' },
  { value: 'USD', label: '$ USD', symbol: '$' },
  { value: 'EUR', label: '€ EUR', symbol: '€' },
  { value: 'JPY', label: '¥ JPY', symbol: '¥' },
  { value: 'GBP', label: '£ GBP', symbol: '£' },
  { value: 'KRW', label: '₩ KRW', symbol: '₩' },
] as const;

export const FUEL_UNITS = [
  { value: 'L', label: 'unit.liter' },
  { value: 'gal', label: 'unit.gallon' },
] as const;

/** 电动车能量单位 */
export const ENERGY_UNITS = [
  { value: 'kWh', label: 'unit.kwh' },
] as const;

export const DISTANCE_UNITS = [
  { value: 'km', label: 'unit.km' },
  { value: 'mi', label: 'unit.mile' },
] as const;

/** 判断是否为纯电车型 */
export const isElectricVehicle = (fuelType: string) => fuelType === 'electric';

/** 判断是否需要显示排量（非电动车型） */
export const hasEngineCC = (fuelType: string) => fuelType !== 'electric';

export const SUPPORTED_LOCALES = [
  { value: 'zh-CN', label: '简体中文' },
  { value: 'en-US', label: 'English' },
  { value: 'ja-JP', label: '日本語' },
] as const;

export const DEFAULT_PAGE_SIZE = 20;
