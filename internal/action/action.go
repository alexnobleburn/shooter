package action

import (
	"time"

	"gitlab.mvk.com/go/vkgo/projects/donut/shooter/internal/stats"
)

type Action interface {
	stats.Statistics
	Do()

	Loader
}

type Loader interface {
	Close() error
}

func DoWrapped(action Action, avgTiming time.Duration) {
	start := time.Now()
	action.Do()
	elapsed := time.Since(start)
	timeWait := time.Duration(avgTiming.Milliseconds()-elapsed.Milliseconds()) * time.Millisecond
	if timeWait > time.Duration(1)*time.Millisecond {
		time.Sleep(timeWait)
	}
}
