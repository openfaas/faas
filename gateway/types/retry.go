package types

import (
	"log"
	"time"
)

type routine func(attempt int) error

func Retry(r routine, label string, attempts int, interval time.Duration) error {
	var err error

	for i := 0; i < attempts; i++ {
		res := r(i)
		if res != nil {
			err = res
			log.Printf("[%s]: %d/%d, error: %s\n", label, i, attempts, res)
		} else {
			err = nil
			break
		}
		time.Sleep(interval)
	}
	return err
}
