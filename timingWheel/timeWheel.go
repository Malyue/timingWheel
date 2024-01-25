package timingWheel

import (
	"container/list"
	"log"
	"sync"
	"time"
)

const (
	defaultSlotNum  = 10
	defaultInterval = time.Second
)

type TimeWheel struct {
	sync.Once
	interval time.Duration
	ticker   *time.Ticker
	// use for stop the instance
	stopc chan struct{}
	// use for adding task
	addTaskCh chan *taskElement
	// use for removing task
	removeTaskCh chan string
	// rounding array
	slots   []*list.List
	curSlot int
	// key to taskNode, use for delete in list
	keyToETask map[string]*list.Element
}

func NewTimeWheel(slotNum int, interval time.Duration) *TimeWheel {
	if slotNum <= 0 {
		slotNum = defaultSlotNum
	}
	if interval <= 0 {
		interval = defaultInterval
	}

	t := TimeWheel{
		interval:     interval,
		ticker:       time.NewTicker(interval),
		stopc:        make(chan struct{}),
		keyToETask:   make(map[string]*list.Element),
		slots:        make([]*list.List, 0, slotNum),
		addTaskCh:    make(chan *taskElement),
		removeTaskCh: make(chan string),
	}

	for i := 0; i < slotNum; i++ {
		t.slots = append(t.slots, list.New())
	}

	go t.run()
	return &t
}

func (t *TimeWheel) run() {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()

	// for + select
	for {
		select {
		case <-t.stopc:
			t.Stop()
			return
		case <-t.ticker.C:
			t.tick()
		case task := <-t.addTaskCh:
			t.addTask(task)
		case removeKey := <-t.removeTaskCh:
			t.removeTask(removeKey)
		}
	}

}

func (t *TimeWheel) tick() {
	list := t.slots[t.curSlot]
	defer t.circularIncr()
	t.execute(list)
}

func (t *TimeWheel) circularIncr() {
	t.curSlot = (t.curSlot + 1) % len(t.slots)
}

func (t *TimeWheel) execute(l *list.List) {
	for e := l.Front(); e != nil; {
		taskElement, _ := e.Value.(*taskElement)
		if taskElement.cycle > 0 {
			taskElement.cycle--
			e = e.Next()
			continue
		}

		go func() {
			defer func() {
				if err := recover(); err != nil {
					log.Println(err)
				}
			}()
			taskElement.task()
		}()

		next := e.Next()
		l.Remove(e)
		delete(t.keyToETask, taskElement.key)
		e = next
	}
}

func (t *TimeWheel) addTask(task *taskElement) {
	list := t.slots[task.pos]
	if _, ok := t.keyToETask[task.key]; ok {
		t.removeTask(task.key)
	}
	eTask := list.PushBack(task)
	t.keyToETask[task.key] = eTask
}

func (t *TimeWheel) Stop() {
	t.Do(func() {
		t.ticker.Stop()
		close(t.stopc)
	})
}

func (t *TimeWheel) AddTask(key string, task func(), executeAt time.Time) {
	pos, cycle := t.getPosAndCircle(executeAt)
	t.addTaskCh <- &taskElement{
		pos:   pos,
		cycle: cycle,
		task:  task,
		key:   key,
	}
}

func (t *TimeWheel) getPosAndCircle(executeAt time.Time) (int, int) {
	delay := int(time.Until(executeAt))
	cycle := delay / (len(t.slots) * int(t.interval))
	pos := (t.curSlot + delay/int(t.interval)) % len(t.slots)
	return pos, cycle
}

func (t *TimeWheel) RemoveTask(key string) {
	t.removeTaskCh <- key
}

func (t *TimeWheel) removeTask(key string) {
	eTask, ok := t.keyToETask[key]
	if !ok {
		return
	}

	delete(t.keyToETask, key)
	task, _ := eTask.Value.(*taskElement)
	_ = t.slots[task.pos].Remove(eTask)
}
