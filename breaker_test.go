package breaker

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestBreak(t *testing.T) {
	breaker := NewBreaker(Config{
		HalfMaxCalls: 3,
		WindowSize:   2 * time.Second,
		Strategy: &BreakStrategyConfig{
			BreakStrategy:    BREAK_STRATEGY_FAILCNT,
			FailCntThreshold: 1,
		},
		CoolingTime: 5 * time.Second,
	})
	var succHandler = func(cnt int) {
		for i := 0; i < cnt; i++ {
			if _, err := breaker.Call(func() (i interface{}, err error) {
				return nil, nil
			}); err != nil {
				fmt.Printf("[%s] SuccCall - %s state:%s \n", time.Now().Format("2006-01-02 15:04:05"), err.Error(), breaker.state.Name())
			} else {
				fmt.Printf("[%s] SuccCall - service is ok  state:%s \n", time.Now().Format("2006-01-02 15:04:05"), breaker.state.Name())
			}
			time.Sleep(1 * time.Second)
		}
	}
	var failHandler = func(cnt int) {
		for i := 0; i < cnt; i++ {
			if _, err := breaker.Call(func() (i interface{}, err error) {
				return nil, errors.New("test err")
			}); err != nil {
				fmt.Printf("[%s] FailCall - %s state:%s \n", time.Now().Format("2006-01-02 15:04:05"), err.Error(), breaker.state.Name())
			} else {
				fmt.Printf("[%s] FailCall - service is ok  state:%s \n", time.Now().Format("2006-01-02 15:04:05"), breaker.state.Name())
			}
			time.Sleep(1 * time.Second)
		}
	}

	succHandler(5)
	if breaker.state != STATE_CLOSED {
		t.Errorf("succ 5 state is %s not %s", STATE_CLOSED.Name(), breaker.state.Name())
	} else {
		t.Logf("succ 5 state is %s", breaker.state.Name())
	}
	failHandler(5)
	if breaker.state != STATE_OPEN {
		t.Errorf("fail 5 state is %s not %s", STATE_OPEN.Name(), breaker.state.Name())
	} else {
		t.Logf("fail 5 state is %s", breaker.state.Name())
	}
	succHandler(2)
	if breaker.state != STATE_HALFOPEN {
		t.Errorf("succ 2 state is %s not %s", STATE_HALFOPEN.Name(), breaker.state.Name())
	} else {
		t.Logf("succ 2 state is %s", breaker.state.Name())
	}
	failHandler(1)
	if breaker.state != STATE_OPEN {
		t.Errorf("fail 1 state is %s not %s", STATE_OPEN.Name(), breaker.state.Name())
	} else {
		t.Logf("fail 1 state is %s", breaker.state.Name())
	}
	succHandler(10)
	if breaker.state != STATE_CLOSED {
		t.Errorf("succ 10 state is %s not %s", STATE_CLOSED.Name(), breaker.state.Name())
	} else {
		t.Logf("succ 10 state is %s", breaker.state.Name())
	}

	t.Log("Done")
}
