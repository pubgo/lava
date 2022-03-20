package vars_type

type Vars func(name string, data func() interface{})

func (v Vars) Do(name string, data func() interface{}) {
	v(name, data)
}

func (v Vars) Publish(name string, data func() interface{}) {
	v(name, data)
}
