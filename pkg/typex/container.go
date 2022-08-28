package typex

type M = map[string]interface{}
type D []Kv

// Map creates a map from the elements of the D.
func (d D) Map() M {
	m := make(M, len(d))
	for _, e := range d {
		m[e.K] = e.V
	}
	return m
}

func (d *D) Append(kv ...Kv) {
	*d = append(*d, kv...)
}

func KvOf(k string, v interface{}) Kv {
	return Kv{K: k, V: v}
}

// Kv represents a BSON element for a D. It is usually used inside a D.
type Kv struct {
	K string      `json:"k"`
	V interface{} `json:"v"`
}

func (kv Kv) Map() M { return M{kv.K: kv.V} }

// An A is an ordered representation of a BSON array.
//
// Example usage:
//
// 		typex.A{"bar", "world", 3.14159, typex.D{{"qux", 12345}}}
type A []interface{}

func (a *A) Append(data ...interface{}) {
	*a = append(*a, data...)
}
