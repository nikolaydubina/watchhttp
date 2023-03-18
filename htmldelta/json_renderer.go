package htmldelta

import (
	_ "embed"
	"encoding/json"
	"io"

	"github.com/nikolaydubina/htmljson"
)

//go:embed delta_json.html
var jsonTemplateHTML []byte

// JSONRenderer renders delta of JSON in HTML with animations.
// Not idempotent.
// Not safe for concurrent use.
type JSONRenderer struct {
	r   io.Reader
	m   *htmljson.PageMarshaler
	num map[string]float64
}

func NewJSONRenderer(title string) *JSONRenderer {
	s := JSONRenderer{
		num: make(map[string]float64),
	}

	r := htmljson.Marshaler{
		Null:   htmljson.NullHTML,
		Bool:   htmljson.BoolHTML,
		String: htmljson.StringHTML,
		Number: s.numberfunc,
		Array:  htmljson.DefaultArrayHTML,
		Map:    htmljson.DefaultMapHTML,
		Row:    htmljson.DefaultRowHTML{Padding: 4}.Marshal,
	}

	m := htmljson.DefaultPageMarshaler
	m.Marshaler = &r
	m.Title = title
	m.Template = jsonTemplateHTML

	s.m = &m
	return &s
}

func (s *JSONRenderer) numberfunc(k string, v float64, sv string) string {
	var class string
	if prev, ok := s.num[k]; ok {
		switch {
		case v > prev:
			class = "number-up"
		case v < prev:
			class = "number-down"
		default:
			class = ""
		}
	}
	s.num[k] = v
	return `<div class="json-value json-number ` + class + `">` + sv + `</div>`
}

func (s *JSONRenderer) From(r io.Reader) *JSONRenderer {
	s.r = r
	return s
}

func (s *JSONRenderer) WriteTo(w io.Writer) (written int64, err error) {
	var v any
	if err := json.NewDecoder(s.r).Decode(&v); err != nil && err != io.EOF {
		return 0, err
	}
	s.m.MarshalTo(w, v)
	return 0, err
}
