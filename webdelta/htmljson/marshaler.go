package htmljson

import (
	"bytes"
	"errors"
	"io"
	"sort"
	"strconv"
	"strings"
)

// Marshaler converts JSON stored as Go `any` object represented into HTML.
// Visually HTML is similar to pretty printed JSON with indentation.
// Rendering is customized by providing renderers for specific JSON elements.
// This facilitates CSS styling, CSS animations, and JavaScript events.
// JSON element renderers receive JSON path and value of element.
// Should be used only for types: bool, float64, string, []any, map[string]any, nil.
// You can get allowed input easily with json.Unmarshal to any.
// Safe for repeated use.
// Not safe for concurrent use.
type Marshaler struct {
	Null   func(key string) string
	Bool   func(key string, v bool) string
	String func(key string, v string) string
	Number func(key string, v float64, s string) string
	Array  ArrayMarshaler
	Map    MapMarshaler
	Row    func(s string, padding int) string

	*rowWriter
	depth int
	key   string
	err   []error
}

type ArrayMarshaler struct {
	OpenBracket  string
	CloseBracket string
	Comma        string
}

type MapMarshaler struct {
	OpenBracket  string
	CloseBracket string
	Comma        string
	Colon        string
	Key          func(key string, v string) string
}

// Marshaler converts JSON stored as Go `any` object represented into HTML.
func (s *Marshaler) Marshal(v any) []byte {
	b := bytes.Buffer{}
	s.MarshalTo(&b, v)
	return b.Bytes()
}

// MarshalTo converts JSON stored as Go `any` object represented into HTML.
func (s *Marshaler) MarshalTo(w io.Writer, v any) error {
	s.depth = 0
	s.key = "$"
	s.rowWriter = &rowWriter{
		b:   strings.Builder{},
		w:   w,
		Row: s.Row,
	}
	s.marshal(v)
	s.flush(s.depth)
	s.err = append(s.err, s.rowWriter.err...)
	return errors.Join(s.err...)
}

func (s *Marshaler) marshal(v any) {
	if v == nil {
		s.encodeNull()
	}
	switch q := v.(type) {
	case bool:
		s.encodeBool(q)
	case string:
		s.encodeString(q)
	case float64:
		s.encodeFloat64(q)
	case map[string]any:
		s.encodeMap(q)
	case []any:
		s.encodeArray(q)
	default:
		s.err = append(s.err, errors.New("skip unsupported type at key("+s.key+")"))
	}
}

func (s *Marshaler) encodeNull() { s.write(s.Null(s.key)) }

func (s *Marshaler) encodeBool(v bool) { s.write(s.Bool(s.key, v)) }

func (s *Marshaler) encodeString(v string) { s.write(s.String(s.key, v)) }

func (s *Marshaler) encodeFloat64(v float64) {
	s.write(s.Number(s.key, v, strconv.FormatFloat(v, 'f', -1, 64)))
}

func (s *Marshaler) encodeArray(v []any) {
	s.write(s.Array.OpenBracket)

	if len(v) == 0 {
		s.write(s.Array.CloseBracket)
		return
	}

	// write array
	k, d := s.key, s.depth
	defer func() { s.key, s.depth = k, d }()
	s.flush(d)

	s.depth = d + 1
	for i, q := range v {
		if i > 0 {
			s.write(s.Array.Comma)
			s.flush(s.depth)
		}

		s.key = k + "[" + strconv.Itoa(i) + "]"

		s.write("") // fake virtual key, to apply same offset logic as JSON Map
		s.marshal(q)
	}

	s.flush(s.depth)
	s.write(s.Array.CloseBracket)
}

func (s *Marshaler) encodeMap(v map[string]any) {
	s.write(s.Map.OpenBracket)

	if len(v) == 0 {
		s.write(s.Map.CloseBracket)
		return
	}

	// extract and sort the keys
	type kv struct {
		k string
		v any
	}
	sv := make([]kv, 0, len(v))
	for k, v := range v {
		sv = append(sv, kv{k: k, v: v})
	}
	sort.Slice(sv, func(i, j int) bool { return sv[i].k < sv[j].k })

	// write map
	k, d := s.key, s.depth
	defer func() { s.key, s.depth = k, d }()

	s.flush(d)

	s.depth = d + 1
	for i, kv := range sv {
		if i > 0 {
			s.write(s.Map.Comma)
			s.flush(s.depth)
		}

		s.key = k + "." + kv.k

		// key
		s.write(s.Map.Key(s.key, kv.k))
		s.write(s.Map.Colon)

		// value
		s.marshal(kv.v)
	}

	s.flush(s.depth)
	s.write(s.Map.CloseBracket)
}
