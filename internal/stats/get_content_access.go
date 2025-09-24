package stats

import (
	"sync/atomic"
	"time"
)

type GetContentAccess struct {
	Stats

	GetContentAccessSuccess uint64
	GetContentAccessFail    uint64

	GetContentAccessSuccessTotal uint64
	GetContentAccessFailTotal    uint64

	GetContentAccessFailStatusCode      map[int64]int
	GetContentAccessFailTotalStatusCode map[int64]int
}

func (s *GetContentAccess) PrintCurrentAndFlush(name string, period time.Duration) {
	s.Stats.PrintCurrentAndFlush(name, period)

	success := atomic.SwapUint64(&s.GetContentAccessSuccess, 0)
	atomic.AddUint64(&s.GetContentAccessSuccessTotal, success)
	fail := atomic.SwapUint64(&s.GetContentAccessFail, 0)
	atomic.AddUint64(&s.GetContentAccessFailTotal, fail)

	printInner("get content access", success, fail, period)
}

func (s *GetContentAccess) PrintTotal(name string, shootingDuration time.Duration) {
	s.Stats.PrintTotal(name, shootingDuration)

	printInner("get content access", s.GetContentAccessSuccessTotal, s.GetContentAccessFailTotal, shootingDuration)
}

func (s *GetContentAccess) Merge(other GetContentAccess) {
	s.Stats.Merge(other.Stats)

	atomic.AddUint64(&s.GetContentAccessSuccess, other.GetContentAccessSuccess)
	atomic.AddUint64(&s.GetContentAccessFail, other.GetContentAccessFail)
}
