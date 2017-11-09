package blackfriday

import (
	"bytes"
	"testing"
)

func TestEsc(t *testing.T) {
	tests := []string{
		"abc", "abc",
		"a&c", "a&amp;c",
		"<", "&lt;",
		"[]:<", "[]:&lt;",
		"Hello <!--", "Hello &lt;!--",
	}
	for i := 0; i < len(tests); i += 2 {
		var b bytes.Buffer
		escapeHTML(&b, []byte(tests[i]))
		if !bytes.Equal(b.Bytes(), []byte(tests[i+1])) {
			t.Errorf("\nInput   [%#v]\nExpected[%#v]\nActual  [%#v]",
				tests[i], tests[i+1], b.String())
		}
	}
}

func BenchmarkEscapeHTML(b *testing.B) {
	tests := [][]byte{
		[]byte(""),
		[]byte("AT&T has an ampersand in their name."),
		[]byte("AT&amp;T is another way to write it."),
		[]byte("This & that."),
		[]byte("4 < 5."),
		[]byte("6 > 5."),
		[]byte("Here's a [link] [1] with an ampersand in the URL."),
		[]byte("Here's a link with an ampersand in the link text: [AT&T] [2]."),
		[]byte("Here's an inline [link](/script?foo=1&bar=2)."),
		[]byte("Here's an inline [link](</script?foo=1&bar=2>)."),
		[]byte("[1]: http://example.com/?foo=1&bar=2"),
		[]byte("[2]: http://att.com/  \"AT&T\""),
	}
	var buf bytes.Buffer
	for n := 0; n < b.N; n++ {
		for _, t := range tests {
			escapeHTML(&buf, t)
			buf.Reset()
		}
	}
}
