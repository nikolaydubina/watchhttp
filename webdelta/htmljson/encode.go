package htmljson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strconv"
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
// Inspired by encoding/json.
// Warning: Should be used only for basic Go types: bool, float64, string, []any, map[string]any, nil.
// Warning: does not handle pointers.
// You can get allowed input easily if you json.Unmarshal of JSON to any or to respective types.
func (s *Marshaller) Marshal(v any) []byte {
	b := bytes.Buffer{}
	s.MarshalTo(&b, v)
	return b.Bytes()
}

func (s *Marshaller) MarshalTo(w io.Writer, v any) error {
	s.w = w
	s.depth = 0
	s.key = "$"
	return s.marshal(v)
}

func (s *Marshaller) marshal(v any) error { return s.valueEncoder(v)(reflect.ValueOf(v)) }

type encoderFn func(v reflect.Value) error

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

func (s *Marshaller) write(v string) error {
	_, err := io.WriteString(s.w, v)
	return err
}

func (s *Marshaller) encodeUnsupported(v reflect.Value) error {
	return fmt.Errorf("skip unsupported type at key(%s) value(%#v) kind(%v)", s.key, v, v.Kind())
}

func (s *Marshaller) encodeNull(v reflect.Value) error {
	return s.write(s.Null(s.key))
}

func (s *Marshaller) encodeBool(v reflect.Value) error {
	return s.write(s.Bool(s.key, v.Bool()))
}

func (s *Marshaller) encodeString(v reflect.Value) error {
	return s.write(s.String(s.key, v.String()))
}

func (s *Marshaller) encodeFloat64(v reflect.Value) error {
	b, err := json.Marshal(v.Float())
	if err != nil {
		return err
	}
	return s.write(s.Number(s.key, v.Float(), string(b)))
}

func (s *Marshaller) encodeArray(v reflect.Value) error {
	s.write(s.Array.OpenBracket)
	s.write("\n")

	// TODO: increment row
	// TODO: increment offset

	k := s.key
	d := s.depth

	n := v.Len()
	for i := 0; i < n; i++ {
		if i > 0 {
			s.write(s.Array.Comma)
		}

		s.key = k + "[" + strconv.Itoa(i) + "]"
		s.depth = d + 1

		s.marshal(v.Index(i).Interface())

		s.write("\n")
	}

	s.key = k
	s.depth = d

	// TODO: increment row

	s.write("\n")
	s.write(s.Array.CloseBracket)

	return nil
}

func (s *Marshaller) encodeMap(v reflect.Value) error {
	s.write(s.Map.OpenBracket)
	s.write("\n")

	// TODO: increment row
	// TODO: increment offset

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

	sort.Slice(sv, func(i, j int) bool { return sv[i].ks < sv[j].ks })

	k := s.key
	d := s.depth

	for i, kv := range sv {
		if i > 0 {
			s.write(s.Map.Comma)
			s.write("\n")
		}

		s.key = k + "." + kv.ks
		s.depth = d + 1

		// key
		s.write(s.Map.Key(s.key, kv.ks))
		s.write(s.Map.Colon)

		// value
		s.marshal(kv.v)

		// TODO: increment row
		// TODO: increment offset
	}

	s.key = k
	s.depth = d

	// TODO: increment row

	s.write("\n")
	s.write(s.Map.CloseBracket)

	return nil
}
