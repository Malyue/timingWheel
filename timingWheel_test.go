package timingWheel

import (
	"testing"
	"time"
)

func TestTimingWheel_AfterFunc(t *testing.T) {
	tw := NewTimingWheel(time.Millisecond, 20)
	tw.Start()
	defer tw.Stop()
}
