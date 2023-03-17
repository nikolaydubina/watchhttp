package htmljson

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// Marshaller convert struct as JSON represented into HTML.
// Visually HTML is similar to pretty printed JSON with indentation.
// This facilitates CSS styling, CSS animations, and JavaScript event hooks.
// Rendering is customized by providing renderers for specific elements.
type Marshaller struct {
	Null   func(key string) string
	Bool   func(key string, v bool) string
	String func(key string, v string) string
	Number func(key string, v float64, s string) string
	Array  ArrayMarshaller
	Map    MapMarshaller
	Row    func(s string, padding int) string

	*RowWriter
	depth int
	key   string
	err   []error
}

type ArrayMarshaller struct {
	OpenBracket  string
	CloseBracket string
	Comma        string
}

type MapMarshaller struct {
	OpenBracket  string
	CloseBracket string
	Comma        string
	Colon        string
	Key          func(key string, v string) string
}

// Marshal convert struct as JSON represented into HTML.
// Passes JSON path to render specific JSON elements.
// Inspired by encoding/json.
// Warning: Should be used only for basic Go types: bool, float64, string, []any, map[string]any, nil.
// Warning: does not handle pointers, does not handle custom struct.
// You can get allowed input easily with json.Unmarshal to any.
func (s *Marshaller) Marshal(v any) []byte {
	b := bytes.Buffer{}
	s.MarshalTo(&b, v)
	return b.Bytes()
}

func (s *Marshaller) MarshalTo(w io.Writer, v any) error {
	s.depth = 0
	s.key = "$"
	s.RowWriter = &RowWriter{
		b:   strings.Builder{},
		w:   w,
		Row: s.Row,
	}
	s.marshal(v)
	s.flush(s.depth)
	s.err = append(s.err, s.RowWriter.err...)
	return errors.Join(s.err...)
}

func (s *Marshaller) marshal(v any) { s.valueEncoder(v)(reflect.ValueOf(v)) }

type encoderFn func(v reflect.Value)

func (s *Marshaller) valueEncoder(v any) encoderFn {
	if v == nil {
		return s.encodeNull
	}

	r := reflect.ValueOf(v)
	if !r.IsValid() {
		return s.encodeUnsupported
	}

	switch r.Type().Kind() {
	// scalars
	case reflect.Bool:
		return s.encodeBool
	case reflect.String:
		return s.encodeString
	case reflect.Float64:
		return s.encodeFloat64

	// containers
	case reflect.Map:
		return s.encodeMap
	case reflect.Slice, reflect.Array:
		return s.encodeArray

	// if we have struct, we did something wrong
	case reflect.Struct:
		return s.encodeUnsupported
	default:
		return s.encodeUnsupported
	}
}

func (s *Marshaller) encodeUnsupported(v reflect.Value) {
	s.err = append(s.err, errors.New("skip unsupported type at key("+s.key+") kind("+v.Kind().String()+")"))
}

func (s *Marshaller) encodeNull(v reflect.Value) { s.write(s.Null(s.key)) }

func (s *Marshaller) encodeBool(v reflect.Value) { s.write(s.Bool(s.key, v.Bool())) }

func (s *Marshaller) encodeString(v reflect.Value) { s.write(s.String(s.key, v.String())) }

func (s *Marshaller) encodeFloat64(v reflect.Value) {
	s.write(s.Number(s.key, v.Float(), strconv.FormatFloat(v.Float(), 'f', -1, 64)))
}

func (s *Marshaller) encodeArray(v reflect.Value) {
	n := v.Len()

	s.write(s.Array.OpenBracket)

	if n == 0 {
		s.write(s.Array.CloseBracket)
		return
	}

	// write array
	k, d := s.key, s.depth
	s.flush(d)

	s.depth = d + 1
	for i := 0; i < n; i++ {
		if i > 0 {
			s.write(s.Array.Comma)
			s.flush(s.depth)
		}

		s.key = k + "[" + strconv.Itoa(i) + "]"

		s.write("") // fake virtual key, to apply same offset logic as JSON Map
		s.marshal(v.Index(i).Interface())
	}

	s.flush(s.depth)
	s.write(s.Array.CloseBracket)

	s.key, s.depth = k, d
}

func (s *Marshaller) encodeMap(v reflect.Value) {
	type mapKV struct {
		rk reflect.Value
		rv reflect.Value
		ks string
		v  any
	}

	// extract and sort the keys
	sv := make([]mapKV, v.Len())
	mi := v.MapRange()
	for i := 0; mi.Next(); i++ {
		sv[i].rk = mi.Key()
		sv[i].rv = mi.Value()

		// key is always string
		if sv[i].rk.Kind() == reflect.String {
			sv[i].ks = sv[i].rk.String()
		}

		sv[i].v = v.MapIndex(mi.Key()).Interface()
	}

	s.write(s.Map.OpenBracket)

	if len(sv) == 0 {
		s.write(s.Map.CloseBracket)
		return
	}

	// write map
	k, d := s.key, s.depth

	s.flush(d)

	sort.Slice(sv, func(i, j int) bool { return sv[i].ks < sv[j].ks })

	s.depth = d + 1
	for i, kv := range sv {
		if i > 0 {
			s.write(s.Map.Comma)
			s.flush(s.depth)
		}

		s.key = k + "." + kv.ks

		// key
		s.write(s.Map.Key(s.key, kv.ks))
		s.write(s.Map.Colon)

		// value
		s.marshal(kv.v)
	}

	s.flush(s.depth)
	s.write(s.Map.CloseBracket)

	s.key, s.depth = k, d
}
