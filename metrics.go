package breaker

import (
	"time"
)

type Metrics struct {
	MetricsID int64
	Win       *Window
	Norm      *Norm
}
type Window struct {
	Size      time.Duration
	StartTime time.Time
}
type Norm struct {
	AllCnt            int64
	SuccCnt           int64
	FailCnt           int64
	ContinuousSuccCnt int64
	ContinuousFailCnt int64
}

func (m *Metrics) Call() {
	m.Norm.AllCnt++
}

func (m *Metrics) Succ() {
	m.Norm.SuccCnt++
	m.Norm.ContinuousSuccCnt++
	m.Norm.ContinuousFailCnt = 0
}

func (m *Metrics) Fail() {
	m.Norm.FailCnt++
	m.Norm.ContinuousFailCnt++
	m.Norm.ContinuousSuccCnt = 0
}

// 重启计数器
func (m *Metrics) Restart(t time.Time) {
	m.MetricsID++

	// 指标重置
	m.Norm.Reset()

	// 滑动时间窗口
	m.Win.Next(t)

	return
}

func (n *Norm) Reset() {
	n.AllCnt = 0
	n.SuccCnt = 0
	n.FailCnt = 0
	n.ContinuousSuccCnt = 0
	n.ContinuousFailCnt = 0
}

func (w *Window) Next(t time.Time) {
	w.StartTime = t
}
