package types

type Vars func(name string, data func() interface{})

func (v Vars) Do(name string, data func() interface{}) {
	v(name, data)
}
