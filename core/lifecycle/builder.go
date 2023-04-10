package lifecycle

type Exporter struct {
	Lifecycle    Lifecycle
	GetLifecycle GetLifecycle
}

func New(handlers []Handler) Exporter {
	lc := new(lifecycleImpl)
	for i := range handlers {
		handlers[i](lc)
	}

	return Exporter{
		Lifecycle:    lc,
		GetLifecycle: lc,
	}
}
