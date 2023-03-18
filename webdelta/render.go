package webdelta

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"io"
	"strings"

	"github.com/nikolaydubina/watchhttp/webdelta/htmljson"
)

//go:embed index.html
var page string

type Renderer struct {
	JSONString []byte
}

func (s Renderer) WriteTo(w io.Writer) (written int64, err error) {
	var b bytes.Buffer
	b.Grow(100)

	var v any
	if err := json.Unmarshal(s.JSONString, &v); err != nil {
		return 0, err
	}

	r := htmljson.DefaultMarshaler
	jsonHTML := r.Marshal(v)
	b.WriteString(strings.ReplaceAll(page, `{{.JSONText}}`, string(jsonHTML)))

	n, err := w.Write(b.Bytes())
	return int64(n), err
}
