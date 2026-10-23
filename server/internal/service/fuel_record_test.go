package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gastrack/internal/model"
)

func fuelRec(odometer, liters float64, full bool) model.FuelRecord {
	return model.FuelRecord{
		FuelAmount:   liters,
		FuelUnit:     "L",
		Odometer:     odometer,
		DistanceUnit: "km",
		IsFullTank:   full,
	}
}

func TestComputeDerivedFields(t *testing.T) {
	t.Run("full to full with partial fill in between", func(t *testing.T) {
		records := []model.FuelRecord{
			fuelRec(1000, 40, true), // 基准满油
			fuelRec(1200, 10, false),
			fuelRec(1500, 30, true), // 区间油量 10+30=40L，里程 500km → 8 L/100km
		}
		changed := computeDerivedFields(records)

		assert.Equal(t, []int{1, 2}, changed)
		assert.Equal(t, 0.0, records[0].TripDistance)
		assert.Equal(t, 0.0, records[0].FuelEfficiency)
		assert.Equal(t, 200.0, records[1].TripDistance)
		assert.Equal(t, 0.0, records[1].FuelEfficiency)
		assert.Equal(t, 300.0, records[2].TripDistance)
		assert.Equal(t, 8.0, records[2].FuelEfficiency)
	})

	t.Run("partial fills before first full tank are ignored", func(t *testing.T) {
		records := []model.FuelRecord{
			fuelRec(900, 15, false),
			fuelRec(1000, 40, true),
			fuelRec(1400, 32, true),
		}
		computeDerivedFields(records)

		assert.Equal(t, 0.0, records[1].FuelEfficiency)
		assert.Equal(t, 8.0, records[2].FuelEfficiency)
	})

	t.Run("stale values are reset", func(t *testing.T) {
		stale := fuelRec(1500, 30, false) // 曾经是加满且有油耗，改为未加满后应清零
		stale.TripDistance = 999
		stale.FuelEfficiency = 6.5
		records := []model.FuelRecord{fuelRec(1000, 40, true), stale}

		changed := computeDerivedFields(records)

		assert.Equal(t, []int{1}, changed)
		assert.Equal(t, 500.0, records[1].TripDistance)
		assert.Equal(t, 0.0, records[1].FuelEfficiency)
	})

	t.Run("odometer going backwards yields no distance or efficiency", func(t *testing.T) {
		records := []model.FuelRecord{fuelRec(1000, 40, true), fuelRec(990, 30, true)}
		computeDerivedFields(records)

		assert.Equal(t, 0.0, records[1].TripDistance)
		assert.Equal(t, 0.0, records[1].FuelEfficiency)
	})

	t.Run("mixed units are normalized to metric", func(t *testing.T) {
		second := model.FuelRecord{FuelAmount: 10, FuelUnit: "gal", Odometer: 1000, DistanceUnit: "mi", IsFullTank: true}
		records := []model.FuelRecord{fuelRec(1300, 40, true), second}
		computeDerivedFields(records)

		// 1000 mi = 1609.3 km → 309.3 km；10 gal = 37.854 L → 12.24 L/100km
		assert.Equal(t, 309.3, records[1].TripDistance)
		assert.Equal(t, 12.24, records[1].FuelEfficiency)
	})

	t.Run("unchanged records are not reported", func(t *testing.T) {
		records := []model.FuelRecord{fuelRec(1000, 40, true), fuelRec(1500, 40, true)}
		computeDerivedFields(records)

		assert.Empty(t, computeDerivedFields(records))
	})
}
