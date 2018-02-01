package nats

import (
	"testing"
	"time"
)

func TestTimerPool(t *testing.T) {
	var tp timerPool

	for i := 0; i < 10; i++ {
		tm := tp.Get(time.Millisecond * 20)

		select {
		case <-tm.C:
			t.Errorf("Timer already expired")
			continue
		default:
		}

		select {
		case <-tm.C:
		case <-time.After(time.Millisecond * 100):
			t.Errorf("Timer didn't expire in time")
		}

		tp.Put(tm)
	}
}
