package tests

import (
	"net/http"
	"testing"

	"github.com/openfaas/faas/gateway/requests"
)

func TestFormattingOfURLWithPath_NoQuery(t *testing.T) {
	req := requests.ForwardRequest{
		RawQuery: "",
		RawPath:  "/encode/utf8/",
		Method:   http.MethodPost,
	}

	url := req.ToURL("markdown", 8080)
	want := "http://markdown:8080/encode/utf8/"
	if url != want {
		t.Logf("Got: %s, want: %s", url, want)
		t.Fail()
	}
}

func TestFormattingOfURLAtRoot_NoQuery(t *testing.T) {
	req := requests.ForwardRequest{
		RawQuery: "",
		RawPath:  "/",
		Method:   http.MethodPost,
	}

	url := req.ToURL("markdown", 8080)
	want := "http://markdown:8080/"
	if url != want {
		t.Logf("Got: %s, want: %s", url, want)
		t.Fail()
	}
}

// experimental test
// func TestMyURL(t *testing.T) {
// 	v, _ := url.Parse("http://markdown/site?query=test")
// 	t.Logf("RequestURI %s", v.RequestURI())
// 	t.Logf("extra %s", v.Path)
// }

func TestUrlForFlask(t *testing.T) {
	req := requests.ForwardRequest{
		RawQuery: "query=uptime",
		RawPath:  "/function/flask",
		Method:   http.MethodPost,
	}

	url := req.ToURL("flask", 8080)
	want := "http://flask:8080/function/flask?query=uptime"
	if url != want {
		t.Logf("Got: %s, want: %s", url, want)
		t.Fail()
	}
}

func TestFormattingOfURL_OneQuery(t *testing.T) {
	req := requests.ForwardRequest{
		RawQuery: "name=alex",
		RawPath:  "/",
		Method:   http.MethodPost,
	}

	url := req.ToURL("flask", 8080)
	want := "http://flask:8080/?name=alex"
	if url != want {
		t.Logf("Got: %s, want: %s", url, want)
		t.Fail()
	}
}
