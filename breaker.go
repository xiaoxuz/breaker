package breaker

import (
	"errors"
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

var (
	DEFAULT_STRATEGY = &BreakStrategyConfig{
		BreakStrategy:              BREAK_STRATEGY_FAILCNT,
		FailCntThreshold:           20,
		ContinuousFailCntThreshold: 0,
		FailRate:                   0,
	}

	// error
	ERR_SERVICE_BREAK          = errors.New("service break")
	ERR_SERVICE_BREAK_HALFOPEN = errors.New("service halfopen break")
)

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
			Win: &Window{
				Size:      c.WindowSize,
				StartTime: time.Now(),
			},
			Norm: &Norm{},
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
	now := time.Now()

	switch b.state {
	case STATE_OPEN:
		//fmt.Println(b.OpenTime)
		// 如果超过冷却期,则调整为半开状态
		if b.OpenTime.Add(b.CoolingTime).Before(now) {
			b.Change(STATE_HALFOPEN, now)
			return nil
		}
		// 如果未过冷却期则拒绝服务
		return ERR_SERVICE_BREAK
		break
	case STATE_HALFOPEN:
		// 如果请求数超过半开上限，则拒绝服务
		if b.Metrics.Norm.AllCnt >= b.HalfMaxCalls {
			return ERR_SERVICE_BREAK_HALFOPEN
		}
		break
	//case STATE_CLOSED:
	default:
		// 如果时间窗口开始时间小于当前时间,则属于执行滑动窗口
		if b.Metrics.Win.StartTime.Before(now) {
			b.Metrics.Restart(now.Add(b.Metrics.Win.Size))
		}
		return nil
	}
	return nil
}

func (b *Breaker) After(response bool) error {
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

		// 如果当前熔断器为半开状态，那么状态机需要流转到开启状态
		if b.state == STATE_HALFOPEN {
			b.Change(STATE_OPEN, time.Now())
		}

		// 如果当前熔断器为关闭状态，那么基于熔断策略判断是否要流转状态
		if b.state == STATE_CLOSED {
			if b.Strategy.Factory().Adapter(b.Metrics) {
				b.Change(STATE_OPEN, time.Now())
			}
		}
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
		// 新窗口时间为增加冷却时间之后
		now = now.Add(b.CoolingTime)
		break
	case STATE_HALFOPEN:
		b.state = state
		now = time.Time{}
	case STATE_CLOSED:
		b.state = state
		// 新窗口时间
		now = now.Add(b.Metrics.Win.Size)
	case b.state:
		return
	default:
		return
	}

	// 重启计数器
	b.Metrics.Restart(now)
}

func (b *Breaker) Info(stdout bool) map[string]string {
	info := map[string]string{
		"breakInfo": fmt.Sprintf("state:%s openTime:%s",
			b.state.Name(),
			b.OpenTime.Format("2006-01-02 15:04:05"),
		),
		"WinInfo": fmt.Sprintf("size:%d startTime:%s",
			b.Metrics.Win.Size.Milliseconds(),
			b.Metrics.Win.StartTime.Format("2006-01-02 15:04:05"),
		),
		"NormInfo": fmt.Sprintf("MetricsID:%d AllCnt:%d SuccCnt:%d FailCnt:%d ContinuousSuccCnt:%d ContinuousFailCnt:%d ",
			b.Metrics.MetricsID,
			b.Metrics.Norm.AllCnt,
			b.Metrics.Norm.SuccCnt,
			b.Metrics.Norm.FailCnt,
			b.Metrics.Norm.ContinuousSuccCnt,
			b.Metrics.Norm.ContinuousFailCnt,
		),
	}

	if true == stdout {
		fmt.Printf("Stdout Break Info: %s \n", time.Now().Format("2006-01-02 15:04:05"))
		for k, v := range info {
			fmt.Printf("[%s] [%s]\n", k, v)
		}
	}
	return info
}
