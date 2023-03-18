package htmldelta

import (
	_ "embed"
	"encoding/json"
	"io"
	"strings"

	"github.com/nikolaydubina/htmljson"
)

//go:embed delta_json.html
var jsonTemplateHTML string

type JSONRenderer struct {
	Title string
	r     io.Reader
	num   map[string]float64
}

func (s *JSONRenderer) From(r io.Reader) *JSONRenderer {
	s.r = r
	return s
}

func (s *JSONRenderer) WriteTo(w io.Writer) (written int64, err error) {
	if s.num == nil {
		s.num = make(map[string]float64)
	}

	var v any
	dec := json.NewDecoder(s.r)
	if err := dec.Decode(&v); err != nil && err != io.EOF {
		return 0, err
	}

	r := htmljson.Marshaler{
		Null:   htmljson.NullHTML,
		Bool:   htmljson.BoolHTML,
		String: htmljson.StringHTML,
		Number: func(k string, v float64, sv string) string {
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
		},
		Array: htmljson.DefaultArrayHTML,
		Map:   htmljson.DefaultMapHTML,
		Row:   htmljson.DefaultRowHTML{Padding: 4}.Marshal,
	}

	htmlPage := jsonTemplateHTML
	htmlPage = strings.ReplaceAll(htmlPage, `{{.Title}}`, s.Title)
	htmlPage = strings.ReplaceAll(htmlPage, `{{.HTMLJSON}}`, string(r.Marshal(v)))

	n, err := io.WriteString(w, htmlPage)
	return int64(n), err
}
