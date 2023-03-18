package htmljson

import (
	"bytes"
	_ "embed"
	"io"
	"strings"
)

//go:embed html_default_page.html
var DefaultHTMLPageTemplate string

func MarshalHTML(v any) []byte {
	b := bytes.Buffer{}
	MarshalHTMLTo(&b, v)
	return b.Bytes()
}

func MarshalHTMLTo(w io.Writer, v any) (written int64, err error) {
	var b bytes.Buffer
	b.Grow(1000)

	jsonHTML := DefaultMarshaler.Marshal(v)
	b.WriteString(strings.ReplaceAll(DefaultHTMLPageTemplate, `{{.HTMLJSON}}`, string(jsonHTML)))

	n, err := w.Write(b.Bytes())
	return int64(n), err
}
