package breaker

import "fmt"

type State int64

const (
	// 关闭
	STATE_CLOSED State = iota
	// 开启
	STATE_OPEN
	// 半开
	STATE_HALFOPEN
)

func (s State) Name() string {
	switch s {
	case STATE_CLOSED:
		return "closed"
		break
	case STATE_OPEN:
		return "open"
		break
	case STATE_HALFOPEN:
		return "half-open"
	default:
		return fmt.Sprintf("unknown state : %d", s)
	}
	return ""
}
