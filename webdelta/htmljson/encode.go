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

	w     io.Writer
	depth int
	key   string
	row   int
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
	s.w = w
	s.depth = 0
	s.key = "$"
	s.row = 0
	s.marshal(v)
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

func (s *Marshaller) writeNoOffset(v string) {
	_, err := io.WriteString(s.w, v)
	s.err = append(s.err, err)
}

func (s *Marshaller) write(v string) {
	s.writeNoOffset(strings.Repeat("    ", s.depth))
	s.writeNoOffset(v)
}

func (s *Marshaller) encodeUnsupported(v reflect.Value) {
	s.err = append(s.err, errors.New("skip unsupported type at key("+s.key+") kind("+v.Kind().String()+")"))
}

func (s *Marshaller) encodeNull(v reflect.Value) { s.writeNoOffset(s.Null(s.key)) }

func (s *Marshaller) encodeBool(v reflect.Value) { s.writeNoOffset(s.Bool(s.key, v.Bool())) }

func (s *Marshaller) encodeString(v reflect.Value) { s.writeNoOffset(s.String(s.key, v.String())) }

func (s *Marshaller) encodeFloat64(v reflect.Value) {
	s.writeNoOffset(s.Number(s.key, v.Float(), strconv.FormatFloat(v.Float(), 'f', -1, 64)))
}

func (s *Marshaller) encodeArray(v reflect.Value) {
	n := v.Len()

	s.writeNoOffset(s.Array.OpenBracket)

	if n == 0 {
		s.writeNoOffset(s.Array.CloseBracket)
		return
	}

	// write array
	k, d := s.key, s.depth

	s.writeNoOffset("\n")
	s.row++

	for i := 0; i < n; i++ {
		if i > 0 {
			s.writeNoOffset(s.Array.Comma)
			s.writeNoOffset("\n")
			s.row++
		}

		s.key = k + "[" + strconv.Itoa(i) + "]"
		s.depth = d + 1

		s.write("") // fake virtual key, to apply same offset logic as JSON Map
		s.marshal(v.Index(i).Interface())
	}

	s.key, s.depth = k, d
	s.row++
	s.writeNoOffset("\n")
	s.write(s.Array.CloseBracket)
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

	s.writeNoOffset(s.Map.OpenBracket)

	if len(sv) == 0 {
		s.writeNoOffset(s.Map.CloseBracket)
		return
	}

	// write map
	k, d := s.key, s.depth

	s.writeNoOffset("\n")
	s.row++

	sort.Slice(sv, func(i, j int) bool { return sv[i].ks < sv[j].ks })

	for i, kv := range sv {
		if i > 0 {
			s.writeNoOffset(s.Map.Comma)
			s.writeNoOffset("\n")
			s.row++
		}

		s.key = k + "." + kv.ks
		s.depth = d + 1

		// key
		s.write(s.Map.Key(s.key, kv.ks))
		s.writeNoOffset(s.Map.Colon)

		// value
		s.marshal(kv.v)
	}

	s.key, s.depth = k, d
	s.row++
	s.writeNoOffset("\n")
	s.write(s.Map.CloseBracket)
}
