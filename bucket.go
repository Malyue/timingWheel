package timingWheel

import (
	"container/list"
	"sync"
	"sync/atomic"
	"unsafe"
)

// Timer represents a single event. When the timer expires, the given task will be executed.
type Timer struct {
	expiration int64
	task       func()
	// the bucket holds the list to which this timer's element belongs
	b       unsafe.Pointer
	element *list.Element
}

func (t *Timer) getBucket() *bucket {
	return (*bucket)(atomic.LoadPointer(&t.b))
}

func (t *Timer) setBucket(b *bucket) {
	atomic.StorePointer(&t.b, unsafe.Pointer(b))
}

// Stop prevents the timer from firing
// it returns true if stop the timer, false if the timer has already expired or been stopped
func (t *Timer) Stop() bool {
	stopped := false
	// if b.Remove is called just after the timing wheel's goroutine has:
	// 	  1. removed t from b (through b.Flush -> b.remove)
	//    2. moved t from b to another bucket ab (through b.Flush -> b.remove and ab.Add)
	// this may fail to remove t due to the change of t's bucket
	// therefore, we should retry until the t's bucket becomes nil,
	// which indicates that t has finally been removed
	for b := t.getBucket(); b != nil; b = t.getBucket() {
		stopped = b.Remove(t)
	}
	return stopped
}

// bucket as the space of the timingWheel
type bucket struct {
	expiration int64
	mu         sync.Mutex
	timers     *list.List
}

func newBucket() *bucket {
	return &bucket{
		timers:     list.New(),
		expiration: -1,
	}
}

func (b *bucket) Expiration() int64 {
	return atomic.LoadInt64(&b.expiration)
}

func (b *bucket) SetExpiration(expiration int64) bool {
	return atomic.SwapInt64(&b.expiration, expiration) != expiration
}

func (b *bucket) Add(t *Timer) {
	b.mu.Lock()
	defer b.mu.Unlock()

	ele := b.timers.PushBack(t)
	t.setBucket(b)
	t.element = ele
}

func (b *bucket) remove(t *Timer) bool {
	if t.getBucket() != b {
		return false
	}

	b.timers.Remove(t.element)
	t.setBucket(nil)
	t.element = nil
	return true
}

func (b *bucket) Remove(t *Timer) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.remove(t)
}

// Flush use to flush the timer in this bucket, and offer to support reinsert method to reinsert the timer
func (b *bucket) Flush(reinsert func(*Timer)) {
	b.mu.Lock()
	var ts []*Timer
	for e := b.timers.Front(); e != nil; {
		next := e.Next()

		t := e.Value.(*Timer)
		b.remove(t)
		ts = append(ts, t)
		e = next
	}
	b.mu.Unlock()
	b.SetExpiration(-1)

	for _, t := range ts {
		if reinsert != nil {
			reinsert(t)
		}
	}
}
