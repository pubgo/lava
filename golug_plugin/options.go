package golug_plugin

// Module will scope the plugin to a specific module, e.g. the "api"
func Module(m string) ManagerOption {
	return func(o *ManagerOptions) {
		o.Module = m
	}
}
