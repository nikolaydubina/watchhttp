package htmljson

// JSONPathCollector provides HTML renderers as its methods
// and collects which JSON path keys have been passed to it and which values.
// This is useful for testing.
type JSONPathCollector struct {
	Keys map[string]any
}

func NewJSONPathCollector() *JSONPathCollector { return &JSONPathCollector{Keys: map[string]any{}} }

func (c *JSONPathCollector) add(k string, v any) string {
	c.Keys[k] = v
	return ""
}

func (c *JSONPathCollector) Null(k string) string { return c.add(k, "null") }

func (c *JSONPathCollector) Bool(k string, v bool) string { return c.add(k, v) }

func (c *JSONPathCollector) String(k string, v string) string { return c.add(k, v) }

func (c *JSONPathCollector) Number(k string, v float64, s string) string { return c.add(k, s) }

func (c *JSONPathCollector) MapKey(k string, v string) string { return c.add(k, v) }
