package cluster

import (
	"sort"
	"sync"
	"time"

	"github.com/Shopify/sarama"
)

// PartitionConsumer allows code to consume individual partitions from the cluster.
//
// See docs for Consumer.Partitions() for more on how to implement this.
type PartitionConsumer interface {

	// Close stops the PartitionConsumer from fetching messages. It will initiate a shutdown, drain
	// the Messages channel, harvest any errors & return them to the caller and trigger a rebalance.
	Close() error

	// Messages returns the read channel for the messages that are returned by
	// the broker.
	Messages() <-chan *sarama.ConsumerMessage

	// HighWaterMarkOffset returns the high water mark offset of the partition,
	// i.e. the offset that will be used for the next message that will be produced.
	// You can use this to determine how far behind the processing is.
	HighWaterMarkOffset() int64

	// Topic returns the consumed topic name
	Topic() string

	// Partition returns the consumed partition
	Partition() int32
}

type partitionConsumer struct {
	sarama.PartitionConsumer

	state partitionState
	mu    sync.Mutex

	topic     string
	partition int32

	once        sync.Once
	dying, dead chan none
}

func newPartitionConsumer(manager sarama.Consumer, topic string, partition int32, info offsetInfo, defaultOffset int64) (*partitionConsumer, error) {
	pcm, err := manager.ConsumePartition(topic, partition, info.NextOffset(defaultOffset))

	// Resume from default offset, if requested offset is out-of-range
	if err == sarama.ErrOffsetOutOfRange {
		info.Offset = -1
		pcm, err = manager.ConsumePartition(topic, partition, defaultOffset)
	}
	if err != nil {
		return nil, err
	}

	return &partitionConsumer{
		PartitionConsumer: pcm,
		state:             partitionState{Info: info},

		topic:     topic,
		partition: partition,

		dying: make(chan none),
		dead:  make(chan none),
	}, nil
}

// Topic implements PartitionConsumer
func (c *partitionConsumer) Topic() string { return c.topic }

// Partition implements PartitionConsumer
func (c *partitionConsumer) Partition() int32 { return c.partition }

func (c *partitionConsumer) WaitFor(stopper <-chan none, errors chan<- error) {
	defer close(c.dead)

	for {
		select {
		case err, ok := <-c.Errors():
			if !ok {
				return
			}
			select {
			case errors <- err:
			case <-stopper:
				return
			case <-c.dying:
				return
			}
		case <-stopper:
			return
		case <-c.dying:
			return
		}
	}
}

func (c *partitionConsumer) Multiplex(stopper <-chan none, messages chan<- *sarama.ConsumerMessage, errors chan<- error) {
	defer close(c.dead)

	for {
		select {
		case msg, ok := <-c.Messages():
			if !ok {
				return
			}
			select {
			case messages <- msg:
			case <-stopper:
				return
			case <-c.dying:
				return
			}
		case err, ok := <-c.Errors():
			if !ok {
				return
			}
			select {
			case errors <- err:
			case <-stopper:
				return
			case <-c.dying:
				return
			}
		case <-stopper:
			return
		case <-c.dying:
			return
		}
	}
}

func (c *partitionConsumer) Close() (err error) {
	c.once.Do(func() {
		err = c.PartitionConsumer.Close()
		close(c.dying)
	})
	<-c.dead
	return err
}

func (c *partitionConsumer) State() partitionState {
	if c == nil {
		return partitionState{}
	}

	c.mu.Lock()
	state := c.state
	c.mu.Unlock()

	return state
}

func (c *partitionConsumer) MarkCommitted(offset int64) {
	if c == nil {
		return
	}

	c.mu.Lock()
	if offset == c.state.Info.Offset {
		c.state.Dirty = false
	}
	c.mu.Unlock()
}

func (c *partitionConsumer) MarkOffset(offset int64, metadata string) {
	if c == nil {
		return
	}

	c.mu.Lock()
	if offset > c.state.Info.Offset {
		c.state.Info.Offset = offset
		c.state.Info.Metadata = metadata
		c.state.Dirty = true
	}
	c.mu.Unlock()
}

// --------------------------------------------------------------------

type partitionState struct {
	Info       offsetInfo
	Dirty      bool
	LastCommit time.Time
}

// --------------------------------------------------------------------

type partitionMap struct {
	data map[topicPartition]*partitionConsumer
	mu   sync.RWMutex
}

func newPartitionMap() *partitionMap {
	return &partitionMap{
		data: make(map[topicPartition]*partitionConsumer),
	}
}

func (m *partitionMap) IsSubscribedTo(topic string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for tp := range m.data {
		if tp.Topic == topic {
			return true
		}
	}
	return false
}

func (m *partitionMap) Fetch(topic string, partition int32) *partitionConsumer {
	m.mu.RLock()
	pc, _ := m.data[topicPartition{topic, partition}]
	m.mu.RUnlock()
	return pc
}

func (m *partitionMap) Store(topic string, partition int32, pc *partitionConsumer) {
	m.mu.Lock()
	m.data[topicPartition{topic, partition}] = pc
	m.mu.Unlock()
}

func (m *partitionMap) Snapshot() map[topicPartition]partitionState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snap := make(map[topicPartition]partitionState, len(m.data))
	for tp, pc := range m.data {
		snap[tp] = pc.State()
	}
	return snap
}

func (m *partitionMap) Stop() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var wg sync.WaitGroup
	for tp := range m.data {
		wg.Add(1)
		go func(p *partitionConsumer) {
			_ = p.Close()
			wg.Done()
		}(m.data[tp])
	}
	wg.Wait()
}

func (m *partitionMap) Clear() {
	m.mu.Lock()
	for tp := range m.data {
		delete(m.data, tp)
	}
	m.mu.Unlock()
}

func (m *partitionMap) Info() map[string][]int32 {
	info := make(map[string][]int32)
	m.mu.RLock()
	for tp := range m.data {
		info[tp.Topic] = append(info[tp.Topic], tp.Partition)
	}
	m.mu.RUnlock()

	for topic := range info {
		sort.Sort(int32Slice(info[topic]))
	}
	return info
}
