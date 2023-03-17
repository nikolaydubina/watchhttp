package webdelta_test

import (
	_ "embed"
	"os"
	"testing"

	"github.com/nikolaydubina/watchhttp/webdelta"
)

//go:embed htmljson/testdata/example.json
var example []byte

func TestJSONHTML(t *testing.T) {
	s := webdelta.Renderer{JSONString: example}

	f, err := os.Create("testdata/example.out.html")
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	s.WriteTo(f)
}
