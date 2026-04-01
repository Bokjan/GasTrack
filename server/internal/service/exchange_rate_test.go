package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"gastrack/internal/config"
)

func TestExchangeRateService_GetRates(t *testing.T) {
	t.Run("success from API", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := frankfurterResponse{
				Base:  "USD",
				Date:  "2026-04-01",
				Rates: map[string]float64{"CNY": 7.25, "EUR": 0.92, "JPY": 150.5, "GBP": 0.79, "KRW": 1350.0},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		svc := NewExchangeRateService(config.ExchangeRateConfig{
			APIURL:          server.URL,
			RefreshInterval: 1 * time.Hour,
			Timeout:         5 * time.Second,
		}, zap.NewNop())

		resp, err := svc.GetRates("USD")
		assert.NoError(t, err)
		assert.Equal(t, "USD", resp.Base)
		assert.Equal(t, "2026-04-01", resp.Date)
		assert.InDelta(t, 7.25, resp.Rates["CNY"], 0.01)
	})

	t.Run("cached response", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			resp := frankfurterResponse{
				Base:  "EUR",
				Date:  "2026-04-01",
				Rates: map[string]float64{"USD": 1.09},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		svc := NewExchangeRateService(config.ExchangeRateConfig{
			APIURL:          server.URL,
			RefreshInterval: 1 * time.Hour,
			Timeout:         5 * time.Second,
		}, zap.NewNop())

		// First call - fetches from API
		_, err := svc.GetRates("EUR")
		assert.NoError(t, err)
		assert.Equal(t, 1, callCount)

		// Second call - uses cache
		resp, err := svc.GetRates("EUR")
		assert.NoError(t, err)
		assert.Equal(t, "EUR", resp.Base)
		assert.Equal(t, 1, callCount) // no extra API call
	})

	t.Run("unsupported currency", func(t *testing.T) {
		svc := NewExchangeRateService(config.ExchangeRateConfig{
			APIURL:          "http://localhost",
			RefreshInterval: 1 * time.Hour,
			Timeout:         5 * time.Second,
		}, zap.NewNop())

		_, err := svc.GetRates("XYZ")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported currency")
	})

	t.Run("API error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		svc := NewExchangeRateService(config.ExchangeRateConfig{
			APIURL:          server.URL,
			RefreshInterval: 1 * time.Hour,
			Timeout:         5 * time.Second,
		}, zap.NewNop())

		_, err := svc.GetRates("USD")
		assert.Error(t, err)
	})
}

func TestExchangeRateService_StartStop(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := frankfurterResponse{Base: "USD", Date: "2026-04-01", Rates: map[string]float64{"CNY": 7.25}}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := NewExchangeRateService(config.ExchangeRateConfig{
		APIURL:          server.URL,
		RefreshInterval: 100 * time.Millisecond,
		Timeout:         5 * time.Second,
	}, zap.NewNop())

	svc.Start()
	time.Sleep(200 * time.Millisecond) // let it run at least once
	svc.Stop()
	// Should not panic or deadlock
}
