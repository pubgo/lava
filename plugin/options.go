package plugin

// Module will scope the plugin to a specific module, e.g. the "api"
func Module(m string) Opt { return func(o *options) { o.Module = m } }
