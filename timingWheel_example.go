package timingWheel

import "time"

type EveryScheduler struct {
	Interval time.Duration
}

func (s *EveryScheduler) Next(prev time.Time) time.Time {
	return prev.Add(s.Interval)
}

//func Example_schedulerTimer() {
//	tw := NewTimingWheel(time.Millisecond, 20)
//	tw.Start()
//	defer tw.Stop()
//
//	exitC := make(chan time.Time)
//	t := tw.ScheduleFunc()
//}
