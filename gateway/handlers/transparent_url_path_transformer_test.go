package handlers

import (
	"net/http"
	"testing"
)

func Test_Transform_DoesntTransformRootPath(t *testing.T) {

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	transformer := TransparentURLPathTransformer{}
	want := req.URL.Path
	got := transformer.Transform(req)

	if want != got {
		t.Errorf("want: %s, got: %s", want, got)
	}
}

func Test_Transform_DoesntTransformAdditionalPath(t *testing.T) {

	req, _ := http.NewRequest(http.MethodGet, "/employees/", nil)
	transformer := TransparentURLPathTransformer{}
	want := req.URL.Path
	got := transformer.Transform(req)

	if want != got {
		t.Errorf("want: %s, got: %s", want, got)
	}

}
