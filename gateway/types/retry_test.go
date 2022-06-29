package types

import (
	"fmt"
	"testing"
	"time"
)

func Test_retry_early_success(t *testing.T) {
	called := 0
	maxRetries := 10
	routine := func(i int) error {

		called++
		if called == 5 {
			return nil
		}
		return fmt.Errorf("not called 5 times yet")
	}

	Retry(routine, "test", maxRetries, time.Millisecond*5)

	want := 5
	if called != want {
		t.Errorf("want: %d, got: %d", want, called)
	}
}

func Test_retry_until_max_attempts(t *testing.T) {
	called := 0
	maxRetries := 10
	routine := func(i int) error {
		called++
		return fmt.Errorf("unable to pass condition for routine")
	}

	Retry(routine, "test", maxRetries, time.Millisecond*5)

	want := maxRetries
	if called != want {
		t.Errorf("want: %d, got: %d", want, called)
	}
}
