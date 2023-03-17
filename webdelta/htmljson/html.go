package htmljson

func NullHTML(k string) string { return `<div>null</div>` }

func BoolHTML(k string, v bool) string {
	if v {
		return `<div>true</div>`
	}
	return `<div>false</div>`
}

func StringHTML(k string, v string) string { return `<div>"` + v + `"</div>` }

func NumberHTML(k string, v float64, s string) string { return `<div>"` + s + `"</div>` }

var DefaultArrayHTML = ArrayMarshaller{
	OpenBracket:  `<div>[</div>`,
	CloseBracket: `<div>]</div>`,
	Comma:        `<div>,</div>`,
}

var DefaultMapHTML = MapMarshaller{
	OpenBracket:  `<div>{</div>`,
	CloseBracket: `<div>}</div>`,
	Comma:        `<div>,</div>`,
	Colon:        `<div>:</div>`,
	Key:          func(key string, v string) string { return `<div>` + v + `</div>` },
}

var DefaultMarshaller = Marshaller{
	Null:   NullHTML,
	Bool:   BoolHTML,
	String: StringHTML,
	Number: NumberHTML,
	Array:  DefaultArrayHTML,
	Map:    DefaultMapHTML,
}
