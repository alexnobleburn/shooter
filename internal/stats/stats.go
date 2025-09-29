package stats

import (
	"fmt"
	"sync/atomic"
	"time"
)

type Statistics interface {
	PrintCurrentAndFlush(name string, period time.Duration)
	PrintTotal(name string, shootingDuration time.Duration)
}

type Stats struct {
	Success uint64
	Fail    uint64

	SuccessTotal uint64
	FailTotal    uint64
}

func (s *Stats) PrintCurrentAndFlush(name string, period time.Duration) {
	success := atomic.SwapUint64(&s.Success, 0)
	atomic.AddUint64(&s.SuccessTotal, success)

	fail := atomic.SwapUint64(&s.Fail, 0)
	atomic.AddUint64(&s.FailTotal, fail)

	printInner(name, success, fail, period)
}

func (s *Stats) PrintTotal(name string, shootingDuration time.Duration) {
	fmt.Printf("total shooting stats\n")
	fmt.Printf("sent %d requests by %s\n", s.SuccessTotal+s.FailTotal, shootingDuration.String())
	printInner(name, s.SuccessTotal, s.FailTotal, shootingDuration)
}

func (s *Stats) Merge(other Stats) {
	atomic.AddUint64(&s.Success, other.Success)
	atomic.AddUint64(&s.Fail, other.Fail)
}

func (s *Stats) RecordSuccess() {
	atomic.AddUint64(&s.Success, 1)
}

func (s *Stats) RecordFailure() {
	atomic.AddUint64(&s.Fail, 1)
}

func printInner(statName string, p1 uint64, p2 uint64, period time.Duration) {
	requests := float64(p1 + p2)
	rps := requests / period.Seconds()
	successRate := float64(p1) / requests
	failRate := float64(p2) / requests
	if requests == 0.0 {
		successRate = 0.0
		failRate = 0.0
	}
	fmt.Printf("[%v] (%s) rps = %6.3f, success = %6.3f, fail = %6.3f, errors absolute = %d\n",
		time.Now().Format("2006-01-02 15:04:05"), statName, rps, successRate, failRate, p2)
}
