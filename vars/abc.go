package vars

type Publisher func(name string, data func() interface{})

func (v Publisher) Do(name string, data func() interface{}) {
	v(name, data)
}

func (v Publisher) Publish(name string, data func() interface{}) {
	v(name, data)
}
