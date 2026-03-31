package handler

import (
	"net/http"

	"go.uber.org/zap"

	"gastrack/internal/middleware"
	"gastrack/internal/pkg/decode"
	"gastrack/internal/pkg/respond"
	"gastrack/internal/service"
)

// ExchangeRateHandler 汇率参考 HTTP 处理器
type ExchangeRateHandler struct {
	exchangeRateService *service.ExchangeRateService
	logger              *zap.Logger
}

// NewExchangeRateHandler 创建 ExchangeRateHandler 实例
func NewExchangeRateHandler(exchangeRateService *service.ExchangeRateService, logger *zap.Logger) *ExchangeRateHandler {
	return &ExchangeRateHandler{exchangeRateService: exchangeRateService, logger: logger}
}

// GetRates 获取汇率参考数据
// GET /api/v1/exchange-rates?base=CNY
func (h *ExchangeRateHandler) GetRates(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserID(r.Context())
	if !ok {
		respond.Unauthorized(w, "missing user identity")
		return
	}

	base := decode.QueryString(r, "base", "USD")

	result, err := h.exchangeRateService.GetRates(base)
	if err != nil {
		h.logger.Warn("failed to get exchange rates", zap.String("base", base), zap.Error(err))
		respond.BadRequest(w, err.Error())
		return
	}

	respond.OK(w, result)
}
