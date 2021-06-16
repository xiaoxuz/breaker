package breaker

import (
	"fmt"
	"sync"
	"time"
)

type Breaker struct {
	m            sync.Mutex
	state        State
	Metrics      *Metrics
	Strategy     *BreakStrategyConfig
	HalfMaxCalls int64
	OpenTime     time.Time
	CoolingTime  time.Duration
}
type Config struct {
	HalfMaxCalls int64
	WindowSize   time.Duration
	Strategy     *BreakStrategyConfig
	CoolingTime  time.Duration
}

const (
	DEFAULT_HALFMAXCALLS int64         = 10
	DEFAULT_WINDOWSIZE   time.Duration = time.Second
	DEFAULT_COOLINGTIME  time.Duration = 100 * time.Millisecond
)

var DEFAULT_STRATEGY = &BreakStrategyConfig{
	BreakStrategy:              BREAK_STRATEGY_FAILCNT,
	FailCntThreshold:           20,
	ContinuousFailCntThreshold: 0,
	FailRate:                   0,
}

func NewBreaker(c Config) *Breaker {
	if c.HalfMaxCalls <= 0 {
		c.HalfMaxCalls = DEFAULT_HALFMAXCALLS
	}
	if c.CoolingTime == 0 {
		c.CoolingTime = DEFAULT_COOLINGTIME
	}
	if c.WindowSize == 0 {
		c.WindowSize = DEFAULT_WINDOWSIZE
	}
	if c.Strategy == nil {
		c.Strategy = DEFAULT_STRATEGY
	}
	b := &Breaker{
		m:     sync.Mutex{},
		state: STATE_CLOSED,
		Metrics: &Metrics{
			MetricsID: 0,
			Win:       &Window{
				WindowSize:      c.WindowSize,
				WindowStartTime: time.Time{},
			},
			Norm:      &Norm{},
		},
		Strategy:     c.Strategy,
		HalfMaxCalls: c.HalfMaxCalls,
		OpenTime:     time.Time{},
		CoolingTime:  c.CoolingTime,
	}

	return b
}

func (b *Breaker) Call(f func() (interface{}, error)) (interface{}, error) {
	// lock
	b.m.Lock()
	defer b.m.Unlock()

	// 前置检查
	if err := b.Before(); err != nil {
		return nil, err
	}

	// call
	b.Metrics.Call()
	response, err := f()

	// 后置处理
	b.After(err == nil)

	return response, nil
}

func (b *Breaker) Before() error {
	fmt.Println("Before Call 前置检查是否需要半开熔断或者当前流量是否命中熔断 or 半开熔断")
	return nil
}

func (b *Breaker) After(response bool) error {
	fmt.Println(fmt.Sprintf("After Call 后置根据请求结果[%t]判断是否需要切换状态", response))

	// 请求失败
	if true == response {
		// Succ 计数+1
		b.Metrics.Succ()

		// 如果当前熔断器为半开状态，并且连续成功数达到阈值，那么状态机需要流转到关闭状态
		if b.state == STATE_HALFOPEN && b.Metrics.Norm.ContinuousSuccCnt >= b.HalfMaxCalls {
			b.Change(STATE_CLOSED, time.Now())
		}
	} else {
		// Fail 计数+1
		b.Metrics.Fail()

		// todo
	}
	return nil
}

// 状态流转
func (b *Breaker) Change(state State, now time.Time) {
	// 切换状态
	switch state {
	case STATE_OPEN:
		b.OpenTime = now // 更新熔断器打开时间
		b.state = state
		break
	case STATE_HALFOPEN:
	case STATE_CLOSED:
		b.state = state
	case b.state:
	default:
		return
	}

	// 重启计数器
	b.Metrics.Restart(now)
}
