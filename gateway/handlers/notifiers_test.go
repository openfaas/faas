package handlers

import "testing"

func Test_urlToLabel_normalizeTrailing(t *testing.T) {
	have := "/system/functions/"
	want := "/system/functions"
	got := urlToLabel(have)

	if got != want {
		t.Errorf("want %s, got %s", want, got)
	}
}

func Test_urlToLabel_retainRoot(t *testing.T) {
	have := "/"
	want := have
	got := urlToLabel(have)

	if got != want {
		t.Errorf("want %s, got %s", want, got)
	}
}
