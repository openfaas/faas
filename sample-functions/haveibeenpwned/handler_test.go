package function

import (
	"encoding/json"
	"testing"
)

func Test_Handle(t *testing.T) {
	res := Handle([]byte("test1234"))

	result := result{}
	err := json.Unmarshal([]byte(res), &result)
	if err != nil {
		t.Errorf("unable to unmarshal response, error: %s", err)
		t.Fail()
	}

	if result.Found == 0 {
		t.Errorf("expected test1234 to be found several times")
		t.Fail()
	}
}
