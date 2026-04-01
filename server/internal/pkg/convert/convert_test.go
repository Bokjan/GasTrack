package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- 容量换算 ---

func TestLitersToGallons(t *testing.T) {
	tests := []struct {
		name   string
		liters float64
		want   float64
	}{
		{"zero", 0, 0},
		{"1 liter", 1, 0.264},
		{"3.78541 liters = 1 gallon", GallonToLiter, 1},
		{"10 liters", 10, 2.642},
		{"50 liters", 50, 13.209},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LitersToGallons(tt.liters)
			assert.InDelta(t, tt.want, got, 0.001)
		})
	}
}

func TestGallonsToLiters(t *testing.T) {
	tests := []struct {
		name    string
		gallons float64
		want    float64
	}{
		{"zero", 0, 0},
		{"1 gallon", 1, 3.785},
		{"5 gallons", 5, 18.927},
		{"10 gallons", 10, 37.854},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GallonsToLiters(tt.gallons)
			assert.InDelta(t, tt.want, got, 0.001)
		})
	}
}

// --- 距离换算 ---

func TestKmToMiles(t *testing.T) {
	tests := []struct {
		name string
		km   float64
		want float64
	}{
		{"zero", 0, 0},
		{"1 km", 1, 0.6},
		{"1.60934 km = 1 mile", MileToKm, 1},
		{"100 km", 100, 62.1},
		{"marathon", 42.195, 26.2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := KmToMiles(tt.km)
			assert.InDelta(t, tt.want, got, 0.1)
		})
	}
}

func TestMilesToKm(t *testing.T) {
	tests := []struct {
		name  string
		miles float64
		want  float64
	}{
		{"zero", 0, 0},
		{"1 mile", 1, 1.6},
		{"60 miles", 60, 96.6},
		{"100 miles", 100, 160.9},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MilesToKm(tt.miles)
			assert.InDelta(t, tt.want, got, 0.1)
		})
	}
}

// --- 油耗计算 ---

func TestCalcL100km(t *testing.T) {
	tests := []struct {
		name   string
		liters float64
		km     float64
		want   float64
	}{
		{"zero km", 10, 0, 0},
		{"negative km", 10, -5, 0},
		{"normal", 50, 600, 8.33},
		{"efficient", 30, 600, 5},
		{"gas guzzler", 80, 400, 20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalcL100km(tt.liters, tt.km)
			assert.InDelta(t, tt.want, got, 0.01)
		})
	}
}

func TestCalcKmL(t *testing.T) {
	tests := []struct {
		name   string
		liters float64
		km     float64
		want   float64
	}{
		{"zero liters", 0, 100, 0},
		{"negative liters", -5, 100, 0},
		{"normal", 50, 600, 12},
		{"efficient", 30, 600, 20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalcKmL(tt.liters, tt.km)
			assert.InDelta(t, tt.want, got, 0.01)
		})
	}
}

func TestCalcMPG(t *testing.T) {
	tests := []struct {
		name    string
		gallons float64
		miles   float64
		want    float64
	}{
		{"zero gallons", 0, 100, 0},
		{"negative gallons", -5, 100, 0},
		{"normal", 10, 300, 30},
		{"efficient", 5, 250, 50},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalcMPG(tt.gallons, tt.miles)
			assert.InDelta(t, tt.want, got, 0.01)
		})
	}
}

// --- 油耗互转 ---

func TestL100kmToKmL(t *testing.T) {
	tests := []struct {
		name   string
		l100km float64
		want   float64
	}{
		{"zero", 0, 0},
		{"negative", -5, 0},
		{"8 L/100km", 8, 12.5},
		{"5 L/100km", 5, 20},
		{"10 L/100km", 10, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := L100kmToKmL(tt.l100km)
			assert.InDelta(t, tt.want, got, 0.01)
		})
	}
}

func TestKmLToL100km(t *testing.T) {
	tests := []struct {
		name string
		kmL  float64
		want float64
	}{
		{"zero", 0, 0},
		{"negative", -5, 0},
		{"12.5 km/L", 12.5, 8},
		{"20 km/L", 20, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := KmLToL100km(tt.kmL)
			assert.InDelta(t, tt.want, got, 0.01)
		})
	}
}

func TestL100kmToMPG(t *testing.T) {
	tests := []struct {
		name   string
		l100km float64
		want   float64
	}{
		{"zero", 0, 0},
		{"negative", -5, 0},
		{"8 L/100km", 8, 29.4},
		{"5 L/100km", 5, 47.04},
		{"10 L/100km", 10, 23.52},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := L100kmToMPG(tt.l100km)
			assert.InDelta(t, tt.want, got, 0.01)
		})
	}
}

