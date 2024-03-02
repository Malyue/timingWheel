package timingWheel

import "time"

// Scheduler determines the execution plan of a task
type Scheduler interface {
	Next(time time.Time) time.Time
}

// ScheduleFunc calls f (in its own goroutine) according to the execution plan scheduled by s.
// It returns a Timer can be used to cancel the call using its Stop method
// If the caller want to terminate the execution plan halfway, it must stop the timer and ensure that the
// timer is stopped actually, since in the current implementation, there is a gap between the expiring
// and the restarting of the timer. The wait time for ensuring is short since the gap is very small
// Internally, ScheduleFunc will ask the first execution time (by calling s.Next()) initially,
// and create a timer if the execution time is non-zero.
// Afterwards, it will ask the next execution time each time f is about ot be executed,
// and f will be called at the next execution time if the time is non-zero.
func (tw *TimingWheel) ScheduleFunc(s Scheduler, f func()) (t *Timer) {
	expiration := s.Next(time.Now().UTC())
	if expiration.IsZero() {
		// No time is scheduled, return nil.
		return
	}

	t = &Timer{
		expiration: timeToMs(expiration),
		task: func() {
			// Schedule the task to execute at the next time if possible
			expiration := s.Next(msToTime(t.expiration))
			if !expiration.IsZero() {
				t.expiration = timeToMs(expiration)
				tw.addOrRun(t)
			}

			// Actually execute the task
			f()
		},
	}

	tw.addOrRun(t)
	return
}
