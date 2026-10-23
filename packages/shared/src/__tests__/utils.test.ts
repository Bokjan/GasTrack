// 在任何 dayjs 调用之前固定本地时区，模拟"设备时区 ≠ 资料时区"
process.env.TZ = 'Europe/Berlin';

import { afterEach, describe, expect, it } from 'vitest';
import {
  convertFuelEfficiency,
  convertViaBase,
  formatCurrency,
  fromPickerValue,
  gallonsToLiters,
  kmToMiles,
  litersToGallons,
  milesToKm,
  toPickerValue,
} from '../utils';
import i18n from '../i18n';

afterEach(async () => {
  await i18n.changeLanguage('zh-CN');
});

describe('unit conversion', () => {
  it('converts between metric and imperial', () => {
    expect(kmToMiles(100)).toBeCloseTo(62.137, 3);
    expect(milesToKm(kmToMiles(5000))).toBeCloseTo(5000, 0);
    expect(gallonsToLiters(litersToGallons(50))).toBeCloseTo(50, 1);
  });

  it('converts fuel efficiency units', () => {
    expect(convertFuelEfficiency(8, 'L/100km', 'km/L')).toBe(12.5);
    expect(convertFuelEfficiency(10, 'L/100km', 'MPG')).toBeCloseTo(23.52, 2);
    expect(convertFuelEfficiency(15, 'kWh/100km', 'MPG')).toBe(15);
  });
});

describe('convertViaBase', () => {
  // 1 CNY = 0.14 USD = 20 JPY
  const rates = { USD: 0.14, JPY: 20 };

  it('converts from and to the base currency', () => {
    expect(convertViaBase(100, 'CNY', 'USD', 'CNY', rates)).toBeCloseTo(14);
    expect(convertViaBase(14, 'USD', 'CNY', 'CNY', rates)).toBeCloseTo(100);
  });

  it('converts between two non-base currencies', () => {
    expect(convertViaBase(2000, 'JPY', 'USD', 'CNY', rates)).toBeCloseTo(14);
  });

  it('returns null when a rate is missing', () => {
    expect(convertViaBase(10, 'EUR', 'CNY', 'CNY', rates)).toBeNull();
    expect(convertViaBase(10, 'EUR', 'EUR', 'CNY', rates)).toBe(10);
  });
});

describe('formatCurrency', () => {
  it('distinguishes CNY and JPY in each locale', async () => {
    await i18n.changeLanguage('zh-CN');
    expect(formatCurrency(1234.5, 'CNY')).toBe('¥1,234.50');
    expect(formatCurrency(1234.5, 'JPY')).toBe('JP¥1,235');
    await i18n.changeLanguage('en-US');
    expect(formatCurrency(1234.5, 'CNY')).toBe('CN¥1,234.50');
    expect(formatCurrency(1234.5, 'JPY')).toBe('¥1,235');
  });

  it('falls back for invalid currency codes', () => {
    expect(formatCurrency(1, 'NOT_A_CODE')).toBe('NOT_A_CODE 1.00');
  });
});

describe('date picker timezone helpers', () => {
  it('shows and submits wall-clock time in the profile timezone', () => {
    const picker = toPickerValue('2026-03-01T15:30:00Z', 'Asia/Tokyo');
    expect(picker.format('YYYY-MM-DD HH:mm')).toBe('2026-03-02 00:30');
    expect(fromPickerValue(picker, 'Asia/Tokyo')).toBe('2026-03-02T00:30:00+09:00');
  });

  it('uses the device timezone without a profile timezone', () => {
    const picker = toPickerValue('2026-03-01T15:30:00Z');
    expect(fromPickerValue(picker)).toBe('2026-03-01T16:30:00+01:00');
  });
});
