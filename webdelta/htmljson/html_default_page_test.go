package htmljson_test

import (
	_ "embed"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/nikolaydubina/watchhttp/webdelta/htmljson"
)

//go:embed testdata/example-page.html
var examplePageHTML string

func TestMarshalHTML(t *testing.T) {
	var v any
	json.Unmarshal(exampleJSON, &v)

	h := htmljson.MarshalHTML(v)

	os.WriteFile("testdata/example-page.out.html", h, 0666)
	if strings.TrimSpace(examplePageHTML) != strings.TrimSpace(string(h)) {
		t.Errorf("wrong output: %s", string(h))
	}
}
