package htmldelta

import (
	_ "embed"
	"io"

	"github.com/nikolaydubina/htmlyaml"
	"gopkg.in/yaml.v3"
)

//go:embed delta_yaml.html
var yamlTemplateHTML []byte

// YAMLRenderer renders delta of YAML in HTML with animations.
// Not idempotent.
// Not safe for concurrent use.
type YAMLRenderer struct {
	m   *htmlyaml.PageMarshaler
	num map[string]float64
}

func NewYAMLRenderer(title string) *YAMLRenderer {
	s := YAMLRenderer{
		num: make(map[string]float64),
	}

	r := htmlyaml.DefaultMarshaler
	r.Number = s.numberfunc

	m := htmlyaml.DefaultPageMarshaler
	m.Marshaler = &r
	m.Title = title
	m.Template = yamlTemplateHTML

	s.m = &m
	return &s
}

func (s *YAMLRenderer) numberfunc(k string, v float64, sv string) string {
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
	return `<div class="yaml-value ` + class + `">` + sv + `</div>`
}

func (s *YAMLRenderer) FromTo(r io.Reader, w io.Writer) (written int64, err error) {
	var v any
	if err := yaml.NewDecoder(r).Decode(&v); err != nil && err != io.EOF {
		return 0, err
	}
	s.m.MarshalTo(w, v)
	return 0, err
}
