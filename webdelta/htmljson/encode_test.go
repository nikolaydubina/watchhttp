package htmljson_test

import (
	_ "embed"
	"encoding/json"
	"os"
	"testing"

	"github.com/nikolaydubina/watchhttp/webdelta/htmljson"
)

//go:embed testdata/example.json
var exampleJSON []byte

//go:embed testdata/example.html
var exampleHTML string

func TestDefaultMarshaller(t *testing.T) {
	var v any
	json.Unmarshal(exampleJSON, &v)

	h := htmljson.DefaultMarshaller.Marshal(v)

	os.WriteFile("testdata/example.out.html", h, 0666)
	if exampleHTML != string(h) {
		t.Errorf("wrong output: %s", string(h))
	}
}

func TestMarshaller_JSONPath(t *testing.T) {
	var v any
	json.Unmarshal(exampleJSON, &v)

	h := htmljson.DefaultMarshaller.Marshal(v)

	os.WriteFile("testdata/example.out.html", h, 0666)
	if exampleHTML != string(h) {
		t.Errorf("wrong output: %s", string(h))
	}
}
