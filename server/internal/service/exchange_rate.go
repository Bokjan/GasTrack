package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"gastrack/internal/config"
	"gastrack/internal/dto"
)

// frankfurterResponse frankfurter.app API 响应结构
type frankfurterResponse struct {
	Base  string             `json:"base"`
	Date  string             `json:"date"`
	Rates map[string]float64 `json:"rates"`
}

// exchangeRateCache 汇率缓存数据
type exchangeRateCache struct {
	base      string
	date      string
	rates     map[string]float64
	fetchedAt time.Time
}

// ExchangeRateService 汇率参考服务（只读展示，不做实时兑换）
type ExchangeRateService struct {
	cfg    config.ExchangeRateConfig
	logger *zap.Logger
	client *http.Client

	mu    sync.RWMutex
	cache map[string]*exchangeRateCache // key: base currency code

	stopCh chan struct{}
}

// supportedCurrencies 系统支持的 6 种货币
var supportedCurrencies = []string{"CNY", "USD", "EUR", "JPY", "GBP", "KRW"}

// NewExchangeRateService 创建 ExchangeRateService 实例
func NewExchangeRateService(cfg config.ExchangeRateConfig, logger *zap.Logger) *ExchangeRateService {
	return &ExchangeRateService{
		cfg:    cfg,
		logger: logger,
		client: &http.Client{Timeout: cfg.Timeout},
		cache:  make(map[string]*exchangeRateCache),
		stopCh: make(chan struct{}),
	}
}

// Start 启动后台定时刷新（异步，不阻塞主流程）
func (s *ExchangeRateService) Start() {
	go func() {
		// 启动时立即拉取一次
		s.refreshAll()

		ticker := time.NewTicker(s.cfg.RefreshInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.refreshAll()
			case <-s.stopCh:
				return
			}
		}
	}()
}

// Stop 停止后台刷新
func (s *ExchangeRateService) Stop() {
	close(s.stopCh)
}

// GetRates 获取以 base 为基准的汇率数据
func (s *ExchangeRateService) GetRates(base string) (*dto.ExchangeRateResponse, error) {
	base = strings.ToUpper(base)

	// 检查是否为支持的币种
	if !isSupportedCurrency(base) {
		return nil, fmt.Errorf("unsupported currency: %s", base)
	}

	s.mu.RLock()
	cached, ok := s.cache[base]
	s.mu.RUnlock()

	if ok && time.Since(cached.fetchedAt) < s.cfg.RefreshInterval*2 {
		return &dto.ExchangeRateResponse{
			Base:  cached.base,
			Date:  cached.date,
			Rates: cached.rates,
		}, nil
	}

	// 缓存不存在或过期，尝试同步拉取
	if err := s.fetchAndCache(base); err != nil {
		// 如果有旧缓存，降级返回旧数据
		if ok {
			s.logger.Warn("using stale exchange rate cache", zap.String("base", base), zap.Error(err))
			return &dto.ExchangeRateResponse{
				Base:  cached.base,
				Date:  cached.date,
				Rates: cached.rates,
			}, nil
		}
		return nil, fmt.Errorf("failed to fetch exchange rates: %w", err)
	}

	s.mu.RLock()
	cached = s.cache[base]
	s.mu.RUnlock()

	return &dto.ExchangeRateResponse{
		Base:  cached.base,
		Date:  cached.date,
		Rates: cached.rates,
	}, nil
}

// refreshAll 刷新所有支持币种的汇率缓存
func (s *ExchangeRateService) refreshAll() {
	for _, currency := range supportedCurrencies {
		if err := s.fetchAndCache(currency); err != nil {
			s.logger.Warn("failed to refresh exchange rate",
				zap.String("base", currency),
				zap.Error(err),
			)
		}
	}
	s.logger.Info("exchange rates refreshed")
}

// fetchAndCache 从 frankfurter.app 拉取汇率并写入缓存
func (s *ExchangeRateService) fetchAndCache(base string) error {
	// 构建目标币种列表（排除自身）
	targets := make([]string, 0, len(supportedCurrencies)-1)
	for _, c := range supportedCurrencies {
		if c != base {
			targets = append(targets, c)
		}
	}

	url := fmt.Sprintf("%s/latest?from=%s&to=%s", s.cfg.APIURL, base, strings.Join(targets, ","))

	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("calling frankfurter API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("frankfurter API returned status %d", resp.StatusCode)
	}

	var result frankfurterResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	s.mu.Lock()
	s.cache[base] = &exchangeRateCache{
		base:      result.Base,
		date:      result.Date,
		rates:     result.Rates,
		fetchedAt: time.Now(),
	}
	s.mu.Unlock()

	return nil
}

// isSupportedCurrency 检查是否为系统支持的币种
func isSupportedCurrency(code string) bool {
	for _, c := range supportedCurrencies {
		if c == code {
			return true
		}
	}
	return false
}
