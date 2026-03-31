// Package convert 提供燃油相关的单位换算功能。
// 支持三种油耗体系：L/100km（公制欧标）、km/L（公制日标）、MPG（英制）。
package convert

import "math"

// 换算常量
const (
	GallonToLiter  = 3.78541 // 1 US gallon = 3.78541 liters
	LiterToGallon  = 1.0 / GallonToLiter
	MileToKm       = 1.60934 // 1 mile = 1.60934 km
	KmToMile       = 1.0 / MileToKm
	L100kmMPGFactor = 235.215 // L/100km ↔ MPG 互转因子: MPG = 235.215 / L100km
)

// FuelEfficiencyUnit 油耗单位类型
type FuelEfficiencyUnit string

const (
	UnitL100km FuelEfficiencyUnit = "L/100km" // 公制欧标：每百公里油耗（越小越省）
	UnitKmL    FuelEfficiencyUnit = "km/L"    // 公制日标：每升行驶公里数（越大越省）
	UnitMPG    FuelEfficiencyUnit = "MPG"     // 英制：每加仑英里数（越大越省）
)

// VolumeUnit 容量单位
type VolumeUnit string

const (
	UnitLiter  VolumeUnit = "L"
	UnitGallon VolumeUnit = "gal"
)

// DistanceUnit 距离单位
type DistanceUnit string

const (
	UnitKm   DistanceUnit = "km"
	UnitMile DistanceUnit = "mi"
)

// --- 容量换算 ---

// LitersToGallons 升 → 加仑
func LitersToGallons(liters float64) float64 {
	return round(liters*LiterToGallon, 3)
}

// GallonsToLiters 加仑 → 升
func GallonsToLiters(gallons float64) float64 {
	return round(gallons*GallonToLiter, 3)
}

// --- 距离换算 ---

// KmToMiles 公里 → 英里
func KmToMiles(km float64) float64 {
	return round(km*KmToMile, 1)
}

// MilesToKm 英里 → 公里
func MilesToKm(miles float64) float64 {
	return round(miles*MileToKm, 1)
}

// --- 油耗换算 ---
// 核心换算以 L/100km 为基准单位（存储统一用此单位）

// CalcL100km 根据油量(L)和距离(km)计算 L/100km
func CalcL100km(liters, km float64) float64 {
	if km <= 0 {
		return 0
	}
	return round(liters/km*100, 2)
}

// CalcKmL 根据油量(L)和距离(km)计算 km/L
func CalcKmL(liters, km float64) float64 {
	if liters <= 0 {
		return 0
	}
	return round(km/liters, 2)
}

// CalcMPG 根据油量(gal)和距离(mi)计算 MPG
func CalcMPG(gallons, miles float64) float64 {
	if gallons <= 0 {
		return 0
	}
	return round(miles/gallons, 2)
}

// L100kmToKmL 将 L/100km 转为 km/L
func L100kmToKmL(l100km float64) float64 {
	if l100km <= 0 {
		return 0
	}
	return round(100/l100km, 2)
}

// KmLToL100km 将 km/L 转为 L/100km
func KmLToL100km(kmL float64) float64 {
	if kmL <= 0 {
		return 0
	}
	return round(100/kmL, 2)
}

// L100kmToMPG 将 L/100km 转为 MPG
func L100kmToMPG(l100km float64) float64 {
	if l100km <= 0 {
		return 0
	}
	// MPG = L100kmMPGFactor / L/100km
	return round(L100kmMPGFactor/l100km, 2)
}

// MPGToL100km 将 MPG 转为 L/100km
func MPGToL100km(mpg float64) float64 {
	if mpg <= 0 {
		return 0
	}
	return round(L100kmMPGFactor/mpg, 2)
}

// ConvertFuelEfficiency 将油耗值从一种单位转为另一种
func ConvertFuelEfficiency(value float64, from, to FuelEfficiencyUnit) float64 {
	if from == to {
		return value
	}

	// 先转为 L/100km 基准
	var l100km float64
	switch from {
	case UnitL100km:
		l100km = value
	case UnitKmL:
		l100km = KmLToL100km(value)
	case UnitMPG:
		l100km = MPGToL100km(value)
	default:
		return value
	}

	// 再从 L/100km 转为目标单位
	switch to {
	case UnitL100km:
		return round(l100km, 2)
	case UnitKmL:
		return L100kmToKmL(l100km)
	case UnitMPG:
		return L100kmToMPG(l100km)
	default:
		return value
	}
}

// NormalizeToMetric 将任意单位的油量和距离统一转为公制（升/公里）
// 用于统计计算前的标准化
func NormalizeToMetric(fuelAmount float64, fuelUnit VolumeUnit, distance float64, distUnit DistanceUnit) (liters, km float64) {
	// 转升
	switch fuelUnit {
	case UnitGallon:
		liters = GallonsToLiters(fuelAmount)
	default:
		liters = fuelAmount
	}

	// 转公里
	switch distUnit {
	case UnitMile:
		km = MilesToKm(distance)
	default:
		km = distance
	}

	return liters, km
}

// ConvertVolume 将容量从一个单位转为另一个单位
func ConvertVolume(value float64, from, to VolumeUnit) float64 {
	if from == to {
		return value
	}
	switch {
	case from == UnitLiter && to == UnitGallon:
		return LitersToGallons(value)
	case from == UnitGallon && to == UnitLiter:
		return GallonsToLiters(value)
	default:
		return value
	}
}

// ConvertDistance 将距离从一个单位转为另一个单位
func ConvertDistance(value float64, from, to DistanceUnit) float64 {
	if from == to {
		return value
	}
	switch {
	case from == UnitKm && to == UnitMile:
		return KmToMiles(value)
	case from == UnitMile && to == UnitKm:
		return MilesToKm(value)
	default:
		return value
	}
}

// round 四舍五入到指定小数位
func round(val float64, precision int) float64 {
	p := math.Pow(10, float64(precision))
	return math.Round(val*p) / p
}
