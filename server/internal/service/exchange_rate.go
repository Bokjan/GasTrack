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

	mu          sync.RWMutex
	cache       map[string]*exchangeRateCache // key: base currency code
	lastFailure map[string]time.Time          // 最近一次同步拉取失败的时间（用于失败退避）

	// fetchLocks 每个币种一把锁：缓存未命中时合并并发请求，只有一个请求真正调用外部 API
	fetchLocks map[string]*sync.Mutex

	stopCh chan struct{}
}

// fetchFailureBackoff 同步拉取失败后的退避时间，期间直接返回错误，避免每个请求都等待外部 API 超时
const fetchFailureBackoff = time.Minute

// supportedCurrencies 系统支持的 6 种货币
var supportedCurrencies = []string{"CNY", "USD", "EUR", "JPY", "GBP", "KRW"}

// NewExchangeRateService 创建 ExchangeRateService 实例
func NewExchangeRateService(cfg config.ExchangeRateConfig, logger *zap.Logger) *ExchangeRateService {
	fetchLocks := make(map[string]*sync.Mutex, len(supportedCurrencies))
	for _, c := range supportedCurrencies {
		fetchLocks[c] = &sync.Mutex{}
	}
	return &ExchangeRateService{
		cfg:         cfg,
		logger:      logger,
		client:      &http.Client{Timeout: cfg.Timeout},
		cache:       make(map[string]*exchangeRateCache),
		lastFailure: make(map[string]time.Time),
		fetchLocks:  fetchLocks,
		stopCh:      make(chan struct{}),
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

	if cached, fresh := s.cachedRates(base); fresh {
		return toRateResponse(cached), nil
	}

	// 缓存不存在或过期：同一币种同时只允许一个请求拉取
	lock := s.fetchLocks[base]
	lock.Lock()
	defer lock.Unlock()

	// 拿到锁后再检查一次：等待期间可能已被其他请求刷新
	cached, fresh := s.cachedRates(base)
	if fresh {
		return toRateResponse(cached), nil
	}

	s.mu.RLock()
	lastFailure := s.lastFailure[base]
	s.mu.RUnlock()

	var fetchErr error
	if time.Since(lastFailure) < fetchFailureBackoff {
		fetchErr = fmt.Errorf("exchange rate API recently failed, retrying after backoff")
	} else if fetchErr = s.fetchAndCache(base); fetchErr != nil {
		s.mu.Lock()
		s.lastFailure[base] = time.Now()
		s.mu.Unlock()
	}

	if fetchErr != nil {
		// 如果有旧缓存，降级返回旧数据
		if cached != nil {
			s.logger.Warn("using stale exchange rate cache", zap.String("base", base), zap.Error(fetchErr))
			return toRateResponse(cached), nil
		}
		return nil, fmt.Errorf("failed to fetch exchange rates: %w", fetchErr)
	}

	cached, _ = s.cachedRates(base)
	return toRateResponse(cached), nil
}

// cachedRates 返回缓存（可能为 nil）及其是否仍在有效期内
func (s *ExchangeRateService) cachedRates(base string) (*exchangeRateCache, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cached, ok := s.cache[base]
	if !ok {
		return nil, false
	}
	return cached, time.Since(cached.fetchedAt) < s.cfg.RefreshInterval*2
}

func toRateResponse(c *exchangeRateCache) *dto.ExchangeRateResponse {
	return &dto.ExchangeRateResponse{
		Base:  c.base,
		Date:  c.date,
		Rates: c.rates,
	}
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
