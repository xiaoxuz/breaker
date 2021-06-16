package breaker

import (
	"fmt"
	"testing"
	"time"
)

func TestBreak(t *testing.T) {
	breaker := NewBreaker(Config{
		HalfMaxCalls: 10,
		WindowSize:   2 * time.Second,
		Strategy:     &BreakStrategyConfig{
			BreakStrategy:              BREAK_STRATEGY_FAILCNT,
			FailCntThreshold:           5,
		},
		CoolingTime:  5 * time.Second,
	})

	resp, err := breaker.Call(func() (i interface{}, err error) {
		return nil, nil
	})
	fmt.Println(resp, err)
	fmt.Println(breaker.Metrics.Norm)
}

