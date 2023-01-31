package encoding

func init() {
	vars.Register(Name, func() interface{} { return Keys() })
}
