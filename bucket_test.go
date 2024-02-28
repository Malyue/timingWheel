package timingWheel

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBucket_Flush(t *testing.T) {
	b := newBucket()

	b.Add(&Timer{})
	b.Add(&Timer{})
	length := b.timers.Len()
	assert.Equal(t, 2, length)
	b.Flush(nil)
	empty := b.timers.Len()
	assert.Equal(t, 0, empty)
}
