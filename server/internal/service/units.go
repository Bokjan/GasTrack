package service

import (
	"gastrack/internal/model"
	"gastrack/internal/pkg/convert"
)

// energyKind 能源类别：燃油（L/gal，基准为升、L/100km）与电（kWh，基准为 kWh/100km）分开统计，不可相加
type energyKind int

const (
	energyLiquid energyKind = iota
	energyElectric
)

// vehicleEnergyKind 车辆的主要能源类别（电动车为电，其余为燃油）
func vehicleEnergyKind(v *model.Vehicle) energyKind {
	if v != nil && v.FuelType == model.FuelTypeElectric {
		return energyElectric
	}
	return energyLiquid
}

// recordEnergyKind 单条记录的能源类别（按加注单位区分）
func recordEnergyKind(fuelUnit string) energyKind {
	if convert.VolumeUnit(fuelUnit) == convert.UnitKWh {
		return energyElectric
	}
	return energyLiquid
}

// unitPrefs 用户的展示单位偏好
type unitPrefs struct {
	effUnit  convert.FuelEfficiencyUnit // 用户的油耗单位偏好
	imperial bool
}

func prefsOf(user *model.User) unitPrefs {
	p := unitPrefs{effUnit: convert.UnitL100km}
	if user == nil {
		return p
	}
	if user.FuelEfficiencyUnit != "" {
		p.effUnit = convert.FuelEfficiencyUnit(user.FuelEfficiencyUnit)
	}
	p.imperial = user.UnitSystem == "imperial"
	return p
}

// efficiency 将基准油耗/电耗转为展示值与单位标签
func (p unitPrefs) efficiency(base float64, kind energyKind) (float64, string) {
	if kind == energyElectric {
		return convert.ConvertEnergyEfficiency(base, p.effUnit), convert.EnergyEfficiencyUnitFor(p.effUnit)
	}
	return convert.ConvertFuelEfficiency(base, convert.UnitL100km, p.effUnit), string(p.effUnit)
}

// efficiencyUnit 展示单位标签
func (p unitPrefs) efficiencyUnit(kind energyKind) string {
	_, unit := p.efficiency(0, kind)
	return unit
}

// amount 将基准加注量（升或 kWh）转为展示值与单位标签
func (p unitPrefs) amount(base float64, kind energyKind) (float64, string) {
	if kind == energyElectric {
		return base, string(convert.UnitKWh)
	}
	if p.imperial {
		return convert.ConvertVolume(base, convert.UnitLiter, convert.UnitGallon), string(convert.UnitGallon)
	}
	return base, string(convert.UnitLiter)
}

// distance 将 km 转为展示距离
func (p unitPrefs) distance(km float64) float64 {
	if p.imperial {
		return convert.ConvertDistance(km, convert.UnitKm, convert.UnitMile)
	}
	return km
}
