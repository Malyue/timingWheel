package delayqueue

import (
	"container/heap"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPriorityQueue(t *testing.T) {
	pq := make(priorityQueue, 0)

	heap.Push(&pq, &item{Value: 1, Priority: 1})
	heap.Push(&pq, &item{Value: 0, Priority: 0})
	heap.Push(&pq, &item{Value: 10, Priority: 3})

	assert.Equal(t, 3, len(pq))
	assert.Equal(t, 0, pq[0].Value)
	assert.Equal(t, 1, pq[1].Value)
	assert.Equal(t, 10, pq[2].Value)

	heap.Pop(&pq)
	assert.Equal(t, 2, len(pq))
	assert.Equal(t, 1, pq[0].Value)
	assert.Equal(t, 10, pq[1].Value)
}

func TestPeekAndShift(t *testing.T) {
	pq := make(priorityQueue, 0)

	heap.Push(&pq, &item{Value: 15, Priority: 15})
	heap.Push(&pq, &item{Value: 20, Priority: 20})
	heap.Push(&pq, &item{Value: 10, Priority: 10})

	var max int64 = 7
	for i := 0; i < 9; i++ {
		pq.PeekAndShift(max)
		max++
		if i == 4 {
			assert.Equal(t, 2, len(pq))
			assert.Equal(t, 15, pq[0].Value)
		}
	}
	assert.Equal(t, 1, len(pq))
	assert.Equal(t, 20, pq[0].Value)
}

func TestDelayQueue(t *testing.T) {
	delayQueue := New(10)
	exitChan := make(chan struct{})
	delayQueue.Poll(exitChan, func() int64 {
		return 15
	})
	delayQueue.Insert(10, int64(10))
	delayQueue.Insert(20, int64(20))

}
