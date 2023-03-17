package htmljson_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/nikolaydubina/watchhttp/webdelta/htmljson"
)

//go:embed testdata/example.json
var example []byte

func TestMarshal(t *testing.T) {
	s := htmljson.DefaultMarshaller

	var v any
	json.Unmarshal(example, &v)
	fmt.Printf("%#v\n", v)

	h, err := s.Marshal(v)
	if err != nil {
		t.Error(err)
	}

	f, _ := os.Create("testdata/example.out.html")
	defer f.Close()
	f.Write(h)
}
