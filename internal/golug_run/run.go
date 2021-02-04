package golug_run

var beforeStarts []func()
var afterStarts []func()
var beforeStops []func()
var afterStops []func()

func GetBeforeStarts() []func() { return beforeStarts }
func GetAfterStarts() []func()  { return afterStarts }
func GetBeforeStops() []func()  { return beforeStops }
func GetAfterStops() []func()   { return afterStops }

func BeforeStart(fn func()) { beforeStarts = append(beforeStarts, fn) }
func AfterStart(fn func())  { afterStarts = append(afterStarts, fn) }
func BeforeStop(fn func())  { beforeStops = append(beforeStops, fn) }
func AfterStop(fn func())   { afterStops = append(afterStops, fn) }
