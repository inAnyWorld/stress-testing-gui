package gohttp

import (
	"sync"
	"time"
)

type LimitRate struct {
	Rate       int
	Interval   time.Duration
	LastAction time.Time
	Lock       sync.Mutex
}

func (l *LimitRate) Limit() bool {
	result := false
	for {
		l.Lock.Lock()
		// 判断最后一次执行的时间与当前的时间间隔是否大于限速速率
		if time.Now().Sub(l.LastAction) > l.Interval {
			l.LastAction = time.Now()
			result = true
		}
		l.Lock.Unlock()
		if result {
			return result
		}
		time.Sleep(l.Interval - 100)
	}
}

//SetRate 设置Rate
func (l *LimitRate) SetRate(r int) {
	l.Rate = r
	l.Interval = time.Microsecond * time.Duration(1000*1000/l.Rate)
}

//GetRate 获取Rate
func (l *LimitRate) GetRate() int {
	return l.Rate
}