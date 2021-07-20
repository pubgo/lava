package entry

var beforeStarts []func()
var afterStarts []func()
var beforeStops []func()
var afterStops []func()

func GetBeforeStartsList() []func() { return beforeStarts }
func GetAfterStartsList() []func()  { return afterStarts }
func GetBeforeStopsList() []func()  { return beforeStops }
func GetAfterStopsList() []func()   { return afterStops }

func BeforeStart(fn func()) { beforeStarts = append(beforeStarts, fn) }
func AfterStart(fn func())  { afterStarts = append(afterStarts, fn) }
func BeforeStop(fn func())  { beforeStops = append(beforeStops, fn) }
func AfterStop(fn func())   { afterStops = append(afterStops, fn) }
