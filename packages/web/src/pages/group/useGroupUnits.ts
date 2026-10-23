import { useAuthStore, convertEfficiencyForDisplay, litersToGallons, kmToMiles } from '@gastrack/shared';

/** 群组统计的单位换算：后端统一返回 L、kWh、km、L/100km（kWh/100km），按用户偏好展示 */
export function useGroupUnits() {
  const user = useAuthStore((s) => s.user);
  const currency = user?.currency_code || 'CNY';
  const isImperial = user?.unit_system === 'imperial';
  const efficiencyUnit = user?.fuel_efficiency_unit || 'L/100km';
  const fuelUnit = isImperial ? 'gal' : 'L';
  const distanceUnit = isImperial ? 'mi' : 'km';

  /** 加注量（燃油为 L，电动车为 kWh） */
  const fuelDisplay = (amount: number, isEv = false): { value: number; unit: string } =>
    isEv ? { value: amount, unit: 'kWh' } : { value: isImperial ? litersToGallons(amount) : amount, unit: fuelUnit };

  /** 油量 (L) */
  const convertFuel = (liters: number): number => fuelDisplay(liters).value;

  /** 距离 (km) */
  const convertDistance = (km: number): number => (isImperial ? kmToMiles(km) : km);

  /** 基准油耗（L/100km）或电耗（kWh/100km） */
  const effDisplay = (base: number, isEv = false): { value: number; unit: string } =>
    convertEfficiencyForDisplay(base, efficiencyUnit, isEv);

  /** 油耗 (L/100km) */
  const convertEfficiency = (l100km: number): number => effDisplay(l100km).value;

  /** 加油站单价：后端统一为每升（或每 kWh），英制显示为每加仑 */
  const stationPrice = (price: number, priceUnit?: string): { value: number; unit: string } => {
    if (priceUnit === 'kWh') return { value: price, unit: 'kWh' };
    return isImperial ? { value: price * 3.78541, unit: 'gal' } : { value: price, unit: 'L' };
  };

  return {
    user,
    currency,
    efficiencyUnit,
    fuelUnit,
    distanceUnit,
    fuelDisplay,
    convertFuel,
    convertDistance,
    effDisplay,
    convertEfficiency,
    stationPrice,
  };
}
