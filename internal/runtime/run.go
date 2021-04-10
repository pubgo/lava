package runtime

var beforeStarts []func()
var afterStarts []func()
var beforeStops []func()
var afterStops []func()

func BeforeStart(fn func()) { beforeStarts = append(beforeStarts, fn) }
func AfterStart(fn func())  { afterStarts = append(afterStarts, fn) }
func BeforeStop(fn func())  { beforeStops = append(beforeStops, fn) }
func AfterStop(fn func())   { afterStops = append(afterStops, fn) }