func TestMPGToL100km(t *testing.T) {
	tests := []struct {
		name string
		mpg  float64
		want float64
	}{
		{"zero", 0, 0},
		{"negative", -5, 0},
		{"30 MPG", 30, 7.84},
		{"50 MPG", 50, 4.7},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MPGToL100km(tt.mpg)
			assert.InDelta(t, tt.want, got, 0.01)
		})
	}
}

// --- 综合换算函数 ---

func TestConvertFuelEfficiency(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		from  FuelEfficiencyUnit
		to    FuelEfficiencyUnit
		want  float64
	}{
		{"same unit", 8, UnitL100km, UnitL100km, 8},
		{"L/100km → km/L", 8, UnitL100km, UnitKmL, 12.5},
		{"L/100km → MPG", 8, UnitL100km, UnitMPG, 29.4},
		{"km/L → L/100km", 12.5, UnitKmL, UnitL100km, 8},
		{"km/L → MPG", 12.5, UnitKmL, UnitMPG, 29.4},
		{"MPG → L/100km", 30, UnitMPG, UnitL100km, 7.84},
		{"MPG → km/L", 30, UnitMPG, UnitKmL, 12.76},
		{"unknown from unit", 10, "unknown", UnitL100km, 10},
		{"unknown to unit", 10, UnitL100km, "unknown", 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertFuelEfficiency(tt.value, tt.from, tt.to)
			assert.InDelta(t, tt.want, got, 0.02)
		})
	}
}

func TestConvertVolume(t *testing.T) {
	tests := []struct {
		name string
		val  float64
		from VolumeUnit
		to   VolumeUnit
		want float64
	}{
		{"same unit", 10, UnitLiter, UnitLiter, 10},
		{"L → gal", 10, UnitLiter, UnitGallon, 2.642},
		{"gal → L", 5, UnitGallon, UnitLiter, 18.927},
		{"unknown combo", 10, "oz", UnitLiter, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertVolume(tt.val, tt.from, tt.to)
			assert.InDelta(t, tt.want, got, 0.001)
		})
	}
}

func TestConvertDistance(t *testing.T) {
	tests := []struct {
		name string
		val  float64
		from DistanceUnit
		to   DistanceUnit
		want float64
	}{
		{"same unit", 100, UnitKm, UnitKm, 100},
		{"km → mi", 100, UnitKm, UnitMile, 62.1},
		{"mi → km", 60, UnitMile, UnitKm, 96.6},
		{"unknown combo", 10, "ft", UnitKm, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertDistance(tt.val, tt.from, tt.to)
			assert.InDelta(t, tt.want, got, 0.1)
		})
	}
}

func TestNormalizeToMetric(t *testing.T) {
	tests := []struct {
		name     string
		fuel     float64
		fuelUnit VolumeUnit
		dist     float64
		distUnit DistanceUnit
		wantL    float64
		wantKm   float64
	}{
		{"already metric", 50, UnitLiter, 600, UnitKm, 50, 600},
		{"gallons to liters", 10, UnitGallon, 600, UnitKm, 37.854, 600},
		{"miles to km", 50, UnitLiter, 100, UnitMile, 50, 160.9},
		{"both imperial", 10, UnitGallon, 100, UnitMile, 37.854, 160.9},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotL, gotKm := NormalizeToMetric(tt.fuel, tt.fuelUnit, tt.dist, tt.distUnit)
			assert.InDelta(t, tt.wantL, gotL, 0.001)
			assert.InDelta(t, tt.wantKm, gotKm, 0.1)
		})
	}
}

// --- 往返换算一致性 ---

func TestRoundTripConversions(t *testing.T) {
	t.Run("liters ↔ gallons", func(t *testing.T) {
		original := 42.5
		gallons := LitersToGallons(original)
		back := GallonsToLiters(gallons)
		assert.InDelta(t, original, back, 0.01)
	})

	t.Run("km ↔ miles", func(t *testing.T) {
		original := 100.0
		miles := KmToMiles(original)
		back := MilesToKm(miles)
		assert.InDelta(t, original, back, 0.2)
	})

	t.Run("L/100km ↔ km/L", func(t *testing.T) {
		original := 8.0
		kmL := L100kmToKmL(original)
		back := KmLToL100km(kmL)
		assert.InDelta(t, original, back, 0.01)
	})

	t.Run("L/100km ↔ MPG", func(t *testing.T) {
		original := 8.0
		mpg := L100kmToMPG(original)
		back := MPGToL100km(mpg)
		assert.InDelta(t, original, back, 0.02)
	})
}
