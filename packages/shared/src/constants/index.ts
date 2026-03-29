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

/** 燃油级别（汽油/柴油）——按地区体系分组 */

/** 中国体系 */
export const FUEL_GRADES_CN = [
  { value: '92', label: 'fuelGrade.92' },
  { value: '95', label: 'fuelGrade.95' },
  { value: '98', label: 'fuelGrade.98' },
  { value: 'diesel_0', label: 'fuelGrade.diesel0' },
  { value: 'diesel_neg10', label: 'fuelGrade.dieselNeg10' },
  { value: 'diesel_neg20', label: 'fuelGrade.dieselNeg20' },
] as const;

/** 日本体系：レギュラー(RON89-90)、ハイオク(RON96-100)、軽油 */
export const FUEL_GRADES_JP = [
  { value: 'jp_regular', label: 'fuelGrade.jpRegular' },
  { value: 'jp_premium', label: 'fuelGrade.jpPremium' },
  { value: 'jp_diesel', label: 'fuelGrade.jpDiesel' },
] as const;

/** 美国/国际体系 */
export const FUEL_GRADES_US = [
  { value: 'regular', label: 'fuelGrade.regular' },
  { value: 'mid_grade', label: 'fuelGrade.midGrade' },
  { value: 'premium', label: 'fuelGrade.premium' },
  { value: 'super_premium', label: 'fuelGrade.superPremium' },
  { value: 'us_diesel', label: 'fuelGrade.usDiesel' },
] as const;

/** 所有燃油级别合集（向后兼容） */
export const FUEL_GRADES = [
  ...FUEL_GRADES_CN,
  ...FUEL_GRADES_JP,
  ...FUEL_GRADES_US,
] as const;

/** 按当前语言获取对应地区的燃油标号 */
export const getFuelGradesByLocale = (locale: string) => {
  switch (locale) {
    case 'ja-JP':
      return FUEL_GRADES_JP;
    case 'zh-CN':
      return FUEL_GRADES_CN;
    case 'en-US':
    default:
      return FUEL_GRADES_US;
  }
};

/** 判断是否为纯电车型 */
export const isElectricVehicle = (fuelType: string) => fuelType === 'electric';

/** 判断是否需要显示排量（非电动车型） */
export const hasEngineCC = (fuelType: string) => fuelType !== 'electric';

export const SUPPORTED_LOCALES = [
  { value: 'zh-CN', label: '简体中文' },
  { value: 'en-US', label: 'English' },
  { value: 'ja-JP', label: '日本語' },
] as const;

