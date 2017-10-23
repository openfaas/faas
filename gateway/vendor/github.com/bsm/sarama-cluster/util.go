package cluster

import (
	"fmt"
	"sort"
	"sync"
)

type none struct{}

type topicPartition struct {
	Topic     string
	Partition int32
}

func (tp *topicPartition) String() string {
	return fmt.Sprintf("%s-%d", tp.Topic, tp.Partition)
}

type offsetInfo struct {
	Offset   int64
	Metadata string
}

func (i offsetInfo) NextOffset(fallback int64) int64 {
	if i.Offset > -1 {
		return i.Offset
	}
	return fallback
}

type int32Slice []int32

func (p int32Slice) Len() int           { return len(p) }
func (p int32Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p int32Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (p int32Slice) Diff(o int32Slice) (res []int32) {
	on := len(o)
	for _, x := range p {
		n := sort.Search(on, func(i int) bool { return o[i] >= x })
		if n < on && o[n] == x {
			continue
		}
		res = append(res, x)
	}
	return
}

// --------------------------------------------------------------------

type loopTomb struct {
	c chan none
	o sync.Once
	w sync.WaitGroup
}

func newLoopTomb() *loopTomb {
	return &loopTomb{c: make(chan none)}
}

func (t *loopTomb) stop()  { t.o.Do(func() { close(t.c) }) }
func (t *loopTomb) Close() { t.stop(); t.w.Wait() }

func (t *loopTomb) Dying() <-chan none { return t.c }
func (t *loopTomb) Go(f func(<-chan none)) {
	t.w.Add(1)

	go func() {
		defer t.stop()
		defer t.w.Done()

		f(t.c)
	}()
}
