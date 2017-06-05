package longpoll

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func describe(lp *Longpoll) string {
	return fmt.Sprintf("%+v", lp.queues)
}

func isEqual(a, b *Longpoll) bool {
	return reflect.DeepEqual(a.queues, b.queues)
}

func TestRegister0(t *testing.T) {
	lp := NewLongpoll()
	lp.Register(127)

	if _, ok := lp.queues[127]; !ok {
		t.Errorf("Queue in lp queues expected %+v", describe(lp))
	}
}

func TestRegister1(t *testing.T) {
	lp := NewLongpoll()
	lp.Register(127)
	lp.Register(127)

	if _, ok := lp.queues[127]; !ok {
		t.Errorf("Queue in lp queues expected %+v", describe(lp))
	}

	if len(lp.queues) != 1 {
		t.Errorf("Should create only 1 queue %+v", describe(lp))
	}
}

func TestSetTimeout0(t *testing.T) {
	lp := NewLongpoll()
	lp.SetTimeout(1 * time.Second)
	if lp.Timeout() != 1*time.Second {
		t.Errorf("Timeout not set %+v", lp)
	}
}

func TestTimeout1(t *testing.T) {
	if testing.Short() {
		t.Skipf("Skip lp expiration test")
	}

	lp := NewLongpoll()
	lp.SetTimeout(1 * time.Second)
	lp.Register(127)
	<-time.After(6 * time.Second)

	if _, ok := lp.queues[127]; ok {
		t.Errorf("Queue in lp queues unexpected %+v", describe(lp))
	}
}

func TestSend0(t *testing.T) {
	lp := NewLongpoll()
	if err := lp.Send(127, 721); err != ErrReceiverNotFound {
		t.Errorf("Send should return error")
	}
}

func TestSend1(t *testing.T) {
	lp := NewLongpoll()
	q := lp.Register(127)
	if err := lp.Send(127, 721); err != nil {
		t.Errorf("Send should not return error %+v", err.Error())
	}

	msg := <-q.Out(0)
	iMsg := msg.Value.(int)
	if iMsg != 721 {
		t.Errorf("LP return wrong message %+v", msg)
	}
}

func TestSetRegisterFunc0(t *testing.T) {
	lp := NewLongpoll()
	okChan := make(chan struct{})
	lp.SetRegisterFunc(func(lp *Longpoll, i interface{}) {
		okChan <- struct{}{}
	})

	lp.Register(127)

	select {
	case <-okChan:
	case <-time.After(1 * time.Second):
		t.Error("Callback function does not called")
	}
}

func TestSetPurgeFunc0(t *testing.T) {
	lp := NewLongpoll()
	lp.SetTimeout(0 * time.Second)
	okChan := make(chan struct{})
	lp.SetPurgeFunc(func(lp *Longpoll, i interface{}) {
		okChan <- struct{}{}
	})

	lp.Register(127)
	lp.purge()

	select {
	case <-okChan:
	case <-time.After(1 * time.Second):
		t.Error("Callback function does not called")
	}
}

func TestPurgeFunc0(t *testing.T) {
	lp := NewLongpoll()
	lp.SetTimeout(0 * time.Second)
	lp.Register(127)
	lp.purge()

	if _, ok := lp.queues[127]; ok {
		t.Errorf("Queue in lp queues unexpected %+v", describe(lp))
	}
}
