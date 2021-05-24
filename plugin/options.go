package plugin

// Module will scope the plugin to a specific module, e.g. the "api"
func Module(m string) ManagerOpt {
	return func(o *managerOpts) {
		o.Module = m
	}
}