/** 全球时区列表（IANA 格式），按 UTC 偏移从西（−12）到东（+14）排列 */
export const TIMEZONES = [
  // UTC−12 ~ UTC−9
  { value: 'Pacific/Midway', label: 'timezone.pacificMidway' },
  { value: 'Pacific/Honolulu', label: 'timezone.pacificHonolulu' },
  { value: 'America/Anchorage', label: 'timezone.americaAnchorage' },
  // UTC−8
  { value: 'America/Los_Angeles', label: 'timezone.americaLosAngeles' },
  { value: 'America/Vancouver', label: 'timezone.americaVancouver' },
  { value: 'America/Tijuana', label: 'timezone.americaTijuana' },
  // UTC−7
  { value: 'America/Denver', label: 'timezone.americaDenver' },
  { value: 'America/Phoenix', label: 'timezone.americaPhoenix' },
  { value: 'America/Edmonton', label: 'timezone.americaEdmonton' },
  // UTC−6
  { value: 'America/Chicago', label: 'timezone.americaChicago' },
  { value: 'America/Mexico_City', label: 'timezone.americaMexicoCity' },
  { value: 'America/Winnipeg', label: 'timezone.americaWinnipeg' },
  { value: 'America/Costa_Rica', label: 'timezone.americaCostaRica' },
  // UTC−5
  { value: 'America/New_York', label: 'timezone.americaNewYork' },
  { value: 'America/Toronto', label: 'timezone.americaToronto' },
  { value: 'America/Bogota', label: 'timezone.americaBogota' },
  { value: 'America/Lima', label: 'timezone.americaLima' },
  { value: 'America/Panama', label: 'timezone.americaPanama' },
  // UTC−4
  { value: 'America/Halifax', label: 'timezone.americaHalifax' },
  { value: 'America/Caracas', label: 'timezone.americaCaracas' },
  { value: 'America/Santiago', label: 'timezone.americaSantiago' },
  { value: 'America/La_Paz', label: 'timezone.americaLaPaz' },
  // UTC−3
  { value: 'America/Sao_Paulo', label: 'timezone.americaSaoPaulo' },
  { value: 'America/Argentina/Buenos_Aires', label: 'timezone.americaBuenosAires' },
  { value: 'America/Montevideo', label: 'timezone.americaMontevideo' },
  // UTC−2 ~ UTC−1
  { value: 'Atlantic/South_Georgia', label: 'timezone.atlanticSouthGeorgia' },
  { value: 'Atlantic/Azores', label: 'timezone.atlanticAzores' },
  { value: 'Atlantic/Cape_Verde', label: 'timezone.atlanticCapeVerde' },
  // UTC+0
  { value: 'UTC', label: 'timezone.utc' },
  { value: 'Europe/London', label: 'timezone.europeLondon' },
  { value: 'Europe/Dublin', label: 'timezone.europeDublin' },
  { value: 'Europe/Lisbon', label: 'timezone.europeLisbon' },
  { value: 'Africa/Casablanca', label: 'timezone.africaCasablanca' },
  { value: 'Africa/Accra', label: 'timezone.africaAccra' },
  // UTC+1
  { value: 'Europe/Paris', label: 'timezone.europeParis' },
  { value: 'Europe/Berlin', label: 'timezone.europeBerlin' },
  { value: 'Europe/Madrid', label: 'timezone.europeMadrid' },
  { value: 'Europe/Rome', label: 'timezone.europeRome' },
  { value: 'Europe/Amsterdam', label: 'timezone.europeAmsterdam' },
  { value: 'Europe/Brussels', label: 'timezone.europeBrussels' },
  { value: 'Europe/Zurich', label: 'timezone.europeZurich' },
  { value: 'Europe/Vienna', label: 'timezone.europeVienna' },
  { value: 'Europe/Stockholm', label: 'timezone.europeStockholm' },
  { value: 'Europe/Oslo', label: 'timezone.europeOslo' },
  { value: 'Europe/Copenhagen', label: 'timezone.europeCopenhagen' },
  { value: 'Europe/Warsaw', label: 'timezone.europeWarsaw' },
  { value: 'Europe/Prague', label: 'timezone.europePrague' },
  { value: 'Africa/Lagos', label: 'timezone.africaLagos' },
  // UTC+2
  { value: 'Europe/Helsinki', label: 'timezone.europeHelsinki' },
  { value: 'Europe/Athens', label: 'timezone.europeAthens' },
  { value: 'Europe/Bucharest', label: 'timezone.europeBucharest' },
  { value: 'Europe/Istanbul', label: 'timezone.europeIstanbul' },
  { value: 'Europe/Kiev', label: 'timezone.europeKiev' },
  { value: 'Africa/Cairo', label: 'timezone.africaCairo' },
  { value: 'Africa/Johannesburg', label: 'timezone.africaJohannesburg' },
  { value: 'Asia/Jerusalem', label: 'timezone.asiaJerusalem' },
  { value: 'Asia/Beirut', label: 'timezone.asiaBeirut' },
  // UTC+3
  { value: 'Europe/Moscow', label: 'timezone.europeMoscow' },
  { value: 'Asia/Riyadh', label: 'timezone.asiaRiyadh' },
  { value: 'Africa/Nairobi', label: 'timezone.africaNairobi' },
  { value: 'Asia/Baghdad', label: 'timezone.asiaBaghdad' },
  { value: 'Asia/Kuwait', label: 'timezone.asiaKuwait' },
  // UTC+3:30
  { value: 'Asia/Tehran', label: 'timezone.asiaTehran' },
  // UTC+4
  { value: 'Asia/Dubai', label: 'timezone.asiaDubai' },
  { value: 'Asia/Baku', label: 'timezone.asiaBaku' },
  { value: 'Asia/Tbilisi', label: 'timezone.asiaTbilisi' },
  { value: 'Indian/Mauritius', label: 'timezone.indianMauritius' },
  // UTC+4:30
  { value: 'Asia/Kabul', label: 'timezone.asiaKabul' },
  // UTC+5
  { value: 'Asia/Karachi', label: 'timezone.asiaKarachi' },
  { value: 'Asia/Tashkent', label: 'timezone.asiaTashkent' },
  { value: 'Asia/Yekaterinburg', label: 'timezone.asiaYekaterinburg' },
  // UTC+5:30
  { value: 'Asia/Kolkata', label: 'timezone.asiaKolkata' },
  { value: 'Asia/Colombo', label: 'timezone.asiaColombo' },
  // UTC+5:45
  { value: 'Asia/Kathmandu', label: 'timezone.asiaKathmandu' },
  // UTC+6
  { value: 'Asia/Dhaka', label: 'timezone.asiaDhaka' },
  { value: 'Asia/Almaty', label: 'timezone.asiaAlmaty' },
  // UTC+6:30
  { value: 'Asia/Yangon', label: 'timezone.asiaYangon' },
  // UTC+7
  { value: 'Asia/Bangkok', label: 'timezone.asiaBangkok' },
  { value: 'Asia/Ho_Chi_Minh', label: 'timezone.asiaHoChiMinh' },
  { value: 'Asia/Jakarta', label: 'timezone.asiaJakarta' },
  { value: 'Asia/Novosibirsk', label: 'timezone.asiaNovosibirsk' },
  // UTC+8
  { value: 'Asia/Shanghai', label: 'timezone.asiaShanghai' },
  { value: 'Asia/Hong_Kong', label: 'timezone.asiaHongKong' },
  { value: 'Asia/Taipei', label: 'timezone.asiaTaipei' },
  { value: 'Asia/Singapore', label: 'timezone.asiaSingapore' },
  { value: 'Asia/Kuala_Lumpur', label: 'timezone.asiaKualaLumpur' },
  { value: 'Asia/Manila', label: 'timezone.asiaManila' },
  { value: 'Asia/Makassar', label: 'timezone.asiaMakassar' },
  { value: 'Australia/Perth', label: 'timezone.australiaPerth' },
  // UTC+9
  { value: 'Asia/Tokyo', label: 'timezone.asiaTokyo' },
  { value: 'Asia/Seoul', label: 'timezone.asiaSeoul' },
  { value: 'Asia/Jayapura', label: 'timezone.asiaJayapura' },
  { value: 'Asia/Yakutsk', label: 'timezone.asiaYakutsk' },
  // UTC+9:30
  { value: 'Australia/Adelaide', label: 'timezone.australiaAdelaide' },
  { value: 'Australia/Darwin', label: 'timezone.australiaDarwin' },
  // UTC+10
  { value: 'Australia/Sydney', label: 'timezone.australiaSydney' },
  { value: 'Australia/Brisbane', label: 'timezone.australiaBrisbane' },
  { value: 'Australia/Melbourne', label: 'timezone.australiaMelbourne' },
  { value: 'Australia/Hobart', label: 'timezone.australiaHobart' },
  { value: 'Pacific/Guam', label: 'timezone.pacificGuam' },
  { value: 'Pacific/Port_Moresby', label: 'timezone.pacificPortMoresby' },
  { value: 'Asia/Vladivostok', label: 'timezone.asiaVladivostok' },
  // UTC+11
  { value: 'Pacific/Noumea', label: 'timezone.pacificNoumea' },
  { value: 'Pacific/Guadalcanal', label: 'timezone.pacificGuadalcanal' },
  // UTC+12
  { value: 'Pacific/Auckland', label: 'timezone.pacificAuckland' },
  { value: 'Pacific/Fiji', label: 'timezone.pacificFiji' },
  { value: 'Asia/Kamchatka', label: 'timezone.asiaKamchatka' },
  // UTC+13
  { value: 'Pacific/Tongatapu', label: 'timezone.pacificTongatapu' },
  { value: 'Pacific/Apia', label: 'timezone.pacificApia' },
] as const;

export const DEFAULT_PAGE_SIZE = 20;
