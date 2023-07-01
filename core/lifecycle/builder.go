package lifecycle

type Provider struct {
	Setter Lifecycle
	Getter Getter
}

func New(handlers []Handler) Provider {
	lc := new(lifecycleImpl)
	for i := range handlers {
		handlers[i](lc)
	}

	return Provider{
		Setter: lc,
		Getter: lc,
	}
}
