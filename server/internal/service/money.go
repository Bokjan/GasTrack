package service

import (
	"sort"
	"strings"

	"gastrack/internal/dto"
)

// RateProvider 参考汇率来源（由 ExchangeRateService 实现）
type RateProvider interface {
	GetRates(base string) (*dto.ExchangeRateResponse, error)
}

// moneyConverter 将不同币种的金额按参考汇率换算为同一基准币种后再汇总。
// 无汇率（服务不可用或币种不受支持）的金额不计入换算结果，并记录在 unconverted 中。
type moneyConverter struct {
	base        string
	rates       map[string]float64 // 1 base = rates[X] 个 X
	unconverted map[string]struct{}
}

func newMoneyConverter(provider RateProvider, base string) *moneyConverter {
	c := &moneyConverter{base: strings.ToUpper(base), unconverted: map[string]struct{}{}}
	if provider != nil && c.base != "" {
		if resp, err := provider.GetRates(c.base); err == nil && resp != nil {
			c.rates = resp.Rates
		}
	}
	return c
}

// convert 换算单笔金额；ok 为 false 表示无法换算
func (c *moneyConverter) convert(amount float64, currency string) (float64, bool) {
	currency = strings.ToUpper(currency)
	if amount == 0 || currency == c.base {
		return amount, true
	}
	if rate, ok := c.rates[currency]; ok && rate > 0 {
		return amount / rate, true
	}
	c.unconverted[currency] = struct{}{}
	return 0, false
}

// sum 换算并汇总按币种分组的金额
func (c *moneyConverter) sum(byCurrency map[string]float64) float64 {
	var total float64
	for cur, amount := range byCurrency {
		if v, ok := c.convert(amount, cur); ok {
			total += v
		}
	}
	return total
}

// unconvertedCurrencies 未能换算（因而未计入总额）的币种，按字母排序
func (c *moneyConverter) unconvertedCurrencies() []string {
	if len(c.unconverted) == 0 {
		return nil
	}
	out := make([]string, 0, len(c.unconverted))
	for cur := range c.unconverted {
		out = append(out, cur)
	}
	sort.Strings(out)
	return out
}
