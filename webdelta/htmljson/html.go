package htmljson

func NullHTML(k string) string { return `<div class="json-lang json-value json-null">null</div>` }

func BoolHTML(k string, v bool) string {
	x := "false"
	if v {
		x = "true"
	}
	return `<div class="json-lang json-value json-bool">` + x + `</div>`
}

func StringHTML(k string, v string) string {
	return `<div class="json-value json-string">"` + v + `"</div>`
}

func NumberHTML(k string, v float64, s string) string {
	return `<div class="json-value json-number">` + s + `</div>`
}

var DefaultArrayHTML = ArrayMarshaller{
	OpenBracket:  `<div class="json-lang">[</div>`,
	CloseBracket: `<div class="json-lang">]</div>`,
	Comma:        `<div class="json-lang">,</div>`,
}

var DefaultMapHTML = MapMarshaller{
	OpenBracket:  `<div class="json-lang">{</div>`,
	CloseBracket: `<div class="json-lang">}</div>`,
	Comma:        `<div class="json-lang">,</div>`,
	Colon:        `<div class="json-lang">:</div>`,
	Key:          func(key string, v string) string { return `<div class="json-key json-string">` + v + `</div>` },
}

// DefaultMarshaller adds basic HTML div classes for further styling.
var DefaultMarshaller = Marshaller{
	Null:   NullHTML,
	Bool:   BoolHTML,
	String: StringHTML,
	Number: NumberHTML,
	Array:  DefaultArrayHTML,
	Map:    DefaultMapHTML,
}
