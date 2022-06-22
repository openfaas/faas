package scaling

import (
	"log"
	"sync"
)

type Call struct {
	wg  *sync.WaitGroup
	res *SingleFlightResult
}

type SingleFlight struct {
	lock  *sync.RWMutex
	calls map[string]*Call
}

type SingleFlightResult struct {
	Result interface{}
	Error  error
}

func NewSingleFlight() *SingleFlight {
	return &SingleFlight{
		lock:  &sync.RWMutex{},
		calls: map[string]*Call{},
	}
}

func (s *SingleFlight) Do(key string, f func() (interface{}, error)) (interface{}, error) {

	s.lock.Lock()

	if call, ok := s.calls[key]; ok {
		s.lock.Unlock()
		call.wg.Wait()

		return call.res.Result, call.res.Error
	}

	var call *Call
	if s.calls[key] == nil {
		call = &Call{
			wg: &sync.WaitGroup{},
		}
		s.calls[key] = call
	}

	call.wg.Add(1)

	s.lock.Unlock()

	go func() {
		log.Printf("Miss, so running: %s", key)
		res, err := f()

		s.lock.Lock()
		call.res = &SingleFlightResult{
			Result: res,
			Error:  err,
		}

		call.wg.Done()

		delete(s.calls, key)

		s.lock.Unlock()
	}()

	call.wg.Wait()

	return call.res.Result, call.res.Error
}
