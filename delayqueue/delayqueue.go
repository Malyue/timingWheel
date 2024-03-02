package delayqueue

import (
	"container/heap"
	"sync"
	"sync/atomic"
	"time"
)

type item struct {
	Value    interface{}
	Priority int64
	Index    int
}

type priorityQueue []*item

// create a priorityQueue with capacity
// and the capacity will add auto
func newPriorityQueue(capacity int) priorityQueue {
	return make(priorityQueue, 0, capacity)
}

func (pq priorityQueue) Len() int {
	return len(pq)
}

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].Priority < pq[j].Priority
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

// Push
func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	c := cap(*pq)
	if n+1 > c {
		if c == 0 {
			c = 1
		}
		npq := make(priorityQueue, n, c*2)
		copy(npq, *pq)
		*pq = npq
	}
	*pq = (*pq)[0 : n+1]
	item := x.(*item)
	item.Index = n
	(*pq)[n] = item
}

func (pq *priorityQueue) Pop() interface{} {
	n := len(*pq)
	c := cap(*pq)
	if n < (c/2) && c > 25 {
		npq := make(priorityQueue, n, c/2)
		copy(npq, *pq)
		*pq = npq
	}
	item := (*pq)[n-1]
	item.Index = -1
	*pq = (*pq)[0 : n-1]
	return item
}

// PeekAndShift pop
func (pq *priorityQueue) PeekAndShift(max int64) (*item, int64) {
	if pq.Len() == 0 {
		return nil, 0
	}

	head := (*pq)[0]
	if head.Priority > max {
		return nil, head.Priority - max
	}
	heap.Remove(pq, 0)

	return head, 0
}

type DelayQueue struct {
	C chan interface{}

	mu       sync.Mutex
	pq       priorityQueue
	sleeping int32
	wakeupC  chan struct{}
}

func New(size int) *DelayQueue {
	return &DelayQueue{
		C:       make(chan interface{}),
		pq:      newPriorityQueue(size),
		wakeupC: make(chan struct{}),
	}
}

// Insert  a elem into the current queue
func (dq *DelayQueue) Insert(elem interface{}, expiration int64) {
	item := &item{Value: elem, Priority: expiration}

	dq.mu.Lock()
	heap.Push(&dq.pq, item)
	index := item.Index
	dq.mu.Unlock()

	if index == 0 {
		if atomic.CompareAndSwapInt32(&dq.sleeping, 1, 0) {
			dq.wakeupC <- struct{}{}
		}
	}
}

// Poll fn is the method to get timestamp
func (dq *DelayQueue) Poll(exitC chan struct{}, fn func() int64) {
	for {
		now := fn()
		dq.mu.Lock()

		item, delta := dq.pq.PeekAndShift(now)
		if item == nil {
			atomic.StoreInt32(&dq.sleeping, 1)
		}
		dq.mu.Unlock()

		if item == nil {
			if delta == 0 {
				// it means pq is null
				select {
				// if it has any inserted elem, continue
				case <-dq.wakeupC:
					continue
				// if exit
				case <-exitC:
					goto exit
				}
			} else if delta > 0 {
				// it means the first elem is not expire
				select {
				// a new item with earlier expiration than current item, we should use the earlier
				case <-dq.wakeupC:
					continue
				case <-time.After(time.Duration(delta) * time.Millisecond):
					// reset the sleeping state since there's no need to receive from wakeupC
					if atomic.SwapInt32(&dq.sleeping, 0) == 0 {
						// a caller of Offer() is being blocked on sending to wakeupC
						// drain wakeupC to unblock the caller
						<-dq.wakeupC
					}
					continue
				case <-exitC:
					goto exit
				}
			}
		}
		select {
		case dq.C <- item.Value:
		case <-exitC:
			goto exit
		}
	}
exit:
	atomic.SwapInt32(&dq.sleeping, 0)
}
