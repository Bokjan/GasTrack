package dto

// --- 汇率参考相关 DTO ---

// ExchangeRateResponse 汇率参考响应
type ExchangeRateResponse struct {
	Base  string             `json:"base"`  // 基准币种 "CNY"
	Date  string             `json:"date"`  // 汇率日期 "2026-03-31"
	Rates map[string]float64 `json:"rates"` // { "USD": 0.138, "EUR": 0.126, ... }
}
