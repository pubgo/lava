package typex

type M = map[string]interface{}
type D []Kv

// Map creates a map from the elements of the D.
func (d D) Map() M {
	m := make(M, len(d))
	for _, e := range d {
		m[e.Key] = e.Value
	}
	return m
}

// Kv represents a BSON element for a D. It is usually used inside a D.
type Kv struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

func (kv Kv) Map() M { return M{kv.Key: kv.Value} }

// An A is an ordered representation of a BSON array.
//
// Example usage:
//
// 		typex.A{"bar", "world", 3.14159, typex.D{{"qux", 12345}}}
type A []interface{}

func (a *A) Append(data ...interface{}) {
	*a = append(*a, data...)
}
