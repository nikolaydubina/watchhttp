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
func (s Marshaller) Marshal(v any) []byte {
	b := bytes.Buffer{}
	s.MarshalTo(&b, v)
	return b.Bytes()
}

func (s Marshaller) MarshalTo(w io.Writer, v any) {
	s.marshal(w, "$", v)
}

func (s *Marshaller) marshal(w io.Writer, key string, v any) {
	valueEncoder(v)(s, w, key, reflect.ValueOf(v))
}

type encoderFn func(s *Marshaller, w io.Writer, key string, v reflect.Value)

func valueEncoder(v any) encoderFn {
	if v == nil {
		return encodeNull
	}

	r := reflect.ValueOf(v)
	if !r.IsValid() {
		return encodeUnsupported
	}

	switch r.Type().Kind() {
	// scalars
	case reflect.Bool:
		return encodeBool
	case reflect.String:
		return encodeString
	case reflect.Float64:
		return encodeFloat64

	// containers
	case reflect.Map:
		return encodeMap
	case reflect.Slice, reflect.Array:
		return encodeArray

	// if we have struct, we did something wrong
	case reflect.Struct:
		return encodeUnsupported
	default:
		return encodeUnsupported
	}
}

func encodeUnsupported(s *Marshaller, w io.Writer, key string, v reflect.Value) {
	panic(fmt.Errorf("skip unsupported type at key(%s) value(%#v) kind(%v)", key, v, v.Kind()))
}

func encodeNull(s *Marshaller, w io.Writer, key string, v reflect.Value) {
	io.WriteString(w, s.Null(key))
}

func encodeBool(s *Marshaller, w io.Writer, key string, v reflect.Value) {
	io.WriteString(w, s.Bool(key, v.Bool()))
}

func encodeString(s *Marshaller, w io.Writer, key string, v reflect.Value) {
	io.WriteString(w, s.String(key, v.String()))
}

func encodeFloat64(s *Marshaller, w io.Writer, key string, v reflect.Value) {
	b, _ := json.Marshal(v.Float())
	io.WriteString(w, s.Number(key, v.Float(), string(b)))
}

func encodeArray(s *Marshaller, w io.Writer, key string, v reflect.Value) {
	io.WriteString(w, s.Array.OpenBracket)

	io.WriteString(w, "\n")
	// TODO: increment row
	// TODO: increment offset

	n := v.Len()
	for i := 0; i < n; i++ {
		if i > 0 {
			io.WriteString(w, s.Array.Comma)
		}
		kk := key + "[" + strconv.Itoa(i) + "]"

		s.marshal(w, kk, v.Index(i).Interface())

		io.WriteString(w, "\n")
	}

	// TODO: increment row

	io.WriteString(w, "\n")
	io.WriteString(w, s.Array.CloseBracket)

}

func encodeMap(s *Marshaller, w io.Writer, key string, v reflect.Value) {
	io.WriteString(w, s.Map.OpenBracket)
	io.WriteString(w, "\n")

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

	for i, kv := range sv {
		if i > 0 {
			io.WriteString(w, s.Map.Comma)
			io.WriteString(w, "\n")
		}

		kk := key + "." + kv.ks

		// key
		io.WriteString(w, s.Map.Key(kk, kv.ks))
		io.WriteString(w, s.Map.Colon)

		// value
		s.marshal(w, kk, kv.v)

		// TODO: increment row
		// TODO: increment offset
	}

	// TODO: increment row

	io.WriteString(w, "\n")
	io.WriteString(w, s.Map.CloseBracket)
}
