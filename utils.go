package timingWheel

import (
	"sync"
	"time"
)

// truncate rounds down a number x to the nearest multiple of m.
// it eliminates any remainder or incomplete portion
// if m is less than or equal to 0, x is returned directly
func truncate(x, m int64) int64 {
	if m <= 0 {
		return x
	}
	return x - x%m
}

func timeToMs(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

// msToTime returns the UTC time corresponding to the given unix time
// t milliseconds since Jan 1, 1970 UTC
func msToTime(t int64) time.Time {
	return time.Unix(0, t*int64(time.Millisecond)).UTC()
}

type waitGroupWrapper struct {
	sync.WaitGroup
}

func (w *waitGroupWrapper) Wrap(fn func()) {
	w.Add(1)
	go func() {
		fn()
		w.Done()
	}()
}
