package longpoll

import (
	"errors"
	"sync"
	"time"

	seqqueue "git.nulana.com/bobrnor/seqqueue.git"
)

const (
	TickTimeout    = 5 * time.Second
	DefaultTimeout = 2 * time.Minute
)

var (
	defaultLongpoll *Longpoll

	ErrReceiverNotFound = errors.New("Received not found")
)

type Longpoll struct {
	sync.RWMutex
	queues  map[interface{}]*Entry
	timeout time.Duration

	registerFunc RegisterFunc
	purgeFunc    PurgeFunc
}

type RegisterFunc func(lp *Longpoll, i interface{})
type PurgeFunc func(lp *Longpoll, i interface{})

func NewLongpoll() *Longpoll {
	lp := Longpoll{
		queues:  map[interface{}]*Entry{},
		timeout: DefaultTimeout,
	}
	go lp.loop()
	return &lp
}

func DefaultLongpoll() *Longpoll {
	if defaultLongpoll == nil {
		defaultLongpoll = NewLongpoll()
	}
	return defaultLongpoll
}

func (lp *Longpoll) Register(i interface{}) *seqqueue.Queue {
	lp.Lock()
	defer lp.Unlock()

	if e, ok := lp.queues[i]; ok {
		e.ts = time.Now()
		return e.queue
	}

	e := Entry{
		queue: seqqueue.NewQueue(),
		ts:    time.Now(),
	}
	lp.queues[i] = &e

	if lp.registerFunc != nil {
		go lp.registerFunc(lp, i)
	}

	return e.queue
}

func (lp *Longpoll) SetTimeout(d time.Duration) {
	lp.Lock()
	defer lp.Unlock()
	lp.timeout = d
}

func (lp *Longpoll) Timeout() time.Duration {
	lp.RLock()
	defer lp.RUnlock()
	return lp.timeout
}

func (lp *Longpoll) Send(i interface{}, msg interface{}) error {
	lp.RLock()
	defer lp.RUnlock()

	e, ok := lp.queues[i]
	if !ok {
		return ErrReceiverNotFound
	}

	e.queue.In() <- msg
	return nil
}

func (lp *Longpoll) SetRegisterFunc(f RegisterFunc) {
	lp.Lock()
	defer lp.Unlock()

	lp.registerFunc = f
}

func (lp *Longpoll) SetPurgeFunc(f PurgeFunc) {
	lp.Lock()
	defer lp.Unlock()

	lp.purgeFunc = f
}

func (lp *Longpoll) loop() {
	for range time.Tick(TickTimeout) {
		lp.purge()
	}
}

func (lp *Longpoll) purge() {
	lp.Lock()
	defer lp.Unlock()

	for k, e := range lp.queues {
		if time.Now().Sub(e.ts) >= lp.timeout {
			delete(lp.queues, k)

			if lp.purgeFunc != nil {
				go lp.purgeFunc(lp, k)
			}
		}
	}
}
