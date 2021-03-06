package breaker

import "fmt"

type BreakStrategy int64

// 熔断策略
const (
	// 根据错误计数，如果一个时间窗口期内失败数 >= n 次，开启熔断。
	BREAK_STRATEGY_FAILCNT BreakStrategy = iota
	// 根据连续错误计数，一个时间窗口期内连续失败 >=n 次，开启熔断。
	BREAK_STRATEGY_CONTINIUOUSFAILCNT
	// 根据错误比例，一个时间窗口期内错误占比 >= n （0 ~ 1），开启熔断.
	BREAK_STRATEGY_FAILRATE
)

type BreakStrategyFunc interface {
	Adapter(metrics *Metrics) bool
}
type BreakStrategyConfig struct {
	BreakStrategy              BreakStrategy
	FailCntThreshold           int64
	ContinuousFailCntThreshold int64
	FailRate                   float64
}
type BsFailCnt struct {
	*BreakStrategyConfig
}
type BsContinuousFailCnt struct {
	*BreakStrategyConfig
}
type BsFailRate struct {
	*BreakStrategyConfig
}

func (bsc BreakStrategyConfig) Factory() BreakStrategyFunc {
	switch bsc.BreakStrategy {
	case BREAK_STRATEGY_FAILCNT:
		return &BsFailCnt{&bsc}
		break
	case BREAK_STRATEGY_CONTINIUOUSFAILCNT:
		return &BsContinuousFailCnt{&bsc}
		break
	case BREAK_STRATEGY_FAILRATE:
		return &BsFailRate{&bsc}
		break
	default:
		panic(fmt.Sprintf("unknown break strategy : %d", bsc.BreakStrategy))
	}
	return nil
}

func (bs *BsFailCnt) Adapter(metrics *Metrics) bool {
	return metrics.Norm.FailCnt >= bs.FailCntThreshold
}
func (bs *BsContinuousFailCnt) Adapter(metrics *Metrics) bool {
	return metrics.Norm.ContinuousFailCnt >= bs.ContinuousFailCntThreshold
}
func (bs *BsFailRate) Adapter(metrics *Metrics) bool {
	rate := float64(metrics.Norm.FailCnt / metrics.Norm.AllCnt)
	return rate >= bs.FailRate
}
