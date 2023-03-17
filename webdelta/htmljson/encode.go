package htmljson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
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
func (s Marshaller) Marshal(v any) ([]byte, error) {
	e := encoder{s: s}

	e.marshal("$", v)

	buf := append([]byte(nil), e.Bytes()...)
	return buf, nil
}

type encoder struct {
	s Marshaller
	bytes.Buffer
}

func (e *encoder) marshal(key string, v any) {
	e.reflectValue(key, reflect.ValueOf(v))
}

func (e *encoder) reflectValue(key string, v reflect.Value) {
	log.Printf("%s: %#v", key, v.Kind().String())
	valueEncoder(v)(e, key, v)
}

func valueEncoder(v reflect.Value) encoderFn {
	if !v.IsValid() {
		return encodeUnsupported
	}

	switch v.Type().Kind() {
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

	// null or recursive
	case reflect.Interface:
		return encodeInterface
	case reflect.Pointer:
		return encodePtr

	// if we have struct, we did something wrong
	case reflect.Struct:
		return encodeUnsupported
	default:
		return encodeUnsupported
	}
}

type encoderFn func(e *encoder, key string, v reflect.Value)

func encodePtr(e *encoder, key string, v reflect.Value) {
	if v.IsNil() {
		e.WriteString("null")
		return
	}
	e.marshal(key, v.Elem())
}

func encodeUnsupported(e *encoder, key string, v reflect.Value) {
	panic(fmt.Errorf("skip unsupported type at key(%s) value(%#v) kind(%v)", key, v, v.Kind()))
}

func encodeBool(e *encoder, key string, v reflect.Value) { e.WriteString(e.s.Bool(key, v.Bool())) }

func encodeString(e *encoder, key string, v reflect.Value) {
	e.WriteString(e.s.String(key, v.String()))
}

func encodeFloat64(e *encoder, key string, v reflect.Value) {
	b, _ := json.Marshal(v.Float())
	e.WriteString(e.s.Number(key, v.Float(), string(b)))
}

func encodeInterface(e *encoder, key string, v reflect.Value) {
	if v.IsNil() {
		e.WriteString(e.s.Null(key))
	}
	e.marshal(key, v.Elem())
}

func encodeArray(e *encoder, key string, v reflect.Value) {
	e.WriteString(e.s.Array.OpenBracket)

	e.WriteByte('\n')
	// TODO: increment row
	// TODO: increment offset

	n := v.Len()
	for i := 0; i < n; i++ {
		if i > 0 {
			e.WriteString(e.s.Array.Comma)
		}
		kk := key + "[" + strconv.Itoa(i) + "]."

		e.marshal(kk, v.Index(i).Interface())

		e.WriteByte('\n')
	}

	// TODO: increment row

	e.WriteByte('\n')
	e.WriteString(e.s.Array.CloseBracket)

}

func encodeMap(e *encoder, key string, v reflect.Value) {
	e.WriteString(e.s.Map.OpenBracket)
	e.WriteByte('\n')

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
			e.WriteString(e.s.Map.Comma)
			e.WriteByte('\n')
		}

		kk := key + "." + kv.ks

		// key
		e.WriteString(e.s.Map.Key(kk, kv.ks))
		e.WriteString(e.s.Map.Colon)

		// value
		e.marshal(kk, kv.v)

		// TODO: increment row
		// TODO: increment offset
	}

	// TODO: increment row

	e.WriteByte('\n')
	e.WriteString(e.s.Map.CloseBracket)
}
