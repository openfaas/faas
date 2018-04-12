package plugin

import "testing"

const fallbackValue = 120

func TestLabelValueWasEmpty(t *testing.T) {
	extractedValue := extractLabelValue("", fallbackValue)

	if extractedValue != fallbackValue {
		t.Log("Expected extractedValue to equal the fallbackValue")
		t.Fail()
	}
}

func TestLabelValueWasValid(t *testing.T) {
	extractedValue := extractLabelValue("42", fallbackValue)

	if extractedValue != 42 {
		t.Log("Expected extractedValue to equal answer to life (42)")
		t.Fail()
	}
}

func TestLabelValueWasInValid(t *testing.T) {
	extractedValue := extractLabelValue("InvalidValue", fallbackValue)

	if extractedValue != fallbackValue {
		t.Log("Expected extractedValue to equal the fallbackValue")
		t.Fail()
	}
}
